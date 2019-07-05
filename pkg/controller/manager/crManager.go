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
		case "Control":
			typedObject := &v1alpha1.Control{}
			typedObject = newObj.(*v1alpha1.Control)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			crLabel := map[string]string{"contrail_manager": "control"}
			typedObject.ObjectMeta.SetLabels(crLabel)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil {
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name + " Created.")
			return nil
		case "Kubemanager":
			typedObject := &v1alpha1.Kubemanager{}
			typedObject = newObj.(*v1alpha1.Kubemanager)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			crLabel := map[string]string{"contrail_manager": "kubemanager"}
			typedObject.ObjectMeta.SetLabels(crLabel)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil {
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name + " Created.")
			return nil
		case "Webui":
			typedObject := &v1alpha1.Webui{}
			typedObject = newObj.(*v1alpha1.Webui)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			crLabel := map[string]string{"contrail_manager": "webui"}
			typedObject.ObjectMeta.SetLabels(crLabel)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil {
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name + " Created.")
			return nil
		case "Vrouter":
			typedObject := &v1alpha1.Vrouter{}
			typedObject = newObj.(*v1alpha1.Vrouter)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			crLabel := map[string]string{"contrail_manager": "vrouter"}
			typedObject.ObjectMeta.SetLabels(crLabel)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil {
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name + " Created.")
			return nil
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
		case "Zookeeper":
			typedObject := &v1alpha1.Zookeeper{}
			typedObject = newObj.(*v1alpha1.Zookeeper)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			crLabel := map[string]string{"contrail_manager": "zookeeper"}
			typedObject.ObjectMeta.SetLabels(crLabel)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil {
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name + " Created.")
			return nil
		case "Rabbitmq":
			typedObject := &v1alpha1.Rabbitmq{}
			typedObject = newObj.(*v1alpha1.Rabbitmq)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			crLabel := map[string]string{"contrail_manager": "rabbitmq"}
			typedObject.ObjectMeta.SetLabels(crLabel)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil {
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name + " Created.")
			return nil
		case "Config":
			typedObject := &v1alpha1.Config{}
			typedObject = newObj.(*v1alpha1.Config)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			crLabel := map[string]string{"contrail_manager": "config"}
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
	controlCreationStatus := false
	if instance.Status.Control != nil {
		if instance.Status.Control.Created != nil {
			controlCreationStatus = *instance.Status.Control.Created
		}
	}

	controlCreationIntent := false
	if instance.Spec.Services.Control != nil {
		if instance.Spec.Services.Control.Create != nil {
			controlCreationIntent = *instance.Spec.Services.Control.Create
		}
	}
	if controlCreationIntent && !controlCreationStatus {
		//Create Control
		cr := cr.GetControlCr()
		cr.Spec = instance.Spec.Services.Control
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
		if instance.Status.Control == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Control = status
		} else {
			instance.Status.Control.Created = &created
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

	if !controlCreationIntent && controlCreationStatus {
		//Delete Control
		cr := cr.GetControlCr()
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
		if instance.Status.Control == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Control = status
		} else {
			instance.Status.Control.Created = &created
		}
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}

	}
	kubemanagerCreationStatus := false
	if instance.Status.Kubemanager != nil {
		if instance.Status.Kubemanager.Created != nil {
			kubemanagerCreationStatus = *instance.Status.Kubemanager.Created
		}
	}

	kubemanagerCreationIntent := false
	if instance.Spec.Services.Kubemanager != nil {
		if instance.Spec.Services.Kubemanager.Create != nil {
			kubemanagerCreationIntent = *instance.Spec.Services.Kubemanager.Create
		}
	}
	if kubemanagerCreationIntent && !kubemanagerCreationStatus {
		//Create Kubemanager
		cr := cr.GetKubemanagerCr()
		cr.Spec = instance.Spec.Services.Kubemanager
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
		if instance.Status.Kubemanager == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Kubemanager = status
		} else {
			instance.Status.Kubemanager.Created = &created
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

	if !kubemanagerCreationIntent && kubemanagerCreationStatus {
		//Delete Kubemanager
		cr := cr.GetKubemanagerCr()
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
		if instance.Status.Kubemanager == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Kubemanager = status
		} else {
			instance.Status.Kubemanager.Created = &created
		}
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}

	}
	webuiCreationStatus := false
	if instance.Status.Webui != nil {
		if instance.Status.Webui.Created != nil {
			webuiCreationStatus = *instance.Status.Webui.Created
		}
	}

	webuiCreationIntent := false
	if instance.Spec.Services.Webui != nil {
		if instance.Spec.Services.Webui.Create != nil {
			webuiCreationIntent = *instance.Spec.Services.Webui.Create
		}
	}
	if webuiCreationIntent && !webuiCreationStatus {
		//Create Webui
		cr := cr.GetWebuiCr()
		cr.Spec = instance.Spec.Services.Webui
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
		if instance.Status.Webui == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Webui = status
		} else {
			instance.Status.Webui.Created = &created
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

	if !webuiCreationIntent && webuiCreationStatus {
		//Delete Webui
		cr := cr.GetWebuiCr()
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
		if instance.Status.Webui == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Webui = status
		} else {
			instance.Status.Webui.Created = &created
		}
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}

	}
	vrouterCreationStatus := false
	if instance.Status.Vrouter != nil {
		if instance.Status.Vrouter.Created != nil {
			vrouterCreationStatus = *instance.Status.Vrouter.Created
		}
	}

	vrouterCreationIntent := false
	if instance.Spec.Services.Vrouter != nil {
		if instance.Spec.Services.Vrouter.Create != nil {
			vrouterCreationIntent = *instance.Spec.Services.Vrouter.Create
		}
	}
	if vrouterCreationIntent && !vrouterCreationStatus {
		//Create Vrouter
		cr := cr.GetVrouterCr()
		cr.Spec = instance.Spec.Services.Vrouter
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
		if instance.Status.Vrouter == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Vrouter = status
		} else {
			instance.Status.Vrouter.Created = &created
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

	if !vrouterCreationIntent && vrouterCreationStatus {
		//Delete Vrouter
		cr := cr.GetVrouterCr()
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
		if instance.Status.Vrouter == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Vrouter = status
		} else {
			instance.Status.Vrouter.Created = &created
		}
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}

	}
	cassandraCreationStatus := false
	if instance.Status.Cassandra != nil {
		if instance.Status.Cassandra.Created != nil {
			cassandraCreationStatus = *instance.Status.Cassandra.Created
		}
	}

	cassandraCreationIntent := false
	if instance.Spec.Services.Cassandra != nil {
		if instance.Spec.Services.Cassandra.Create != nil {
			cassandraCreationIntent = *instance.Spec.Services.Cassandra.Create
		}
	}
	if cassandraCreationIntent && !cassandraCreationStatus {
		//Create Cassandra
		cr := cr.GetCassandraCr()
		cr.Spec = instance.Spec.Services.Cassandra
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
	zookeeperCreationStatus := false
	if instance.Status.Zookeeper != nil {
		if instance.Status.Zookeeper.Created != nil {
			zookeeperCreationStatus = *instance.Status.Zookeeper.Created
		}
	}

	zookeeperCreationIntent := false
	if instance.Spec.Services.Zookeeper != nil {
		if instance.Spec.Services.Zookeeper.Create != nil {
			zookeeperCreationIntent = *instance.Spec.Services.Zookeeper.Create
		}
	}
	if zookeeperCreationIntent && !zookeeperCreationStatus {
		//Create Zookeeper
		cr := cr.GetZookeeperCr()
		cr.Spec = instance.Spec.Services.Zookeeper
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
		if instance.Status.Zookeeper == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Zookeeper = status
		} else {
			instance.Status.Zookeeper.Created = &created
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

	if !zookeeperCreationIntent && zookeeperCreationStatus {
		//Delete Zookeeper
		cr := cr.GetZookeeperCr()
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
		if instance.Status.Zookeeper == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Zookeeper = status
		} else {
			instance.Status.Zookeeper.Created = &created
		}
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}

	}
	rabbitmqCreationStatus := false
	if instance.Status.Rabbitmq != nil {
		if instance.Status.Rabbitmq.Created != nil {
			rabbitmqCreationStatus = *instance.Status.Rabbitmq.Created
		}
	}

	rabbitmqCreationIntent := false
	if instance.Spec.Services.Rabbitmq != nil {
		if instance.Spec.Services.Rabbitmq.Create != nil {
			rabbitmqCreationIntent = *instance.Spec.Services.Rabbitmq.Create
		}
	}
	if rabbitmqCreationIntent && !rabbitmqCreationStatus {
		//Create Rabbitmq
		cr := cr.GetRabbitmqCr()
		cr.Spec = instance.Spec.Services.Rabbitmq
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
		if instance.Status.Rabbitmq == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Rabbitmq = status
		} else {
			instance.Status.Rabbitmq.Created = &created
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

	if !rabbitmqCreationIntent && rabbitmqCreationStatus {
		//Delete Rabbitmq
		cr := cr.GetRabbitmqCr()
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
		if instance.Status.Rabbitmq == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Rabbitmq = status
		} else {
			instance.Status.Rabbitmq.Created = &created
		}
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}

	}
	configCreationStatus := false
	if instance.Status.Config != nil {
		if instance.Status.Config.Created != nil {
			configCreationStatus = *instance.Status.Config.Created
		}
	}

	configCreationIntent := false
	if instance.Spec.Services.Config != nil {
		if instance.Spec.Services.Config.Create != nil {
			configCreationIntent = *instance.Spec.Services.Config.Create
		}
	}
	if configCreationIntent && !configCreationStatus {
		//Create Config
		cr := cr.GetConfigCr()
		cr.Spec = instance.Spec.Services.Config
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
		if instance.Status.Config == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Config = status
		} else {
			instance.Status.Config.Created = &created
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

	if !configCreationIntent && configCreationStatus {
		//Delete Config
		cr := cr.GetConfigCr()
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
		if instance.Status.Config == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.Config = status
		} else {
			instance.Status.Config.Created = &created
		}
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}

	}
	return nil
}
	
	