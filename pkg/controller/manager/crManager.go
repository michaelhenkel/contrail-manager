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

		switch groupVersionKind.Kind{
		case "Cassandra":
			typedObject := &v1alpha1.Cassandra{}
			typedObject = newObj.(*v1alpha1.Cassandra)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" Created.")
			/*
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, typedObject)
			if err != nil {
				reqLogger.Info("Failed to get created CR " + name)
				return err
			}
			
			err = r.client.Update(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to update created CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" updated.")
			*/
		case "Zookeeper":
			typedObject := &v1alpha1.Zookeeper{}
			typedObject = newObj.(*v1alpha1.Zookeeper)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" Created.")
			/*
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, typedObject)
			if err != nil {
				reqLogger.Info("Failed to get created CR " + name)
				return err
			}
			
			err = r.client.Update(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to update created CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" updated.")
			*/
		case "Rabbitmq":
			typedObject := &v1alpha1.Rabbitmq{}
			typedObject = newObj.(*v1alpha1.Rabbitmq)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" Created.")
			/*
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, typedObject)
			if err != nil {
				reqLogger.Info("Failed to get created CR " + name)
				return err
			}
			
			err = r.client.Update(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to update created CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" updated.")
			*/
		case "Config":
			typedObject := &v1alpha1.Config{}
			typedObject = newObj.(*v1alpha1.Config)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" Created.")
			/*
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, typedObject)
			if err != nil {
				reqLogger.Info("Failed to get created CR " + name)
				return err
			}
			
			err = r.client.Update(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to update created CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" updated.")
			*/
		case "Control":
			typedObject := &v1alpha1.Control{}
			typedObject = newObj.(*v1alpha1.Control)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" Created.")
			/*
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, typedObject)
			if err != nil {
				reqLogger.Info("Failed to get created CR " + name)
				return err
			}
			
			err = r.client.Update(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to update created CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" updated.")
			*/
		case "Kubemanager":
			typedObject := &v1alpha1.Kubemanager{}
			typedObject = newObj.(*v1alpha1.Kubemanager)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" Created.")
			/*
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, typedObject)
			if err != nil {
				reqLogger.Info("Failed to get created CR " + name)
				return err
			}
			
			err = r.client.Update(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to update created CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" updated.")
			*/
		case "Webui":
			typedObject := &v1alpha1.Webui{}
			typedObject = newObj.(*v1alpha1.Webui)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" Created.")
			/*
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, typedObject)
			if err != nil {
				reqLogger.Info("Failed to get created CR " + name)
				return err
			}
			
			err = r.client.Update(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to update created CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" updated.")
			*/
		case "Vrouter":
			typedObject := &v1alpha1.Vrouter{}
			typedObject = newObj.(*v1alpha1.Vrouter)
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" Created.")
			/*
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, typedObject)
			if err != nil {
				reqLogger.Info("Failed to get created CR " + name)
				return err
			}
			
			err = r.client.Update(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to update created CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" updated.")
			*/
		}
	}
	return nil
}
/*
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
*/

