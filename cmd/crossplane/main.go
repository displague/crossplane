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

package main

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	"github.com/spf13/afero"
	"gopkg.in/alecthomas/kingpin.v2"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	runtimelog "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/manager/signals"

	"github.com/crossplaneio/crossplane-runtime/pkg/logging"
	"github.com/crossplaneio/crossplane/apis"
	stacksController "github.com/crossplaneio/crossplane/pkg/controller/stacks"
	templatesController "github.com/crossplaneio/crossplane/pkg/controller/stacks/templates"
	"github.com/crossplaneio/crossplane/pkg/controller/workload"
	"github.com/crossplaneio/crossplane/pkg/stacks"
	"github.com/crossplaneio/crossplane/pkg/stacks/walker"
)

func main() {
	var (
		log = logging.Logger

		// top level app definition
		app        = kingpin.New(filepath.Base(os.Args[0]), "An open source multicloud control plane.").DefaultEnvars()
		debug      = app.Flag("debug", "Run with debug logging.").Short('d').Bool()
		syncPeriod = app.Flag("sync", "Controller manager sync period duration such as 300ms, 1.5h or 2h45m").
				Short('s').Default("1h").Duration()

		// default crossplane command and args, this is the default main entry point for Crossplane's
		// multi-cloud control plane functionality
		crossplaneCmd = app.Command(filepath.Base(os.Args[0]), "An open source multicloud control plane.").Default()

		// stacks  commands and args, these are the main entry points for Crossplane's stack manager (SM).
		// The SM runs as a separate pod from the main Crossplane pod because in order to install stacks that
		// have arbitrary permissions, the SM itself must have cluster-admin permissions.  We isolate these elevated
		// permissions as much as possible by running the Crossplane stack manager in its own isolate deployment.
		stackCmd = app.Command("stack", "Perform operations on stacks")

		// stack manage - adds the stack manager controllers and starts their reconcile loops
		stackManageCmd       = stackCmd.Command("manage", "Manage stacks (run stack manager controllers)")
		stackManageGroups    = stackManageCmd.Flag("allowed-groups", "A comma separated list of GVK groups that Stacks may depend on. When omitted, all groups are permitted.").Default("").String()
		stackManageTemplates = stackManageCmd.Flag("templates", "Enable support for template stacks").Bool()

		// stack unpack - performs the unpacking operation for the given stack package content
		// directory. This command is expected to parse the content and generate manifests for stack
		// related artifacts to stdout so that the SM can read the output and use the Kubernetes API to
		// create the artifacts.
		//
		// Users are not expected to run this command themselves, the stack manager itself should
		// execute this command.
		stackUnpackCmd             = stackCmd.Command("unpack", "Unpack a stack")
		stackUnpackDir             = stackUnpackCmd.Flag("content-dir", "The absolute path of the directory that contains the stack contents").Required().String()
		stackUnpackOutfile         = stackUnpackCmd.Flag("outfile", "The file where the YAML Stack record and CRD artifacts will be written").String()
		stackUnpackPermissionScope = stackUnpackCmd.Flag("permission-scope", "The permission-scope that the stack must request (Namespaced, Cluster)").Default("Namespaced").String()
		stackUnpackGroups          = stackUnpackCmd.Flag("allowed-groups", "A comma separated list of GVK groups that Stacks may depend on. When omitted, all groups are permitted.").Default("").String()
	)
	cmd := kingpin.MustParse(app.Parse(os.Args[1:]))

	zl := zap.New(zap.UseDevMode(*debug))
	logging.SetLogger(zl)
	if *debug {
		// The controller-runtime runs with a no-op logger by default. It is
		// *very* verbose even at info level, so we only provide it a real
		// logger when we're running in debug mode.
		runtimelog.SetLogger(zl)
	}

	// Determine the command being called and execute the corresponding logic
	switch cmd {
	case crossplaneCmd.FullCommand():
		// the default Crossplane command is being run, add all the regular controllers to the manager
		mgr := setupManager(log, syncPeriod)

		log.Info("Adding controllers")
		kingpin.FatalIfError(controllerSetupWithManager(mgr), "Cannot add controllers to manager")

		// Start the Cmd
		log.Info("Starting the manager")
		kingpin.FatalIfError(mgr.Start(signals.SetupSignalHandler()), "Cannot start controller")
	case stackManageCmd.FullCommand():
		// the "stacks manage" command is being run, the only controllers we should add to the
		// manager are the stacks controllers

		mgr := setupManager(log, syncPeriod)

		if *stackManageGroups > "" {
			log.Info("Restricting Stack dependsOn types to: %q", stackManageGroups)
		}

		log.Info("Adding controllers")
		kingpin.FatalIfError(stacksControllerSetupWithManager(mgr, *stackManageGroups), "Cannot add controllers to manager")

		if *stackManageTemplates {
			log.Info("Adding template controllers")
			kingpin.FatalIfError(stacksTemplateControllerSetupWithManager(mgr), "Cannot add template controllers to manager")
		}

		// Start the Cmd
		log.Info("Starting the manager")
		kingpin.FatalIfError(mgr.Start(signals.SetupSignalHandler()), "Cannot start controller")
	case stackUnpackCmd.FullCommand():
		var outFile io.StringWriter
		// stack unpack command was called, run the stack unpacking logic
		if stackUnpackOutfile == nil || *stackUnpackOutfile == "" {
			outFile = os.Stdout
		} else {
			openFile, err := os.Create(*stackUnpackOutfile)
			kingpin.FatalIfError(err, "Cannot create outfile")
			defer closeOrError(openFile)
			outFile = openFile
		}

		// TODO(displague) afero.NewBasePathFs could avoid the need to track Base
		fs := afero.NewOsFs()
		rd := &walker.ResourceDir{Base: filepath.Clean(*stackUnpackDir), Walker: afero.Afero{Fs: fs}}
		kingpin.FatalIfError(stacks.Unpack(rd, outFile, rd.Base, *stackUnpackPermissionScope, *stackUnpackGroups), "failed to unpack stacks")
	default:
		kingpin.FatalUsage("unknown command %s", cmd)
	}

}

