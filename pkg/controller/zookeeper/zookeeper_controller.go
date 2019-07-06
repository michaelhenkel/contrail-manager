package zookeeper

import (
	"context"
	"strconv"
	"strings"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_zookeeper")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileZookeeper{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("zookeeper-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Zookeeper
	err = c.Watch(&source.Kind{Type: &v1alpha1.Zookeeper{}}, &handler.EnqueueRequestForObject{})
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
			labels := e.MetaOld.GetLabels()
			if v, ok := labels["contrail_manager"]; ok {
				if v == "zookeeper" {
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

	predInitStatus := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			labels := e.MetaOld.GetLabels()
			if v, ok := labels["contrail_manager"]; ok {
				if v == "zookeeper" {
					oldPod := e.ObjectOld.(*corev1.Pod)
					newPod := e.ObjectNew.(*corev1.Pod)
					newPodReady := true
					if newPod.Status.InitContainerStatuses == nil {
						return true
					}
					if oldPod.Status.InitContainerStatuses == nil {
						return true
					}
					for _, initContainerStatus := range newPod.Status.InitContainerStatuses {
						if initContainerStatus.Name == "init" {
							if !initContainerStatus.Ready {
								newPodReady = false
							}
						}
					}
					oldPodReady := true
					for _, initContainerStatus := range oldPod.Status.InitContainerStatuses {
						if initContainerStatus.Name == "init" {
							if !initContainerStatus.Ready {
								oldPodReady = false
							}
						}
					}
					if !newPodReady || !oldPodReady {
						return true
					}
				}
			}
			return false
		},
	}
	// Watch for Pod events.
	err = c.Watch(srcPod, podHandler, predPodIp, predInitStatus)
	if err != nil {
		return err
	}

	srcManager := &source.Kind{Type: &v1alpha1.Manager{}}

	managerHandler := &handler.EnqueueRequestForObject{}
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {

			oldManager := e.ObjectOld.(*v1alpha1.Manager)
			newManager := e.ObjectNew.(*v1alpha1.Manager)
			var oldSize, newSize int32
			if oldManager.Spec.Services.Zookeeper.Size != nil {
				oldSize = *oldManager.Spec.Services.Zookeeper.Size
			} else {
				oldSize = *oldManager.Spec.Size
			}
			if newManager.Spec.Services.Zookeeper.Size != nil {
				newSize = *newManager.Spec.Services.Zookeeper.Size
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
	srcDeployment := &source.Kind{Type: &appsv1.Deployment{}}

	deploymentHandler := &handler.EnqueueRequestForObject{}
	deploymentPred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldDeployment := e.ObjectOld.(*appsv1.Deployment)
			newDeployment := e.ObjectNew.(*appsv1.Deployment)
			isOwner := false
			for _, owner := range newDeployment.ObjectMeta.OwnerReferences {
				if owner.Kind == "Zookeeper" {
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
	err = c.Watch(srcDeployment, deploymentHandler, deploymentPred)
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileZookeeper implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileZookeeper{}

// ReconcileZookeeper reconciles a Zookeeper object
type ReconcileZookeeper struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client
	Scheme *runtime.Scheme
}

func (r *ReconcileZookeeper) GetRequestObject(request reconcile.Request) (ro runtime.Object) {
	zookeeperInstance := &v1alpha1.Zookeeper{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, zookeeperInstance)
	if err == nil {
		return zookeeperInstance
	}

	managerInstance := &v1alpha1.Manager{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, managerInstance)
	if err == nil {
		return managerInstance
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

func (r *ReconcileZookeeper) ZookeeperReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Zookeeper Object")

	zookeeperInstance := &v1alpha1.Zookeeper{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, zookeeperInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Zookeeper Instance")
		}
	}
	managerInstance := &v1alpha1.Manager{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, managerInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Manager Instance")
		}
	} else {
		zookeeperInstance.Spec = managerInstance.Spec.Services.Zookeeper
		if managerInstance.Spec.Services.Zookeeper.Size != nil {
			zookeeperInstance.Spec.Size = managerInstance.Spec.Services.Zookeeper.Size
		} else {
			zookeeperInstance.Spec.Size = managerInstance.Spec.Size
		}
		if managerInstance.Spec.HostNetwork != nil {
			zookeeperInstance.Spec.HostNetwork = managerInstance.Spec.HostNetwork
		}
	}
	err = r.Client.Update(context.TODO(), zookeeperInstance)
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
			Name:      "zookeeper-" + zookeeperInstance.Name,
			Namespace: zookeeperInstance.Namespace,
			Labels:    map[string]string{"contrail_manager": "zookeeper"},
		},
		Data: zookeeperInstance.Spec.Configuration,
	}
	controllerutil.SetControllerReference(zookeeperInstance, &configMap, r.Scheme)
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "zookeeper-" + zookeeperInstance.Name, Namespace: zookeeperInstance.Namespace}, &configMap)
	if err != nil && errors.IsNotFound(err) {
		err = r.Client.Create(context.TODO(), &configMap)
		if err != nil {
			reqLogger.Error(err, "Failed to create ConfigMap", "Namespace", zookeeperInstance.Namespace, "Name", "zookeeper-"+zookeeperInstance.Name)
			return reconcile.Result{}, err
		}
	}

	// Set Deployment Name & Namespace

	deployment.ObjectMeta.Name = "zookeeper-" + zookeeperInstance.Name
	deployment.ObjectMeta.Namespace = zookeeperInstance.Namespace

	// Configure Containers
	for idx, container := range deployment.Spec.Template.Spec.Containers {
		for containerName, image := range zookeeperInstance.Spec.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.Containers[idx]).Image = image
			}
			if containerName == "zookeeper" {
				(&deployment.Spec.Template.Spec.Containers[idx]).EnvFrom[0].ConfigMapRef.Name = "zookeeper-" + zookeeperInstance.Name
			}
		}
	}

	// Configure InitContainers
	for idx, container := range deployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range zookeeperInstance.Spec.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.InitContainers[idx]).Image = image
			}
		}
	}

	// Set HostNetwork
	deployment.Spec.Template.Spec.HostNetwork = *zookeeperInstance.Spec.HostNetwork

	// Set Selector and Label
	deployment.Spec.Selector.MatchLabels["app"] = "zookeeper-" + zookeeperInstance.Name
	deployment.Spec.Template.ObjectMeta.Labels["app"] = "zookeeper-" + zookeeperInstance.Name

	// Set Size
	deployment.Spec.Replicas = zookeeperInstance.Spec.Size
	// Create Deployment

	controllerutil.SetControllerReference(zookeeperInstance, deployment, r.Scheme)
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "zookeeper-" + zookeeperInstance.Name, Namespace: zookeeperInstance.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		deployment.Spec.Template.ObjectMeta.Labels["version"] = "1"
		err = r.Client.Create(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create Deployment", "Namespace", zookeeperInstance.Namespace, "Name", "zookeeper-"+zookeeperInstance.Name)
			return reconcile.Result{}, err
		}
	} else if err == nil && *deployment.Spec.Replicas != *zookeeperInstance.Spec.Size {
		deployment.Spec.Replicas = zookeeperInstance.Spec.Size
		versionInt, _ := strconv.Atoi(deployment.Spec.Template.ObjectMeta.Labels["version"])
		newVersion := versionInt + 1
		newVersionString := strconv.Itoa(newVersion)
		deployment.Spec.Template.ObjectMeta.Labels["version"] = newVersionString
		err = r.Client.Update(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Namespace", zookeeperInstance.Namespace, "Name", "zookeeper-"+zookeeperInstance.Name)
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileZookeeper) ManagerReconcile(instance *v1alpha1.Manager) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}

