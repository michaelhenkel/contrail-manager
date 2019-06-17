package webui

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

var log = logf.Log.WithName("controller_webui")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Webui Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWebui{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("webui-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Webui
	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Webui{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Webui
	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Cassandra{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Cassandra{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Zookeeper{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Zookeeper{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Rabbitmq{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Rabbitmq{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Config{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Config{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileWebui implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileWebui{}

// ReconcileWebui reconciles a Webui object
type ReconcileWebui struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Webui object and makes changes based on the state read
// and what is in the Webui.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWebui) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Webui")

	// Fetch the Webui instance
	instance := &contrailv1alpha1.Webui{}
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

	configInstance := &contrailv1alpha1.Config{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, configInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Config Instance")
			return reconcile.Result{}, err
		}
	}
	configStatus := false
	if configInstance.Status.Active != nil {
		configStatus = *configInstance.Status.Active
	}

	if !rabbitmqStatus || !zookeeperStatus || !cassandraStatus || !configStatus {
		reqLogger.Info("cassandra, zookeeper, rmq or config not ready")
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

	var configNodes []string
	for _, ip := range configInstance.Status.Nodes {
		configNodes = append(configNodes, ip)
	}
	configNodeList := strings.Join(configNodes, ",")

	managerInstance := &contrailv1alpha1.Manager{}
	err = r.client.Get(context.TODO(), request.NamespacedName, managerInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Manager Instance")
		}
	} else {
		instance.Spec.Service = managerInstance.Spec.Webui
		if managerInstance.Spec.Webui.Size != nil {
			instance.Spec.Service.Size = managerInstance.Spec.Webui.Size
		} else {
			instance.Spec.Service.Size = managerInstance.Spec.Size
		}
		if managerInstance.Spec.HostNetwork != nil {
			instance.Spec.HostNetwork = managerInstance.Spec.HostNetwork
		}
	}

	// Get default Deployment
	deployment := GetDeployment()

	if instance.Spec.Service.Configuration == nil {
		instance.Spec.Service.Configuration = make(map[string]string)
		reqLogger.Info("config map empty, initializing it")
	}

	instance.Spec.Service.Configuration["RABBITMQ_NODES"] = rabbitmqNodeList
	instance.Spec.Service.Configuration["ZOOKEEPER_NODES"] = zookeeperNodeList
	instance.Spec.Service.Configuration["CONFIGDB_NODES"] = cassandraNodeList
	instance.Spec.Service.Configuration["CONFIG_NODES"] = configNodeList
	instance.Spec.Service.Configuration["CONTROLLER_NODES"] = configNodeList
	instance.Spec.Service.Configuration["ANALYTICS_NODES"] = configNodeList
	instance.Spec.Service.Configuration["CONFIGDB_CQL_PORT"] = cassandraInstance.Status.Ports["cqlPort"]
	instance.Spec.Service.Configuration["CONFIGDB_PORT"] = cassandraInstance.Status.Ports["port"]
	instance.Spec.Service.Configuration["RABBITMQ_NODE_PORT"] = rabbitmqInstance.Status.Ports["port"]
	instance.Spec.Service.Configuration["ZOOKEEPER_NODE_PORT"] = zookeeperInstance.Status.Ports["port"]

	// Create initial ConfigMap
	configMap := corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "webui-" + instance.Name,
			Namespace: instance.Namespace,
		},
		Data: instance.Spec.Service.Configuration,
	}
	controllerutil.SetControllerReference(instance, &configMap, r.scheme)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "webui-" + instance.Name, Namespace: instance.Namespace}, &configMap)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), &configMap)
		if err != nil {
			reqLogger.Error(err, "Failed to create ConfigMap", "Namespace", instance.Namespace, "Name", "webui-"+instance.Name)
			return reconcile.Result{}, err
		}
	}

	// Set Deployment Name & Namespace

	deployment.ObjectMeta.Name = "webui-" + instance.Name
	deployment.ObjectMeta.Namespace = instance.Namespace

	// Configure Containers
	for idx, container := range deployment.Spec.Template.Spec.Containers {
		for containerName, image := range instance.Spec.Service.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.Containers[idx]).Image = image
				(&deployment.Spec.Template.Spec.Containers[idx]).EnvFrom[0].ConfigMapRef.Name = "webui-" + instance.Name
			}
		}
	}

	// Configure InitContainers
	for idx, container := range deployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range instance.Spec.Service.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.InitContainers[idx]).Image = image
				(&deployment.Spec.Template.Spec.InitContainers[idx]).EnvFrom[0].ConfigMapRef.Name = "webui-" + instance.Name
			}
		}
	}

	// Set HostNetwork
	deployment.Spec.Template.Spec.HostNetwork = *instance.Spec.HostNetwork

	// Set Selector and Label
	deployment.Spec.Selector.MatchLabels["app"] = "webui-" + instance.Name
	deployment.Spec.Template.ObjectMeta.Labels["app"] = "webui-" + instance.Name

	// Set Size
	deployment.Spec.Replicas = instance.Spec.Service.Size

	// Create Deployment
	controllerutil.SetControllerReference(instance, deployment, r.scheme)
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: "webui-" + instance.Name, Namespace: instance.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		err = r.client.Create(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create Deployment", "Namespace", instance.Namespace, "Name", "webui-"+instance.Name)
			return reconcile.Result{}, err
		}
	}

	return reconcile.Result{}, nil
}
