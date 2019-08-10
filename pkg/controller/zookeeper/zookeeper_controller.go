package zookeeper

import (
	"context"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_zookeeper")
var err error

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileZookeeper{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller

	c, err := controller.New("zookeeper-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	// Watch for changes to primary resource Zookeeper
	err = c.Watch(&source.Kind{Type: &v1alpha1.Zookeeper{}},
		&handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to PODs
	srcPod := &source.Kind{Type: &corev1.Pod{}}
	podHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1.ReplicaSet{},
	}
	predInitStatus := utils.PodInitStatusChange(map[string]string{"contrail_manager": "zookeeper"})
	predPodIPChange := utils.PodIPChange(map[string]string{"contrail_manager": "zookeeper"})
	err = c.Watch(srcPod, podHandler, predPodIPChange)
	if err != nil {
		return err
	}
	err = c.Watch(srcPod, podHandler, predInitStatus)
	if err != nil {
		return err
	}

	// Watch for changes to Manager
	srcManager := &source.Kind{Type: &v1alpha1.Manager{}}
	managerHandler := &handler.EnqueueRequestForObject{}
	predManagerSizeChange := utils.ManagerSizeChange(utils.ZookeeperGroupKind())
	// Watch for Manager events.
	err = c.Watch(srcManager, managerHandler, predManagerSizeChange)
	if err != nil {
		return err
	}

	srcDeployment := &source.Kind{Type: &appsv1.Deployment{}}
	deploymentHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Zookeeper{},
	}
	deploymentPred := utils.DeploymentStatusChange(utils.ZookeeperGroupKind())
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

// Reconcile reconciles zookeeper
func (r *ReconcileZookeeper) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Zookeeper")
	instanceType := "zookeeper"

	// A reconcile.Request for Zookeeper can be triggered by 4 different types:
	// 1. Any changes on the Zookeeper instance
	// --> reconcile.Request is Zookeeper instance name/namespace
	// 2. IP Status change on the Pods
	// --> reconcile.Request is Replicaset name/namespace
	// --> we need to evaluate the label to get the Zookeeper instance
	// 3. Status change on the Deployment
	// --> reconcile.Request is Zookeeper instance name/namespace
	// 4. Zookeepers changes on the Manager instance
	// --> reconcile.Request is Manager instance name/namespace
	instance := &v1alpha1.Zookeeper{}
	var resourceIdentification v1alpha1.ResourceIdentification = instance
	var resourceObject v1alpha1.ResourceObject = instance
	var resourceConfiguration v1alpha1.ResourceConfiguration = instance
	var resourceStatus v1alpha1.ResourceStatus = instance
	err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
	// if not found we expect it a change in replicaset
	// and get the zookeeper instance via label
	if err != nil && errors.IsNotFound(err) {
		replicaSet := &appsv1.ReplicaSet{}
		err = r.Client.Get(context.TODO(), request.NamespacedName, replicaSet)
		if err != nil {
			return reconcile.Result{}, nil
		}
		request.Name = replicaSet.Labels[instanceType]
		err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
		if err != nil {
			return reconcile.Result{}, nil
		}
	}

	managerInstance, err := resourceIdentification.OwnedByManager(r.Client, request)
	if err != nil {
		return reconcile.Result{}, err
	}

	if managerInstance != nil {
		if managerInstance.Spec.Services.Zookeepers != nil {
			for _, zookeeperManagerInstance := range managerInstance.Spec.Services.Zookeepers {
				if zookeeperManagerInstance.Name == request.Name {
					instance.Spec.CommonConfiguration = utils.MergeCommonConfiguration(
						managerInstance.Spec.CommonConfiguration,
						zookeeperManagerInstance.Spec.CommonConfiguration)
					err = r.Client.Update(context.TODO(), instance)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		}
	}

	configMap, err := resourceObject.CreateConfigMap(request.Name+"-"+instanceType+"-configmap",
		r.Client,
		r.Scheme,
		request)
	if err != nil {
		return reconcile.Result{}, err
	}
	configMap2, err := resourceObject.CreateConfigMap(request.Name+"-"+instanceType+"-configmap-1",
		r.Client,
		r.Scheme,
		request)
	if err != nil {
		return reconcile.Result{}, err
	}

	intendedDeployment, err := resourceObject.PrepareIntendedDeployment(GetDeployment(),
		&instance.Spec.CommonConfiguration,
		request,
		r.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	resourceObject.AddVolumesToIntendedDeployments(intendedDeployment,
		map[string]string{configMap.Name: request.Name + "-" + instanceType + "-volume",
			configMap2.Name: request.Name + "-" + instanceType + "-volume-1"})

	var revisionHistoryLimit int32
	intendedDeployment.Spec.RevisionHistoryLimit = &revisionHistoryLimit

	for idx, container := range intendedDeployment.Spec.Template.Spec.Containers {
		for containerName, image := range instance.Spec.ServiceConfiguration.Images {
			if containerName == container.Name {
				(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Image = image
			}
			if containerName == "zookeeper" {

				command := []string{"bash", "-c", "myid=$(cat /mydata/${POD_IP}) && echo ${myid} > /data/myid && cp /conf-1/* /conf/ && sed -i \"s/clientPortAddress=.*/clientPortAddress=${POD_IP}/g\" /conf/zoo.cfg && zkServer.sh --config /conf start-foreground"}
				//command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
				(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Command = command

				volumeMountList := []corev1.VolumeMount{}
				volumeMount := corev1.VolumeMount{
					Name:      request.Name + "-" + instanceType + "-volume-1",
					MountPath: "/conf-1",
				}
				volumeMountList = append(volumeMountList, volumeMount)

				volumeMount = corev1.VolumeMount{
					Name:      request.Name + "-" + instanceType + "-volume",
					MountPath: "/mydata",
				}
				volumeMountList = append(volumeMountList, volumeMount)
				(&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList
			}
		}
	}

	// Configure InitContainers
	for idx, container := range intendedDeployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range instance.Spec.ServiceConfiguration.Images {
			if containerName == container.Name {
				(&intendedDeployment.Spec.Template.Spec.InitContainers[idx]).Image = image
			}
		}
	}

	increaseVersion := false
	currentDeployment := &appsv1.Deployment{}
	err = r.Client.Get(context.TODO(),
		types.NamespacedName{Name: intendedDeployment.Name, Namespace: request.Namespace},
		currentDeployment)
	if err == nil {
		if *currentDeployment.Spec.Replicas == 1 {
			increaseVersion = true
		}
	}

	err = resourceConfiguration.CompareIntendedWithCurrentDeployment(intendedDeployment,
		&instance.Spec.CommonConfiguration,
		request,
		r.Scheme,
		r.Client,
		increaseVersion)
	if err != nil {
		return reconcile.Result{}, err
	}

	podIPList, podIPMap, err := resourceConfiguration.PodIPListAndIPMap(request, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}
	if len(podIPList.Items) > 0 {
		err = resourceConfiguration.InstanceConfiguration(request,
			podIPList,
			r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = resourceStatus.SetPodsToReady(podIPList, r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = resourceStatus.ManageNodeStatus(podIPMap, r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	err = resourceStatus.SetInstanceActive(r.Client, &instance.Status, intendedDeployment, request)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
