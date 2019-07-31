package manager

import (
	"context"
	"fmt"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/cassandra"
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


		gvk := schema.GroupVersionKind{
			Group:   "contrail.juniper.net",
			Version: "v1alpha1",
			Kind:    "Cassandra",
		}
		cassandraCrd.SetGroupVersionKind(gvk)
		cassandraCrd.TypeMeta.APIVersion = "contrail.juniper.net/v1alpha1"
		cassandraCrd.TypeMeta.Kind = "Cassandra"
	*/
	cassandraResource := v1alpha1.Cassandra{}
	cassandraCrd := cassandraResource.GetCrd()
	err = r.createCrd(instance, cassandraCrd)
	if err != nil {
		return reconcile.Result{}, err
	}
	controllerRunning := false
	cassandraSharedIndexInformer, err := r.cache.GetInformerForKind(cassandraResource.GroupVersionKind())
	fmt.Println("err", err)
	if err == nil {
		controller := cassandraSharedIndexInformer.GetController()
		if controller == nil {
			controllerRunning = true
		}
	}
	if !controllerRunning {
		err = cassandra.Add(r.manager)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	/*
		cassandraControllerRunning := false
		cassandraSharedIndexInformer, err := r.cache.GetInformerForKind(cassandraResource.GroupVersionKind())
		if err == nil {
			controller := cassandraSharedIndexInformer.GetController()
			if controller != nil {
				cassandraControllerRunning = true
			}
		}
	*/
	/*
		cassandraPred := utils.AppSizeChange(utils.CassandraGroupKind())
		err = r.controller.Watch(&source.Kind{Type: &v1alpha1.Cassandra{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &v1alpha1.Manager{},
		}, cassandraPred)
		if err != nil {
			return reconcile.Result{}, err
		}
	*/

	/*
		if !cassandraControllerRunning {
			fmt.Println("CONTROLLER NOT RUNNING, STARTING NOW")
			err = cassandra.Add(r.manager)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	*/

	// Create CRs
	for _, cassandraService := range instance.Spec.Services.Cassandras {
		fmt.Println("CREATE CR for ", cassandraService.Name)
		create := true
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
		fmt.Println("create ", create)
		fmt.Println("update ", update)
		fmt.Println("delete ", delete)
		cr := cr.GetCassandraCr()
		cr.ObjectMeta = cassandraService.ObjectMeta
		cr.Labels = cassandraService.ObjectMeta.Labels
		cr.Namespace = instance.Namespace
		cr.Spec.ServiceConfiguration = cassandraService.Spec.ServiceConfiguration
		cr.TypeMeta.APIVersion = "contrail.juniper.net/v1alpha1"
		cr.TypeMeta.Kind = "Cassandra"
		if create {
			fmt.Println("CREATE")
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
			fmt.Println("UPDATE")
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
					//if reflect.DeepEqual(cr.Spec.ServiceConfiguration, cassandraService.Spec.ServiceConfiguration)
					err = r.client.Update(context.TODO(), cr)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		}
		if delete {
			fmt.Println("DELETE")
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: cr.Name, Namespace: cr.Namespace}, cr)
			if err != nil {
				return reconcile.Result{}, err
			}
			err = r.client.Delete(context.TODO(), cr)
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
