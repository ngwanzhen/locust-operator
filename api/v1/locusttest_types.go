/*
Copyright 2023.

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

// LocustTestSpec defines the desired state of LocustTest
type LocustTestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	// NumPods    int32  `json:"numPods,omitempty"`
	Image      string `json:"image,omitempty"`
	ConfigMap  string `json:"configMap,omitempty"`
	Secret     string `json:"secret,omitempty"`
	HostUrl    string `json:"hostUrl,omitempty"`
	NumWorkers int32  `json:"numWorkers,omitempty"`
	// +kubebuilder:default=100
	Users int32 `json:"users"`
	// +kubebuilder:default=3
	SpawnRate int32 `json:"spawnRate"`
	// +kubebuilder:default="5m"
	RunTime string `json:"runTime"`
	// +kubebuilder:default=false
	AutoStart bool `json:"autoStart"`
}

// LocustTestStatus defines the observed state of LocustTest
type LocustTestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	DeploymentHappy bool `json:"deploymentHappy,omitempty"`
	PodsHappy       bool `json:"podsHappy,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// LocustTest is the Schema for the locusttests API
type LocustTest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   LocustTestSpec   `json:"spec,omitempty"`
	Status LocustTestStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// LocustTestList contains a list of LocustTest
type LocustTestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []LocustTest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&LocustTest{}, &LocustTestList{})
}
