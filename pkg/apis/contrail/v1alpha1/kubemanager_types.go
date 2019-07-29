package v1alpha1

import (
	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Kubemanager is the Schema for the kubemanagers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Kubemanager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubemanagerSpec `json:"spec,omitempty"`
	Status Status          `json:"status,omitempty"`
}

// KubemanagerSpec is the Spec for the cassandras API
// +k8s:openapi-gen=true
type KubemanagerSpec struct {
	CommonConfiguration  CommonConfiguration      `json:"commonConfiguration"`
	ServiceConfiguration KubemanagerConfiguration `json:"serviceConfiguration"`
}

// KubemanagerConfiguration is the Spec for the cassandras API
// +k8s:openapi-gen=true
type KubemanagerConfiguration struct {
	Images map[string]string `json:"images"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// KubemanagerList contains a list of Kubemanager
type KubemanagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kubemanager `json:"items"`
}

func (c Kubemanager) GetCrd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetKubemanagerCrd()
}

func init() {
	SchemeBuilder.Register(&Kubemanager{}, &KubemanagerList{})
}