func setupManager(log logr.Logger, syncPeriod *time.Duration) manager.Manager {
	// Get a config to talk to the apiserver
	cfg, err := config.GetConfig()
	kingpin.FatalIfError(err, "Cannot get config")

	log.Info("Sync period", "duration", syncPeriod.String())

	// Create a new Cmd to provide shared dependencies and start components
	mgr, err := manager.New(cfg, manager.Options{SyncPeriod: syncPeriod})
	kingpin.FatalIfError(err, "Cannot create manager")

	// add all resources to the manager's runtime scheme
	log.Info("Adding schemes")
	kingpin.FatalIfError(addToScheme(mgr.GetScheme()), "Cannot add APIs to scheme")

	return mgr
}

func closeOrError(c io.Closer) {
	err := c.Close()
	kingpin.FatalIfError(err, "Cannot close file")
}

func controllerSetupWithManager(mgr manager.Manager) error {
	c := &workload.Controllers{}
	return c.SetupWithManager(mgr)
}

func stacksControllerSetupWithManager(mgr manager.Manager, allowedGroups string) error {
	c := stacksController.Controllers{AllowedGroups: allowedGroups}
	return c.SetupWithManager(mgr)
}

func stacksTemplateControllerSetupWithManager(mgr manager.Manager) error {
	c := templatesController.Controllers{}
	return c.SetupWithManager(mgr)
}

// addToScheme adds all resources to the runtime scheme.
func addToScheme(scheme *runtime.Scheme) error {
	if err := apis.AddToScheme(scheme); err != nil {
		return err
	}

	if err := apiextensionsv1beta1.AddToScheme(scheme); err != nil {
		return err
	}

	return nil
}
