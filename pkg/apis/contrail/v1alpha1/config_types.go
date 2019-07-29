package v1alpha1

import (
	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Config is the Schema for the configs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec `json:"spec,omitempty"`
	Status Status     `json:"status,omitempty"`
}

// ConfigSpec is the Spec for the cassandras API
// +k8s:openapi-gen=true
type ConfigSpec struct {
	CommonConfiguration  CommonConfiguration `json:"commonConfiguration"`
	ServiceConfiguration ConfigConfiguration `json:"serviceConfiguration"`
}

// ConfigConfiguration is the Spec for the cassandras API
// +k8s:openapi-gen=true
type ConfigConfiguration struct {
	Images map[string]string `json:"images"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigList contains a list of Config
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func (c Config) GetCrd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetConfigCrd()
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}
