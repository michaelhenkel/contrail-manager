package cassandra

import (
	"context"
	"sort"
	"strings"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/enqueue"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_cassandra")
var err error

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
	err = c.Watch(&source.Kind{Type: &v1alpha1.Cassandra{}}, &enqueue.RequestForObjectGroupKind{NewGroupKind: utils.CassandraGroupKind()})
	if err != nil {
		return err
	}

	// Watch for changes to PODs
	srcPod := &source.Kind{Type: &corev1.Pod{}}
	podHandler := &enqueue.RequestForOwnerGroupKind{
		IsController: true,
		OwnerType:    &appsv1.ReplicaSet{},
		NewGroupKind: utils.ReplicaSetGroupKind(),
	}
	predInitStatus := utils.PodInitStatusChange(map[string]string{"contrail_manager": "cassandra"})
	predPodIPChange := utils.PodIPChange(map[string]string{"contrail_manager": "cassandra"})
	err = c.Watch(srcPod, podHandler, predPodIPChange, predInitStatus)
	if err != nil {
		return err
	}

	// Watch for changes to Manager
	srcManager := &source.Kind{Type: &v1alpha1.Manager{}}
	managerHandler := &enqueue.RequestForObjectGroupKind{NewGroupKind: utils.ManagerGroupKind()}
	predManagerSizeChange := utils.ManagerSizeChange(utils.CassandraGroupKind())
	// Watch for Manager events.
	err = c.Watch(srcManager, managerHandler, predManagerSizeChange)
	if err != nil {
		return err
	}

	srcDeployment := &source.Kind{Type: &appsv1.Deployment{}}
	deploymentHandler := &enqueue.RequestForObjectGroupKind{NewGroupKind: utils.DeploymentGroupKind()}
	deploymentPred := utils.DeploymentStatusChange(utils.CassandraGroupKind())
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

