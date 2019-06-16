package manager
	
import(
	"context"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/api/errors"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/config"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/control"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/cassandra"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/zookeeper"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/rabbitmq"
)

func (r *ReconcileManager) createCrd(instance *contrailv1alpha1.Manager, crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Reconciling Manager")
	newCrd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crd.Spec.Names.Plural + "." + crd.Spec.Group, Namespace: newCrd.Namespace}, &newCrd)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group +" not found. Creating it.")
		err = r.client.Create(context.TODO(), crd)	
		if err != nil {
			reqLogger.Error(err, "Failed to create new crd.", "crd.Namespace", crd.Namespace, "crd.Name", crd.Name)
			return err
		}
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group +" created.")
	}
	return nil
}

func (r *ReconcileManager) updateCrd(instance *contrailv1alpha1.Manager, crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Reconciling Manager")
	newCrd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crd.Spec.Names.Plural + "." + crd.Spec.Group, Namespace: newCrd.Namespace}, &newCrd)
	if err != nil{
		reqLogger.Info("Failed to get CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group)
		return err
	}
	controllerutil.SetControllerReference(instance, &newCrd, r.scheme)
	err = r.client.Update(context.TODO(), &newCrd)
	if err != nil{
		reqLogger.Info("Resource version: " + crd.ObjectMeta.ResourceVersion)
		reqLogger.Info("Failed to update CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group)
		return err
	}
	reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group +" updated.")
	return nil
}


func (r *ReconcileManager) ActivateResource(instance *contrailv1alpha1.Manager,
	ro runtime.Object,
	crd *apiextensionsv1beta1.CustomResourceDefinition) error {
		err := r.createCrd(instance, crd)
		if err != nil {
			return err
		}
		err = r.updateCrd(instance, crd)
		if err != nil {
			return err
		}
		err = r.addWatch(ro)
		if err != nil {
			return err
		}	
		return nil
}


func (r *ReconcileManager) ActivateService(instance *contrailv1alpha1.Manager) error{
	var err error
	var ConfigStatus contrailv1alpha1.ServiceStatus
	var ControlStatus contrailv1alpha1.ServiceStatus
	var CassandraStatus contrailv1alpha1.ServiceStatus
	var ZookeeperStatus contrailv1alpha1.ServiceStatus
	var RabbitmqStatus contrailv1alpha1.ServiceStatus
	ConfigActive := true
	if instance.Status.Config == nil {
		ConfigActive = false
		active := true
		ConfigStatus = contrailv1alpha1.ServiceStatus{
			Active: &active,
		}
	} else if instance.Status.Config.Active == nil {
		ConfigActive = false
		active := true
		ConfigStatus = *instance.Status.Config
		ConfigStatus.Active = &active

	}
	ControlActive := true
	if instance.Status.Control == nil {
		ControlActive = false
		active := true
		ControlStatus = contrailv1alpha1.ServiceStatus{
			Active: &active,
		}
	} else if instance.Status.Control.Active == nil {
		ControlActive = false
		active := true
		ControlStatus = *instance.Status.Control
		ControlStatus.Active = &active

	}
	CassandraActive := true
	if instance.Status.Cassandra == nil {
		CassandraActive = false
		active := true
		CassandraStatus = contrailv1alpha1.ServiceStatus{
			Active: &active,
		}
	} else if instance.Status.Cassandra.Active == nil {
		CassandraActive = false
		active := true
		CassandraStatus = *instance.Status.Cassandra
		CassandraStatus.Active = &active

	}
	ZookeeperActive := true
	if instance.Status.Zookeeper == nil {
		ZookeeperActive = false
		active := true
		ZookeeperStatus = contrailv1alpha1.ServiceStatus{
			Active: &active,
		}
	} else if instance.Status.Zookeeper.Active == nil {
		ZookeeperActive = false
		active := true
		ZookeeperStatus = *instance.Status.Zookeeper
		ZookeeperStatus.Active = &active

	}
	RabbitmqActive := true
	if instance.Status.Rabbitmq == nil {
		RabbitmqActive = false
		active := true
		RabbitmqStatus = contrailv1alpha1.ServiceStatus{
			Active: &active,
		}
	} else if instance.Status.Rabbitmq.Active == nil {
		RabbitmqActive = false
		active := true
		RabbitmqStatus = *instance.Status.Rabbitmq
		RabbitmqStatus.Active = &active

	}
	if !ConfigActive{
		if instance.Spec.Config != nil {
			ConfigActivated := instance.Spec.Config.Activate
			if *ConfigActivated{
				resource := contrailv1alpha1.Config{}
				err = r.ActivateResource(instance, &resource, crds.GetConfigCrd())
				if err != nil {
					return err
				}
			}
		}
	}
	if !ControlActive{
		if instance.Spec.Control != nil {
			ControlActivated := instance.Spec.Control.Activate
			if *ControlActivated{
				resource := contrailv1alpha1.Control{}
				err = r.ActivateResource(instance, &resource, crds.GetControlCrd())
				if err != nil {
					return err
				}
			}
		}
	}
	if !CassandraActive{
		if instance.Spec.Cassandra != nil {
			CassandraActivated := instance.Spec.Cassandra.Activate
			if *CassandraActivated{
				resource := contrailv1alpha1.Cassandra{}
				err = r.ActivateResource(instance, &resource, crds.GetCassandraCrd())
				if err != nil {
					return err
				}
			}
		}
	}
	if !ZookeeperActive{
		if instance.Spec.Zookeeper != nil {
			ZookeeperActivated := instance.Spec.Zookeeper.Activate
			if *ZookeeperActivated{
				resource := contrailv1alpha1.Zookeeper{}
				err = r.ActivateResource(instance, &resource, crds.GetZookeeperCrd())
				if err != nil {
					return err
				}
			}
		}
	}
	if !RabbitmqActive{
		if instance.Spec.Rabbitmq != nil {
			RabbitmqActivated := instance.Spec.Rabbitmq.Activate
			if *RabbitmqActivated{
				resource := contrailv1alpha1.Rabbitmq{}
				err = r.ActivateResource(instance, &resource, crds.GetRabbitmqCrd())
				if err != nil {
					return err
				}
			}
		}
	}
	if !ConfigActive{
		if instance.Spec.Config != nil {
			ConfigActivated := instance.Spec.Config.Activate
			if *ConfigActivated{
				err = config.Add(r.manager)
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
	if !ControlActive{
		if instance.Spec.Control != nil {
			ControlActivated := instance.Spec.Control.Activate
			if *ControlActivated{
				err = control.Add(r.manager)
				if err != nil {
					return err
				}
			}
			instance.Status.Control = &ControlStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}
		}
	}
	if !CassandraActive{
		if instance.Spec.Cassandra != nil {
			CassandraActivated := instance.Spec.Cassandra.Activate
			if *CassandraActivated{
				err = cassandra.Add(r.manager)
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
	if !ZookeeperActive{
		if instance.Spec.Zookeeper != nil {
			ZookeeperActivated := instance.Spec.Zookeeper.Activate
			if *ZookeeperActivated{
				err = zookeeper.Add(r.manager)
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
	if !RabbitmqActive{
		if instance.Spec.Rabbitmq != nil {
			RabbitmqActivated := instance.Spec.Rabbitmq.Activate
			if *RabbitmqActivated{
				err = rabbitmq.Add(r.manager)
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
	
	