func (r *ReconcileManager) ManageCr(instance *v1alpha1.Manager) error{
	var err error
	var CassandraStatus v1alpha1.ServiceStatus
	var ZookeeperStatus v1alpha1.ServiceStatus
	var RabbitmqStatus v1alpha1.ServiceStatus
	var ConfigStatus v1alpha1.ServiceStatus
	var ControlStatus v1alpha1.ServiceStatus
	var KubemanagerStatus v1alpha1.ServiceStatus
	var WebuiStatus v1alpha1.ServiceStatus
	var VrouterStatus v1alpha1.ServiceStatus
	CassandraCreated := true
	if instance.Status.Cassandra == nil {
		CassandraCreated = false
		CassandraStatus = v1alpha1.ServiceStatus{
			Created: instance.Spec.Cassandra.Create,
		}
	} else if instance.Status.Cassandra.Created == nil {
		CassandraCreated = false
		CassandraStatus = *instance.Status.Cassandra
		CassandraStatus.Created = instance.Spec.Cassandra.Create
	} else if *instance.Status.Cassandra.Created && !*instance.Spec.Cassandra.Create {
		cr := cr.GetCassandraCr()
		cr.Name = instance.Name
		cr.Namespace = instance.Namespace
		err := r.client.Delete(context.TODO(), cr)
		if err != nil {
			return err
		}
		instance.Status.Cassandra.Created = instance.Spec.Cassandra.Create
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	} else if !*instance.Status.Cassandra.Created && *instance.Spec.Cassandra.Create {
		CassandraCreated = false
	}
	ZookeeperCreated := true
	if instance.Status.Zookeeper == nil {
		ZookeeperCreated = false
		ZookeeperStatus = v1alpha1.ServiceStatus{
			Created: instance.Spec.Zookeeper.Create,
		}
	} else if instance.Status.Zookeeper.Created == nil {
		ZookeeperCreated = false
		ZookeeperStatus = *instance.Status.Zookeeper
		ZookeeperStatus.Created = instance.Spec.Zookeeper.Create
	} else if *instance.Status.Zookeeper.Created && !*instance.Spec.Zookeeper.Create {
		cr := cr.GetZookeeperCr()
		cr.Name = instance.Name
		cr.Namespace = instance.Namespace
		err := r.client.Delete(context.TODO(), cr)
		if err != nil {
			return err
		}
		instance.Status.Zookeeper.Created = instance.Spec.Zookeeper.Create
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	} else if !*instance.Status.Zookeeper.Created && *instance.Spec.Zookeeper.Create {
		ZookeeperCreated = false
	}
	RabbitmqCreated := true
	if instance.Status.Rabbitmq == nil {
		RabbitmqCreated = false
		RabbitmqStatus = v1alpha1.ServiceStatus{
			Created: instance.Spec.Rabbitmq.Create,
		}
	} else if instance.Status.Rabbitmq.Created == nil {
		RabbitmqCreated = false
		RabbitmqStatus = *instance.Status.Rabbitmq
		RabbitmqStatus.Created = instance.Spec.Rabbitmq.Create
	} else if *instance.Status.Rabbitmq.Created && !*instance.Spec.Rabbitmq.Create {
		cr := cr.GetRabbitmqCr()
		cr.Name = instance.Name
		cr.Namespace = instance.Namespace
		err := r.client.Delete(context.TODO(), cr)
		if err != nil {
			return err
		}
		instance.Status.Rabbitmq.Created = instance.Spec.Rabbitmq.Create
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	} else if !*instance.Status.Rabbitmq.Created && *instance.Spec.Rabbitmq.Create {
		RabbitmqCreated = false
	}
	ConfigCreated := true
	if instance.Status.Config == nil {
		ConfigCreated = false
		ConfigStatus = v1alpha1.ServiceStatus{
			Created: instance.Spec.Config.Create,
		}
	} else if instance.Status.Config.Created == nil {
		ConfigCreated = false
		ConfigStatus = *instance.Status.Config
		ConfigStatus.Created = instance.Spec.Config.Create
	} else if *instance.Status.Config.Created && !*instance.Spec.Config.Create {
		cr := cr.GetConfigCr()
		cr.Name = instance.Name
		cr.Namespace = instance.Namespace
		err := r.client.Delete(context.TODO(), cr)
		if err != nil {
			return err
		}
		instance.Status.Config.Created = instance.Spec.Config.Create
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	} else if !*instance.Status.Config.Created && *instance.Spec.Config.Create {
		ConfigCreated = false
	}
	ControlCreated := true
	if instance.Status.Control == nil {
		ControlCreated = false
		ControlStatus = v1alpha1.ServiceStatus{
			Created: instance.Spec.Control.Create,
		}
	} else if instance.Status.Control.Created == nil {
		ControlCreated = false
		ControlStatus = *instance.Status.Control
		ControlStatus.Created = instance.Spec.Control.Create
	} else if *instance.Status.Control.Created && !*instance.Spec.Control.Create {
		cr := cr.GetControlCr()
		cr.Name = instance.Name
		cr.Namespace = instance.Namespace
		err := r.client.Delete(context.TODO(), cr)
		if err != nil {
			return err
		}
		instance.Status.Control.Created = instance.Spec.Control.Create
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	} else if !*instance.Status.Control.Created && *instance.Spec.Control.Create {
		ControlCreated = false
	}
	KubemanagerCreated := true
	if instance.Status.Kubemanager == nil {
		KubemanagerCreated = false
		KubemanagerStatus = v1alpha1.ServiceStatus{
			Created: instance.Spec.Kubemanager.Create,
		}
	} else if instance.Status.Kubemanager.Created == nil {
		KubemanagerCreated = false
		KubemanagerStatus = *instance.Status.Kubemanager
		KubemanagerStatus.Created = instance.Spec.Kubemanager.Create
	} else if *instance.Status.Kubemanager.Created && !*instance.Spec.Kubemanager.Create {
		cr := cr.GetKubemanagerCr()
		cr.Name = instance.Name
		cr.Namespace = instance.Namespace
		err := r.client.Delete(context.TODO(), cr)
		if err != nil {
			return err
		}
		instance.Status.Kubemanager.Created = instance.Spec.Kubemanager.Create
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	} else if !*instance.Status.Kubemanager.Created && *instance.Spec.Kubemanager.Create {
		KubemanagerCreated = false
	}
	WebuiCreated := true
	if instance.Status.Webui == nil {
		WebuiCreated = false
		WebuiStatus = v1alpha1.ServiceStatus{
			Created: instance.Spec.Webui.Create,
		}
	} else if instance.Status.Webui.Created == nil {
		WebuiCreated = false
		WebuiStatus = *instance.Status.Webui
		WebuiStatus.Created = instance.Spec.Webui.Create
	} else if *instance.Status.Webui.Created && !*instance.Spec.Webui.Create {
		cr := cr.GetWebuiCr()
		cr.Name = instance.Name
		cr.Namespace = instance.Namespace
		err := r.client.Delete(context.TODO(), cr)
		if err != nil {
			return err
		}
		instance.Status.Webui.Created = instance.Spec.Webui.Create
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	} else if !*instance.Status.Webui.Created && *instance.Spec.Webui.Create {
		WebuiCreated = false
	}
	VrouterCreated := true
	if instance.Status.Vrouter == nil {
		VrouterCreated = false
		VrouterStatus = v1alpha1.ServiceStatus{
			Created: instance.Spec.Vrouter.Create,
		}
	} else if instance.Status.Vrouter.Created == nil {
		VrouterCreated = false
		VrouterStatus = *instance.Status.Vrouter
		VrouterStatus.Created = instance.Spec.Vrouter.Create
	} else if *instance.Status.Vrouter.Created && !*instance.Spec.Vrouter.Create {
		cr := cr.GetVrouterCr()
		cr.Name = instance.Name
		cr.Namespace = instance.Namespace
		err := r.client.Delete(context.TODO(), cr)
		if err != nil {
			return err
		}
		instance.Status.Vrouter.Created = instance.Spec.Vrouter.Create
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	} else if !*instance.Status.Vrouter.Created && *instance.Spec.Vrouter.Create {
		VrouterCreated = false
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
				/*
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				*/	
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
				/*
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				*/	
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
				/*
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				*/	
			}
			instance.Status.Rabbitmq = &RabbitmqStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}			
		}
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
				/*
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				*/	
			}
			instance.Status.Config = &ConfigStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}			
		}
	}
	if !ControlCreated{
		if instance.Spec.Control != nil{
			ControlCreated := instance.Spec.Control.Create
			if *ControlCreated{
				cr := cr.GetControlCr()
				cr.Spec.Service = instance.Spec.Control
				cr.Name = instance.Name
				cr.Namespace = instance.Namespace
				err = r.CreateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				/*
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				*/	
			}
			instance.Status.Control = &ControlStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}			
		}
	}
	if !KubemanagerCreated{
		if instance.Spec.Kubemanager != nil{
			KubemanagerCreated := instance.Spec.Kubemanager.Create
			if *KubemanagerCreated{
				cr := cr.GetKubemanagerCr()
				cr.Spec.Service = instance.Spec.Kubemanager
				cr.Name = instance.Name
				cr.Namespace = instance.Namespace
				err = r.CreateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				/*
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				*/	
			}
			instance.Status.Kubemanager = &KubemanagerStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}			
		}
	}
	if !WebuiCreated{
		if instance.Spec.Webui != nil{
			WebuiCreated := instance.Spec.Webui.Create
			if *WebuiCreated{
				cr := cr.GetWebuiCr()
				cr.Spec.Service = instance.Spec.Webui
				cr.Name = instance.Name
				cr.Namespace = instance.Namespace
				err = r.CreateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				/*
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				*/	
			}
			instance.Status.Webui = &WebuiStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}			
		}
	}
	if !VrouterCreated{
		if instance.Spec.Vrouter != nil{
			VrouterCreated := instance.Spec.Vrouter.Create
			if *VrouterCreated{
				cr := cr.GetVrouterCr()
				cr.Spec.Service = instance.Spec.Vrouter
				cr.Name = instance.Name
				cr.Namespace = instance.Namespace
				err = r.CreateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				/*
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				*/	
			}
			instance.Status.Vrouter = &VrouterStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}			
		}
	}
	return nil
}
	
	