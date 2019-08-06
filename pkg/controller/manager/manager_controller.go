package manager

import (
	"context"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	cr "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crs"

	//crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	//"github.com/michaelhenkel/contrail-manager/pkg/controller/config"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"

	//corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_manager")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Manager Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	apiextensionsv1beta1.AddToScheme(scheme.Scheme)
	var r reconcile.Reconciler
	reconcileManager := ReconcileManager{client: mgr.GetClient(), scheme: mgr.GetScheme(), manager: mgr, cache: mgr.GetCache()}
	r = &reconcileManager
	//r := newReconciler(mgr)
	c, err := createController(mgr, r)
	if err != nil {
		return err
	}
	reconcileManager.controller = c
	return addManagerWatch(c)
}

func createController(mgr manager.Manager, r reconcile.Reconciler) (controller.Controller, error) {
	c, err := controller.New("manager-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return c, err
	}
	return c, nil
}

func addManagerWatch(c controller.Controller) error {
	err := c.Watch(&source.Kind{Type: &v1alpha1.Manager{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}
	return nil
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) (controller.Controller, error) {
	// Create a new controller
	c, err := controller.New("manager-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return c, err
	}

	err = c.Watch(&source.Kind{Type: &v1alpha1.Manager{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return c, err
	}
	return c, nil
}

func (r *ReconcileManager) addWatch(ro runtime.Object) error {

	controller := r.controller
	//err := controller.Watch(&source.Kind{Type: &v1alpha1.Config{}}, &handler.EnqueueRequestForOwner{
	err := controller.Watch(&source.Kind{Type: ro}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Manager{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileManager implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileManager{}

// ReconcileManager reconciles a Manager object
type ReconcileManager struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client     client.Client
	scheme     *runtime.Scheme
	manager    manager.Manager
	controller controller.Controller
	cache      cache.Cache
}

func (r *ReconcileManager) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.NamespaceX", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Manager")
	instance := &v1alpha1.Manager{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	// Create CRDs
	/*
		cassandraCrdActive := false
		for _, crdStatus := range instance.Status.CrdStatus {
			if crdStatus.Name == "Cassandra" {
				if crdStatus.Active != nil {
					if *crdStatus.Active {
						cassandraCrdActive = true
					}
				}
			}
		}
		if !cassandraCrdActive && len(instance.Spec.Services.Cassandras) > 0 {
			cassandraResource := v1alpha1.Cassandra{}
			cassandraCrd := cassandraResource.GetCrd()
			err = r.createCrd(instance, cassandraCrd)
			if err != nil {
				return reconcile.Result{}, err
			}
			controllerRunning := false
			c := r.manager.GetCache()
			sharedIndexInformer, err := c.GetInformer(&v1alpha1.Cassandra{})
			if err == nil {
				store := sharedIndexInformer.GetStore()
				if store != nil {
					fmt.Println("STORE NOT NIL")
				} else {
					fmt.Println("STORE NIL")
				}
				if sharedIndexInformer.HasSynced() {
					fmt.Println("has synced")
				} else {
					fmt.Println("has not synced")
				}

				controller := sharedIndexInformer.GetController()
				if controller != nil {
					controllerRunning = true
				}
			}
			if !controllerRunning {
				err = cassandra.Add(r.manager)
				if err != nil {
					return reconcile.Result{}, err
				}
			}
			active := true
			crdStatus := &v1alpha1.CrdStatus{
				Name:   "Cassandra",
				Active: &active,
			}
			var crdStatusList []v1alpha1.CrdStatus
			if len(instance.Status.CrdStatus) > 0 {
				crdStatusList = instance.Status.CrdStatus
			}
			crdStatusList = append(crdStatusList, *crdStatus)
			instance.Status.CrdStatus = crdStatusList
			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
	*/
	// Create CRs
	for _, cassandraService := range instance.Spec.Services.Cassandras {
		create := *cassandraService.Spec.CommonConfiguration.Create
		delete := false
		update := false
		for _, cassandraStatus := range instance.Status.Cassandras {
			if cassandraService.Name == *cassandraStatus.Name {
				if *cassandraService.Spec.CommonConfiguration.Create && *cassandraStatus.Created {
					create = false
					delete = false
					update = true
				}
				if !*cassandraService.Spec.CommonConfiguration.Create && *cassandraStatus.Created {
					create = false
					delete = true
					update = false
				}
			}
		}
		cr := cr.GetCassandraCr()
		cr.ObjectMeta = cassandraService.ObjectMeta
		cr.Labels = cassandraService.ObjectMeta.Labels
		cr.Namespace = instance.Namespace
		cr.Spec.ServiceConfiguration = cassandraService.Spec.ServiceConfiguration
		cr.TypeMeta.APIVersion = "contrail.juniper.net/v1alpha1"
		cr.TypeMeta.Kind = "Cassandra"
		if create {
			err = r.client.Get(context.TODO(), request.NamespacedName, instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				if errors.IsNotFound(err) {
					controllerutil.SetControllerReference(instance, cr, r.scheme)
					err = r.client.Create(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}

			status := &v1alpha1.ServiceStatus{}
			cassandraStatusList := []*v1alpha1.ServiceStatus{}
			if instance.Status.Cassandras != nil {
				for _, cassandraStatus := range instance.Status.Cassandras {
					if cassandraService.Name == *cassandraStatus.Name {
						status = cassandraStatus
						status.Created = &create
					}
				}
			} else {
				status.Name = &cr.Name
				status.Created = &create
				cassandraStatusList = append(cassandraStatusList, status)
				instance.Status.Cassandras = cassandraStatusList
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
		if update {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			replicas := instance.Spec.CommonConfiguration.Replicas
			if cassandraService.Spec.CommonConfiguration.Replicas != nil {
				replicas = cassandraService.Spec.CommonConfiguration.Replicas
			}
			if cr.Spec.CommonConfiguration.Replicas != nil {
				if *replicas != *cr.Spec.CommonConfiguration.Replicas {
					cr.Spec.CommonConfiguration.Replicas = replicas
					err = r.client.Update(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		}
		if delete {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Delete(context.TODO(), cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			status := &v1alpha1.ServiceStatus{}
			cassandraStatusList := []*v1alpha1.ServiceStatus{}
			if instance.Status.Cassandras != nil {
				for _, cassandraStatus := range instance.Status.Cassandras {
					if cassandraService.Name == *cassandraStatus.Name {
						status = cassandraStatus
						status.Created = &create
					}
				}
			} else {
				status.Name = &cr.Name
				status.Created = &create
				cassandraStatusList = append(cassandraStatusList, status)
				instance.Status.Cassandras = cassandraStatusList
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
	}

	for _, zookeeperService := range instance.Spec.Services.Zookeepers {
		create := *zookeeperService.Spec.CommonConfiguration.Create
		delete := false
		update := false
		for _, zookeeperStatus := range instance.Status.Zookeepers {
			if zookeeperService.Name == *zookeeperStatus.Name {
				if *zookeeperService.Spec.CommonConfiguration.Create && *zookeeperStatus.Created {
					create = false
					delete = false
					update = true
				}
				if !*zookeeperService.Spec.CommonConfiguration.Create && *zookeeperStatus.Created {
					create = false
					delete = true
					update = false
				}
			}
		}
		cr := cr.GetZookeeperCr()
		cr.ObjectMeta = zookeeperService.ObjectMeta
		cr.Labels = zookeeperService.ObjectMeta.Labels
		cr.Namespace = instance.Namespace
		cr.Spec.ServiceConfiguration = zookeeperService.Spec.ServiceConfiguration
		cr.TypeMeta.APIVersion = "contrail.juniper.net/v1alpha1"
		cr.TypeMeta.Kind = "Zookeeper"
		if create {
			err = r.client.Get(context.TODO(), request.NamespacedName, instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				if errors.IsNotFound(err) {
					controllerutil.SetControllerReference(instance, cr, r.scheme)
					err = r.client.Create(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}

			status := &v1alpha1.ServiceStatus{}
			zookeeperStatusList := []*v1alpha1.ServiceStatus{}
			if instance.Status.Zookeepers != nil {
				for _, zookeeperStatus := range instance.Status.Zookeepers {
					if zookeeperService.Name == *zookeeperStatus.Name {
						status = zookeeperStatus
						status.Created = &create
					}
				}
			} else {
				status.Name = &cr.Name
				status.Created = &create
				zookeeperStatusList = append(zookeeperStatusList, status)
				instance.Status.Zookeepers = zookeeperStatusList
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
		if update {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			replicas := instance.Spec.CommonConfiguration.Replicas
			if zookeeperService.Spec.CommonConfiguration.Replicas != nil {
				replicas = zookeeperService.Spec.CommonConfiguration.Replicas
			}
			if cr.Spec.CommonConfiguration.Replicas != nil {
				if *replicas != *cr.Spec.CommonConfiguration.Replicas {
					cr.Spec.CommonConfiguration.Replicas = replicas
					err = r.client.Update(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		}
		if delete {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Delete(context.TODO(), cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			status := &v1alpha1.ServiceStatus{}
			zookeeperStatusList := []*v1alpha1.ServiceStatus{}
			if instance.Status.Zookeepers != nil {
				for _, zookeeperStatus := range instance.Status.Zookeepers {
					if zookeeperService.Name == *zookeeperStatus.Name {
						status = zookeeperStatus
						status.Created = &create
					}
				}
			} else {
				status.Name = &cr.Name
				status.Created = &create
				zookeeperStatusList = append(zookeeperStatusList, status)
				instance.Status.Zookeepers = zookeeperStatusList
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
	}

	if instance.Spec.Services.Webui != nil {
		webuiService := instance.Spec.Services.Webui
		create := *webuiService.Spec.CommonConfiguration.Create
		delete := false
		update := false
		if instance.Status.Webui != nil {
			if *webuiService.Spec.CommonConfiguration.Create && *instance.Status.Webui.Created {
				create = false
				delete = false
				update = true
			}
			if !*webuiService.Spec.CommonConfiguration.Create && *instance.Status.Webui.Created {
				create = false
				delete = true
				update = false
			}
		}

		cr := cr.GetWebuiCr()
		cr.ObjectMeta = webuiService.ObjectMeta
		cr.Labels = webuiService.ObjectMeta.Labels
		cr.Namespace = instance.Namespace
		cr.Spec.ServiceConfiguration = webuiService.Spec.ServiceConfiguration
		cr.TypeMeta.APIVersion = "contrail.juniper.net/v1alpha1"
		cr.TypeMeta.Kind = "Webui"
		if create {
			err = r.client.Get(context.TODO(), request.NamespacedName, instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				if errors.IsNotFound(err) {
					controllerutil.SetControllerReference(instance, cr, r.scheme)
					err = r.client.Create(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}

			status := &v1alpha1.ServiceStatus{}
			if instance.Status.Webui != nil {
				status = instance.Status.Webui
				status.Created = &create
			} else {
				status.Name = &cr.Name
				status.Created = &create
				instance.Status.Webui = status
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
		if update {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			replicas := instance.Spec.CommonConfiguration.Replicas
			if webuiService.Spec.CommonConfiguration.Replicas != nil {
				replicas = webuiService.Spec.CommonConfiguration.Replicas
			}
			if cr.Spec.CommonConfiguration.Replicas != nil {
				if *replicas != *cr.Spec.CommonConfiguration.Replicas {
					cr.Spec.CommonConfiguration.Replicas = replicas
					err = r.client.Update(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		}
		if delete {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Delete(context.TODO(), cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			status := &v1alpha1.ServiceStatus{}
			if instance.Status.Webui != nil {
				status = instance.Status.Webui
				status.Created = &create
			} else {
				status.Name = &cr.Name
				status.Created = &create
				instance.Status.Webui = status
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}

	if instance.Spec.Services.Config != nil {
		configService := instance.Spec.Services.Config
		create := *configService.Spec.CommonConfiguration.Create
		delete := false
		update := false
		if instance.Status.Config != nil {
			if *configService.Spec.CommonConfiguration.Create && *instance.Status.Config.Created {
				create = false
				delete = false
				update = true
			}
			if !*configService.Spec.CommonConfiguration.Create && *instance.Status.Config.Created {
				create = false
				delete = true
				update = false
			}
		}

		cr := cr.GetConfigCr()
		cr.ObjectMeta = configService.ObjectMeta
		cr.Labels = configService.ObjectMeta.Labels
		cr.Namespace = instance.Namespace
		cr.Spec.ServiceConfiguration = configService.Spec.ServiceConfiguration
		cr.TypeMeta.APIVersion = "contrail.juniper.net/v1alpha1"
		cr.TypeMeta.Kind = "Config"
		if create {
			err = r.client.Get(context.TODO(), request.NamespacedName, instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				if errors.IsNotFound(err) {
					controllerutil.SetControllerReference(instance, cr, r.scheme)
					err = r.client.Create(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}

			status := &v1alpha1.ServiceStatus{}
			if instance.Status.Config != nil {
				status = instance.Status.Config
				status.Created = &create
			} else {
				status.Name = &cr.Name
				status.Created = &create
				instance.Status.Config = status
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
		if update {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			replicas := instance.Spec.CommonConfiguration.Replicas
			if configService.Spec.CommonConfiguration.Replicas != nil {
				replicas = configService.Spec.CommonConfiguration.Replicas
			}
			if cr.Spec.CommonConfiguration.Replicas != nil {
				if *replicas != *cr.Spec.CommonConfiguration.Replicas {
					cr.Spec.CommonConfiguration.Replicas = replicas
					err = r.client.Update(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		}
		if delete {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Delete(context.TODO(), cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			status := &v1alpha1.ServiceStatus{}
			if instance.Status.Config != nil {
				status = instance.Status.Config
				status.Created = &create
			} else {
				status.Name = &cr.Name
				status.Created = &create
				instance.Status.Config = status
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
	}

	for _, kubemanagerService := range instance.Spec.Services.Kubemanagers {
		create := *kubemanagerService.Spec.CommonConfiguration.Create
		delete := false
		update := false
		for _, kubemanagerStatus := range instance.Status.Kubemanagers {
			if kubemanagerService.Name == *kubemanagerStatus.Name {
				if *kubemanagerService.Spec.CommonConfiguration.Create && *kubemanagerStatus.Created {
					create = false
					delete = false
					update = true
				}
				if !*kubemanagerService.Spec.CommonConfiguration.Create && *kubemanagerStatus.Created {
					create = false
					delete = true
					update = false
				}
			}
		}
		cr := cr.GetKubemanagerCr()
		cr.ObjectMeta = kubemanagerService.ObjectMeta
		cr.Labels = kubemanagerService.ObjectMeta.Labels
		cr.Namespace = instance.Namespace
		cr.Spec.ServiceConfiguration = kubemanagerService.Spec.ServiceConfiguration
		cr.TypeMeta.APIVersion = "contrail.juniper.net/v1alpha1"
		cr.TypeMeta.Kind = "Kubemanager"
		if create {
			err = r.client.Get(context.TODO(), request.NamespacedName, instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				if errors.IsNotFound(err) {
					controllerutil.SetControllerReference(instance, cr, r.scheme)
					err = r.client.Create(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}

			status := &v1alpha1.ServiceStatus{}
			kubemanagerStatusList := []*v1alpha1.ServiceStatus{}
			if instance.Status.Kubemanagers != nil {
				for _, kubemanagerStatus := range instance.Status.Kubemanagers {
					if kubemanagerService.Name == *kubemanagerStatus.Name {
						status = kubemanagerStatus
						status.Created = &create
					}
				}
			} else {
				status.Name = &cr.Name
				status.Created = &create
				kubemanagerStatusList = append(kubemanagerStatusList, status)
				instance.Status.Kubemanagers = kubemanagerStatusList
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
		if update {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			replicas := instance.Spec.CommonConfiguration.Replicas
			if kubemanagerService.Spec.CommonConfiguration.Replicas != nil {
				replicas = kubemanagerService.Spec.CommonConfiguration.Replicas
			}
			if cr.Spec.CommonConfiguration.Replicas != nil {
				if *replicas != *cr.Spec.CommonConfiguration.Replicas {
					cr.Spec.CommonConfiguration.Replicas = replicas
					err = r.client.Update(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		}
		if delete {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Delete(context.TODO(), cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			status := &v1alpha1.ServiceStatus{}
			kubemanagerStatusList := []*v1alpha1.ServiceStatus{}
			if instance.Status.Kubemanagers != nil {
				for _, kubemanagerStatus := range instance.Status.Kubemanagers {
					if kubemanagerService.Name == *kubemanagerStatus.Name {
						status = kubemanagerStatus
						status.Created = &create
					}
				}
			} else {
				status.Name = &cr.Name
				status.Created = &create
				kubemanagerStatusList = append(kubemanagerStatusList, status)
				instance.Status.Kubemanagers = kubemanagerStatusList
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
	}

	for _, controlService := range instance.Spec.Services.Controls {
		create := *controlService.Spec.CommonConfiguration.Create
		delete := false
		update := false
		for _, controlStatus := range instance.Status.Controls {
			if controlService.Name == *controlStatus.Name {
				if *controlService.Spec.CommonConfiguration.Create && *controlStatus.Created {
					create = false
					delete = false
					update = true
				}
				if !*controlService.Spec.CommonConfiguration.Create && *controlStatus.Created {
					create = false
					delete = true
					update = false
				}
			}
		}
		cr := cr.GetControlCr()
		cr.ObjectMeta = controlService.ObjectMeta
		cr.Labels = controlService.ObjectMeta.Labels
		cr.Namespace = instance.Namespace
		cr.Spec.ServiceConfiguration = controlService.Spec.ServiceConfiguration
		cr.TypeMeta.APIVersion = "contrail.juniper.net/v1alpha1"
		cr.TypeMeta.Kind = "Control"
		if create {
			err = r.client.Get(context.TODO(), request.NamespacedName, instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				if errors.IsNotFound(err) {
					controllerutil.SetControllerReference(instance, cr, r.scheme)
					err = r.client.Create(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}

			status := &v1alpha1.ServiceStatus{}
			controlStatusList := []*v1alpha1.ServiceStatus{}
			if instance.Status.Controls != nil {
				for _, controlStatus := range instance.Status.Controls {
					if controlService.Name == *controlStatus.Name {
						status = controlStatus
						status.Created = &create
					}
				}
			} else {
				status.Name = &cr.Name
				status.Created = &create
				controlStatusList = append(controlStatusList, status)
				instance.Status.Controls = controlStatusList
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
		if update {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			replicas := instance.Spec.CommonConfiguration.Replicas
			if controlService.Spec.CommonConfiguration.Replicas != nil {
				replicas = controlService.Spec.CommonConfiguration.Replicas
			}
			if cr.Spec.CommonConfiguration.Replicas != nil {
				if *replicas != *cr.Spec.CommonConfiguration.Replicas {
					cr.Spec.CommonConfiguration.Replicas = replicas
					err = r.client.Update(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		}
		if delete {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Delete(context.TODO(), cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			status := &v1alpha1.ServiceStatus{}
			controlStatusList := []*v1alpha1.ServiceStatus{}
			if instance.Status.Controls != nil {
				for _, controlStatus := range instance.Status.Controls {
					if controlService.Name == *controlStatus.Name {
						status = controlStatus
						status.Created = &create
					}
				}
			} else {
				status.Name = &cr.Name
				status.Created = &create
				controlStatusList = append(controlStatusList, status)
				instance.Status.Controls = controlStatusList
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
	}

	if instance.Spec.Services.Rabbitmq != nil {
		rabbitmqService := instance.Spec.Services.Rabbitmq
		create := *rabbitmqService.Spec.CommonConfiguration.Create
		delete := false
		update := false
		if instance.Status.Rabbitmq != nil {
			if *rabbitmqService.Spec.CommonConfiguration.Create && *instance.Status.Rabbitmq.Created {
				create = false
				delete = false
				update = true
			}
			if !*rabbitmqService.Spec.CommonConfiguration.Create && *instance.Status.Rabbitmq.Created {
				create = false
				delete = true
				update = false
			}
		}

		cr := cr.GetRabbitmqCr()
		cr.ObjectMeta = rabbitmqService.ObjectMeta
		cr.Labels = rabbitmqService.ObjectMeta.Labels
		cr.Namespace = instance.Namespace
		cr.Spec.ServiceConfiguration = rabbitmqService.Spec.ServiceConfiguration
		cr.TypeMeta.APIVersion = "contrail.juniper.net/v1alpha1"
		cr.TypeMeta.Kind = "Rabbitmq"
		if create {
			err = r.client.Get(context.TODO(), request.NamespacedName, instance)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				if errors.IsNotFound(err) {
					controllerutil.SetControllerReference(instance, cr, r.scheme)
					err = r.client.Create(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}

			status := &v1alpha1.ServiceStatus{}
			if instance.Status.Rabbitmq != nil {
				status = instance.Status.Rabbitmq
				status.Created = &create
			} else {
				status.Name = &cr.Name
				status.Created = &create
				instance.Status.Rabbitmq = status
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
		if update {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			replicas := instance.Spec.CommonConfiguration.Replicas
			if rabbitmqService.Spec.CommonConfiguration.Replicas != nil {
				replicas = rabbitmqService.Spec.CommonConfiguration.Replicas
			}
			if cr.Spec.CommonConfiguration.Replicas != nil {
				if *replicas != *cr.Spec.CommonConfiguration.Replicas {
					cr.Spec.CommonConfiguration.Replicas = replicas
					err = r.client.Update(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		}
		if delete {
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Delete(context.TODO(), cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			status := &v1alpha1.ServiceStatus{}
			if instance.Status.Rabbitmq != nil {
				status = instance.Status.Rabbitmq
				status.Created = &create
			} else {
				status.Name = &cr.Name
				status.Created = &create
				instance.Status.Rabbitmq = status
			}

			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}

		}
	}

	return reconcile.Result{}, nil
}

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
