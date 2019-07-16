package v1alpha1

import (
	"context"

	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CassandraStatus defines the observed state of Cassandra

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cassandra is the Schema for the cassandras API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   *Service `json:"spec,omitempty"`
	Status Status   `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CassandraList contains a list of Cassandra
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

// Get implements Service Get
/*
func (c Cassandra) Get(request reconcile.Request, client client.Client) (*Cassandra, error) {
	err := client.Get(context.TODO(), request.NamespacedName, &c)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil
		}
	}
	return &c, nil
}
*/
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
