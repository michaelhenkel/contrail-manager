package manager

import (
	"context"
	//"fmt"

	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	//crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	//"github.com/michaelhenkel/contrail-manager/pkg/controller/config"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"

	//corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	//"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
	reconcileManager := ReconcileManager{client: mgr.GetClient(), scheme: mgr.GetScheme(), manager: mgr}
	r = &reconcileManager
	//r := newReconciler(mgr) 
	c, err := createController(mgr, r)
	if err != nil{
		return err
	}
	reconcileManager.controller = c
	return addManagerWatch(c)
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	apiextensionsv1beta1.AddToScheme(scheme.Scheme)
	return &ReconcileManager{client: mgr.GetClient(), scheme: mgr.GetScheme(), manager: mgr}
}

func createController(mgr manager.Manager, r reconcile.Reconciler) (controller.Controller, error) {
	c, err := controller.New("manager-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return c, err
	}
	return c, nil
}

func addManagerWatch(c controller.Controller) (error) {
	err := c.Watch(&source.Kind{Type: &contrailv1alpha1.Manager{}}, &handler.EnqueueRequestForObject{})
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

	//var reconcileManager = &r
	//reconcileManager = r.(ReconcileManager)

	// Watch for changes to primary resource Manager
	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Manager{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return c, err
	}
	return c, nil
}

func (r *ReconcileManager) addWatch (ro runtime.Object) error {

	controller := r.controller
	//err := controller.Watch(&source.Kind{Type: &contrailv1alpha1.Config{}}, &handler.EnqueueRequestForOwner{
	err := controller.Watch(&source.Kind{Type: ro}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Manager{},
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
	client client.Client
	scheme *runtime.Scheme
	manager manager.Manager
	controller controller.Controller
}

func (r *ReconcileManager) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Manager")
	instance := &contrailv1alpha1.Manager{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	err = r.ActivateService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = r.CreateService(instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}