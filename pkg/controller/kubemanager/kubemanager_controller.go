package kubemanager

import (
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_kubemanager")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Kubemanager Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKubemanager{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kubemanager-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Kubemanager
	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Kubemanager{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Kubemanager
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

	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Config{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Manager{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileKubemanager implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKubemanager{}

// ReconcileKubemanager reconciles a Kubemanager object
type ReconcileKubemanager struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Kubemanager object and makes changes based on the state read
// and what is in the Kubemanager.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKubemanager) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	/*
		reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
		reqLogger.Info("Reconciling Kubemanager")

		// Fetch the Kubemanager instance
		instance := &contrailv1alpha1.Kubemanager{}
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
			instance.Spec = managerInstance.Spec.Services.Kubemanager
			if managerInstance.Spec.Services.Kubemanager.Size != nil {
				instance.Spec.Size = managerInstance.Spec.Services.Kubemanager.Size
			} else {
				instance.Spec.Size = managerInstance.Spec.Size
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

		if instance.Spec.Configuration == nil {
			instance.Spec.Configuration = make(map[string]string)
			reqLogger.Info("config map empty, initializing it")
		}

		instance.Spec.Configuration["RABBITMQ_NODES"] = rabbitmqNodeList
		instance.Spec.Configuration["ZOOKEEPER_NODES"] = zookeeperNodeList
		instance.Spec.Configuration["CONFIGDB_NODES"] = cassandraNodeList
		instance.Spec.Configuration["CONFIG_NODES"] = configNodeList
		instance.Spec.Configuration["CONTROLLER_NODES"] = configNodeList
		instance.Spec.Configuration["ANALYTICS_NODES"] = configNodeList
		instance.Spec.Configuration["CONFIGDB_CQL_PORT"] = cassandraInstance.Status.Ports["cqlPort"]
		instance.Spec.Configuration["CONFIGDB_PORT"] = cassandraInstance.Status.Ports["port"]
		instance.Spec.Configuration["RABBITMQ_NODE_PORT"] = rabbitmqInstance.Status.Ports["port"]
		instance.Spec.Configuration["ZOOKEEPER_NODE_PORT"] = zookeeperInstance.Status.Ports["port"]

		if instance.Spec.Configuration["USE_KUBEADM_CONFIG"] == "true" {
			controlPlaneEndpoint := ""
			clusterName := "kubernetes"
			podSubnet := "10.32.0.0/12"
			serviceSubnet := "10.96.0.0/12"
			controlPlaneEndpointHost := "10.96.0.1"
			controlPlaneEndpointPort := "443"

			config, err := rest.InClusterConfig()
			if err == nil {
				clientset, err := kubernetes.NewForConfig(config)
				if err == nil {
					kubeadmConfigMapClient := clientset.CoreV1().ConfigMaps("kube-system")
					kcm, _ := kubeadmConfigMapClient.Get("kubeadm-config", metav1.GetOptions{})
					clusterConfig := kcm.Data["ClusterConfiguration"]
					clusterConfigByte := []byte(clusterConfig)
					clusterConfigMap := make(map[interface{}]interface{})
					err = yaml.Unmarshal(clusterConfigByte, &clusterConfigMap)
					if err != nil {
						return reconcile.Result{}, err
					}
					controlPlaneEndpoint = clusterConfigMap["controlPlaneEndpoint"].(string)
					controlPlaneEndpointHost, controlPlaneEndpointPort, _ = net.SplitHostPort(controlPlaneEndpoint)
					clusterName = clusterConfigMap["clusterName"].(string)
					networkConfig := make(map[interface{}]interface{})
					networkConfig = clusterConfigMap["networking"].(map[interface{}]interface{})
					podSubnet = networkConfig["podSubnet"].(string)
					serviceSubnet = networkConfig["serviceSubnet"].(string)
				}
			}
			if kubeApiServer, ok := instance.Spec.Configuration["KUBERNETES_API_SERVER"]; ok {
				instance.Spec.Configuration["KUBERNETES_API_SERVER"] = kubeApiServer
			} else {
				instance.Spec.Configuration["KUBERNETES_API_SERVER"] = controlPlaneEndpointHost
			}
			if kubeApiSecurePort, ok := instance.Spec.Configuration["KUBERNETES_API_SECURE_PORT"]; ok {
				instance.Spec.Configuration["KUBERNETES_API_SECURE_PORT"] = kubeApiSecurePort
			} else {
				instance.Spec.Configuration["KUBERNETES_API_SECURE_PORT"] = controlPlaneEndpointPort
			}
			if kubePodSubnets, ok := instance.Spec.Configuration["KUBERNETES_POD_SUBNETS"]; ok {
				instance.Spec.Configuration["KUBERNETES_POD_SUBNETS"] = kubePodSubnets
			} else {
				instance.Spec.Configuration["KUBERNETES_POD_SUBNETS"] = podSubnet
			}
			if kubeServiceSubnets, ok := instance.Spec.Configuration["KUBERNETES_SERVICE_SUBNETS"]; ok {
				instance.Spec.Configuration["KUBERNETES_SERVICE_SUBNETS"] = kubeServiceSubnets
			} else {
				instance.Spec.Configuration["KUBERNETES_SERVICE_SUBNETS"] = serviceSubnet
			}
			if kubeClusterName, ok := instance.Spec.Configuration["KUBERNETES_CLUSTER_NAME"]; ok {
				instance.Spec.Configuration["KUBERNETES_CLUSTER_NAME"] = kubeClusterName
			} else {
				instance.Spec.Configuration["KUBERNETES_CLUSTER_NAME"] = clusterName
			}
		}

		// Create initial ConfigMap
		configMap := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "kubemanager-" + instance.Name,
				Namespace: instance.Namespace,
			},
			Data: instance.Spec.Configuration,
		}
		controllerutil.SetControllerReference(instance, &configMap, r.scheme)
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: "kubemanager-" + instance.Name, Namespace: instance.Namespace}, &configMap)
		if err != nil && errors.IsNotFound(err) {
			err = r.client.Create(context.TODO(), &configMap)
			if err != nil {
				reqLogger.Error(err, "Failed to create ConfigMap", "Namespace", instance.Namespace, "Name", "kubemanager-"+instance.Name)
				return reconcile.Result{}, err
			}
		}
		var serviceAccountName string
		if serviceAccount, ok := instance.Spec.Configuration["serviceAccount"]; ok {
			serviceAccountName = serviceAccount
		} else {
			serviceAccountName = "contrail-service-account"
		}

		var clusterRoleName string
		if clusterRole, ok := instance.Spec.Configuration["clusterRole"]; ok {
			clusterRoleName = clusterRole
		} else {
			clusterRoleName = "contrail-cluster-role"
		}

		var clusterRoleBindingName string
		if clusterRoleBinding, ok := instance.Spec.Configuration["clusterRoleBinding"]; ok {
			clusterRoleBindingName = clusterRoleBinding
		} else {
			clusterRoleBindingName = "contrail-cluster-role-binding"
		}

		existingServiceAccount := &corev1.ServiceAccount{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: serviceAccountName, Namespace: instance.Namespace}, existingServiceAccount)
		if err != nil && errors.IsNotFound(err) {
			serviceAccount := &corev1.ServiceAccount{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "ServiceAccount",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      serviceAccountName,
					Namespace: instance.Namespace,
				},
			}
			controllerutil.SetControllerReference(instance, serviceAccount, r.scheme)
			err = r.client.Create(context.TODO(), serviceAccount)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		existingClusterRole := &rbacv1.ClusterRole{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: clusterRoleName}, existingClusterRole)
		if err != nil && errors.IsNotFound(err) {
			clusterRole := &rbacv1.ClusterRole{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "rbac/v1",
					Kind:       "ClusterRole",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterRoleName,
					Namespace: instance.Namespace,
				},
				Rules: []rbacv1.PolicyRule{{
					Verbs: []string{
						"*",
					},
					APIGroups: []string{
						"*",
					},
					Resources: []string{
						"*",
					},
				}},
			}
			controllerutil.SetControllerReference(instance, clusterRole, r.scheme)
			err = r.client.Create(context.TODO(), clusterRole)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		existingClusterRoleBinding := &rbacv1.ClusterRoleBinding{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBindingName}, existingClusterRoleBinding)
		if err != nil && errors.IsNotFound(err) {
			clusterRoleBinding := &rbacv1.ClusterRoleBinding{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "rbac/v1",
					Kind:       "ClusterRoleBinding",
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      clusterRoleBindingName,
					Namespace: instance.Namespace,
				},
				Subjects: []rbacv1.Subject{{
					Kind:      "ServiceAccount",
					Name:      serviceAccountName,
					Namespace: instance.Namespace,
				}},
				RoleRef: rbacv1.RoleRef{
					APIGroup: "rbac.authorization.k8s.io",
					Kind:     "ClusterRole",
					Name:     clusterRoleName,
				},
			}
			controllerutil.SetControllerReference(instance, clusterRoleBinding, r.scheme)
			err = r.client.Create(context.TODO(), clusterRoleBinding)
			if err != nil {
				return reconcile.Result{}, err
			}
		}

		deployment.Spec.Template.Spec.ServiceAccountName = serviceAccountName
		// Set Deployment Name & Namespace

		deployment.ObjectMeta.Name = "kubemanager-" + instance.Name
		deployment.ObjectMeta.Namespace = instance.Namespace

		// Configure Containers
		for idx, container := range deployment.Spec.Template.Spec.Containers {
			for containerName, image := range instance.Spec.Images {
				if containerName == container.Name {
					(&deployment.Spec.Template.Spec.Containers[idx]).Image = image
					(&deployment.Spec.Template.Spec.Containers[idx]).EnvFrom[0].ConfigMapRef.Name = "kubemanager-" + instance.Name
				}
			}
		}

		// Configure InitContainers
		for idx, container := range deployment.Spec.Template.Spec.InitContainers {
			for containerName, image := range instance.Spec.Images {
				if containerName == container.Name {
					(&deployment.Spec.Template.Spec.InitContainers[idx]).Image = image
					(&deployment.Spec.Template.Spec.InitContainers[idx]).EnvFrom[0].ConfigMapRef.Name = "kubemanager-" + instance.Name
				}
			}
		}

		// Set HostNetwork
		deployment.Spec.Template.Spec.HostNetwork = *instance.Spec.HostNetwork

		// Set Selector and Label
		deployment.Spec.Selector.MatchLabels["app"] = "kubemanager-" + instance.Name
		deployment.Spec.Template.ObjectMeta.Labels["app"] = "kubemanager-" + instance.Name

		// Set Size
		deployment.Spec.Replicas = instance.Spec.Size

		// Create Deployment
		controllerutil.SetControllerReference(instance, deployment, r.scheme)
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: "kubemanager-" + instance.Name, Namespace: instance.Namespace}, deployment)
		if err != nil && errors.IsNotFound(err) {
			err = r.client.Create(context.TODO(), deployment)
			if err != nil {
				reqLogger.Error(err, "Failed to create Deployment", "Namespace", instance.Namespace, "Name", "kubemanager-"+instance.Name)
				return reconcile.Result{}, err
			}
		}

		// Check if Init Containers are running
		_, err = contrailv1alpha1.InitContainerRunning(r.client,
			instance.ObjectMeta,
			"kubemanager",
			instance,
			*instance.Spec,
			&instance.Status)

		if err != nil {
			reqLogger.Error(err, "Err Init Pods not ready, requeing")
			return reconcile.Result{}, err
		}

		err = contrailv1alpha1.MarkInitPodsReady(r.client, instance.ObjectMeta, "kubemanager")

		if err != nil {
			reqLogger.Error(err, "Failed to mark Pods ready")
			return reconcile.Result{}, err
		}

		err = contrailv1alpha1.SetServiceStatus(r.client,
			instance.ObjectMeta,
			"kubemanager",
			instance,
			&deployment.Status,
			&instance.Status)

		if err != nil {
			reqLogger.Error(err, "Failed to set Service status")
			return reconcile.Result{}, err
		} else {
			reqLogger.Info("set service status")
		}
	*/
	return reconcile.Result{}, nil
}
