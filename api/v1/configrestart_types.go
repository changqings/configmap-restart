/*
Copyright 2024.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ConfigrestartSpec defines the desired state of Configrestart.
type ConfigrestartSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// if this field is true, the operator will suspend the reconciliation of the resource.
	// +kubebuilder:validation:Optional
	// +kubebuilder:default=false
	Suspend bool `json:"suspend,omitempty"`

	// +kubebuilder:validation:Required
	ConfigName string `json:"configName"`

	// list of deployments to restart
	// if empty, will restart all related deployments
	// +kubebuilder:validation:Optional
	Deployments []string `json:"deployments,omitempty"`
}

// ConfigrestartStatus defines the observed state of Configrestart.
type ConfigrestartStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Configrestart is the Schema for the configrestarts API.
type Configrestart struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigrestartSpec   `json:"spec,omitempty"`
	Status ConfigrestartStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ConfigrestartList contains a list of Configrestart.
type ConfigrestartList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Configrestart `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Configrestart{}, &ConfigrestartList{})
}