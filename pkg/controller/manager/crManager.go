package manager
	
import(
	"context"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/api/errors"
	"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	cr "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crs"
)

func (r *ReconcileManager) UpdateResource(instance *v1alpha1.Manager, obj runtime.Object, name string, namespace string) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Reconciling Manager")
	
	objectKind := obj.GetObjectKind()
	groupVersionKind := objectKind.GroupVersionKind()

	gkv := schema.FromAPIVersionAndKind(groupVersionKind.Group + "/" + groupVersionKind.Version, groupVersionKind.Kind)
	newObj, err := scheme.Scheme.New(gkv)
	if err != nil {
		return err
	}

	newObj = obj

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, newObj)
	if err != nil{
		reqLogger.Info("Failed to get CR " + name)
		return err
	}

	switch groupVersionKind.Kind{
	case "Config":
		var typedObject *v1alpha1.Config
		typedObject = newObj.(*v1alpha1.Config)
		controllerutil.SetControllerReference(typedObject, typedObject, r.scheme)
	case "Cassandra":
		var typedObject *v1alpha1.Cassandra
		typedObject = newObj.(*v1alpha1.Cassandra)
		controllerutil.SetControllerReference(typedObject, typedObject, r.scheme)
	case "Zookeeper":
		var typedObject *v1alpha1.Zookeeper
		typedObject = newObj.(*v1alpha1.Zookeeper)
		controllerutil.SetControllerReference(typedObject, typedObject, r.scheme)
	case "Rabbitmq":
		var typedObject *v1alpha1.Rabbitmq
		typedObject = newObj.(*v1alpha1.Rabbitmq)
		controllerutil.SetControllerReference(typedObject, typedObject, r.scheme)
	}


	err = r.client.Update(context.TODO(), newObj)
	if err != nil{
		reqLogger.Info("Failed to update CR " + name)
		return err
	}
	reqLogger.Info("CR " + name +" updated.")
	return nil
}

func (r *ReconcileManager) CreateResource(instance *v1alpha1.Manager, obj runtime.Object, name string, namespace string) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Reconciling Manager")
	
	objectKind := obj.GetObjectKind()
	groupVersionKind := objectKind.GroupVersionKind()

	gkv := schema.FromAPIVersionAndKind(groupVersionKind.Group + "/" + groupVersionKind.Version, groupVersionKind.Kind)
	newObj, err := scheme.Scheme.New(gkv)
	if err != nil {
		return err
	}

	newObj = obj

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, newObj)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("CR " + name +" not found. Creating it.")
		err = r.client.Create(context.TODO(), newObj)	
		if err != nil {
			reqLogger.Error(err, "Failed to create new newObj.", "resource.Namespace", namespace, "resource.Name", name)
			return err
		}
		reqLogger.Info("CR " + namespace +" created.")
	}
	return nil
}

func (r *ReconcileManager) CreateService(instance *v1alpha1.Manager) error{
	var err error
	var ConfigStatus v1alpha1.ServiceStatus
	var CassandraStatus v1alpha1.ServiceStatus
	var ZookeeperStatus v1alpha1.ServiceStatus
	var RabbitmqStatus v1alpha1.ServiceStatus
	ConfigCreated := true
	if instance.Status.Config == nil {
		ConfigCreated = false
		active := true
		ConfigStatus = v1alpha1.ServiceStatus{
			Created: &active,
		}
	} else if instance.Status.Config.Created == nil {
		ConfigCreated = false
		active := true
		ConfigStatus = *instance.Status.Config
		ConfigStatus.Created = &active
	}
	CassandraCreated := true
	if instance.Status.Cassandra == nil {
		CassandraCreated = false
		active := true
		CassandraStatus = v1alpha1.ServiceStatus{
			Created: &active,
		}
	} else if instance.Status.Cassandra.Created == nil {
		CassandraCreated = false
		active := true
		CassandraStatus = *instance.Status.Cassandra
		CassandraStatus.Created = &active
	}
	ZookeeperCreated := true
	if instance.Status.Zookeeper == nil {
		ZookeeperCreated = false
		active := true
		ZookeeperStatus = v1alpha1.ServiceStatus{
			Created: &active,
		}
	} else if instance.Status.Zookeeper.Created == nil {
		ZookeeperCreated = false
		active := true
		ZookeeperStatus = *instance.Status.Zookeeper
		ZookeeperStatus.Created = &active
	}
	RabbitmqCreated := true
	if instance.Status.Rabbitmq == nil {
		RabbitmqCreated = false
		active := true
		RabbitmqStatus = v1alpha1.ServiceStatus{
			Created: &active,
		}
	} else if instance.Status.Rabbitmq.Created == nil {
		RabbitmqCreated = false
		active := true
		RabbitmqStatus = *instance.Status.Rabbitmq
		RabbitmqStatus.Created = &active
	}
	if !ConfigCreated{
		if instance.Spec.Config != nil{
			ConfigCreated := instance.Spec.Config.Create
			if *ConfigCreated{
				cr := cr.GetConfigCr()
				cr.Spec.Service = instance.Spec.Config
				cr.Name = instance.Name
				cr.Namespace = instance.Namespace
				err = r.CreateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}		
			}
			instance.Status.Config = &ConfigStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}			
		}
	}
	if !CassandraCreated{
		if instance.Spec.Cassandra != nil{
			CassandraCreated := instance.Spec.Cassandra.Create
			if *CassandraCreated{
				cr := cr.GetCassandraCr()
				cr.Spec.Service = instance.Spec.Cassandra
				cr.Name = instance.Name
				cr.Namespace = instance.Namespace
				err = r.CreateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}		
			}
			instance.Status.Cassandra = &CassandraStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}			
		}
	}
	if !ZookeeperCreated{
		if instance.Spec.Zookeeper != nil{
			ZookeeperCreated := instance.Spec.Zookeeper.Create
			if *ZookeeperCreated{
				cr := cr.GetZookeeperCr()
				cr.Spec.Service = instance.Spec.Zookeeper
				cr.Name = instance.Name
				cr.Namespace = instance.Namespace
				err = r.CreateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}		
			}
			instance.Status.Zookeeper = &ZookeeperStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}			
		}
	}
	if !RabbitmqCreated{
		if instance.Spec.Rabbitmq != nil{
			RabbitmqCreated := instance.Spec.Rabbitmq.Create
			if *RabbitmqCreated{
				cr := cr.GetRabbitmqCr()
				cr.Spec.Service = instance.Spec.Rabbitmq
				cr.Name = instance.Name
				cr.Namespace = instance.Namespace
				err = r.CreateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}		
			}
			instance.Status.Rabbitmq = &RabbitmqStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}			
		}
	}
	return nil
}
	
	