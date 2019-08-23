// +build !ignore_autogenerated

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudMemorystoreInstance) DeepCopyInto(out *CloudMemorystoreInstance) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudMemorystoreInstance.
func (in *CloudMemorystoreInstance) DeepCopy() *CloudMemorystoreInstance {
	if in == nil {
		return nil
	}
	out := new(CloudMemorystoreInstance)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CloudMemorystoreInstance) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudMemorystoreInstanceClass) DeepCopyInto(out *CloudMemorystoreInstanceClass) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.SpecTemplate.DeepCopyInto(&out.SpecTemplate)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudMemorystoreInstanceClass.
func (in *CloudMemorystoreInstanceClass) DeepCopy() *CloudMemorystoreInstanceClass {
	if in == nil {
		return nil
	}
	out := new(CloudMemorystoreInstanceClass)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CloudMemorystoreInstanceClass) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudMemorystoreInstanceClassList) DeepCopyInto(out *CloudMemorystoreInstanceClassList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CloudMemorystoreInstanceClass, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudMemorystoreInstanceClassList.
func (in *CloudMemorystoreInstanceClassList) DeepCopy() *CloudMemorystoreInstanceClassList {
	if in == nil {
		return nil
	}
	out := new(CloudMemorystoreInstanceClassList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CloudMemorystoreInstanceClassList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudMemorystoreInstanceClassSpecTemplate) DeepCopyInto(out *CloudMemorystoreInstanceClassSpecTemplate) {
	*out = *in
	in.ResourceClassSpecTemplate.DeepCopyInto(&out.ResourceClassSpecTemplate)
	in.CloudMemorystoreInstanceParameters.DeepCopyInto(&out.CloudMemorystoreInstanceParameters)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudMemorystoreInstanceClassSpecTemplate.
func (in *CloudMemorystoreInstanceClassSpecTemplate) DeepCopy() *CloudMemorystoreInstanceClassSpecTemplate {
	if in == nil {
		return nil
	}
	out := new(CloudMemorystoreInstanceClassSpecTemplate)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudMemorystoreInstanceList) DeepCopyInto(out *CloudMemorystoreInstanceList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]CloudMemorystoreInstance, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudMemorystoreInstanceList.
func (in *CloudMemorystoreInstanceList) DeepCopy() *CloudMemorystoreInstanceList {
	if in == nil {
		return nil
	}
	out := new(CloudMemorystoreInstanceList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *CloudMemorystoreInstanceList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudMemorystoreInstanceParameters) DeepCopyInto(out *CloudMemorystoreInstanceParameters) {
	*out = *in
	if in.RedisConfigs != nil {
		in, out := &in.RedisConfigs, &out.RedisConfigs
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudMemorystoreInstanceParameters.
func (in *CloudMemorystoreInstanceParameters) DeepCopy() *CloudMemorystoreInstanceParameters {
	if in == nil {
		return nil
	}
	out := new(CloudMemorystoreInstanceParameters)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudMemorystoreInstanceSpec) DeepCopyInto(out *CloudMemorystoreInstanceSpec) {
	*out = *in
	in.ResourceSpec.DeepCopyInto(&out.ResourceSpec)
	in.CloudMemorystoreInstanceParameters.DeepCopyInto(&out.CloudMemorystoreInstanceParameters)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudMemorystoreInstanceSpec.
func (in *CloudMemorystoreInstanceSpec) DeepCopy() *CloudMemorystoreInstanceSpec {
	if in == nil {
		return nil
	}
	out := new(CloudMemorystoreInstanceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CloudMemorystoreInstanceStatus) DeepCopyInto(out *CloudMemorystoreInstanceStatus) {
	*out = *in
	in.ResourceStatus.DeepCopyInto(&out.ResourceStatus)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CloudMemorystoreInstanceStatus.
func (in *CloudMemorystoreInstanceStatus) DeepCopy() *CloudMemorystoreInstanceStatus {
	if in == nil {
		return nil
	}
	out := new(CloudMemorystoreInstanceStatus)
	in.DeepCopyInto(out)
	return out
}
