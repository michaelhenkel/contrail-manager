package config

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_config")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {

	return &ReconcileConfig{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("config-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Config
	err = c.Watch(&source.Kind{Type: &v1alpha1.Config{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	srcCassandra := &source.Kind{Type: &v1alpha1.Cassandra{}}
	cassandraHandler := &handler.EnqueueRequestForOwner{
		OwnerType:    &v1alpha1.Manager{},
		IsController: true,
	}
	cassandraPred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			//return true
			oldCassandra := e.ObjectOld.(*v1alpha1.Cassandra)
			newCassandra := e.ObjectNew.(*v1alpha1.Cassandra)
			oldCassandraActive := false
			newCassandraActive := false
			if oldCassandra.Status.Active != nil {
				if *oldCassandra.Status.Active {
					oldCassandraActive = true
				}
			}
			if newCassandra.Status.Active != nil {
				if *newCassandra.Status.Active {
					newCassandraActive = true
				}
			}
			if oldCassandraActive != newCassandraActive {
				return true
			}
			return false
		},
	}
	// Watch for Manager events.
	err = c.Watch(srcCassandra, cassandraHandler, cassandraPred)
	if err != nil {
		return err
	}

	srcPod := &source.Kind{Type: &corev1.Pod{}}
	podHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1.ReplicaSet{},
	}
	predPodIp := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			labels := e.MetaNew.GetLabels()
			if v, ok := labels["contrail_manager"]; ok {
				if v == "config" {
					oldPod := e.ObjectOld.(*corev1.Pod)
					newPod := e.ObjectNew.(*corev1.Pod)
					if oldPod.Status.PodIP != newPod.Status.PodIP {
						return true
					}
				}
			}
			return false
		},
	}

	// Watch for Pod events.
	//err = c.Watch(srcPod, podHandler, predPodIp, predInitStatus)
	err = c.Watch(srcPod, podHandler, predPodIp)
	if err != nil {
		return err
	}
	/*
		srcManager := &source.Kind{Type: &v1alpha1.Manager{}}

		managerHandler := &handler.EnqueueRequestForObject{}
		pred := predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {

				oldManager := e.ObjectOld.(*v1alpha1.Manager)
				newManager := e.ObjectNew.(*v1alpha1.Manager)
				var oldSize, newSize int32
				if oldManager.Spec.Services.Config.Size != nil {
					oldSize = *oldManager.Spec.Services.Config.Size
				} else {
					oldSize = *oldManager.Spec.Size
				}
				if newManager.Spec.Services.Config.Size != nil {
					newSize = *newManager.Spec.Services.Config.Size
				} else {
					newSize = *newManager.Spec.Size
				}

				if oldSize != newSize {
					return true
				}
				return false
			},
		}
		// Watch for Manager events.
		err = c.Watch(srcManager, managerHandler, pred)
		if err != nil {
			return err
		}
	*/
	srcDeployment := &source.Kind{Type: &appsv1.Deployment{}}
	deploymentHandler := &handler.EnqueueRequestForObject{}
	deploymentPred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldDeployment := e.ObjectOld.(*appsv1.Deployment)
			newDeployment := e.ObjectNew.(*appsv1.Deployment)
			isOwner := false
			for _, owner := range newDeployment.ObjectMeta.OwnerReferences {
				if owner.Kind == "Config" {
					isOwner = true
				}
			}
			if (oldDeployment.Status.ReadyReplicas != newDeployment.Status.ReadyReplicas) && isOwner {
				return true
			}
			return false
		},
	}
	// Watch for Manager events.

	/*
			srcCassandra := &source.Kind{Type: &v1alpha1.Cassandra{}}
		cassandraHandler := &handler.EnqueueRequestForOwner{
			OwnerType:    &v1alpha1.Manager{},
			IsController: true,
		}
		cassandraPred := predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				//return true
				oldCassandra := e.ObjectOld.(*v1alpha1.Cassandra)
				newCassandra := e.ObjectNew.(*v1alpha1.Cassandra)
				oldCassandraActive := false
				newCassandraActive := false
				if oldCassandra.Status.Active != nil {
					if *oldCassandra.Status.Active {
						oldCassandraActive = true
					}
				}
				if newCassandra.Status.Active != nil {
					if *newCassandra.Status.Active {
						newCassandraActive = true
					}
				}
				if oldCassandraActive != newCassandraActive {
					return true
				}
				return false
			},
		}
		// Watch for Manager events.
		err = c.Watch(srcCassandra, cassandraHandler, cassandraPred)
		if err != nil {
			return err
		}
	*/
	err = c.Watch(srcDeployment, deploymentHandler, deploymentPred)
	if err != nil {
		return err
	}

	srcRabbitmq := &source.Kind{Type: &v1alpha1.Rabbitmq{}}
	rabbitmqHandler := &handler.EnqueueRequestForOwner{
		OwnerType:    &v1alpha1.Manager{},
		IsController: true,
	}
	rabbitmqPred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldRabbitmq := e.ObjectOld.(*v1alpha1.Rabbitmq)
			newRabbitmq := e.ObjectNew.(*v1alpha1.Rabbitmq)
			oldRabbitmqActive := false
			newRabbitmqActive := false
			if oldRabbitmq.Status.Active != nil {
				if *oldRabbitmq.Status.Active {
					oldRabbitmqActive = true
				}
			}
			if newRabbitmq.Status.Active != nil {
				if *newRabbitmq.Status.Active {
					newRabbitmqActive = true
				}
			}
			if oldRabbitmqActive != newRabbitmqActive {
				return true
			}
			return false
		},
	}
	// Watch for Manager events.
	err = c.Watch(srcRabbitmq, rabbitmqHandler, rabbitmqPred)
	if err != nil {
		return err
	}

	srcZookeeper := &source.Kind{Type: &v1alpha1.Zookeeper{}}
	zookeeperHandler := &handler.EnqueueRequestForOwner{
		OwnerType:    &v1alpha1.Manager{},
		IsController: true,
	}
	zookeeperPred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldZookeeper := e.ObjectOld.(*v1alpha1.Zookeeper)
			newZookeeper := e.ObjectNew.(*v1alpha1.Zookeeper)
			oldActiveStatus := false
			newActiveStatus := false
			if oldZookeeper.Status.Active != nil {
				oldActiveStatus = *oldZookeeper.Status.Active
			}
			if newZookeeper.Status.Active != nil {
				newActiveStatus = *newZookeeper.Status.Active
			}
			if oldActiveStatus != newActiveStatus {
				return true
			}

			return false
		},
	}
	// Watch for Manager events.
	err = c.Watch(srcZookeeper, zookeeperHandler, zookeeperPred)
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
	Client    client.Client
	Scheme    *runtime.Scheme
	podsReady *bool
}

