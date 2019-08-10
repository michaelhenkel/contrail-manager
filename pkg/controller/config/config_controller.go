package config

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

	// Watch for changes to PODs
	srcPod := &source.Kind{Type: &corev1.Pod{}}
	podHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1.ReplicaSet{},
	}
	predInitStatus := utils.PodInitStatusChange(map[string]string{"contrail_manager": "config"})
	predPodIPChange := utils.PodIPChange(map[string]string{"contrail_manager": "config"})
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
	predManagerSizeChange := utils.ManagerSizeChange(utils.ConfigGroupKind())
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
		OwnerType:    &v1alpha1.Config{},
	}
	deploymentPred := utils.DeploymentStatusChange(utils.ConfigGroupKind())
	err = c.Watch(srcDeployment, deploymentHandler, deploymentPred)
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

func (r *ReconcileConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	var err error
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Config")
	instanceType := "config"
	instance := &v1alpha1.Config{}
	cassandraInstance := &v1alpha1.Cassandra{}
	zookeeperInstance := &v1alpha1.Zookeeper{}
	rabbitmqInstance := &v1alpha1.Rabbitmq{}
	var resourceIdentification v1alpha1.ResourceIdentification = instance
	var resourceObject v1alpha1.ResourceObject = instance
	var resourceConfiguration v1alpha1.ResourceConfiguration = instance
	var resourceStatus v1alpha1.ResourceStatus = instance

	err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil && errors.IsNotFound(err) {
		isReplicaset := resourceIdentification.IsReplicaset(&request, instanceType, r.Client)
		isManager := resourceIdentification.IsManager(&request, r.Client)
		isZookeeper := resourceIdentification.IsZookeeper(&request, r.Client)
		isRabbitmq := resourceIdentification.IsRabbitmq(&request, r.Client)
		isCassandra := resourceIdentification.IsCassandra(&request, r.Client)
		if !isCassandra && !isRabbitmq && !isZookeeper && !isReplicaset && !isManager {
			return reconcile.Result{}, nil
		}
		err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
		if err != nil {
			return reconcile.Result{}, nil
		}
	}

	cassandraActive := cassandraInstance.IsActive(instance.Spec.ServiceConfiguration.CassandraInstance,
		request.Namespace, r.Client)
	zookeeperActive := zookeeperInstance.IsActive(instance.Spec.ServiceConfiguration.ZookeeperInstance,
		request.Namespace, r.Client)
	rabbitmqActive := rabbitmqInstance.IsActive(instance.Labels["contrail_cluster"],
		request.Namespace, r.Client)
	/*
		cassandraActive = utils.IsCassandraActive(instance.Spec.ServiceConfiguration.CassandraInstance,
			request.Namespace, r.Client)
		zookeeperActive = utils.IsZookeeperActive(instance.Spec.ServiceConfiguration.ZookeeperInstance,
			request.Namespace, r.Client)
		rabbitmqActive = utils.IsRabbitmqActive(instance.Labels["contrail_cluster"],
			request.Namespace, r.Client)
	*/
	if !cassandraActive || !rabbitmqActive || !zookeeperActive {
		return reconcile.Result{}, nil
	}

	managerInstance, err := resourceIdentification.OwnedByManager(r.Client, request)
	if err != nil {
		return reconcile.Result{}, err
	}
	if managerInstance != nil {
		if managerInstance.Spec.Services.Config != nil {
			configManagerInstance := managerInstance.Spec.Services.Config
			if configManagerInstance.Name == request.Name {
				instance.Spec.CommonConfiguration = utils.MergeCommonConfiguration(
					managerInstance.Spec.CommonConfiguration,
					configManagerInstance.Spec.CommonConfiguration)
				err = r.Client.Update(context.TODO(), instance)
				if err != nil {
					return reconcile.Result{}, err
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

	intendedDeployment, err := resourceObject.PrepareIntendedDeployment(GetDeployment(),
		&instance.Spec.CommonConfiguration,
		request,
		r.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	resourceObject.AddVolumesToIntendedDeployments(intendedDeployment,
		map[string]string{configMap.Name: request.Name + "-" + instanceType + "-volume"})

	for idx, container := range intendedDeployment.Spec.Template.Spec.Containers {
		if container.Name == "api" {
			command := []string{"bash", "-c",
				"/usr/bin/python /usr/bin/contrail-api --conf_file /etc/mycontrail/api.${POD_IP} --conf_file /etc/contrail/contrail-keystone-auth.conf --worker_id 0"}
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
		if container.Name == "devicemanager" {
			command := []string{"bash", "-c",
				"/usr/bin/python /usr/bin/contrail-device-manager --conf_file /etc/mycontrail/devicemanager.${POD_IP} --conf_file /etc/contrail/contrail-keystone-auth.conf"}
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
		if container.Name == "servicemonitor" {
			command := []string{"bash", "-c",
				"/usr/bin/python /usr/bin/contrail-svc-monitor --conf_file /etc/mycontrail/servicemonitor.${POD_IP} --conf_file /etc/contrail/contrail-keystone-auth.conf"}
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
		if container.Name == "schematransformer" {
			command := []string{"bash", "-c",
				"/usr/bin/python /usr/bin/contrail-schema --conf_file /etc/mycontrail/schematransformer.${POD_IP}  --conf_file /etc/contrail/contrail-keystone-auth.conf"}
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
		if container.Name == "analyticsapi" {
			command := []string{"bash", "-c",
				"/usr/bin/python /usr/bin/contrail-analytics-api -c /etc/mycontrail/analyticsapi.${POD_IP} -c /etc/contrail/contrail-keystone-auth.conf"}
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
		if container.Name == "collector" {
			command := []string{"bash", "-c",
				"/usr/bin/contrail-collector --conf_file /etc/mycontrail/collector.${POD_IP}"}
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
		if container.Name == "redis" {
			command := []string{"bash", "-c",
				"redis-server --lua-time-limit 15000 --dbfilename '' --bind 127.0.0.1 ${POD_IP} --port 6379"}
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
		if container.Name == "nodemanagerconfig" {
			command := []string{"bash", "-c",
				"sed \"s/hostip=.*/hostip=${POD_IP}/g\" /etc/mycontrail/nodemanagerconfig.${POD_IP} > /etc/contrail/contrail-config-nodemgr.conf; /usr/bin/python /usr/bin/contrail-nodemgr --nodetype=contrail-config"}
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
		if container.Name == "nodemanageranalytics" {
			command := []string{"bash", "-c",
				"sed \"s/hostip=.*/hostip=${POD_IP}/g\" /etc/mycontrail/nodemanageranalytics.${POD_IP} > /etc/contrail/contrail-analytics-nodemgr.conf;/usr/bin/python /usr/bin/contrail-nodemgr --nodetype=contrail-analytics"}
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

	// Configure InitContainers
	for idx, container := range intendedDeployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range instance.Spec.ServiceConfiguration.Images {
			if containerName == container.Name {
				(&intendedDeployment.Spec.Template.Spec.InitContainers[idx]).Image = image
			}
		}
	}

	err = resourceConfiguration.CompareIntendedWithCurrentDeployment(intendedDeployment,
		&instance.Spec.CommonConfiguration,
		request,
		r.Scheme,
		r.Client,
		false)
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