/*
func (r *ReconcileCassandra) GetRequestObject(request reconcile.Request) (ro runtime.Object) {
	instance := &v1alpha1.Cassandra{}
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err == nil {
		return instance
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
	var err error

	instance := &v1alpha1.Cassandra{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
	utils.ReconcileNilNotFound(err)

	managerInstance := &v1alpha1.Manager{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, managerInstance)

	deployment := GetDeployment()

	if err == nil {
		managerInstance.GetObjectFromObjectList()


		instance.Spec.CommonConfiguration = managerInstance.Spec.Services.Cassandras
		if managerInstance.Spec.Services.Cassandra.Size != nil {
			instance.Spec.Size = managerInstance.Spec.Services.Cassandra.Size
		} else {
			instance.Spec.Size = managerInstance.Spec.Size
		}
		if managerInstance.Spec.HostNetwork != nil {
			instance.Spec.HostNetwork = managerInstance.Spec.HostNetwork
		}
		err = r.Client.Update(context.TODO(), instance)
		utils.ReconcileErr(err)

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
	}
	// Create initial ConfigMaps
	volumeList := deployment.Spec.Template.Spec.Volumes
	configMap := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cassandra-" + instance.Name,
			Namespace: request.Namespace,
			Labels:    map[string]string{"contrail_manager": "cassandra"},
		},
		Data: instance.Spec.Configuration,
	}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "cassandra-" + instance.Name, Namespace: request.Namespace}, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			err = r.Client.Create(context.TODO(), configMap)
			utils.ReconcileErr(err)
		}
	}
	controllerutil.SetControllerReference(instance, configMap, r.Scheme)

	volume := corev1.Volume{
		Name: "cassandra-" + instance.Name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "cassandra-" + instance.Name,
				},
			},
		},
	}
	volumeList = append(volumeList, volume)

	deployment.Spec.Template.Spec.Volumes = volumeList

	// Set Deployment Name & Namespace

	deployment.ObjectMeta.Name = "cassandra-" + instance.Name
	deployment.ObjectMeta.Namespace = instance.Namespace

	// Configure Containers
	for idx, container := range deployment.Spec.Template.Spec.Containers {
		for containerName, image := range instance.Spec.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.Containers[idx]).Image = image
			}
			if containerName == "cassandra" {
				command := []string{"bash", "-c", "/docker-entrypoint.sh cassandra -f -Dcassandra.config=file:///mydata/${POD_IP}.yaml"}
				//command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
				(&deployment.Spec.Template.Spec.Containers[idx]).Command = command
				volumeMountList := []corev1.VolumeMount{}

				volumeMount := corev1.VolumeMount{
					Name:      "cassandra-" + instance.Name,
					MountPath: "/mydata",
				}
				volumeMountList = append(volumeMountList, volumeMount)

				(&deployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList

			}
		}
	}

	// Configure InitContainers
	for idx, container := range deployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range instance.Spec.Images {
			if containerName == container.Name {
				(&deployment.Spec.Template.Spec.InitContainers[idx]).Image = image
			}
		}
	}

	// Set HostNetwork
	deployment.Spec.Template.Spec.HostNetwork = *instance.Spec.HostNetwork

	// Set Selector and Label
	deployment.Spec.Selector.MatchLabels["app"] = "cassandra-" + instance.Name
	deployment.Spec.Template.ObjectMeta.Labels["app"] = "cassandra-" + instance.Name

	// Set Size
	deployment.Spec.Replicas = instance.Spec.Size
	// Create Deployment
	controllerutil.SetControllerReference(instance, deployment, r.Scheme)

	i = deployment
	err = i.Get(r.Client, deployemntRequest)
	if err != nil && errors.IsNotFound(err) {
		err = i.Create(r.Client)
		if err != nil {
			reqLogger.Error(err, "Failed to create Deployment", "Namespace", instance.Namespace, "Name", "cassandra-"+instance.Name)
			return reconcile.Result{}, err
		}
	} else if err == nil && *deployment.Spec.Replicas != *instance.Spec.Size {
		deployment.Spec.Replicas = instance.Spec.Size
		err = i.Update(r.Client)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Namespace", instance.Namespace, "Name", "cassandra-"+instance.Name)
			return reconcile.Result{}, err
		}
		active := false
		instance.Status.Active = &active
		err = r.Client.Status().Update(context.TODO(), instance)
		if err != nil {
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
		reqLogger.Info("Cassandra Deployment is ready")

	}
	return reconcile.Result{}, nil
}
func (r *ReconcileCassandra) ReplicaSetReconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Cassandra due to ReplicaSet changes")
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": "cassandra"})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	replicaSetList := &appsv1.ReplicaSetList{}
	err := r.Client.List(context.TODO(), listOps, replicaSetList)
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(replicaSetList.Items) > 0 {
		replicaSet := &appsv1.ReplicaSet{}
		for _, rs := range replicaSetList.Items {
			if *rs.Spec.Replicas > 0 {
				replicaSet = &rs
			} else {
				replicaSet = &rs
			}
		}
		instanceList := &v1alpha1.CassandraList{}
		err := r.Client.List(context.TODO(), listOps, instanceList)
		if err != nil {
			return reconcile.Result{}, err
		}
		instance := instanceList.Items[0]
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

			configMapInstanceDynamicConfig := &corev1.ConfigMap{}
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "cassandra-" + instance.Name, Namespace: request.Namespace}, configMapInstanceDynamicConfig)
			if err != nil {
				return reconcile.Result{}, err
			}

			sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })

			for idx := range podList.Items {
				var seeds []string
				for idx2 := range podList.Items {
					seeds = append(seeds, podList.Items[idx2].Status.PodIP)
				}
				cassandraConfigString := `cluster_name: ` + instance.Spec.Configuration["CASSANDRA_CLUSTER_NAME"] + `
num_tokens: 32
hinted_handoff_enabled: true
max_hint_window_in_ms: 10800000 # 3 hours
hinted_handoff_throttle_in_kb: 1024
max_hints_delivery_threads: 2
hints_directory: /var/lib/cassandra/hints
hints_flush_period_in_ms: 10000
max_hints_file_size_in_mb: 128
batchlog_replay_throttle_in_kb: 1024
authenticator: AllowAllAuthenticator
authorizer: AllowAllAuthorizer
role_manager: CassandraRoleManager
roles_validity_in_ms: 2000
permissions_validity_in_ms: 2000
credentials_validity_in_ms: 2000
partitioner: org.apache.cassandra.dht.Murmur3Partitioner
data_file_directories:
- /var/lib/cassandra/data
commitlog_directory: /var/lib/cassandra/commitlog
disk_failure_policy: stop
commit_failure_policy: stop
key_cache_size_in_mb:
key_cache_save_period: 14400
row_cache_size_in_mb: 0
row_cache_save_period: 0
counter_cache_size_in_mb:
counter_cache_save_period: 7200
saved_caches_directory: /var/lib/cassandra/saved_caches
commitlog_sync: periodic
commitlog_sync_period_in_ms: 10000
commitlog_segment_size_in_mb: 32
seed_provider:
- class_name: org.apache.cassandra.locator.SimpleSeedProvider
  parameters:
  - seeds: ` + strings.Join(seeds, ",") + `
concurrent_reads: 32
concurrent_writes: 32
concurrent_counter_writes: 32
concurrent_materialized_view_writes: 32
disk_optimization_strategy: ssd
memtable_allocation_type: heap_buffers
index_summary_capacity_in_mb:
index_summary_resize_interval_in_minutes: 60
trickle_fsync: false
trickle_fsync_interval_in_kb: 10240
storage_port: ` + instance.Spec.Configuration["CASSANDRA_STORAGE_PORT"] + `
ssl_storage_port: ` + instance.Spec.Configuration["CASSANDRA_SSL_STORAGE_PORT"] + `
listen_address: ` + podList.Items[idx].Status.PodIP + `
broadcast_address: ` + podList.Items[idx].Status.PodIP + `
start_native_transport: true
native_transport_port: ` + instance.Spec.Configuration["CASSANDRA_CQL_PORT"] + `
start_rpc: true
rpc_address: 0.0.0.0
rpc_port: ` + instance.Spec.Configuration["CASSANDRA_PORT"] + `
broadcast_rpc_address: ` + podList.Items[idx].Status.PodIP + `
rpc_keepalive: true
rpc_server_type: sync
thrift_framed_transport_size_in_mb: 15
incremental_backups: false
snapshot_before_compaction: false
auto_snapshot: true
tombstone_warn_threshold: 1000
tombstone_failure_threshold: 100000
column_index_size_in_kb: 64
batch_size_warn_threshold_in_kb: 5
batch_size_fail_threshold_in_kb: 50
compaction_throughput_mb_per_sec: 16
compaction_large_partition_warning_threshold_mb: 100
sstable_preemptive_open_interval_in_mb: 50
read_request_timeout_in_ms: 5000
range_request_timeout_in_ms: 10000
write_request_timeout_in_ms: 2000
counter_write_request_timeout_in_ms: 5000
cas_contention_timeout_in_ms: 1000
truncate_request_timeout_in_ms: 60000
request_timeout_in_ms: 10000
cross_node_timeout: false
endpoint_snitch: SimpleSnitch
dynamic_snitch_update_interval_in_ms: 100
dynamic_snitch_reset_interval_in_ms: 600000
dynamic_snitch_badness_threshold: 0.1
request_scheduler: org.apache.cassandra.scheduler.NoScheduler
server_encryption_options:
  internode_encryption: none
  keystore: conf/.keystore
  keystore_password: cassandra
  truststore: conf/.truststore
  truststore_password: cassandra
client_encryption_options:
  enabled: false
  optional: false
  keystore: conf/.keystore
  keystore_password: cassandra
internode_compression: all
inter_dc_tcp_nodelay: false
tracetype_query_ttl: 86400
tracetype_repair_ttl: 604800
gc_warn_threshold_in_ms: 1000
enable_user_defined_functions: false
enable_scripted_user_defined_functions: false
windows_timer_interval: 1
transparent_data_encryption_options:
  enabled: false
  chunk_length_kb: 64
  cipher: AES/CBC/PKCS5Padding
  key_alias: testing:1
  key_provider:
  - class_name: org.apache.cassandra.security.JKSKeyProvider
    parameters:
    - keystore: conf/.keystore
      keystore_password: cassandra
      store_type: JCEKS
      key_password: cassandra
auto_bootstrap: true
`
				if configMapInstanceDynamicConfig.Data == nil {
					data := map[string]string{podList.Items[idx].Status.PodIP + ".yaml": cassandraConfigString}
					configMapInstanceDynamicConfig.Data = data
				} else {
					configMapInstanceDynamicConfig.Data[podList.Items[idx].Status.PodIP+".yaml"] = cassandraConfigString
				}
				err = r.Client.Update(context.TODO(), configMapInstanceDynamicConfig)
				if err != nil {
					return reconcile.Result{}, err
				}

			}

			var podIpList []string
			for _, ip := range podNameIpMap {
				podIpList = append(podIpList, ip)
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
			portMap := map[string]string{"port": cassandra.Spec.Configuration["CASSANDRA_PORT"], "cqlPort": cassandra.Spec.Configuration["CASSANDRA_CQL_PORT"]}

			cassandra.Status.Ports = portMap
			err = r.Client.Status().Update(context.TODO(), &cassandra)
			if err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("All POD IPs available CASSANDRA: " + replicaSet.ObjectMeta.Labels["contrail_manager"])

		}
	}
	return reconcile.Result{}, nil
}

// Service is the interface to manage services
type Service interface {
	Get(reconcile.Request, runtime.Object) runtime.Object
}

// Get implements Service Get
func (r *ReconcileCassandra) Get(request reconcile.Request, object runtime.Object) *v1alpha1.Cassandra {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	instance := object.(*v1alpha1.Cassandra)
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No Cassandra Instance")
		}
	}
	return instance
}
*/
func PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment,
	commonConfiguration *v1alpha1.CommonConfiguration,
	instanceType string,
	request reconcile.Request,
	scheme *runtime.Scheme,
	instanceInterface interface{}) *appsv1.Deployment {
	instanceDeploymentName := request.Name + "-" + instanceType + "-deployment"
	instanceConfigMapName := request.Name + "-" + instanceType + "-configmap"
	instanceVolumeName := request.Name + "-" + instanceType + "-volume"
	intendedDeployment := utils.SetDeploymentCommonConfiguration(instanceDeployment, commonConfiguration)
	intendedDeployment.SetName(instanceDeploymentName)
	intendedDeployment.SetNamespace(request.Namespace)
	intendedDeployment.SetLabels(map[string]string{"contrail_manager": instanceType})
	switch instanceInterface.(type) {
	case v1alpha1.Cassandra:
		instance := instanceInterface.(*v1alpha1.Cassandra)
		controllerutil.SetControllerReference(instance, intendedDeployment, scheme)
	}

	volumeList := intendedDeployment.Spec.Template.Spec.Volumes
	volume := corev1.Volume{
		Name: instanceVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: instanceConfigMapName,
				},
			},
		},
	}
	volumeList = append(volumeList, volume)

	intendedDeployment.Spec.Template.Spec.Volumes = volumeList
	return intendedDeployment
}
func CreateInitialConfigMap(request reconcile.Request,
	instanceType string,
	instanceInterface interface{},
	scheme *runtime.Scheme,
	client client.Client) error {
	instanceConfigMapName := request.Name + "-" + instanceType + "-configmap"
	configMap := &corev1.ConfigMap{}
	configMap.SetName(instanceConfigMapName)
	configMap.SetNamespace(request.Namespace)
	configMap.SetLabels(map[string]string{"contrail_manager": instanceType})
	configMap.Data = make(map[string]string)
	switch instanceInterface.(type) {
	case v1alpha1.Cassandra:
		instance := instanceInterface.(*v1alpha1.Cassandra)
		controllerutil.SetControllerReference(instance, configMap, scheme)
	}
	err = client.Get(context.TODO(),
		types.NamespacedName{Name: instanceConfigMapName, Namespace: request.Namespace},
		configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			err = client.Create(context.TODO(), configMap)
			return err
		}
	}
	return nil
}

func CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment,
	commonConfiguration *v1alpha1.CommonConfiguration,
	instanceType string,
	request reconcile.Request,
	scheme *runtime.Scheme,
	client client.Client,
	instanceInterface interface{}) error {
	currentDeployment := &appsv1.Deployment{}
	err = client.Get(context.TODO(),
		types.NamespacedName{Name: request.Name, Namespace: request.Namespace},
		currentDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			switch instanceInterface.(type) {
			case v1alpha1.Cassandra:
				instance := instanceInterface.(*v1alpha1.Cassandra)
				controllerutil.SetControllerReference(instance, intendedDeployment, scheme)
			}
			err = client.Create(context.TODO(), intendedDeployment)
			if err != nil {
				return err
			}
		}
	} else {
		update := false
		if intendedDeployment.Spec.Replicas != currentDeployment.Spec.Replicas {
			update = true
		}
		for _, intendedContainer := range intendedDeployment.Spec.Template.Spec.Containers {
			for _, currentContainer := range currentDeployment.Spec.Template.Spec.Containers {
				if intendedContainer.Name == currentContainer.Name {
					if intendedContainer.Image != currentContainer.Image {
						update = true
					}
				}
			}
		}
		if update {
			err = client.Update(context.TODO(), intendedDeployment)
			if err != nil {
				return err
			}
		}
	}
	err = ManageActiveStatus(&currentDeployment.Status.ReadyReplicas,
		intendedDeployment.Spec.Replicas,
		instanceInterface,
		client)

	if err != nil {
		return err
	}
	return nil
}

