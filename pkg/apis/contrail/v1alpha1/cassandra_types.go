package v1alpha1

import (
	"context"

	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// CassandraStatus defines the observed state of Cassandra

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cassandra is the Schema for the cassandras API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraSpec `json:"spec,omitempty"`
	Status Status        `json:"status,omitempty"`
}

// CassandraSpec is the Spec for the cassandras API
// +k8s:openapi-gen=true
type CassandraSpec struct {
	CommonConfiguration  CommonConfiguration    `json:"commonConfiguration"`
	ServiceConfiguration CassandraConfiguration `json:"serviceConfiguration"`
}

// CassandraConfiguration is the Spec for the cassandras API
// +k8s:openapi-gen=true
type CassandraConfiguration struct {
	Images         map[string]string `json:"images"`
	ClusterName    string            `json:"clusterName,omitempty"`
	ListenAddress  string            `json:"listenAddress,omitempty"`
	Port           int               `json:"port,omitempty"`
	CqlPort        int               `json:"cqlPort,omitempty"`
	SslStoragePort int               `json:"sslStoragePort,omitempty"`
	StoragePort    int               `json:"storagePort,omitempty"`
	JmxLocalPort   int               `json:"jmxLocalPort,omitempty"`
}

// CassandraList contains a list of Cassandra
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CassandraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cassandra `json:"items"`
}

func (c Cassandra) Crd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetCassandraCrd()
}

func (c Cassandra) GetCrd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetCassandraCrd()
}

func init() {
	SchemeBuilder.Register(&Cassandra{}, &CassandraList{})
}

func (c *Cassandra) Get(client client.Client, request reconcile.Request) error {
	err := client.Get(context.TODO(), request.NamespacedName, c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cassandra) Update(client client.Client) error {
	err := client.Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cassandra) Create(client client.Client) error {
	err := client.Create(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cassandra) Delete(client client.Client) error {
	err := client.Delete(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}
