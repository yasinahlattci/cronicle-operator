//go:build !ignore_autogenerated

/*
Copyright 2024 Yasin AHLATCI.

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

package cronicle_client

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CronicleParams) DeepCopyInto(out *CronicleParams) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CronicleParams.
func (in *CronicleParams) DeepCopy() *CronicleParams {
	if in == nil {
		return nil
	}
	out := new(CronicleParams)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *CronicleTiming) DeepCopyInto(out *CronicleTiming) {
	*out = *in
	if in.Hours != nil {
		in, out := &in.Hours, &out.Hours
		*out = make([]int, len(*in))
		copy(*out, *in)
	}
	if in.Days != nil {
		in, out := &in.Days, &out.Days
		*out = make([]int, len(*in))
		copy(*out, *in)
	}
	if in.Months != nil {
		in, out := &in.Months, &out.Months
		*out = make([]int, len(*in))
		copy(*out, *in)
	}
	if in.Weekdays != nil {
		in, out := &in.Weekdays, &out.Weekdays
		*out = make([]int, len(*in))
		copy(*out, *in)
	}
	if in.Years != nil {
		in, out := &in.Years, &out.Years
		*out = make([]int, len(*in))
		copy(*out, *in)
	}
	if in.Minutes != nil {
		in, out := &in.Minutes, &out.Minutes
		*out = make([]int, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new CronicleTiming.
func (in *CronicleTiming) DeepCopy() *CronicleTiming {
	if in == nil {
		return nil
	}
	out := new(CronicleTiming)
	in.DeepCopyInto(out)
	return out
}
