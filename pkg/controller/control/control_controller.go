package control

import (
	"context"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_control")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Control Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileControl{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("control-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Control
	err = c.Watch(&source.Kind{Type: &v1alpha1.Control{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to PODs
	srcPod := &source.Kind{Type: &corev1.Pod{}}
	podHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1.ReplicaSet{},
	}
	predInitStatus := utils.PodInitStatusChange(map[string]string{"contrail_manager": "control"})
	predPodIPChange := utils.PodIPChange(map[string]string{"contrail_manager": "control"})
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
	predManagerSizeChange := utils.ManagerSizeChange(utils.ControlGroupKind())
	// Watch for Manager events.
	err = c.Watch(srcManager, managerHandler, predManagerSizeChange)
	if err != nil {
		return err
	}

	// Watch for changes to Manager
	srcCassandra := &source.Kind{Type: &v1alpha1.Cassandra{}}
	cassandraHandler := &handler.EnqueueRequestForObject{}
	predCassandraSizeChange := utils.CassandraActiveChange()
	// Watch for Manager events.
	err = c.Watch(srcCassandra, cassandraHandler, predCassandraSizeChange)
	if err != nil {
		return err
	}

	// Watch for changes to Manager
	srcConfig := &source.Kind{Type: &v1alpha1.Config{}}
	configHandler := &handler.EnqueueRequestForObject{}
	predConfigSizeChange := utils.ConfigActiveChange()
	// Watch for Manager events.
	err = c.Watch(srcConfig, configHandler, predConfigSizeChange)
	if err != nil {
		return err
	}

	// Watch for changes to Manager
	srcRabbitmq := &source.Kind{Type: &v1alpha1.Rabbitmq{}}
	rabbitmqHandler := &handler.EnqueueRequestForObject{}
	predRabbitmqSizeChange := utils.RabbitmqActiveChange()
	// Watch for Manager events.
	err = c.Watch(srcRabbitmq, rabbitmqHandler, predRabbitmqSizeChange)
	if err != nil {
		return err
	}

	// Watch for changes to Manager
	srcZookeeper := &source.Kind{Type: &v1alpha1.Zookeeper{}}
	zookeeperHandler := &handler.EnqueueRequestForObject{}
	predZookeeperSizeChange := utils.ZookeeperActiveChange()
	// Watch for Manager events.
	err = c.Watch(srcZookeeper, zookeeperHandler, predZookeeperSizeChange)
	if err != nil {
		return err
	}

	srcDeployment := &source.Kind{Type: &appsv1.Deployment{}}
	deploymentHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Control{},
	}
	deploymentPred := utils.DeploymentStatusChange(utils.ControlGroupKind())
	err = c.Watch(srcDeployment, deploymentHandler, deploymentPred)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileControl implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileControl{}

// ReconcileControl reconciles a Control object
type ReconcileControl struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Control object and makes changes based on the state read
// and what is in the Control.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileControl) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	var err error
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Control")
	instanceType := "control"
	instance := &v1alpha1.Control{}
	var i v1alpha1.Instance = instance
	err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil && errors.IsNotFound(err) {
		isReplicaset := i.IsReplicaset(&request, instanceType, r.Client)
		isManager := i.IsManager(&request, r.Client)
		isZookeeper := i.IsZookeeper(&request, r.Client)
		isRabbitmq := i.IsRabbitmq(&request, r.Client)
		isCassandra := i.IsCassandra(&request, r.Client)
		isConfig := i.IsConfig(&request, r.Client)
		if !isConfig && !isCassandra && !isRabbitmq && !isZookeeper && !isReplicaset && !isManager {
			return reconcile.Result{}, nil
		}
		err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
		if err != nil {
			return reconcile.Result{}, nil
		}
	}
	cassandraActive := false
	zookeeperActive := false
	rabbitmqActive := false
	configActive := false
	cassandraActive = utils.IsCassandraActive(instance.Spec.ServiceConfiguration.CassandraInstance,
		request.Namespace, r.Client)
	zookeeperActive = utils.IsZookeeperActive(instance.Spec.ServiceConfiguration.ZookeeperInstance,
		request.Namespace, r.Client)
	rabbitmqActive = utils.IsRabbitmqActive(instance.Labels["contrail_cluster"],
		request.Namespace, r.Client)
	configActive = utils.IsConfigActive(instance.Labels["contrail_cluster"],
		request.Namespace, r.Client)
	if !configActive || !cassandraActive || !rabbitmqActive || !zookeeperActive {
		return reconcile.Result{}, nil
	}

	managerInstance, err := i.OwnedByManager(r.Client, request)
	if err != nil {
		return reconcile.Result{}, err
	}
	if managerInstance != nil {
		if managerInstance.Spec.Services.Controls != nil {
			for _, controlManagerInstance := range managerInstance.Spec.Services.Controls {
				if controlManagerInstance.Name == request.Name {
					instance.Spec.CommonConfiguration = utils.MergeCommonConfiguration(
						managerInstance.Spec.CommonConfiguration,
						controlManagerInstance.Spec.CommonConfiguration)
					err = r.Client.Update(context.TODO(), instance)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		}
	}
	configMap, err := i.CreateConfigMap(request.Name+"-"+instanceType+"-configmap",
		r.Client,
		r.Scheme,
		request)
	if err != nil {
		return reconcile.Result{}, err
	}

	intendedDeployment, err := i.PrepareIntendedDeployment(GetDeployment(),
		&instance.Spec.CommonConfiguration,
		request,
		r.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	i.AddVolumesToIntendedDeployments(intendedDeployment,
		map[string]string{configMap.Name: request.Name + "-" + instanceType + "-volume"})

	for idx, container := range intendedDeployment.Spec.Template.Spec.Containers {
		if container.Name == "control" {
			command := []string{"bash", "-c",
				"/usr/bin/contrail-control --conf_file /etc/mycontrail/control.${POD_IP}"}
			//command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Command = command

			volumeMountList := []corev1.VolumeMount{}
			if len((&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts) > 0 {
				volumeMountList = (&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts
			}
			volumeMount := corev1.VolumeMount{
				Name:      request.Name + "-" + instanceType + "-volume",
				MountPath: "/etc/mycontrail",
			}
			volumeMountList = append(volumeMountList, volumeMount)
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Image = instance.Spec.ServiceConfiguration.Images[container.Name]
		}
		if container.Name == "named" {
			command := []string{"bash", "-c",
				"/usr/bin/contrail-named -f -g -u contrail -c /etc/mycontrail/named.${POD_IP}"}
			//command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Command = command

			volumeMountList := []corev1.VolumeMount{}
			if len((&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts) > 0 {
				volumeMountList = (&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts
			}
			volumeMount := corev1.VolumeMount{
				Name:      request.Name + "-" + instanceType + "-volume",
				MountPath: "/etc/mycontrail",
			}
			volumeMountList = append(volumeMountList, volumeMount)
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Image = instance.Spec.ServiceConfiguration.Images[container.Name]
		}
		if container.Name == "dns" {
			command := []string{"bash", "-c",
				"/usr/bin/contrail-dns --conf_file /etc/mycontrail/dns.${POD_IP}"}
			//command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Command = command

			volumeMountList := []corev1.VolumeMount{}
			if len((&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts) > 0 {
				volumeMountList = (&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts
			}
			volumeMount := corev1.VolumeMount{
				Name:      request.Name + "-" + instanceType + "-volume",
				MountPath: "/etc/mycontrail",
			}
			volumeMountList = append(volumeMountList, volumeMount)
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Image = instance.Spec.ServiceConfiguration.Images[container.Name]
		}
		if container.Name == "nodemanager" {
			command := []string{"bash", "-c",
				"sed \"s/hostip=.*/hostip=${POD_IP}/g\" /etc/mycontrail/nodemanager.${POD_IP} > /etc/contrail/contrail-control-nodemgr.conf; /usr/bin/python /usr/bin/contrail-nodemgr --nodetype=contrail-control"}
			//command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Command = command

			volumeMountList := []corev1.VolumeMount{}
			if len((&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts) > 0 {
				volumeMountList = (&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts
			}
			volumeMount := corev1.VolumeMount{
				Name:      request.Name + "-" + instanceType + "-volume",
				MountPath: "/etc/mycontrail",
			}
			volumeMountList = append(volumeMountList, volumeMount)
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Image = instance.Spec.ServiceConfiguration.Images[container.Name]
		}
	}

	for idx, container := range intendedDeployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range instance.Spec.ServiceConfiguration.Images {
			if containerName == container.Name {
				(&intendedDeployment.Spec.Template.Spec.InitContainers[idx]).Image = image
			}
		}
	}

	err = i.CompareIntendedWithCurrentDeployment(intendedDeployment,
		&instance.Spec.CommonConfiguration,
		request,
		r.Scheme,
		r.Client,
		false)
	if err != nil {
		return reconcile.Result{}, err
	}

	podIPList, podIPMap, err := i.GetPodIPListAndIPMap(request, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}
	if len(podIPList.Items) > 0 {
		err = i.CreateInstanceConfiguration(request,
			podIPList,
			r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = i.SetPodsToReady(podIPList, r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = i.ManageNodeStatus(podIPMap, r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	err = i.SetInstanceActive(r.Client, &instance.Status, intendedDeployment, request)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
