package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

/*
type Service struct {
	Activate      *bool `json:"activate,omitempty"`
	Image         string `json:"image,omitempty"`
	Size          *int `json:"size,omitempty"`
	Configuration map[string]string `json:"configuration,omitempty"`
}

type Global struct {

}
// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.
*/
// ManagerSpec defines the desired state of Manager
// +k8s:openapi-gen=true
type ManagerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Config              *Service `json:"config,omitempty"`
	Cassandra           *Service `json:"cassandra,omitempty"`
	Zookeeper           *Service `json:"zookeeper,omitempty"`
	Rabbitmq            *Service `json:"rabbitmq,omitempty"`
	Size                *int32   `json:"size,omitempty"`
	HostNetwork         *bool    `json:"hostNetwork,omitempty"`
	ContrailStatusImage string   `json:"contrailStatusImage,omitempty"`
}

// ManagerStatus defines the observed state of Manager
// +k8s:openapi-gen=true
type ManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Config    *ServiceStatus `json:"config"`
	Cassandra *ServiceStatus `json:"cassandra"`
	Zookeeper *ServiceStatus `json:"zookeeper"`
	Rabbitmq  *ServiceStatus `json:"rabbitmq"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Manager is the Schema for the managers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Manager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagerSpec   `json:"spec,omitempty"`
	Status ManagerStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ManagerList contains a list of Manager
type ManagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Manager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Manager{}, &ManagerList{})
}