func ManageActiveStatus(currentDeploymentReadyReplicas *int32,
	intendedDeploymentReplicas *int32,
	instanceInterface interface{},
	client client.Client) error {
	active := false
	if currentDeploymentReadyReplicas == intendedDeploymentReplicas {
		active = true
		switch instanceInterface.(type) {
		case v1alpha1.Cassandra:
			instance := instanceInterface.(*v1alpha1.Cassandra)
			instance.Status.Active = &active
		}
	}
	switch instanceInterface.(type) {
	case v1alpha1.Cassandra:
		instance := instanceInterface.(*v1alpha1.Cassandra)
		instance.Status.Active = &active
		err = client.Status().Update(context.TODO(), instance)
	}
	if err != nil {
		return err
	}
	return nil

}

// Reconcile reconciles cassandra
func (r *ReconcileCassandra) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Cassandra")
	instanceType := "cassandra"
	instanceDeploymentName := request.Name + "-" + instanceType + "-deployment"
	instanceVolumeName := request.Name + "-" + instanceType + "-volume"

	instance := &v1alpha1.Cassandra{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil && errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}

	managerInstance := &v1alpha1.Manager{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, managerInstance)
	if err == nil {
		if managerInstance.Spec.Services.Cassandras != nil {
			for _, cassandraManagerInstance := range managerInstance.Spec.Services.Cassandras {
				if cassandraManagerInstance.Name == request.Name {
					instance.Spec.CommonConfiguration = utils.MergeCommonConfiguration(
						managerInstance.Spec.CommonConfiguration,
						cassandraManagerInstance.Spec.CommonConfiguration)
				}
			}
		}
	}

	err = CreateInitialConfigMap(request,
		instanceType,
		instance,
		r.Scheme,
		r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	intendedDeployment := PrepareIntendedDeployment(GetDeployment(),
		&instance.Spec.CommonConfiguration,
		instanceType,
		request,
		r.Scheme,
		r.Client)

	for idx, container := range intendedDeployment.Spec.Template.Spec.Containers {
		for containerName, image := range instance.Spec.ServiceConfiguration.Images {
			if containerName == container.Name {
				(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Image = image
			}
			if containerName == "cassandra" {
				command := []string{"bash", "-c",
					"/docker-entrypoint.sh cassandra -f -Dcassandra.config=file:///mydata/${POD_IP}.yaml"}
				//command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
				(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Command = command

				volumeMountList := []corev1.VolumeMount{}
				volumeMount := corev1.VolumeMount{
					Name:      instanceVolumeName,
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

	err = CompareIntendedWithCurrentDeployment(intendedDeployment,
		&instance.Spec.CommonConfiguration,
		instanceType,
		request,
		r.Scheme,
		r.Client,
		instance)

	if err != nil {
		return reconcile.Result{}, err
	}

	currentDeployment := &appsv1.Deployment{}
	err = r.Client.Get(context.TODO(),
		types.NamespacedName{Name: instanceDeploymentName, Namespace: request.Namespace},
		currentDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			controllerutil.SetControllerReference(instance, intendedDeployment, r.Scheme)
			err = r.Client.Create(context.TODO(), intendedDeployment)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	} else {
		update := false
		if intendedDeployment.Spec.Replicas != currentDeployment.Spec.Replicas {
			update = true
		}
		for _, intendedContainer := range intendedDeployment.Spec.Template.Spec.Containers {
			for _, currentContainer := range currentDeployment.Spec.Template.Spec.Containers {
				if intendedContainer.Name == currentContainer.Name {
					if intendedContainer.Image != currentContainer.Image {
						update = true
					}
				}
			}
		}
		if update {
			err = r.Client.Update(context.TODO(), intendedDeployment)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}
	active := false
	if currentDeployment.Status.ReadyReplicas == *intendedDeployment.Spec.Replicas {
		active = true
		instance.Status.Active = &active
		reqLogger.Info("Cassandra Deployment is ready")
	}
	err = r.Client.Status().Update(context.TODO(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": instanceType})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	replicaSetList := &appsv1.ReplicaSetList{}
	err = r.Client.List(context.TODO(), listOps, replicaSetList)
	if err != nil {
		return reconcile.Result{}, err
	}

	if len(replicaSetList.Items) > 0 {
		replicaSet := &appsv1.ReplicaSet{}
		for _, rs := range replicaSetList.Items {
			if *rs.Spec.Replicas > 0 {
				replicaSet = &rs
			}
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
		var podNameIPMap = make(map[string]string)

		if int32(len(podList.Items)) == *replicaSet.Spec.Replicas {
			for _, pod := range podList.Items {
				if pod.Status.PodIP != "" {
					podNameIPMap[pod.Name] = pod.Status.PodIP
				}
			}
		}

		if int32(len(podNameIPMap)) == *replicaSet.Spec.Replicas {

			configMapInstanceDynamicConfig := &corev1.ConfigMap{}
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "cassandra-" + instance.Name, Namespace: request.Namespace}, configMapInstanceDynamicConfig)
			if err != nil {
				return reconcile.Result{}, err
			}

			sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })

			for idx := range podList.Items {
				var seeds []string
				for idx2 := range podList.Items {
					seeds = append(seeds, podList.Items[idx2].Status.PodIP)
				}
				cassandraConfigString := `cluster_name: ` + instance.Spec.ServiceConfiguration.ClusterName + `
num_tokens: 32
hinted_handoff_enabled: true
max_hint_window_in_ms: 10800000 # 3 hours
hinted_handoff_throttle_in_kb: 1024
max_hints_delivery_threads: 2
hints_directory: /var/lib/cassandra/hints
hints_flush_period_in_ms: 10000
max_hints_file_size_in_mb: 128
batchlog_replay_throttle_in_kb: 1024
authenticator: AllowAllAuthenticator
authorizer: AllowAllAuthorizer
role_manager: CassandraRoleManager
roles_validity_in_ms: 2000
permissions_validity_in_ms: 2000
credentials_validity_in_ms: 2000
partitioner: org.apache.cassandra.dht.Murmur3Partitioner
data_file_directories:
- /var/lib/cassandra/data
commitlog_directory: /var/lib/cassandra/commitlog
disk_failure_policy: stop
commit_failure_policy: stop
key_cache_size_in_mb:
key_cache_save_period: 14400
row_cache_size_in_mb: 0
row_cache_save_period: 0
counter_cache_size_in_mb:
counter_cache_save_period: 7200
saved_caches_directory: /var/lib/cassandra/saved_caches
commitlog_sync: periodic
commitlog_sync_period_in_ms: 10000
commitlog_segment_size_in_mb: 32
seed_provider:
- class_name: org.apache.cassandra.locator.SimpleSeedProvider
  parameters:
  - seeds: ` + strings.Join(seeds, ",") + `
concurrent_reads: 32
concurrent_writes: 32
concurrent_counter_writes: 32
concurrent_materialized_view_writes: 32
disk_optimization_strategy: ssd
memtable_allocation_type: heap_buffers
index_summary_capacity_in_mb:
index_summary_resize_interval_in_minutes: 60
trickle_fsync: false
trickle_fsync_interval_in_kb: 10240
storage_port: ` + instance.Spec.ServiceConfiguration.StoragePort + `
ssl_storage_port: ` + instance.Spec.ServiceConfiguration.SslStoragePort + `
listen_address: ` + podList.Items[idx].Status.PodIP + `
broadcast_address: ` + podList.Items[idx].Status.PodIP + `
start_native_transport: true
native_transport_port: ` + instance.Spec.ServiceConfiguration.CqlPort + `
start_rpc: true
rpc_address: 0.0.0.0
rpc_port: ` + instance.Spec.ServiceConfiguration.Port + `
broadcast_rpc_address: ` + podList.Items[idx].Status.PodIP + `
rpc_keepalive: true
rpc_server_type: sync
thrift_framed_transport_size_in_mb: 15
incremental_backups: false
snapshot_before_compaction: false
auto_snapshot: true
tombstone_warn_threshold: 1000
tombstone_failure_threshold: 100000
column_index_size_in_kb: 64
batch_size_warn_threshold_in_kb: 5
batch_size_fail_threshold_in_kb: 50
compaction_throughput_mb_per_sec: 16
compaction_large_partition_warning_threshold_mb: 100
sstable_preemptive_open_interval_in_mb: 50
read_request_timeout_in_ms: 5000
range_request_timeout_in_ms: 10000
write_request_timeout_in_ms: 2000
counter_write_request_timeout_in_ms: 5000
cas_contention_timeout_in_ms: 1000
truncate_request_timeout_in_ms: 60000
request_timeout_in_ms: 10000
cross_node_timeout: false
endpoint_snitch: SimpleSnitch
dynamic_snitch_update_interval_in_ms: 100
dynamic_snitch_reset_interval_in_ms: 600000
dynamic_snitch_badness_threshold: 0.1
request_scheduler: org.apache.cassandra.scheduler.NoScheduler
server_encryption_options:
  internode_encryption: none
  keystore: conf/.keystore
  keystore_password: cassandra
  truststore: conf/.truststore
  truststore_password: cassandra
client_encryption_options:
  enabled: false
  optional: false
  keystore: conf/.keystore
  keystore_password: cassandra
internode_compression: all
inter_dc_tcp_nodelay: false
tracetype_query_ttl: 86400
tracetype_repair_ttl: 604800
gc_warn_threshold_in_ms: 1000
enable_user_defined_functions: false
enable_scripted_user_defined_functions: false
windows_timer_interval: 1
transparent_data_encryption_options:
  enabled: false
  chunk_length_kb: 64
  cipher: AES/CBC/PKCS5Padding
  key_alias: testing:1
  key_provider:
  - class_name: org.apache.cassandra.security.JKSKeyProvider
    parameters:
    - keystore: conf/.keystore
      keystore_password: cassandra
      store_type: JCEKS
      key_password: cassandra
auto_bootstrap: true
`
				if configMapInstanceDynamicConfig.Data == nil {
					data := map[string]string{podList.Items[idx].Status.PodIP + ".yaml": cassandraConfigString}
					configMapInstanceDynamicConfig.Data = data
				} else {
					configMapInstanceDynamicConfig.Data[podList.Items[idx].Status.PodIP+".yaml"] = cassandraConfigString
				}
				err = r.Client.Update(context.TODO(), configMapInstanceDynamicConfig)
				if err != nil {
					return reconcile.Result{}, err
				}

			}

			var podIpList []string
			for _, ip := range podNameIPMap {
				podIpList = append(podIpList, ip)
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
			cassandra.Status.Nodes = podNameIPMap
			portMap := map[string]string{"port": cassandra.Spec.ServiceConfiguration.Port, "cqlPort": cassandra.Spec.ServiceConfiguration.CqlPort}

			cassandra.Status.Ports = portMap
			err = r.Client.Status().Update(context.TODO(), &cassandra)
			if err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("All POD IPs available CASSANDRA: " + replicaSet.ObjectMeta.Labels["contrail_manager"])

		}
	}

	return reconcile.Result{}, nil
}
