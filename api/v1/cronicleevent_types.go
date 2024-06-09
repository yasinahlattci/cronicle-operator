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

package v1

import (
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronicleEventSpec defines the desired state of CronicleEvent
type CronicleEventSpec struct {
	// +kubebuilder:default=0
	CatchUp int `json:"catchUp,omitempty"`

	// +kubebuilder:validation:Required
	Category string `json:"category,omitempty"`

	CpuLimit   int `json:"cpuLimit,omitempty"`
	CpuSustain int `json:"cpuSustain,omitempty"`
	Detached   int `json:"detached,omitempty"`

	// +kubebuilder:default=1
	// +kubebuilder:validation:Required
	Enabled int `json:"enabled,omitempty"`

	LogMaxSize    int    `json:"logMaxSize,omitempty"`
	MaxChildren   int    `json:"maxChildren,omitempty"`
	MemoryLimit   int    `json:"memoryLimit,omitempty"`
	MemorySustain int    `json:"memorySustain,omitempty"`
	Multiplex     int    `json:"multiplex,omitempty"`
	Notes         string `json:"notes,omitempty"`
	NotifyFail    string `json:"notifyFail,omitempty"`
	NotifySuccess string `json:"notifySuccess,omitempty"`

	// +kubebuilder:validation:Required
	Params v1.JSON `json:"params,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:default="shellplug"
	Plugin string `json:"plugin,omitempty"`

	Retries    int `json:"retries,omitempty"`
	RetryDelay int `json:"retryDelay,omitempty"`

	// +kubebuilder:validation:Required
	Target string `json:"target,omitempty"`

	Timeout int `json:"timeout,omitempty"`

	// +kubebuilder:validation:Required
	Timezone string `json:"timezone,omitempty"`

	// +kubebuilder:validation:Required
	Timing v1.JSON `json:"timing,omitempty"`

	// +kubebuilder:validation:Required
	Title string `json:"title,omitempty"`

	Webhook          string                `json:"webhook,omitempty"`
	InstanceSelector *metav1.LabelSelector `json:"instanceSelector,omitempty"`
}

// CronicleEventStatus defines the observed state of CronicleEvent
type CronicleEventStatus struct {
	EventId  string `json:"eventId,omitempty"`
	Modified int64  `json:"modified,omitempty"`

	EventStatus string `json:"eventStatus,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// CronicleEvent is the Schema for the cronicleevents API
type CronicleEvent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CronicleEventSpec   `json:"spec,omitempty"`
	Status CronicleEventStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CronicleEventList contains a list of CronicleEvent
type CronicleEventList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CronicleEvent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CronicleEvent{}, &CronicleEventList{})
}
