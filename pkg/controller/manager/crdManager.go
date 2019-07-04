package manager

import (
	"context"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/cassandra"
	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func (r *ReconcileManager) createCrd(instance *v1alpha1.Manager, crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Creating CRD")
	newCrd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crd.Spec.Names.Plural + "." + crd.Spec.Group, Namespace: newCrd.Namespace}, &newCrd)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group + " not found. Creating it.")
		//controllerutil.SetControllerReference(&newCrd, crd, r.scheme)
		err = r.client.Create(context.TODO(), crd)
		if err != nil {
			reqLogger.Error(err, "Failed to create new crd.", "crd.Namespace", crd.Namespace, "crd.Name", crd.Name)
			return err
		}
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group + " created.")
	} else if err == nil {
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group + " already exists.")
	}
	return nil
}

func (r *ReconcileManager) ManageCrd(request reconcile.Request) error {
	var err error
	instance := &v1alpha1.Manager{}
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		return err
	}
	activationStatus := false
	if instance.Status.Cassandra != nil {
		if instance.Status.Cassandra.Active != nil {
			activationStatus = *instance.Status.Cassandra.Active
		}
	}

	activationIntent := false
	if instance.Spec.Cassandra != nil {
		if instance.Spec.Cassandra.Activate != nil {
			activationIntent = *instance.Spec.Cassandra.Activate
		}
	}
	if activationIntent && !activationStatus {
		crd := crds.GetCassandraCrd()

		err = r.createCrd(instance, crd)
		if err != nil {
			return err
		}

		controllerRunning := false
		sharedIndexInformer, err := r.cache.GetInformerForKind(crd.GroupVersionKind())
		if err == nil {
			controller := sharedIndexInformer.GetController()
			if controller != nil {
				controllerRunning = true
			}
		}

		if !controllerRunning {
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
