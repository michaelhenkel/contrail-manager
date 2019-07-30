package manager

import (
	"context"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/cassandra"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/*
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
*/

func (r *ReconcileManager) ManageCrd(request reconcile.Request) error {
	var err error
	instance := &v1alpha1.Manager{}
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		return err
	}
	/*
		configActivationStatus := false
		if instance.Status.Config != nil {
			if instance.Status.Config.Active != nil {
				configActivationStatus = *instance.Status.Config.Active
			}
		}

		configActivationIntent := false
		if instance.Spec.Services.Config != nil {
			if instance.Spec.Services.Config.Activate != nil {
				configActivationIntent = *instance.Spec.Services.Config.Activate
			}
		}
		configResource := v1alpha1.Config{}
		configCrd := configResource.GetCrd()
		configControllerActivate := false
		if configActivationIntent && !configActivationStatus {
			err = r.createCrd(instance, configCrd)
			if err != nil {
				return err
			}

			controllerRunning := false
			sharedIndexInformer, err := r.cache.GetInformerForKind(configResource.GroupVersionKind())
			if err == nil {
				controller := sharedIndexInformer.GetController()
				if controller != nil {
					controllerRunning = true
				}
			}
			if !controllerRunning {
				configControllerActivate = true
			}

			err = r.controller.Watch(&source.Kind{Type: &v1alpha1.Config{}}, &handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &v1alpha1.Manager{},
			})
			if err != nil {
				return err
			}

			active := true
			if instance.Status.Config == nil {
				status := &v1alpha1.ServiceStatus{
					Active: &active,
				}
				instance.Status.Config = status
			} else {
				instance.Status.Config.Active = &active
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}
		}
		controlActivationStatus := false
		if instance.Status.Control != nil {
			if instance.Status.Control.Active != nil {
				controlActivationStatus = *instance.Status.Control.Active
			}
		}

		controlActivationIntent := false
		if instance.Spec.Services.Control != nil {
			if instance.Spec.Services.Control.Activate != nil {
				controlActivationIntent = *instance.Spec.Services.Control.Activate
			}
		}
		controlResource := v1alpha1.Control{}
		controlCrd := controlResource.GetCrd()
		controlControllerActivate := false
		if controlActivationIntent && !controlActivationStatus {
			err = r.createCrd(instance, controlCrd)
			if err != nil {
				return err
			}

			controllerRunning := false
			sharedIndexInformer, err := r.cache.GetInformerForKind(controlResource.GroupVersionKind())
			if err == nil {
				controller := sharedIndexInformer.GetController()
				if controller != nil {
					controllerRunning = true
				}
			}
			if !controllerRunning {
				controlControllerActivate = true
			}

			err = r.controller.Watch(&source.Kind{Type: &v1alpha1.Control{}}, &handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &v1alpha1.Manager{},
			})
			if err != nil {
				return err
			}
			active := true
			if instance.Status.Control == nil {
				status := &v1alpha1.ServiceStatus{
					Active: &active,
				}
				instance.Status.Control = status
			} else {
				instance.Status.Control.Active = &active
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}
		}
		kubemanagerActivationStatus := false
		if instance.Status.Kubemanager != nil {
			if instance.Status.Kubemanager.Active != nil {
				kubemanagerActivationStatus = *instance.Status.Kubemanager.Active
			}
		}

		kubemanagerActivationIntent := false
		if instance.Spec.Services.Kubemanager != nil {
			if instance.Spec.Services.Kubemanager.Activate != nil {
				kubemanagerActivationIntent = *instance.Spec.Services.Kubemanager.Activate
			}
		}
		kubemanagerResource := v1alpha1.Kubemanager{}
		kubemanagerCrd := kubemanagerResource.GetCrd()
		kubemanagerControllerActivate := false
		if kubemanagerActivationIntent && !kubemanagerActivationStatus {
			err = r.createCrd(instance, kubemanagerCrd)
			if err != nil {
				return err
			}

			controllerRunning := false
			sharedIndexInformer, err := r.cache.GetInformerForKind(kubemanagerResource.GroupVersionKind())
			if err == nil {
				controller := sharedIndexInformer.GetController()
				if controller != nil {
					controllerRunning = true
				}
			}
			if !controllerRunning {
				kubemanagerControllerActivate = true
			}

			err = r.controller.Watch(&source.Kind{Type: &v1alpha1.Kubemanager{}}, &handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &v1alpha1.Manager{},
			})
			if err != nil {
				return err
			}
			active := true
			if instance.Status.Kubemanager == nil {
				status := &v1alpha1.ServiceStatus{
					Active: &active,
				}
				instance.Status.Kubemanager = status
			} else {
				instance.Status.Kubemanager.Active = &active
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}
		}
		webuiActivationStatus := false
		if instance.Status.Webui != nil {
			if instance.Status.Webui.Active != nil {
				webuiActivationStatus = *instance.Status.Webui.Active
			}
		}

		webuiActivationIntent := false
		if instance.Spec.Services.Webui != nil {
			if instance.Spec.Services.Webui.Activate != nil {
				webuiActivationIntent = *instance.Spec.Services.Webui.Activate
			}
		}
		webuiResource := v1alpha1.Webui{}
		webuiCrd := webuiResource.GetCrd()
		webuiControllerActivate := false
		if webuiActivationIntent && !webuiActivationStatus {
			err = r.createCrd(instance, webuiCrd)
			if err != nil {
				return err
			}

			controllerRunning := false
			sharedIndexInformer, err := r.cache.GetInformerForKind(webuiResource.GroupVersionKind())
			if err == nil {
				controller := sharedIndexInformer.GetController()
				if controller != nil {
					controllerRunning = true
				}
			}
			if !controllerRunning {
				webuiControllerActivate = true
			}

			err = r.controller.Watch(&source.Kind{Type: &v1alpha1.Webui{}}, &handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &v1alpha1.Manager{},
			})
			if err != nil {
				return err
			}
			active := true
			if instance.Status.Webui == nil {
				status := &v1alpha1.ServiceStatus{
					Active: &active,
				}
				instance.Status.Webui = status
			} else {
				instance.Status.Webui.Active = &active
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}
		}
		vrouterActivationStatus := false
		if instance.Status.Vrouter != nil {
			if instance.Status.Vrouter.Active != nil {
				vrouterActivationStatus = *instance.Status.Vrouter.Active
			}
		}

		vrouterActivationIntent := false
		if instance.Spec.Services.Vrouter != nil {
			if instance.Spec.Services.Vrouter.Activate != nil {
				vrouterActivationIntent = *instance.Spec.Services.Vrouter.Activate
			}
		}
		vrouterResource := v1alpha1.Vrouter{}
		vrouterCrd := vrouterResource.GetCrd()
		vrouterControllerActivate := false
		if vrouterActivationIntent && !vrouterActivationStatus {
			err = r.createCrd(instance, vrouterCrd)
			if err != nil {
				return err
			}

			controllerRunning := false
			sharedIndexInformer, err := r.cache.GetInformerForKind(vrouterResource.GroupVersionKind())
			if err == nil {
				controller := sharedIndexInformer.GetController()
				if controller != nil {
					controllerRunning = true
				}
			}
			if !controllerRunning {
				vrouterControllerActivate = true
			}

			err = r.controller.Watch(&source.Kind{Type: &v1alpha1.Vrouter{}}, &handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &v1alpha1.Manager{},
			})
			if err != nil {
				return err
			}
			active := true
			if instance.Status.Vrouter == nil {
				status := &v1alpha1.ServiceStatus{
					Active: &active,
				}
				instance.Status.Vrouter = status
			} else {
				instance.Status.Vrouter.Active = &active
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}
		}
	*/
	cassandraResource := v1alpha1.Cassandra{}
	cassandraCrd := cassandraResource.GetCrd()
	err = r.createCrd(instance, cassandraCrd)
	if err != nil {
		return err
	}
	cassandraControllerRunning := false
	cassandraSharedIndexInformer, err := r.cache.GetInformerForKind(cassandraResource.GroupVersionKind())
	if err == nil {
		controller := cassandraSharedIndexInformer.GetController()
		if controller != nil {
			cassandraControllerRunning = true
		}
	}
	cassandraControllerActivate := false
	if !cassandraControllerRunning {
		cassandraControllerActivate = true
	}

	err = r.controller.Watch(&source.Kind{Type: &v1alpha1.Cassandra{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Manager{},
	})
	if err != nil {
		return err
	}

	if cassandraControllerActivate {
		err = cassandra.Add(r.manager)
		if err != nil {
			return err
		}
	}

	/*
		zookeeperActivationStatus := false
		if instance.Status.Zookeeper != nil {
			if instance.Status.Zookeeper.Active != nil {
				zookeeperActivationStatus = *instance.Status.Zookeeper.Active
			}
		}

		zookeeperActivationIntent := false
		if instance.Spec.Services.Zookeeper != nil {
			if instance.Spec.Services.Zookeeper.Activate != nil {
				zookeeperActivationIntent = *instance.Spec.Services.Zookeeper.Activate
			}
		}
		zookeeperResource := v1alpha1.Zookeeper{}
		zookeeperCrd := zookeeperResource.GetCrd()
		zookeeperControllerActivate := false
		if zookeeperActivationIntent && !zookeeperActivationStatus {
			err = r.createCrd(instance, zookeeperCrd)
			if err != nil {
				return err
			}

			controllerRunning := false
			sharedIndexInformer, err := r.cache.GetInformerForKind(zookeeperResource.GroupVersionKind())
			if err == nil {
				controller := sharedIndexInformer.GetController()
				if controller != nil {
					controllerRunning = true
				}
			}
			if !controllerRunning {
				zookeeperControllerActivate = true
			}

			err = r.controller.Watch(&source.Kind{Type: &v1alpha1.Zookeeper{}}, &handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &v1alpha1.Manager{},
			})
			if err != nil {
				return err
			}
			active := true
			if instance.Status.Zookeeper == nil {
				status := &v1alpha1.ServiceStatus{
					Active: &active,
				}
				instance.Status.Zookeeper = status
			} else {
				instance.Status.Zookeeper.Active = &active
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}
		}
		rabbitmqActivationStatus := false
		if instance.Status.Rabbitmq != nil {
			if instance.Status.Rabbitmq.Active != nil {
				rabbitmqActivationStatus = *instance.Status.Rabbitmq.Active
			}
		}

		rabbitmqActivationIntent := false
		if instance.Spec.Services.Rabbitmq != nil {
			if instance.Spec.Services.Rabbitmq.Activate != nil {
				rabbitmqActivationIntent = *instance.Spec.Services.Rabbitmq.Activate
			}
		}
		rabbitmqResource := v1alpha1.Rabbitmq{}
		rabbitmqCrd := rabbitmqResource.GetCrd()
		rabbitmqControllerActivate := false
		if rabbitmqActivationIntent && !rabbitmqActivationStatus {
			err = r.createCrd(instance, rabbitmqCrd)
			if err != nil {
				return err
			}

			controllerRunning := false
			sharedIndexInformer, err := r.cache.GetInformerForKind(rabbitmqResource.GroupVersionKind())
			if err == nil {
				controller := sharedIndexInformer.GetController()
				if controller != nil {
					controllerRunning = true
				}
			}
			if !controllerRunning {
				rabbitmqControllerActivate = true
			}

			err = r.controller.Watch(&source.Kind{Type: &v1alpha1.Rabbitmq{}}, &handler.EnqueueRequestForOwner{
				IsController: true,
				OwnerType:    &v1alpha1.Manager{},
			})
			if err != nil {
				return err
			}
			active := true
			if instance.Status.Rabbitmq == nil {
				status := &v1alpha1.ServiceStatus{
					Active: &active,
				}
				instance.Status.Rabbitmq = status
			} else {
				instance.Status.Rabbitmq.Active = &active
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}
		}

		if configControllerActivate {
			err = config.Add(r.manager)
			if err != nil {
				return err
			}
		}
		if controlControllerActivate {
			//err = control.Add(r.manager)
			if err != nil {
				return err
			}
		}
		if kubemanagerControllerActivate {
			//err = kubemanager.Add(r.manager)
			if err != nil {
				return err
			}
		}
		if webuiControllerActivate {
			//err = webui.Add(r.manager)
			if err != nil {
				return err
			}
		}
		if vrouterControllerActivate {
			//err = vrouter.Add(r.manager)
			if err != nil {
				return err
			}
		}
	*/

	/*
		if zookeeperControllerActivate {
			err = zookeeper.Add(r.manager)
			if err != nil {
				return err
			}
		}
		if rabbitmqControllerActivate {
			err = rabbitmq.Add(r.manager)
			if err != nil {
				return err
			}
		}
	*/

	return nil
}
