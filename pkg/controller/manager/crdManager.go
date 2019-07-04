package manager

import (
	"context"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/cassandra"
	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func (r *ReconcileManager) createCrd(instance *v1alpha1.Manager, crd *apiextensionsv1beta1.CustomResourceDefinition) (bool, error) {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Creating CRD")
	newCrd := apiextensionsv1beta1.CustomResourceDefinition{}
	crdCreated := false
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crd.Spec.Names.Plural + "." + crd.Spec.Group, Namespace: newCrd.Namespace}, &newCrd)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group + " not found. Creating it.")
		controllerutil.SetControllerReference(instance, crd, r.scheme)
		err = r.client.Create(context.TODO(), crd)
		if err != nil {
			reqLogger.Error(err, "Failed to create new crd.", "crd.Namespace", crd.Namespace, "crd.Name", crd.Name)
			return crdCreated, err
		}
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group + " created.")
		crdCreated = true
	} else if err == nil {
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group + " already exists.")
	}
	return crdCreated, nil
}

func (r *ReconcileManager) ManageCrd(request reconcile.Request) error {
	var err error
	instance := &v1alpha1.Manager{}
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		return err
	}
	cassandraActivationStatus := false
	if instance.Status.Cassandra != nil {
		if instance.Status.Cassandra.Active != nil {
			cassandraActivationStatus = *instance.Status.Cassandra.Active
		}
	}

	cassandraActivationIntent := false
	if instance.Spec.Cassandra != nil {
		if instance.Spec.Cassandra.Activate != nil {
			cassandraActivationIntent = *instance.Spec.Cassandra.Activate
		}
	}
	if cassandraActivationIntent && !cassandraActivationStatus {
		crdCreated := false
		crdCreated, err = r.createCrd(instance, crds.GetCassandraCrd())
		if err != nil {
			return err
		}
		if crdCreated {
			err = cassandra.Add(r.manager)
			if err != nil {
				return err
			}
		}
		err = r.controller.Watch(&source.Kind{Type: &v1alpha1.Cassandra{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &v1alpha1.Manager{},
		})
		if err != nil {
			return err
		}
		active := true
		if instance.Status.Cassandra == nil {
			status := &v1alpha1.ServiceStatus{
				Active: &active,
			}
			instance.Status.Cassandra = status
		} else {
			instance.Status.Cassandra.Active = &active
		}

		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	}
	return nil
}
