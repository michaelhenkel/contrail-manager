package rabbitmq

import (
	"context"
	"strings"

	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_rabbitmq")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Rabbitmq Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRabbitmq{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rabbitmq-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Rabbitmq
	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Rabbitmq{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Rabbitmq
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Rabbitmq{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileRabbitmq implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRabbitmq{}

// ReconcileRabbitmq reconciles a Rabbitmq object
type ReconcileRabbitmq struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Rabbitmq object and makes changes based on the state read
// and what is in the Rabbitmq.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileRabbitmq) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Rabbitmq")

	// Fetch the Rabbitmq instance
	instance := &contrailv1alpha1.Rabbitmq{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	// Fetch Manager instance
	managerInstance := &contrailv1alpha1.Manager{}
	err = r.client.Get(context.TODO(), request.NamespacedName, managerInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Manager Instance")
		}
	} else {
		instance.Spec.Service = managerInstance.Spec.Rabbitmq
		if managerInstance.Spec.Rabbitmq.Size != nil {
			instance.Spec.Service.Size = managerInstance.Spec.Rabbitmq.Size
		} else {
			instance.Spec.Service.Size = managerInstance.Spec.Size
		}
		if managerInstance.Spec.HostNetwork != nil {
			instance.Spec.HostNetwork = managerInstance.Spec.HostNetwork
		}
	}

	// Get default Deployment
	deployment := GetDeployment()

	// Create initial ConfigMap
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq-" + instance.Name,
			Namespace: instance.Namespace,
		},
		Data: instance.Spec.Service.Configuration,
	}
	controllerutil.SetControllerReference(instance, &configMap, r.scheme)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "rabbitmq-" + instance.Name, Namespace: instance.Namespace}, &configMap)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), &configMap)
		if err != nil {
			reqLogger.Error(err, "Failed to create ConfigMap", "Namespace", instance.Namespace, "Name", "rabbitmq-"+instance.Name)
			return reconcile.Result{}, err
		}
	}

	// Set Deployment Name & Namespace

	deployment.ObjectMeta.Name = "rabbitmq-" + instance.Name
	deployment.ObjectMeta.Namespace = instance.Namespace

	//command := []string{"/bin/sh", "-c", "while true; do echo hello; sleep 10;done"}

	//readinessCommand := []string{"/bin/bash", "-c", "OK=$(echo ruok | nc 127.0.0.1 2181); if [[ ${OK} == \"imok\" ]]; then exit 0; else exit 1;fi"}

	// Configure Containers
	for idx, container := range deployment.Spec.Template.Spec.Containers {
		for containerName, image := range instance.Spec.Service.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.Containers[idx]).Image = image
			}
			if containerName == "rabbitmq" {
				(&deployment.Spec.Template.Spec.Containers[idx]).EnvFrom[0].ConfigMapRef.Name = "rabbitmq-" + instance.Name
			}
			/*
				if containerName == "rabbitmq" {
					(&deployment.Spec.Template.Spec.Containers[idx]).Command = command
				}
			*/

		}
	}

	// Configure InitContainers
	for idx, container := range deployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range instance.Spec.Service.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.InitContainers[idx]).Image = image
			}
		}
	}

	// Set HostNetwork
	deployment.Spec.Template.Spec.HostNetwork = *instance.Spec.HostNetwork

	// Set Selector and Label
	deployment.Spec.Selector.MatchLabels["app"] = "rabbitmq-" + instance.Name
	deployment.Spec.Template.ObjectMeta.Labels["app"] = "rabbitmq-" + instance.Name

	// Set Size
	deployment.Spec.Replicas = instance.Spec.Service.Size

	// Create Deployment
	controllerutil.SetControllerReference(instance, deployment, r.scheme)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "rabbitmq-" + instance.Name, Namespace: instance.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create Deployment", "Namespace", instance.Namespace, "Name", "rabbitmq-"+instance.Name)
			return reconcile.Result{}, err
		}
	}

	// Check if Init Containers are running
	_, err = contrailv1alpha1.InitContainerRunning(r.client,
		instance.ObjectMeta,
		"rabbitmq",
		instance,
		*instance.Spec.Service,
		&instance.Status)

	if err != nil {
		reqLogger.Error(err, "Err Init Pods not ready, requeing")
		return reconcile.Result{}, err
	}

	// Update ConfigMap

	var podIpList []string
	for _, ip := range instance.Status.Nodes {
		podIpList = append(podIpList, ip)
	}
	nodeList := strings.Join(podIpList, ",")
	configMap.Data["RABBITMQ_NODES"] = nodeList
	configMap.Data["CONTROLLER_NODES"] = nodeList
	err = r.client.Update(context.TODO(), &configMap)
	if err != nil {
		reqLogger.Error(err, "Failed to update ConfigMap", "Namespace", instance.Namespace, "Name", "cassandra-"+instance.Name)
		return reconcile.Result{}, err
	}

	err = contrailv1alpha1.MarkInitPodsReady(r.client, instance.ObjectMeta, "rabbitmq")

	if err != nil {
		reqLogger.Error(err, "Failed to mark Pods ready")
		return reconcile.Result{}, err
	}

	err = contrailv1alpha1.SetServiceStatus(r.client,
		instance.ObjectMeta,
		"rabbitmq",
		instance,
		&deployment.Status,
		&instance.Status)

	if err != nil {
		reqLogger.Error(err, "Failed to set Service status")
		return reconcile.Result{}, err
	} else {
		reqLogger.Info("set service status")
	}
	return reconcile.Result{}, nil
}
