package rabbitmq

import (
	"context"
	"fmt"
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

var log = logf.Log.WithName("controller_rabbitmq")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileRabbitmq{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("rabbitmq-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Rabbitmq
	err = c.Watch(&source.Kind{Type: &v1alpha1.Rabbitmq{}}, &handler.EnqueueRequestForObject{})
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
				if v == "rabbitmq" {
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
				if v == "rabbitmq" {
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
			if oldManager.Spec.Services.Rabbitmq.Size != nil {
				oldSize = *oldManager.Spec.Services.Rabbitmq.Size
			} else {
				oldSize = *oldManager.Spec.Size
			}
			if newManager.Spec.Services.Rabbitmq.Size != nil {
				newSize = *newManager.Spec.Services.Rabbitmq.Size
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
				if owner.Kind == "Rabbitmq" {
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

// blank assignment to verify that ReconcileRabbitmq implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileRabbitmq{}

// ReconcileRabbitmq reconciles a Rabbitmq object
type ReconcileRabbitmq struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client
	Scheme *runtime.Scheme
}

func (r *ReconcileRabbitmq) GetRequestObject(request reconcile.Request) (ro runtime.Object) {
	rabbitmqInstance := &v1alpha1.Rabbitmq{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, rabbitmqInstance)
	if err == nil {
		return rabbitmqInstance
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
func (r *ReconcileRabbitmq) RabbitmqReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Rabbitmq Object")

	rabbitmqInstance := &v1alpha1.Rabbitmq{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, rabbitmqInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Rabbitmq Instance")
		}
	}
	managerInstance := &v1alpha1.Manager{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, managerInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Manager Instance")
		}
	} else {
		rabbitmqInstance.Spec = managerInstance.Spec.Services.Rabbitmq
		if managerInstance.Spec.Services.Rabbitmq.Size != nil {
			rabbitmqInstance.Spec.Size = managerInstance.Spec.Services.Rabbitmq.Size
		} else {
			rabbitmqInstance.Spec.Size = managerInstance.Spec.Size
		}
		if managerInstance.Spec.HostNetwork != nil {
			rabbitmqInstance.Spec.HostNetwork = managerInstance.Spec.HostNetwork
		}
	}
	err = r.Client.Update(context.TODO(), rabbitmqInstance)
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
			Name:      "rabbitmq-" + rabbitmqInstance.Name,
			Namespace: rabbitmqInstance.Namespace,
			Labels:    map[string]string{"contrail_manager": "rabbitmq"},
		},
		Data: rabbitmqInstance.Spec.Configuration,
	}
	controllerutil.SetControllerReference(rabbitmqInstance, &configMap, r.Scheme)
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "rabbitmq-" + rabbitmqInstance.Name, Namespace: rabbitmqInstance.Namespace}, &configMap)
	if err != nil && errors.IsNotFound(err) {
		err = r.Client.Create(context.TODO(), &configMap)
		if err != nil {
			reqLogger.Error(err, "Failed to create ConfigMap", "Namespace", rabbitmqInstance.Namespace, "Name", "rabbitmq-"+rabbitmqInstance.Name)
			return reconcile.Result{}, err
		}
	}

	// Set Deployment Name & Namespace

	deployment.ObjectMeta.Name = "rabbitmq-" + rabbitmqInstance.Name
	deployment.ObjectMeta.Namespace = rabbitmqInstance.Namespace

	// Configure Containers
	for idx, container := range deployment.Spec.Template.Spec.Containers {
		for containerName, image := range rabbitmqInstance.Spec.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.Containers[idx]).Image = image
			}
			if containerName == "rabbitmq" {
				(&deployment.Spec.Template.Spec.Containers[idx]).EnvFrom[0].ConfigMapRef.Name = "rabbitmq-" + rabbitmqInstance.Name
			}
		}
	}

	// Configure InitContainers
	for idx, container := range deployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range rabbitmqInstance.Spec.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.InitContainers[idx]).Image = image
			}
		}
	}

	// Set HostNetwork
	deployment.Spec.Template.Spec.HostNetwork = *rabbitmqInstance.Spec.HostNetwork

	// Set Selector and Label
	deployment.Spec.Selector.MatchLabels["app"] = "rabbitmq-" + rabbitmqInstance.Name
	deployment.Spec.Template.ObjectMeta.Labels["app"] = "rabbitmq-" + rabbitmqInstance.Name

	// Set Size
	deployment.Spec.Replicas = rabbitmqInstance.Spec.Size
	// Create Deployment

	controllerutil.SetControllerReference(rabbitmqInstance, deployment, r.Scheme)
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "rabbitmq-" + rabbitmqInstance.Name, Namespace: rabbitmqInstance.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		err = r.Client.Create(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create Deployment", "Namespace", rabbitmqInstance.Namespace, "Name", "rabbitmq-"+rabbitmqInstance.Name)
			return reconcile.Result{}, err
		}
	} else if err == nil && *deployment.Spec.Replicas != *rabbitmqInstance.Spec.Size {
		deployment.Spec.Replicas = rabbitmqInstance.Spec.Size
		err = r.Client.Update(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Namespace", rabbitmqInstance.Namespace, "Name", "rabbitmq-"+rabbitmqInstance.Name)
			return reconcile.Result{}, err
		}
		active := false
		rabbitmqInstance.Status.Active = &active
		err = r.Client.Status().Update(context.TODO(), rabbitmqInstance)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileRabbitmq) ManagerReconcile(instance *v1alpha1.Manager) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