func (r *ReconcileConfig) GetRequestObject(request reconcile.Request) (ro runtime.Object) {

	configInstance := &v1alpha1.Config{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, configInstance)
	if err == nil {
		return configInstance
	}

	replicaSetInstance := &appsv1.ReplicaSet{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, replicaSetInstance)
	if err == nil {
		return replicaSetInstance
	}

	deploymentInstance := &appsv1.Deployment{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, deploymentInstance)
	if err == nil {
		return deploymentInstance
	}

	return nil
}

func (r *ReconcileConfig) ConfigReconcile(request reconcile.Request) (reconcile.Result, error) {
	/*
		reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
		reqLogger.Info("Reconciling Config Object")

		configInstance := &v1alpha1.Config{}
		err := r.Client.Get(context.TODO(), request.NamespacedName, configInstance)
		if err != nil {
			if errors.IsNotFound(err) {
				reqLogger.Info("No Config Instance")
			}
		}

		cassandraActive := false
		cassandraInstance := &v1alpha1.Cassandra{}
		err = r.Client.Get(context.TODO(), request.NamespacedName, cassandraInstance)
		if err == nil {
			if cassandraInstance.Status.Active != nil {
				cassandraActive = *cassandraInstance.Status.Active
			}
		}

		zookeeperActive := false
		zookeeperInstance := &v1alpha1.Zookeeper{}
		err = r.Client.Get(context.TODO(), request.NamespacedName, zookeeperInstance)
		if err == nil {
			if zookeeperInstance.Status.Active != nil {
				zookeeperActive = *zookeeperInstance.Status.Active
			}
		}

		rabbitmqActive := false
		rabbitmqInstance := &v1alpha1.Rabbitmq{}
		err = r.Client.Get(context.TODO(), request.NamespacedName, rabbitmqInstance)
		if err == nil {
			if rabbitmqInstance.Status.Active != nil {
				rabbitmqActive = *rabbitmqInstance.Status.Active
			}
		}

		managerInstance := &v1alpha1.Manager{}
		err = r.Client.Get(context.TODO(), request.NamespacedName, managerInstance)
		if err != nil {
			if errors.IsNotFound(err) {
				reqLogger.Info("No Manager Instance")
			}
		} else {
			configInstance.Spec = managerInstance.Spec.Services.Config
			if managerInstance.Spec.Services.Config.Size != nil {
				configInstance.Spec.Size = managerInstance.Spec.Services.Config.Size
			} else {
				configInstance.Spec.Size = managerInstance.Spec.Size
			}
			if managerInstance.Spec.HostNetwork != nil {
				configInstance.Spec.HostNetwork = managerInstance.Spec.HostNetwork
			}
		}
		err = r.Client.Update(context.TODO(), configInstance)
		if err != nil {
			reqLogger.Error(err, "Failed to update instance")
			return reconcile.Result{}, err
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
				Name:      "config-" + configInstance.Name,
				Namespace: configInstance.Namespace,
				Labels:    map[string]string{"contrail_manager": "config"},
			},
			Data: configInstance.Spec.Configuration,
		}
		controllerutil.SetControllerReference(configInstance, &configMap, r.Scheme)
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "config-" + configInstance.Name, Namespace: configInstance.Namespace}, &configMap)
		if err != nil && errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), &configMap)
			if err != nil {
				reqLogger.Error(err, "Failed to create ConfigMap", "Namespace", configInstance.Namespace, "Name", "config-"+configInstance.Name)
				return reconcile.Result{}, err
			}
		}

		// Set Deployment Name & Namespace

		deployment.ObjectMeta.Name = "config-" + configInstance.Name
		deployment.ObjectMeta.Namespace = configInstance.Namespace

		for idx, container := range deployment.Spec.Template.Spec.Containers {
			for containerName, image := range configInstance.Spec.Images {
				if containerName == container.Name {
					(&deployment.Spec.Template.Spec.Containers[idx]).Image = image
					(&deployment.Spec.Template.Spec.Containers[idx]).EnvFrom[0].ConfigMapRef.Name = "config-" + configInstance.Name
				}
			}
		}

		// Configure InitContainers
		for idx, container := range deployment.Spec.Template.Spec.InitContainers {
			for containerName, image := range configInstance.Spec.Images {
				if containerName == container.Name {
					(&deployment.Spec.Template.Spec.InitContainers[idx]).Image = image
					(&deployment.Spec.Template.Spec.InitContainers[idx]).EnvFrom[0].ConfigMapRef.Name = "config-" + configInstance.Name
				}
			}
		}

		// Set HostNetwork
		deployment.Spec.Template.Spec.HostNetwork = *configInstance.Spec.HostNetwork

		// Set Selector and Label
		deployment.Spec.Selector.MatchLabels["app"] = "config-" + configInstance.Name
		deployment.Spec.Template.ObjectMeta.Labels["app"] = "config-" + configInstance.Name

		// Set Size
		deployment.Spec.Replicas = configInstance.Spec.Size
		// Create Deployment

		controllerutil.SetControllerReference(configInstance, deployment, r.Scheme)
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "config-" + configInstance.Name, Namespace: configInstance.Namespace}, deployment)
		if err != nil && errors.IsNotFound(err) {
			deployment.Spec.Template.ObjectMeta.Labels["version"] = "1"
			err = r.Client.Create(context.TODO(), deployment)
			if err != nil {
				reqLogger.Error(err, "Failed to create Deployment", "Namespace", configInstance.Namespace, "Name", "config-"+configInstance.Name)
				return reconcile.Result{}, err
			}
		} else if err == nil && *deployment.Spec.Replicas != *configInstance.Spec.Size {
			deployment.Spec.Replicas = configInstance.Spec.Size
			versionInt, _ := strconv.Atoi(deployment.Spec.Template.ObjectMeta.Labels["version"])
			newVersion := versionInt + 1
			newVersionString := strconv.Itoa(newVersion)
			deployment.Spec.Template.ObjectMeta.Labels["version"] = newVersionString
			err = r.Client.Update(context.TODO(), deployment)
			if err != nil {
				reqLogger.Error(err, "Failed to update Deployment", "Namespace", configInstance.Namespace, "Name", "config-"+configInstance.Name)
				return reconcile.Result{}, err
			}
		}

		replicaSetList := &appsv1.ReplicaSetList{}
		labelSelector := labels.SelectorFromSet(map[string]string{"app": deployment.Name})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		err = r.Client.List(context.TODO(), listOps, replicaSetList)
		if err != nil {
			return reconcile.Result{}, err
		}
		if len(replicaSetList.Items) > 0 {
			replicaSetNameSpacedName := types.NamespacedName{Name: replicaSetList.Items[0].Name, Namespace: request.Namespace}
			err := r.ReplicaSetReconcile(reconcile.Request{replicaSetNameSpacedName})
			if err != nil {
				return reconcile.Result{}, err
			}
		}
		podsReady := false
		if r.podsReady != nil {
			podsReady = *r.podsReady
		}
		fmt.Println("ZK active: ", zookeeperActive)
		fmt.Println("RMQ active: ", rabbitmqActive)
		fmt.Println("Cass active: ", cassandraActive)
		fmt.Println("Pods active: ", podsReady)
		if zookeeperActive && rabbitmqActive && cassandraActive && podsReady {
			r.finalize(request)
		}
	*/
	return reconcile.Result{}, nil
}

