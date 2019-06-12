package cassandra

import (
	"context"
	"fmt"
	"strings"
	"k8s.io/apimachinery/pkg/types"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var log = logf.Log.WithName("controller_cassandra")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Cassandra Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCassandra{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("cassandra-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Cassandra
	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Cassandra{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCassandra implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCassandra{}

// ReconcileCassandra reconciles a Cassandra object
type ReconcileCassandra struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Cassandra object and makes changes based on the state read
// and what is in the Cassandra.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCassandra) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Cassandra")

	// Fetch the Cassandra instance
	instance := &contrailv1alpha1.Cassandra{}
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
		instance.Spec.Service = managerInstance.Spec.Cassandra
		if managerInstance.Spec.Cassandra.Size != nil {
			instance.Spec.Service.Size = managerInstance.Spec.Cassandra.Size
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
			Name:      "cassandra-" + instance.Name,
			Namespace: instance.Namespace,
		},
		Data: instance.Spec.Service.Configuration,
	}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "cassandra-" + instance.Name, Namespace: instance.Namespace}, &configMap)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), &configMap)	
		if err != nil {
			reqLogger.Error(err, "Failed to create ConfigMap", "Namespace", instance.Namespace, "Name", "cassandra-" + instance.Name)
			return reconcile.Result{}, err
		}
	}

	// Set Deployment Name & Namespace

	deployment.ObjectMeta.Name = "cassandra-" + instance.Name 
	deployment.ObjectMeta.Namespace = instance.Namespace 

	// Configure Containers
	for idx, container := range(deployment.Spec.Template.Spec.Containers){
		for containerName, image := range(instance.Spec.Service.Images){
			if containerName == container.Name{
				(&deployment.Spec.Template.Spec.Containers[idx]).Image = image
			}
			if containerName == "cassandra"{
				(&deployment.Spec.Template.Spec.Containers[idx]).EnvFrom[0].ConfigMapRef.Name = "cassandra-" + instance.Name
			}
		}
	}

	// Configure InitContainers
	for idx, container := range(deployment.Spec.Template.Spec.InitContainers){
		for containerName, image := range(instance.Spec.Service.Images){
			if containerName == container.Name{
				(&deployment.Spec.Template.Spec.InitContainers[idx]).Image = image
			}
		}
	}

	// Set HostNetwork
	deployment.Spec.Template.Spec.HostNetwork = *instance.Spec.HostNetwork

	// Set Selector and Label
	deployment.Spec.Selector.MatchLabels["app"] = "cassandra-" + instance.Name
	deployment.Spec.Template.ObjectMeta.Labels["app"] = "cassandra-" + instance.Name

	// Set Size
	deployment.Spec.Replicas = instance.Spec.Service.Size



	fmt.Printf("%+v\n", deployment.Spec.Template.Spec.HostNetwork)

	// Create Deployment
	controllerutil.SetControllerReference(instance, deployment, r.scheme)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "cassandra-" + instance.Name, Namespace: instance.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), deployment)	
		if err != nil {
			reqLogger.Error(err, "Failed to create Deployment", "Namespace", instance.Namespace, "Name", "cassandra-" + instance.Name)
			return reconcile.Result{}, err
		}
	}
	
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"app": "cassandra-" + instance.Name})
	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	err = r.client.List(context.TODO(), listOps, podList)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods", "Namespace", instance.Namespace, "Memcached.Name", instance.Name)
		return reconcile.Result{}, err
	}

	initContainerRunning := true
	var podIpMap = make(map[string]string)
	for _, pod := range(podList.Items){
		if pod.Status.PodIP != ""{
			podIpMap[pod.Name] = pod.Status.PodIP
		}
		for _, initContainer := range(pod.Status.InitContainerStatuses){
			if initContainer.Name == "init"{
				if !initContainer.Ready{
					initContainerRunning = false
				}
			}
		}
	}
	var podIpList []string
	if len(podIpMap) == int(*instance.Spec.Service.Size){
		for _, podIp := range(podIpMap){
			podIpList = append(podIpList, podIp)
		}
		for _, pod := range(podList.Items){
			pod.ObjectMeta.Labels["status"] = "ready"
			err = r.client.Update(context.TODO(), &pod)
			if err != nil {
				reqLogger.Error(err, "Failed to update pod", "Namespace", pod.Name, "Memcached.Name", instance.Name)
				return reconcile.Result{}, err
			}	
		}
		nodeList := strings.Join(podIpList,",")
		configMap.Data["CASSANDRA_SEEDS"] = nodeList
		configMap.Data["CONTROLLER_NODES"] = nodeList
		err = r.client.Update(context.TODO(), &configMap)	
		if err != nil {
			reqLogger.Error(err, "Failed to update ConfigMap", "Namespace", instance.Namespace, "Name", "cassandra-" + instance.Name)
			return reconcile.Result{}, err
		}

	}

	if initContainerRunning{
		instance.Status.Nodes = podIpMap
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update status.")
			return reconcile.Result{}, err
		}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: "cassandra-" + instance.Name, Namespace: instance.Namespace}, deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to get Deployment")
			return reconcile.Result{}, err
		}
		if deployment.Status.Replicas == deployment.Status.ReadyReplicas { 
			active := true
			instance.Status.Active = &active
			err = r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				reqLogger.Error(err, "Failed to update status.")
				return reconcile.Result{}, err
			}
		} else {
			reqLogger.Info("Waiting for Deployment to be complete, requeing")
			return reconcile.Result{Requeue: true}, nil
		}
		if err != nil {
			reqLogger.Error(err, "Failed to update status.")
			return reconcile.Result{}, err
		}
	} else {
		reqLogger.Info("Init Pods not ready, requeing")
		return reconcile.Result{Requeue: true}, nil
	}



	return reconcile.Result{}, nil
}

func labelsForService(name string) map[string]string {
	return map[string]string{"app": name}
}

func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}