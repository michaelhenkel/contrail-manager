package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Vrouter is the Schema for the vrouters API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Vrouter struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VrouterSpec `json:"spec,omitempty"`
	Status Status      `json:"status,omitempty"`
}

// VrouterSpec is the Spec for the cassandras API
// +k8s:openapi-gen=true
type VrouterSpec struct {
	CommonConfiguration  CommonConfiguration  `json:"commonConfiguration"`
	ServiceConfiguration VrouterConfiguration `json:"serviceConfiguration"`
}

// VrouterConfiguration is the Spec for the cassandras API
// +k8s:openapi-gen=true
type VrouterConfiguration struct {
	Images map[string]string `json:"images"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VrouterList contains a list of Vrouter
type VrouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Vrouter `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Vrouter{}, &VrouterList{})
}
