/*

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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// MachineFinalizer allows the reconciler to clean up resources associated with InceptionMachine before
	// removing it from the apiserver.
	MachineFinalizer = "inceptionmachine.infrastructure.cluster.x-k8s.io"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// InceptionMachineSpec defines the desired state of InceptionMachine
type InceptionMachineSpec struct {
	// ProviderID will be the only detail (todo: something else)
	// +optional
	ProviderID *string `json:"providerID,omitempty"`
}

// InceptionMachineStatus defines the observed state of InceptionMachine
type InceptionMachineStatus struct {
	// Ready denotes that the machine is ready
	Ready bool `json:"ready"`
}

// +kubebuilder:object:root=true

// InceptionMachine is the Schema for the inceptionmachines API
type InceptionMachine struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   InceptionMachineSpec   `json:"spec,omitempty"`
	Status InceptionMachineStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// InceptionMachineList contains a list of InceptionMachine
type InceptionMachineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []InceptionMachine `json:"items"`
}

func init() {
	SchemeBuilder.Register(&InceptionMachine{}, &InceptionMachineList{})
}
