package v1alpha1

import (
	"context"

	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
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
	Services            *Services `json:"services,omitempty"`
	Size                *int32    `json:"size,omitempty"`
	HostNetwork         *bool     `json:"hostNetwork,omitempty"`
	ContrailStatusImage string    `json:"contrailStatusImage,omitempty"`
	ImagePullSecrets    []string  `json:"imagePullSecrets,omitempty"`
}

type Services struct {
	Config      *Service `json:"config,omitempty"`
	Control     *Service `json:"control,omitempty"`
	Kubemanager *Service `json:"kubemanager,omitempty"`
	Webui       *Service `json:"webui,omitempty"`
	Vrouter     *Service `json:"vrouter,omitempty"`
	Cassandra   *Service `json:"cassandra,omitempty"`
	Zookeeper   *Service `json:"zookeeper,omitempty"`
	Rabbitmq    *Service `json:"rabbitmq,omitempty"`
}

// ManagerStatus defines the observed state of Manager
// +k8s:openapi-gen=true
type ManagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Config      *ServiceStatus `json:"config,omitempty"`
	Control     *ServiceStatus `json:"control,omitempty"`
	Kubemanager *ServiceStatus `json:"kubemanager,omitempty"`
	Webui       *ServiceStatus `json:"webui,omitempty"`
	Vrouter     *ServiceStatus `json:"vrouter,omitempty"`
	Cassandra   *ServiceStatus `json:"cassandra,omitempty"`
	Zookeeper   *ServiceStatus `json:"zookeeper,omitempty"`
	Rabbitmq    *ServiceStatus `json:"rabbitmq,omitempty"`
}

func (m *Manager) CassandraCrd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetCassandraCrd()
}

func (m *Manager) Cassandra() *Cassandra {
	return &Cassandra{}
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

func (m *Manager) Get(client client.Client, request reconcile.Request) error {
	err := client.Get(context.TODO(), request.NamespacedName, m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) Create(client client.Client) error {
	err := client.Create(context.TODO(), m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) Update(client client.Client) error {
	err := client.Update(context.TODO(), m)
	if err != nil {
		return err
	}
	return nil
}

func (m *Manager) Delete(client client.Client) error {
	err := client.Delete(context.TODO(), m)
	if err != nil {
		return err
	}
	return nil
}

func init() {
	SchemeBuilder.Register(&Manager{}, &ManagerList{})
}
