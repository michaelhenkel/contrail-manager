package v1alpha1

import (
	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ZookeeperSpec is the Spec for the cassandras API
// +k8s:openapi-gen=true
type ZookeeperSpec struct {
	CommonConfiguration  CommonConfiguration    `json:"commonConfiguration"`
	ServiceConfiguration ZookeeperConfiguration `json:"serviceConfiguration"`
}

// ZookeeperConfiguration is the Spec for the cassandras API
// +k8s:openapi-gen=true
type ZookeeperConfiguration struct {
	Images       map[string]string `json:"images"`
	ClientPort   int               `json:"clientPort,omitempty"`
	ElectionPort int               `json:"electionPort,omitempty"`
	ServerPort   int               `json:"serverPort,omitempty"`
	HeapSize     string            `json:"heapSize,omitempty"`
}

// ZookeeperStatus defines the observed state of Zookeeper
// +k8s:openapi-gen=true
type ZookeeperStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Active *bool             `json:"active"`
	Nodes  map[string]string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Zookeeper is the Schema for the zookeepers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Zookeeper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZookeeperSpec `json:"spec,omitempty"`
	Status Status        `json:"status,omitempty"`
}

func (z Zookeeper) GetCrd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetZookeeperCrd()
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZookeeperList contains a list of Zookeeper
type ZookeeperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zookeeper `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Zookeeper{}, &ZookeeperList{})
}
