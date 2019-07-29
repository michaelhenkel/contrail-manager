package v1alpha1

import (
	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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

	Spec   *Service `json:"spec,omitempty"`
	Status Status   `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// VrouterList contains a list of Vrouter
type VrouterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Vrouter `json:"items"`
}

func (v Vrouter) GetCrd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetVrouterCrd()
}

func init() {
	SchemeBuilder.Register(&Vrouter{}, &VrouterList{})
}