func (r *ReconcileRabbitmq) DeploymentReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Rabbitmq due to Deployment changes")
	deployment := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, deployment)
	if err != nil {
		return reconcile.Result{}, err
	}
	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
		var ownerName string
		for _, owner := range deployment.ObjectMeta.OwnerReferences {
			if owner.Kind == "Rabbitmq" {
				ownerName = owner.Name
			}
		}
		owner := &v1alpha1.Rabbitmq{}
		err := r.Client.Get(context.TODO(), types.NamespacedName{Name: ownerName, Namespace: request.Namespace}, owner)
		if err != nil {
			return reconcile.Result{}, err
		}
		active := true
		owner.Status.Active = &active
		err = r.Client.Status().Update(context.TODO(), owner)
		if err != nil {
			return reconcile.Result{}, err
		}
		fmt.Println("Ready Replicas: ", deployment.Status.ReadyReplicas)
		fmt.Println("Spec Replicas: ", *deployment.Spec.Replicas)
		reqLogger.Info("Rabbitmq Deployment is ready")

	}
	return reconcile.Result{}, nil
}
func (r *ReconcileRabbitmq) ReplicaSetReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Rabbitmq due to ReplicaSet changes")
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": "rabbitmq"})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	replicaSetList := &appsv1.ReplicaSetList{}
	err := r.Client.List(context.TODO(), listOps, replicaSetList)
	if err != nil {
		return reconcile.Result{}, err
	}
	if len(replicaSetList.Items) > 0 {
		replicaSet := &replicaSetList.Items[0]
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
			configMap.Data["RABBITMQ_NODES"] = nodeList
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
			rabbitmqList := &v1alpha1.RabbitmqList{}
			rabbitmqListOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
			err = r.Client.List(context.TODO(), rabbitmqListOps, rabbitmqList)
			if err != nil {
				return reconcile.Result{}, err
			}
			rabbitmq := rabbitmqList.Items[0]
			rabbitmq.Status.Nodes = podNameIpMap
			portMap := map[string]string{"port": rabbitmq.Spec.Configuration["RABBITMQ_NODE_PORT"]}
			rabbitmq.Status.Ports = portMap
			err = r.Client.Status().Update(context.TODO(), &rabbitmq)
			if err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("All POD IPs available CASSANDRA: " + replicaSet.ObjectMeta.Labels["contrail_manager"])

		}
	}
	return reconcile.Result{}, nil
}
func (r *ReconcileRabbitmq) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Rabbitmq")
	requestObject := r.GetRequestObject(request)
	if requestObject == nil {
		return reconcile.Result{}, nil
	}

	objectKind := requestObject.GetObjectKind()
	objectGVK := objectKind.GroupVersionKind()
	kind := objectGVK.Kind
	switch kind {
	case "Rabbitmq":
		r.RabbitmqReconcile(request)
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