func (r *ReconcileConfig) finalize(request reconcile.Request) (reconcile.Result, error) {
	configMapList := &corev1.ConfigMapList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": "config"})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	err := r.Client.List(context.TODO(), listOps, configMapList)
	if err != nil {
		return reconcile.Result{}, err
	}
	configMap := configMapList.Items[0]

	configInstance := &v1alpha1.Config{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, configInstance)
	if err != nil {
		return reconcile.Result{}, err
	}

	cassandraInstance := &v1alpha1.Cassandra{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, cassandraInstance)
	if err != nil {
		return reconcile.Result{}, err
	}
	zookeeperInstance := &v1alpha1.Zookeeper{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, zookeeperInstance)
	if err != nil {
		return reconcile.Result{}, err
	}
	rabbitmqInstance := &v1alpha1.Rabbitmq{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, rabbitmqInstance)
	if err != nil {
		return reconcile.Result{}, err
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

	serviceChanged := false
	if v, ok := configMap.Data["CONFIGDB_NODES"]; ok {
		cassandraNodeSlice := strings.Split(cassandraNodeList, ",")
		configDBSlice := strings.Split(v, ",")
		sort.Strings(cassandraNodeSlice)
		sort.Strings(configDBSlice)
		reflect.DeepEqual(cassandraNodeSlice, configDBSlice)
		if !reflect.DeepEqual(cassandraNodeSlice, configDBSlice) {
			serviceChanged = true
		}
		fmt.Println("cassandraNodeList", cassandraNodeList)
		fmt.Println("CONFIGDB_NODES", configMap.Data["CONFIGDB_NODES"])

	}
	if v, ok := configMap.Data["ZOOKEEPER_NODES"]; ok {
		zookeeperNodeSlice := strings.Split(zookeeperNodeList, ",")
		configDBSlice := strings.Split(v, ",")
		sort.Strings(zookeeperNodeSlice)
		sort.Strings(configDBSlice)
		reflect.DeepEqual(zookeeperNodeSlice, configDBSlice)
		if !reflect.DeepEqual(zookeeperNodeSlice, configDBSlice) {
			serviceChanged = true
		}
		fmt.Println("zkNodeList", zookeeperNodeList)
		fmt.Println("ZOOKEEPER_NODES", configMap.Data["ZOOKEEPER_NODES"])
	}
	if v, ok := configMap.Data["RABBITMQ_NODES"]; ok {
		rabbitmqNodeSlice := strings.Split(rabbitmqNodeList, ",")
		configDBSlice := strings.Split(v, ",")
		sort.Strings(rabbitmqNodeSlice)
		sort.Strings(configDBSlice)
		reflect.DeepEqual(rabbitmqNodeSlice, configDBSlice)
		if !reflect.DeepEqual(rabbitmqNodeSlice, configDBSlice) {
			serviceChanged = true
		}
		fmt.Println("rmqNodeList", zookeeperNodeList)
		fmt.Println("RABBITMQ_NODES", configMap.Data["RABBITMQ_NODES"])
	}
	configMap.Data["RABBITMQ_NODES"] = rabbitmqNodeList
	configMap.Data["ZOOKEEPER_NODES"] = zookeeperNodeList
	configMap.Data["CONFIGDB_NODES"] = cassandraNodeList
	configMap.Data["CONFIGDB_CQL_PORT"] = cassandraInstance.Status.Ports["cqlPort"]
	configMap.Data["CONFIGDB_PORT"] = cassandraInstance.Status.Ports["port"]
	configMap.Data["RABBITMQ_NODE_PORT"] = rabbitmqInstance.Status.Ports["port"]
	configMap.Data["ZOOKEEPER_NODE_PORT"] = zookeeperInstance.Status.Ports["port"]

	if serviceChanged {
		err = r.Client.Update(context.TODO(), &configMap)
		if err != nil {
			return reconcile.Result{}, err
		}
		deployment := &appsv1.Deployment{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "config-" + configInstance.Name, Namespace: configInstance.Namespace}, deployment)
		if err != nil {
			return reconcile.Result{}, err
		}
		versionInt, _ := strconv.Atoi(deployment.Spec.Template.ObjectMeta.Labels["version"])
		newVersion := versionInt + 1
		newVersionString := strconv.Itoa(newVersion)
		deployment.Spec.Template.ObjectMeta.Labels["version"] = newVersionString
		err = r.Client.Update(context.TODO(), deployment)
		if err != nil {
			return reconcile.Result{}, err
		}
		return reconcile.Result{}, nil
	}

	podList := &corev1.PodList{}
	err = r.Client.List(context.TODO(), listOps, podList)
	if err != nil {
		return reconcile.Result{}, err
	}

	var podNameIPMap = make(map[string]string)
	for _, pod := range podList.Items {
		if pod.Status.PodIP != "" {
			podNameIPMap[pod.Name] = pod.Status.PodIP
		}
	}

	var podIPList []string
	for _, ip := range podNameIPMap {
		podIPList = append(podIPList, ip)
	}

	nodeList := strings.Join(podIPList, ",")
	configMap.Data["CONTROLLER_NODES"] = nodeList
	err = r.Client.Update(context.TODO(), &configMap)
	if err != nil {
		return reconcile.Result{}, err
	}
	for _, pod := range podList.Items {
		pod.ObjectMeta.Labels["status"] = "ready"
		err = r.Client.Update(context.TODO(), &pod)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	configList := &v1alpha1.ConfigList{}
	configListOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	err = r.Client.List(context.TODO(), configListOps, configList)
	if err != nil {
		return reconcile.Result{}, err
	}
	config := configList.Items[0]
	config.Status.Nodes = podNameIPMap
	err = r.Client.Status().Update(context.TODO(), &config)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileConfig) ManagerReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Manager Object")
	configInstanceList := &v1alpha1.ConfigList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": "config"})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	err := r.Client.List(context.TODO(), listOps, configInstanceList)
	if err != nil {
		return reconcile.Result{}, err
	}
	configInstance := v1alpha1.Config{}
	if len(configInstanceList.Items) > 0 {
		configInstance = configInstanceList.Items[0]
		cassandra := &v1alpha1.Cassandra{}
		err = r.Client.Get(context.TODO(), request.NamespacedName, cassandra)
		if err != nil {
			return reconcile.Result{}, err
		}
		configInstance.Labels["cassandraActive"] = "false"
		if cassandra.Status.Active != nil {
			if *cassandra.Status.Active {
				configInstance.Labels["cassandraActive"] = "true"
			}
		}
		zookeeper := &v1alpha1.Zookeeper{}
		err = r.Client.Get(context.TODO(), request.NamespacedName, zookeeper)
		if err != nil {
			return reconcile.Result{}, err
		}
		configInstance.Labels["zookeeperActive"] = "false"
		if zookeeper.Status.Active != nil {
			if *zookeeper.Status.Active {
				configInstance.Labels["zookeeperActive"] = "true"
			}
		}
		rabbitmq := &v1alpha1.Rabbitmq{}
		err = r.Client.Get(context.TODO(), request.NamespacedName, rabbitmq)
		if err != nil {
			return reconcile.Result{}, err
		}
		configInstance.Labels["rabbitmqActive"] = "false"
		if rabbitmq.Status.Active != nil {
			if *rabbitmq.Status.Active {
				configInstance.Labels["rabbitmqActive"] = "true"
			}
		}
		err = r.Client.Update(context.TODO(), &configInstance)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileConfig) DeploymentReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Config due to Deployment changes")
	deployment := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, deployment)
	if err != nil {
		return reconcile.Result{}, err
	}
	var ownerName string
	for _, owner := range deployment.ObjectMeta.OwnerReferences {
		if owner.Kind == "Config" {
			ownerName = owner.Name
		}
	}
	owner := &v1alpha1.Config{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: ownerName, Namespace: request.Namespace}, owner)
	if err != nil {
		return reconcile.Result{}, err
	}
	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {

		active := true
		owner.Status.Active = &active
		err = r.Client.Status().Update(context.TODO(), owner)
		if err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Config Deployment is ready")

	} else {
		owner.Status.Active = nil
		err = r.Client.Status().Update(context.TODO(), owner)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}
