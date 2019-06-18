package config

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

var log = logf.Log.WithName("controller_config")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Config Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileConfig{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("config-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Config
	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Config{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Config

	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Cassandra{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Manager{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Zookeeper{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Manager{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Rabbitmq{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Manager{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileConfig implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileConfig{}

// ReconcileConfig reconciles a Config object
type ReconcileConfig struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Config object and makes changes based on the state read
// and what is in the Config.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Config")

	// Fetch the Config instance
	instance := &contrailv1alpha1.Config{}
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

	// Get cassandra, zk and rmq status

	cassandraInstance := &contrailv1alpha1.Cassandra{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, cassandraInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Cassandra Instance")
			return reconcile.Result{}, err
		}
	}
	cassandraStatus := false
	if cassandraInstance.Status.Active != nil {
		cassandraStatus = *cassandraInstance.Status.Active
	}

	zookeeperInstance := &contrailv1alpha1.Zookeeper{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, zookeeperInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Zookeeper Instance")
			return reconcile.Result{}, err
		}
	}
	zookeeperStatus := false
	if zookeeperInstance.Status.Active != nil {
		zookeeperStatus = *zookeeperInstance.Status.Active
	}

	rabbitmqInstance := &contrailv1alpha1.Rabbitmq{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, rabbitmqInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Rabbitmq Instance")
			return reconcile.Result{}, err
		}
	}
	rabbitmqStatus := false
	if rabbitmqInstance.Status.Active != nil {
		rabbitmqStatus = *rabbitmqInstance.Status.Active
	}

	if !rabbitmqStatus || !zookeeperStatus || !cassandraStatus {
		reqLogger.Info("cassandra, zookeeper or rmq not ready")
		return reconcile.Result{}, nil
	}

	var rabbitmqNodes []string
	for _, ip := range rabbitmqInstance.Status.Nodes {
		rabbitmqNodes = append(rabbitmqNodes, ip)
	}
	rabbitmqNodeList := strings.Join(rabbitmqNodes, ",")

	var zookeeperNodes []string
	for _, ip := range zookeeperInstance.Status.Nodes {
		zookeeperNodes = append(zookeeperNodes, ip)
	}
	zookeeperNodeList := strings.Join(zookeeperNodes, ",")

	var cassandraNodes []string
	for _, ip := range cassandraInstance.Status.Nodes {
		cassandraNodes = append(cassandraNodes, ip)
	}
	cassandraNodeList := strings.Join(cassandraNodes, ",")

	managerInstance := &contrailv1alpha1.Manager{}
	err = r.client.Get(context.TODO(), request.NamespacedName, managerInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Manager Instance")
		}
	} else {
		instance.Spec.Service = managerInstance.Spec.Config
		if managerInstance.Spec.Config.Size != nil {
			instance.Spec.Service.Size = managerInstance.Spec.Config.Size
		} else {
			instance.Spec.Service.Size = managerInstance.Spec.Size
		}
		if managerInstance.Spec.HostNetwork != nil {
			instance.Spec.HostNetwork = managerInstance.Spec.HostNetwork
		}
	}

	// Get default Deployment
	deployment := GetDeployment()

	if managerInstance.Spec.ImagePullSecrets != nil {
		var imagePullSecretsList []corev1.LocalObjectReference
		for _, imagePullSecretName := range managerInstance.Spec.ImagePullSecrets {
			imagePullSecret := corev1.LocalObjectReference{
				Name: imagePullSecretName,
			}
			imagePullSecretsList = append(imagePullSecretsList, imagePullSecret)
		}
		deployment.Spec.Template.Spec.ImagePullSecrets = imagePullSecretsList
	}

	// Create initial ConfigMap
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "config-" + instance.Name,
			Namespace: instance.Namespace,
		},
		Data: instance.Spec.Service.Configuration,
	}
	controllerutil.SetControllerReference(instance, &configMap, r.scheme)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "config-" + instance.Name, Namespace: instance.Namespace}, &configMap)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), &configMap)
		if err != nil {
			reqLogger.Error(err, "Failed to create ConfigMap", "Namespace", instance.Namespace, "Name", "config-"+instance.Name)
			return reconcile.Result{}, err
		}
	}

	// Set Deployment Name & Namespace

	deployment.ObjectMeta.Name = "config-" + instance.Name
	deployment.ObjectMeta.Namespace = instance.Namespace

	// Configure Containers
	for idx, container := range deployment.Spec.Template.Spec.Containers {
		for containerName, image := range instance.Spec.Service.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.Containers[idx]).Image = image
				(&deployment.Spec.Template.Spec.Containers[idx]).EnvFrom[0].ConfigMapRef.Name = "config-" + instance.Name
			}
		}
	}

	// Configure InitContainers
	for idx, container := range deployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range instance.Spec.Service.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.InitContainers[idx]).Image = image
				(&deployment.Spec.Template.Spec.InitContainers[idx]).EnvFrom[0].ConfigMapRef.Name = "config-" + instance.Name
			}
		}
	}

	// Set HostNetwork
	deployment.Spec.Template.Spec.HostNetwork = *instance.Spec.HostNetwork

	// Set Selector and Label
	deployment.Spec.Selector.MatchLabels["app"] = "config-" + instance.Name
	deployment.Spec.Template.ObjectMeta.Labels["app"] = "config-" + instance.Name

	// Set Size
	deployment.Spec.Replicas = instance.Spec.Service.Size

	// Create Deployment
	controllerutil.SetControllerReference(instance, deployment, r.scheme)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "config-" + instance.Name, Namespace: instance.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create Deployment", "Namespace", instance.Namespace, "Name", "config-"+instance.Name)
			return reconcile.Result{}, err
		}
	}

	// Check if Init Containers are running
	_, err = contrailv1alpha1.InitContainerRunning(r.client,
		instance.ObjectMeta,
		"config",
		instance,
		*instance.Spec.Service,
		&instance.Status)

	if err != nil {
		reqLogger.Error(err, "Err Init Pods not ready, requeing")
		return reconcile.Result{}, err
	}

	configMap.Data["RABBITMQ_NODES"] = rabbitmqNodeList
	configMap.Data["ZOOKEEPER_NODES"] = zookeeperNodeList
	configMap.Data["CONFIGDB_NODES"] = cassandraNodeList
	configMap.Data["CONFIGDB_CQL_PORT"] = cassandraInstance.Status.Ports["cqlPort"]
	configMap.Data["CONFIGDB_PORT"] = cassandraInstance.Status.Ports["port"]
	configMap.Data["RABBITMQ_NODE_PORT"] = rabbitmqInstance.Status.Ports["port"]
	configMap.Data["ZOOKEEPER_NODE_PORT"] = zookeeperInstance.Status.Ports["port"]

	var podIpList []string
	for _, ip := range instance.Status.Nodes {
		podIpList = append(podIpList, ip)
	}
	nodeList := strings.Join(podIpList, ",")
	configMap.Data["CONTROLLER_NODES"] = nodeList
	err = r.client.Update(context.TODO(), &configMap)
	if err != nil {
		reqLogger.Error(err, "Failed to update ConfigMap", "Namespace", instance.Namespace, "Name", "config-"+instance.Name)
		return reconcile.Result{}, err
	}

	err = contrailv1alpha1.MarkInitPodsReady(r.client, instance.ObjectMeta, "config")

	if err != nil {
		reqLogger.Error(err, "Failed to mark Pods ready")
		return reconcile.Result{}, err
	}

	err = contrailv1alpha1.SetServiceStatus(r.client,
		instance.ObjectMeta,
		"config",
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
