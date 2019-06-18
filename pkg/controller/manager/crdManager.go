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
	"github.com/michaelhenkel/contrail-manager/pkg/controller/vrouter"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/cassandra"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/zookeeper"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/rabbitmq"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/config"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/control"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/kubemanager"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/webui"
)

func (r *ReconcileManager) createCrd(instance *contrailv1alpha1.Manager, crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Reconciling Manager")
	newCrd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crd.Spec.Names.Plural + "." + crd.Spec.Group, Namespace: newCrd.Namespace}, &newCrd)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group +" not found. Creating it.")
		controllerutil.SetControllerReference(&newCrd, &newCrd, r.scheme)
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
		/*
		err = r.updateCrd(instance, crd)
		if err != nil {
			return err
		}
		*/
		err = r.addWatch(ro)
		if err != nil {
			return err
		}	
		return nil
}


func (r *ReconcileManager) ManageCrd(instance *contrailv1alpha1.Manager) error{
	var err error
	var VrouterStatus contrailv1alpha1.ServiceStatus
	var CassandraStatus contrailv1alpha1.ServiceStatus
	var ZookeeperStatus contrailv1alpha1.ServiceStatus
	var RabbitmqStatus contrailv1alpha1.ServiceStatus
	var ConfigStatus contrailv1alpha1.ServiceStatus
	var ControlStatus contrailv1alpha1.ServiceStatus
	var KubemanagerStatus contrailv1alpha1.ServiceStatus
	var WebuiStatus contrailv1alpha1.ServiceStatus
	VrouterActive := true
	if instance.Status.Vrouter == nil {
		VrouterActive = false
		active := true
		VrouterStatus = contrailv1alpha1.ServiceStatus{
			Active: &active,
		}
	} else if instance.Status.Vrouter.Active == nil {
		VrouterActive = false
		active := true
		VrouterStatus = *instance.Status.Vrouter
		VrouterStatus.Active = &active

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
	KubemanagerActive := true
	if instance.Status.Kubemanager == nil {
		KubemanagerActive = false
		active := true
		KubemanagerStatus = contrailv1alpha1.ServiceStatus{
			Active: &active,
		}
	} else if instance.Status.Kubemanager.Active == nil {
		KubemanagerActive = false
		active := true
		KubemanagerStatus = *instance.Status.Kubemanager
		KubemanagerStatus.Active = &active

	}
	WebuiActive := true
	if instance.Status.Webui == nil {
		WebuiActive = false
		active := true
		WebuiStatus = contrailv1alpha1.ServiceStatus{
			Active: &active,
		}
	} else if instance.Status.Webui.Active == nil {
		WebuiActive = false
		active := true
		WebuiStatus = *instance.Status.Webui
		WebuiStatus.Active = &active

	}
	if !VrouterActive{
		if instance.Spec.Vrouter != nil {
			VrouterActivated := instance.Spec.Vrouter.Activate
			if *VrouterActivated{
				resource := contrailv1alpha1.Vrouter{}
				err = r.ActivateResource(instance, &resource, crds.GetVrouterCrd())
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
	if !KubemanagerActive{
		if instance.Spec.Kubemanager != nil {
			KubemanagerActivated := instance.Spec.Kubemanager.Activate
			if *KubemanagerActivated{
				resource := contrailv1alpha1.Kubemanager{}
				err = r.ActivateResource(instance, &resource, crds.GetKubemanagerCrd())
				if err != nil {
					return err
				}
			}
		}
	}
	if !WebuiActive{
		if instance.Spec.Webui != nil {
			WebuiActivated := instance.Spec.Webui.Activate
			if *WebuiActivated{
				resource := contrailv1alpha1.Webui{}
				err = r.ActivateResource(instance, &resource, crds.GetWebuiCrd())
				if err != nil {
					return err
				}
			}
		}
	}
	if !VrouterActive{
		if instance.Spec.Vrouter != nil {
			VrouterActivated := instance.Spec.Vrouter.Activate
			if *VrouterActivated{
				err = vrouter.Add(r.manager)
				if err != nil {
					return err
				}
			}
			instance.Status.Vrouter = &VrouterStatus
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
	if !KubemanagerActive{
		if instance.Spec.Kubemanager != nil {
			KubemanagerActivated := instance.Spec.Kubemanager.Activate
			if *KubemanagerActivated{
				err = kubemanager.Add(r.manager)
				if err != nil {
					return err
				}
			}
			instance.Status.Kubemanager = &KubemanagerStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}
		}
	}
	if !WebuiActive{
		if instance.Spec.Webui != nil {
			WebuiActivated := instance.Spec.Webui.Activate
			if *WebuiActivated{
				err = webui.Add(r.manager)
				if err != nil {
					return err
				}
			}
			instance.Status.Webui = &WebuiStatus
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
	
	