func (r *ReconcileZookeeper) DeploymentReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Zookeeper due to Deployment changes")
	deployment := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, deployment)
	if err != nil {
		return reconcile.Result{}, err
	}
	var ownerName string
	for _, owner := range deployment.ObjectMeta.OwnerReferences {
		if owner.Kind == "Zookeeper" {
			ownerName = owner.Name
		}
	}
	owner := &v1alpha1.Zookeeper{}
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
		reqLogger.Info("Zookeeper Deployment is ready")
	} else {
		owner.Status.Active = nil
		err = r.Client.Status().Update(context.TODO(), owner)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileZookeeper) ReplicaSetReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Zookeeper due to ReplicaSet changes")
	replicaSet := &appsv1.ReplicaSet{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, replicaSet)
	if err != nil {
		return reconcile.Result{}, err
	}
	podList := &corev1.PodList{}
	if podHash, ok := replicaSet.ObjectMeta.Labels["pod-template-hash"]; ok {
		labelSelector := labels.SelectorFromSet(map[string]string{"pod-template-hash": podHash})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		err := r.Client.List(context.TODO(), listOps, podList)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	var podNameIpMap = make(map[string]string)
	if int32(len(podList.Items)) == *replicaSet.Spec.Replicas {
		for _, pod := range podList.Items {
			if pod.Status.PodIP != "" {
				podNameIpMap[pod.Name] = pod.Status.PodIP
			}
		}
	}

	if int32(len(podNameIpMap)) == *replicaSet.Spec.Replicas {

		configMapList := &corev1.ConfigMapList{}
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": "zookeeper"})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		err := r.Client.List(context.TODO(), listOps, configMapList)
		if err != nil {
			return reconcile.Result{}, err
		}
		configMap := configMapList.Items[0]
		var podIpList []string
		for _, ip := range podNameIpMap {
			podIpList = append(podIpList, ip)
		}
		nodeList := strings.Join(podIpList, ",")
		configMap.Data["ZOOKEEPER_NODES"] = nodeList
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
		zookeeperList := &v1alpha1.ZookeeperList{}
		zookeeperListOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		err = r.Client.List(context.TODO(), zookeeperListOps, zookeeperList)
		if err != nil {
			return reconcile.Result{}, err
		}
		zookeeper := zookeeperList.Items[0]
		zookeeper.Status.Nodes = podNameIpMap

		portMap := map[string]string{"port": zookeeper.Spec.Configuration["ZOOKEEPER_PORT"]}
		zookeeper.Status.Ports = portMap
		err = r.Client.Status().Update(context.TODO(), &zookeeper)
		if err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("All POD IPs available")

	}
	return reconcile.Result{}, nil
}
func (r *ReconcileZookeeper) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Zookeeper")
	requestObject := r.GetRequestObject(request)
	if requestObject == nil {
		return reconcile.Result{}, nil
	}

	objectKind := requestObject.GetObjectKind()
	objectGVK := objectKind.GroupVersionKind()
	kind := objectGVK.Kind
	switch kind {
	case "Zookeeper":
		r.ZookeeperReconcile(request)
	case "Manager":
		instance := requestObject.(*v1alpha1.Manager)
		r.ManagerReconcile(instance)
	case "ReplicaSet":
		r.ReplicaSetReconcile(request)
	case "Deployment":
		r.DeploymentReconcile(request)
	default:
		return reconcile.Result{}, nil
	}
	return reconcile.Result{}, nil
}
