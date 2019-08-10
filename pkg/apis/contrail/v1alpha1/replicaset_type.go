package v1alpha1

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ReplicaSet is an extension of the appsv1 ReplicaSet resource
type ReplicaSet struct {
	*appsv1.ReplicaSet
}

// IsInstance returns true if it is a replicaset
func (c *ReplicaSet) IsInstance(request *reconcile.Request, instanceType string, client client.Client) bool {
	replicaSet := &appsv1.ReplicaSet{}
	err := client.Get(context.TODO(), request.NamespacedName, replicaSet)
	if err == nil {
		request.Name = replicaSet.Labels[instanceType]
		return true
	}
	return false
}
