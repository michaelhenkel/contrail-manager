package v1alpha1

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Deployment struct {
	*appsv1.Deployment
}

func (d *Deployment) Get(client client.Client, request reconcile.Request) error {
	err := client.Get(context.TODO(), request.NamespacedName, d)
	if err != nil {
		return err
	}
	return nil
}

func (d *Deployment) Create(client client.Client) error {
	err := client.Create(context.TODO(), d)
	if err != nil {
		return err
	}
	return nil
}

func (d *Deployment) Update(client client.Client) error {
	err := client.Update(context.TODO(), d)
	if err != nil {
		return err
	}
	return nil
}

func (d *Deployment) Delete(client client.Client) error {
	err := client.Delete(context.TODO(), d)
	if err != nil {
		return err
	}
	return nil
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ReplicaSet struct {
	*appsv1.ReplicaSet
}

func (r *ReplicaSet) Get(client client.Client, request reconcile.Request) error {
	err := client.Get(context.TODO(), request.NamespacedName, r)
	if err != nil {
		return err
	}
	return nil
}

func (r *ReplicaSet) Create(client client.Client) error {
	err := client.Create(context.TODO(), r)
	if err != nil {
		return err
	}
	return nil
}

func (r *ReplicaSet) Update(client client.Client) error {
	err := client.Update(context.TODO(), r)
	if err != nil {
		return err
	}
	return nil
}

func (r *ReplicaSet) Delete(client client.Client) error {
	err := client.Delete(context.TODO(), r)
	if err != nil {
		return err
	}
	return nil
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Pod struct {
	*corev1.Pod
}

func (p *Pod) Get(client client.Client, request reconcile.Request) error {
	err := client.Get(context.TODO(), request.NamespacedName, p)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pod) Create(client client.Client) error {
	err := client.Create(context.TODO(), p)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pod) Update(client client.Client) error {
	err := client.Update(context.TODO(), p)
	if err != nil {
		return err
	}
	return nil
}

func (p *Pod) Delete(client client.Client) error {
	err := client.Delete(context.TODO(), p)
	if err != nil {
		return err
	}
	return nil
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ConfigMap struct {
	*corev1.ConfigMap
}

func (c *ConfigMap) Get(client client.Client, request reconcile.Request) error {
	err := client.Get(context.TODO(), request.NamespacedName, c)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConfigMap) Create(client client.Client) error {
	err := client.Create(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConfigMap) Update(client client.Client) error {
	err := client.Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *ConfigMap) Delete(client client.Client) error {
	err := client.Delete(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}
