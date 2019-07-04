package manager

import (
	"context"

	"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	cr "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crs"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileManager) CreateResource(instance *v1alpha1.Manager, obj runtime.Object, name string, namespace string) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Create CR")

	objectKind := obj.GetObjectKind()
	groupVersionKind := objectKind.GroupVersionKind()

	gkv := schema.FromAPIVersionAndKind(groupVersionKind.Group+"/"+groupVersionKind.Version, groupVersionKind.Kind)
	newObj, err := scheme.Scheme.New(gkv)
	if err != nil {
		return err
	}

	newObj = obj

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, newObj)
	if err != nil && errors.IsNotFound(err) {

		switch groupVersionKind.Kind {

		case "Cassandra":
			typedObject := &v1alpha1.Cassandra{}
			typedObject = newObj.(*v1alpha1.Cassandra)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			crLabel := map[string]string{"contrail_manager": "cassandra"}
			typedObject.ObjectMeta.SetLabels(crLabel)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil {
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name + " Created.")
			return nil

		}
	}
	return nil
}

func (r *ReconcileManager) ManageCr(request reconcile.Request) error {
	var err error
	instance := &v1alpha1.Manager{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return err
		}
		return err
	}
	cassandraCreationStatus := false
	if instance.Status.Cassandra != nil {
		if instance.Status.Cassandra.Created != nil {
			cassandraCreationStatus = *instance.Status.Cassandra.Created
		}
	}

	cassandraCreationIntent := false
	if instance.Spec.Cassandra != nil {
		if instance.Spec.Cassandra.Create != nil {
			cassandraCreationIntent = *instance.Spec.Cassandra.Create
		}
	}
	if cassandraCreationIntent && !cassandraCreationStatus {
		//Create Cassandra
		cr := cr.GetCassandraCr()
		cr.Spec = instance.Spec.Cassandra
		cr.Name = instance.Name
		cr.Namespace = instance.Namespace
		if instance.Spec.Size != nil && cr.Spec.Size == nil {
			cr.Spec.Size = instance.Spec.Size
		}
		if instance.Spec.HostNetwork != nil && cr.Spec.HostNetwork == nil {
			cr.Spec.HostNetwork = instance.Spec.HostNetwork
		}
		err := r.CreateResource(instance, cr, cr.Name, cr.Namespace)
		if err != nil {
			return err
		}
		created := true
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, instance)
		if err != nil {
			if errors.IsNotFound(err) {
				return err
			}
			return err
		}
		if instance.Status.Cassandra == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Cassandra = status
		} else {
			instance.Status.Cassandra.Created = &created
		}
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
		err = r.client.Update(context.TODO(), instance)
		if err != nil {
			return err
		}

	}

	if !cassandraCreationIntent && cassandraCreationStatus {
		//Delete Cassandra
		cr := cr.GetCassandraCr()
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, cr)
		if err != nil && errors.IsNotFound(err) {
			return nil
		}
		err = r.client.Delete(context.TODO(), cr)
		if err != nil {
			return err
		}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, instance)
		if err != nil {
			if errors.IsNotFound(err) {
				return err
			}
			return err
		}
		created := false
		if instance.Status.Cassandra == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Cassandra = status
		} else {
			instance.Status.Cassandra.Created = &created
		}
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}

	}
	return nil
}