func (r *ReconcileConfig) CassandraReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Cassandra Object")
	configInstanceList := &v1alpha1.ConfigList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": "config"})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	err := r.Client.List(context.TODO(), listOps, configInstanceList)
	if err != nil {
		return reconcile.Result{}, err
	}
	configInstance := configInstanceList.Items[0]
	cassandra := &v1alpha1.Cassandra{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, cassandra)
	if err != nil {
		return reconcile.Result{}, err
	}
	configInstance.Labels["cassandraActive"] = "false"
	if cassandra.Status.Active != nil {
		if *cassandra.Status.Active {
			configInstance.Labels["cassandraActive"] = "true"
		}
	}
	err = r.Client.Update(context.TODO(), &configInstance)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileConfig) RabbitmqReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Rabbitmq Object")
	configInstanceList := &v1alpha1.ConfigList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": "config"})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	err := r.Client.List(context.TODO(), listOps, configInstanceList)
	if err != nil {
		return reconcile.Result{}, err
	}
	configInstance := configInstanceList.Items[0]
	rabbitmq := &v1alpha1.Rabbitmq{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, rabbitmq)
	if err != nil {
		return reconcile.Result{}, err
	}
	configInstance.Labels["rabbitmqActive"] = "false"
	if rabbitmq.Status.Active != nil {
		if *rabbitmq.Status.Active {
			configInstance.Labels["rabbitmqActive"] = "true"
		}
	}
	err = r.Client.Update(context.TODO(), &configInstance)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileConfig) ZookeeperReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Zookeeper Object !!!!!")
	configInstanceList := &v1alpha1.ConfigList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": "config"})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	err := r.Client.List(context.TODO(), listOps, configInstanceList)
	if err != nil {
		return reconcile.Result{}, err
	}
	configInstance := configInstanceList.Items[0]
	zookeeper := &v1alpha1.Zookeeper{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, zookeeper)
	if err != nil {
		return reconcile.Result{}, err
	}
	configInstance.Labels["zookeeperActive"] = "false"
	if zookeeper.Status.Active != nil {
		if *zookeeper.Status.Active {
			configInstance.Labels["zookeeperActive"] = "true"
		}
	}
	err = r.Client.Update(context.TODO(), &configInstance)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileConfig) ReplicaSetReconcile(request reconcile.Request) error {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Config due to ReplicaSet changes")
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": "config"})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	replicaSetList := &appsv1.ReplicaSetList{}
	err := r.Client.List(context.TODO(), listOps, replicaSetList)
	if err != nil {
		return err
	}
	if len(replicaSetList.Items) > 0 {
		replicaSet := &replicaSetList.Items[0]
		var podNameIpMap = make(map[string]string)
		podList := &corev1.PodList{}

		err := r.Client.Get(context.TODO(), request.NamespacedName, replicaSet)
		if err != nil {
			return err
		}
		if podHash, ok := replicaSet.ObjectMeta.Labels["pod-template-hash"]; ok {
			labelSelector := labels.SelectorFromSet(map[string]string{"pod-template-hash": podHash})
			listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
			err := r.Client.List(context.TODO(), listOps, podList)
			if err != nil {
				return err
			}
		}
		if int32(len(podList.Items)) == *replicaSet.Spec.Replicas {
			for _, pod := range podList.Items {
				if pod.Status.PodIP != "" {
					podNameIpMap[pod.Name] = pod.Status.PodIP
				}
			}
		}
		if int32(len(podNameIpMap)) == *replicaSet.Spec.Replicas {
			ready := true
			r.podsReady = &ready
		} else {
			ready := false
			r.podsReady = &ready
		}
	}
	return nil
}

