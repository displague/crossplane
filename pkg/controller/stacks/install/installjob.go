/*
Copyright 2019 The Crossplane Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package install

import (
	"bytes"
	"context"
	"fmt"
	"path/filepath"

	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplaneio/crossplane-runtime/pkg/logging"
	"github.com/crossplaneio/crossplane-runtime/pkg/meta"
	"github.com/crossplaneio/crossplane/apis/stacks/v1alpha1"
	"github.com/crossplaneio/crossplane/pkg/controller/stacks/hosted"
	"github.com/crossplaneio/crossplane/pkg/stacks"
)

var (
	jobBackoff                = int32(0)
	registryDirName           = "/.registry"
	packageContentsVolumeName = "package-contents"
)

// JobCompleter is an interface for handling job completion
type jobCompleter interface {
	handleJobCompletion(ctx context.Context, i stacks.KindlyIdentifier, job *batchv1.Job) error
}

// StackInstallJobCompleter is a concrete implementation of the jobCompleter interface
type stackInstallJobCompleter struct {
	client       client.Client
	hostClient   client.Client
	podLogReader Reader
	log          logging.Logger
}

type metaV1RuntimeObject interface {
	metav1.Object
	runtime.Object
}

func createInstallJob(i v1alpha1.StackInstaller, executorInfo *stacks.ExecutorInfo, hCfg *hosted.Config, tscImage string) *batchv1.Job {
	name := i.GetName()
	namespace := i.GetNamespace()

	if hCfg != nil {
		// In Hosted Mode, we need to map all install jobs on tenant Kubernetes into a single namespace on host cluster.
		o := hCfg.ObjectReferenceOnHost(name, namespace)
		name = o.Name
		namespace = o.Namespace
	}

	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    stacks.ParentLabels(i),
		},
		Spec: batchv1.JobSpec{
			BackoffLimit: &jobBackoff,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					RestartPolicy: corev1.RestartPolicyNever,
					InitContainers: []corev1.Container{
						{
							Name:    "stack-copy-to-volume",
							Image:   i.Image(),
							Command: []string{"cp", "-R", registryDirName, "/ext-pkg/"},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      packageContentsVolumeName,
									MountPath: "/ext-pkg",
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "stack-unpack-and-output",
							Image: executorInfo.Image,
							// "--debug" can be added to this list of Args to get debug output from the job,
							// but note that will be included in the stdout from the pod, which makes it
							// impossible to create the resources that the job unpacks.
							Args: []string{
								"stack",
								"unpack",
								fmt.Sprintf("--content-dir=%s", filepath.Join("/ext-pkg", registryDirName)),
								"--permission-scope=" + i.PermissionScope(),
								"--templating-controller-image=" + tscImage,
							},
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      packageContentsVolumeName,
									MountPath: "/ext-pkg",
								},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: packageContentsVolumeName,
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
		},
	}
}

func (jc *stackInstallJobCompleter) handleJobCompletion(ctx context.Context, i stacks.KindlyIdentifier, job *batchv1.Job) error {
	// find the pod associated with the given job
	podName, err := jc.findPodNameForJob(ctx, job)
	if err != nil {
		return err
	}

	podLogs, err := jc.podLogReader.GetReader(job.Namespace, podName)

	if err != nil {
		return err
	}

	buf := &bytes.Buffer{}
	if _, err := buf.ReadFrom(podLogs); err != nil {
		return err
	}
	b := buf.Bytes()

	o := &stacks.UnpackJobOutput{}
	// decode and process all resources from job output
	if err := yaml.Unmarshal(b, o); err != nil {
		return err
	}

	// process and create the object that we just decoded
	if err := jc.createJobOutputObjects(ctx, o, i, job); err != nil {
		return err
	}

	return nil
}

// findPodNameForJob finds the pod name associated with the given job.  Note that this functions
// assumes only a single pod will be associated with the job.
func (jc *stackInstallJobCompleter) findPodNameForJob(ctx context.Context, job *batchv1.Job) (string, error) {
	podList, err := jc.findPodsForJob(ctx, job)
	if err != nil {
		return "", err
	}

	if len(podList.Items) != 1 {
		return "", errors.Errorf("pod list for job %s should only have 1 item, actual: %d", job.Name, len(podList.Items))
	}

	return podList.Items[0].Name, nil
}

func (jc *stackInstallJobCompleter) findPodsForJob(ctx context.Context, job *batchv1.Job) (*corev1.PodList, error) {
	podList := &corev1.PodList{}
	labelSelector := client.MatchingLabels{
		"job-name": job.Name,
	}
	nsSelector := client.InNamespace(job.Namespace)
	if err := jc.hostClient.List(ctx, podList, labelSelector, nsSelector); err != nil {
		return nil, err
	}

	return podList, nil
}

// createJobOutputObjects names, labels, and creates resources in the API
// Expected resources are CRD, Stack, StackDefinition, & StackConfiguration
// nolint:gocyclo
func (jc *stackInstallJobCompleter) createJobOutputObjects(ctx context.Context, obj *stacks.UnpackJobOutput,
	i stacks.KindlyIdentifier, job *batchv1.Job) error {

	// if we decoded a non-nil unstructured object, try to create it now
	if obj == nil {
		return nil
	}

	// We want to clean up any installed resources when we're deleted. We can't
	// rely on garbage collection because a namespaced object (StackInstall)
	// can't own a cluster scoped object (CustomResourceDefinition), so we use
	// labels instead.
	labels := stacks.ParentLabels(i)

	// make sure the name and namespace are set to match the current
	// StackInstall (if they haven't already been set)
	if obj.Stack != nil {
		o := obj.Stack
		if o.GetName() == "" {
			o.SetName(i.GetName())
		}
		if o.GetNamespace() == "" {
			o.SetNamespace(i.GetNamespace())
		}

		meta.AddLabels(o, labels)
	}

	if obj.StackDefinition != nil {
		o := obj.StackDefinition
		if o.GetName() == "" {
			o.SetName(i.GetName())
		}
		if o.GetNamespace() == "" {
			o.SetNamespace(i.GetNamespace())
		}

		meta.AddLabels(o, labels)
	}

	// StackDefinition controllers need the name of the StackDefinition
	// which, by design, matches the StackInstall
	if obj.StackDefinition != nil {
		setStackDefinitionControllerEnv(obj.StackDefinition, i.GetNamespace(), i.GetName())
	}

	// CRDs are labeled with the namespaces of the stacks they are managed by.
	// This will allow for a single Namespaced stack to be installed in multiple
	// namespaces, or different stacks (possibly only differing by versions) to
	// provide the same CRDs without the risk that a single StackInstall removal
	// will delete a CRD until there are no remaining namespace labels.
	for _, crd := range obj.CRDs {
		if crd == nil {
			continue
		}
		meta.AddLabels(crd, labels)

		labelNamespace := fmt.Sprintf(stacks.LabelNamespaceFmt, i.GetNamespace())

		nsLabels := map[string]string{labelNamespace: "true"}
		meta.AddLabels(crd, nsLabels)
	}

	objs := append([]metaV1RuntimeObject{}, obj.Stack, obj.StackDefinition)

	for _, c := range obj.CRDs {
		objs = append(objs, c)
	}

	for _, o := range objs {
		if o == nil {
			continue
		}

		/*
			jc.log.Debug(
				"creating object from job output",
				"job", job.Name,
				"name", o.GetName(),
				"namespace", o.GetNamespace(),
				"apiVersion", o.GetAPIVersion(),
				"kind", o.GetKind(),
				"labels", o.GetLabels(),
			)
		*/

		if err := jc.client.Create(ctx, o); err != nil && !kerrors.IsAlreadyExists(err) {
			return errors.Wrapf(err, "failed to create object %s from job output %s", o.GetName(), job.Name)
		}
	}
	return nil
}

func setStackDefinitionControllerEnv(sd *v1alpha1.StackDefinition, namespace, name string) {
	env := []corev1.EnvVar{{
		Name:  stacks.StackDefinitionNamespaceEnv,
		Value: namespace,
	}, {
		Name:  stacks.StackDefinitionNameEnv,
		Value: name,
	}}

	if d := sd.Spec.Controller.Deployment; d != nil {
		c := d.Spec.Template.Spec.Containers
		c[0].Env = append(c[0].Env, env...)
	}

}
