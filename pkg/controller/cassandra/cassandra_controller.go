package cassandra

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

var log = logf.Log.WithName("controller_cassandra")

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCassandra{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("cassandra-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Cassandra
	err = c.Watch(&source.Kind{Type: &v1alpha1.Cassandra{}}, &handler.EnqueueRequestForObject{})
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
				if v == "cassandra" {
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
				if v == "cassandra" {
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
			if oldManager.Spec.Services.Cassandra.Size != nil {
				oldSize = *oldManager.Spec.Services.Cassandra.Size
			} else {
				oldSize = *oldManager.Spec.Size
			}
			if newManager.Spec.Services.Cassandra.Size != nil {
				newSize = *newManager.Spec.Services.Cassandra.Size
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
				if owner.Kind == "Cassandra" {
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

// blank assignment to verify that ReconcileCassandra implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCassandra{}

// ReconcileCassandra reconciles a Cassandra object
type ReconcileCassandra struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client
	Scheme *runtime.Scheme
}

func (r *ReconcileCassandra) GetRequestObject(request reconcile.Request) (ro runtime.Object) {
	cassandraInstance := &v1alpha1.Cassandra{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, cassandraInstance)
	if err == nil {
		return cassandraInstance
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
func (r *ReconcileCassandra) CassandraReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Cassandra Object")

	cassandraInstance := &v1alpha1.Cassandra{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, cassandraInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Cassandra Instance")
		}
	}
	managerInstance := &v1alpha1.Manager{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, managerInstance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Manager Instance")
		}
	} else {
		cassandraInstance.Spec = managerInstance.Spec.Services.Cassandra
		if managerInstance.Spec.Services.Cassandra.Size != nil {
			cassandraInstance.Spec.Size = managerInstance.Spec.Services.Cassandra.Size
		} else {
			cassandraInstance.Spec.Size = managerInstance.Spec.Size
		}
		if managerInstance.Spec.HostNetwork != nil {
			cassandraInstance.Spec.HostNetwork = managerInstance.Spec.HostNetwork
		}
	}
	err = r.Client.Update(context.TODO(), cassandraInstance)
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
			Name:      "cassandra-" + cassandraInstance.Name,
			Namespace: cassandraInstance.Namespace,
			Labels:    map[string]string{"contrail_manager": "cassandra"},
		},
		Data: cassandraInstance.Spec.Configuration,
	}
	controllerutil.SetControllerReference(cassandraInstance, &configMap, r.Scheme)
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "cassandra-" + cassandraInstance.Name, Namespace: cassandraInstance.Namespace}, &configMap)
	if err != nil && errors.IsNotFound(err) {
		err = r.Client.Create(context.TODO(), &configMap)
		if err != nil {
			reqLogger.Error(err, "Failed to create ConfigMap", "Namespace", cassandraInstance.Namespace, "Name", "cassandra-"+cassandraInstance.Name)
			return reconcile.Result{}, err
		}
	}

	// Set Deployment Name & Namespace

	deployment.ObjectMeta.Name = "cassandra-" + cassandraInstance.Name
	deployment.ObjectMeta.Namespace = cassandraInstance.Namespace

	// Configure Containers
	for idx, container := range deployment.Spec.Template.Spec.Containers {
		for containerName, image := range cassandraInstance.Spec.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.Containers[idx]).Image = image
			}
			if containerName == "cassandra" {
				(&deployment.Spec.Template.Spec.Containers[idx]).EnvFrom[0].ConfigMapRef.Name = "cassandra-" + cassandraInstance.Name
			}
		}
	}

	// Configure InitContainers
	for idx, container := range deployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range cassandraInstance.Spec.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.InitContainers[idx]).Image = image
			}
		}
	}

	// Set HostNetwork
	deployment.Spec.Template.Spec.HostNetwork = *cassandraInstance.Spec.HostNetwork

	// Set Selector and Label
	deployment.Spec.Selector.MatchLabels["app"] = "cassandra-" + cassandraInstance.Name
	deployment.Spec.Template.ObjectMeta.Labels["app"] = "cassandra-" + cassandraInstance.Name

	// Set Size
	deployment.Spec.Replicas = cassandraInstance.Spec.Size
	// Create Deployment

	controllerutil.SetControllerReference(cassandraInstance, deployment, r.Scheme)
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "cassandra-" + cassandraInstance.Name, Namespace: cassandraInstance.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		err = r.Client.Create(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create Deployment", "Namespace", cassandraInstance.Namespace, "Name", "cassandra-"+cassandraInstance.Name)
			return reconcile.Result{}, err
		}
	} else if err == nil && *deployment.Spec.Replicas != *cassandraInstance.Spec.Size {
		deployment.Spec.Replicas = cassandraInstance.Spec.Size
		err = r.Client.Update(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Namespace", cassandraInstance.Namespace, "Name", "cassandra-"+cassandraInstance.Name)
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileCassandra) ManagerReconcile(instance *v1alpha1.Manager) (reconcile.Result, error) {
	return reconcile.Result{}, nil
}
func (r *ReconcileCassandra) DeploymentReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Cassandra due to Deployment changes")
	deployment := &appsv1.Deployment{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, deployment)
	if err != nil {
		return reconcile.Result{}, err
	}
	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
		var ownerName string
		for _, owner := range deployment.ObjectMeta.OwnerReferences {
			if owner.Kind == "Cassandra" {
				ownerName = owner.Name
			}
		}
		owner := &v1alpha1.Cassandra{}
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
		reqLogger.Info("Cassandra Deployment is ready")

	}
	return reconcile.Result{}, nil
}
func (r *ReconcileCassandra) ReplicaSetReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Cassandra due to ReplicaSet changes")
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
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": "cassandra"})
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
		configMap.Data["CASSANDRA_SEEDS"] = nodeList
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
		cassandraList := &v1alpha1.CassandraList{}
		cassandraListOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		err = r.Client.List(context.TODO(), cassandraListOps, cassandraList)
		if err != nil {
			return reconcile.Result{}, err
		}
		cassandra := cassandraList.Items[0]
		cassandra.Status.Nodes = podNameIpMap
		portMap := map[string]string{"port": cassandra.Spec.Configuration["CASSANDRA_PORT"],
			"cqlPort": cassandra.Spec.Configuration["CASSANDRA_CQL_PORT"]}
		cassandra.Status.Ports = portMap
		err = r.Client.Status().Update(context.TODO(), &cassandra)
		if err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("All POD IPs available")

	}
	return reconcile.Result{}, nil
}
func (r *ReconcileCassandra) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Cassandra")
	requestObject := r.GetRequestObject(request)
	if requestObject == nil {
		return reconcile.Result{}, nil
	}

	objectKind := requestObject.GetObjectKind()
	objectGVK := objectKind.GroupVersionKind()
	kind := objectGVK.Kind
	switch kind {
	case "Cassandra":
		r.CassandraReconcile(request)
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