func (r *ReconcileConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	var err error
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Config")
	requestObject := r.GetRequestObject(request)
	if requestObject == nil {
		return reconcile.Result{}, nil
	}

	objectKind := requestObject.GetObjectKind()
	objectGVK := objectKind.GroupVersionKind()
	kind := objectGVK.Kind
	switch kind {
	case "ReplicaSet":
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": "config"})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		replicaSetList := &appsv1.ReplicaSetList{}
		err := r.Client.List(context.TODO(), listOps, replicaSetList)
		if err != nil {
			return reconcile.Result{}, err
		}
		if len(replicaSetList.Items) > 0 {

			err := r.ReplicaSetReconcile(request)
			if err != nil {
				return reconcile.Result{}, err
			}
			replicaSetInstance := &replicaSetList.Items[0]
			err = r.Client.Get(context.TODO(), request.NamespacedName, replicaSetInstance)
			if err != nil {
				return reconcile.Result{}, err
			}
			for _, deploymentOwner := range replicaSetInstance.OwnerReferences {
				if *deploymentOwner.Controller {
					deploymentInstance := &appsv1.Deployment{}
					err = r.Client.Get(context.TODO(), types.NamespacedName{Name: deploymentOwner.Name, Namespace: request.Namespace}, deploymentInstance)
					if err != nil {
						return reconcile.Result{}, err
					}
					for _, configOwner := range deploymentInstance.OwnerReferences {
						if *configOwner.Controller {
							configInstance := &v1alpha1.Config{}
							err = r.Client.Get(context.TODO(), types.NamespacedName{Name: configOwner.Name, Namespace: request.Namespace}, configInstance)
							if err != nil {
								return reconcile.Result{}, err
							}
							request.Name = configOwner.Name
							r.ConfigReconcile(request)
						}
					}
				}
			}
		}
	case "Deployment":
		r.DeploymentReconcile(request)
	default:
		r.ConfigReconcile(request)
	}

	if err != nil {
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}
