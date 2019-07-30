package cassandra

import (
	"context"
	"sort"
	"strconv"
	"strings"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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
	err = c.Watch(&source.Kind{Type: &v1alpha1.Cassandra{}},
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
	predInitStatus := utils.PodInitStatusChange(map[string]string{"contrail_manager": "cassandra"})
	predPodIPChange := utils.PodIPChange(map[string]string{"contrail_manager": "cassandra"})
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
	predManagerSizeChange := utils.ManagerSizeChange(utils.CassandraGroupKind())
	// Watch for Manager events.
	err = c.Watch(srcManager, managerHandler, predManagerSizeChange)
	if err != nil {
		return err
	}

	srcDeployment := &source.Kind{Type: &appsv1.Deployment{}}
	deploymentHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Cassandra{},
	}
	deploymentPred := utils.DeploymentStatusChange(utils.CassandraGroupKind())
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

func CreateInstanceConfiguration(request reconcile.Request,
	podList *corev1.PodList,
	instanceType string,
	client client.Client,
	cassandraConfiguration v1alpha1.CassandraConfiguration) error {
	instanceConfigMapName := request.Name + "-" + instanceType + "-configmap"
	configMapInstanceDynamicConfig := &corev1.ConfigMap{}
	err = client.Get(context.TODO(),
		types.NamespacedName{Name: instanceConfigMapName, Namespace: request.Namespace},
		configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}

	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })

	for idx := range podList.Items {
		var seeds []string
		for idx2 := range podList.Items {
			seeds = append(seeds, podList.Items[idx2].Status.PodIP)
		}
		cassandraConfigString := `cluster_name: ` + cassandraConfiguration.ClusterName + `
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
storage_port: ` + strconv.Itoa(cassandraConfiguration.StoragePort) + `
ssl_storage_port: ` + strconv.Itoa(cassandraConfiguration.SslStoragePort) + `
listen_address: ` + podList.Items[idx].Status.PodIP + `
broadcast_address: ` + podList.Items[idx].Status.PodIP + `
start_native_transport: true
native_transport_port: ` + strconv.Itoa(cassandraConfiguration.CqlPort) + `
start_rpc: ` + strconv.FormatBool(cassandraConfiguration.StartRpc) + `
rpc_address: ` + podList.Items[idx].Status.PodIP + `
rpc_port: ` + strconv.Itoa(cassandraConfiguration.Port) + `
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
		err = client.Update(context.TODO(), configMapInstanceDynamicConfig)
		if err != nil {
			return err
		}

	}
	return nil
}

func ManageNodeStatus(podNameIPMap map[string]string,
	client client.Client,
	instance *v1alpha1.Cassandra) error {

	instance.Status.Nodes = podNameIPMap
	portMap := map[string]string{"port": strconv.Itoa(instance.Spec.ServiceConfiguration.Port),
		"cqlPort": strconv.Itoa(instance.Spec.ServiceConfiguration.CqlPort)}

	instance.Status.Ports = portMap
	err = client.Status().Update(context.TODO(), instance)
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

	// A reconcile.Request for Cassandra can be triggered by 4 different types:
	// 1. Any changes on the Cassandra instance
	// --> reconcile.Request is Cassandra instance name/namespace
	// 2. IP Status change on the Pods
	// --> reconcile.Request is Replicaset name/namespace
	// --> we need to evaluate the label to get the Cassandra instance
	// 3. Status change on the Deployment
	// --> reconcile.Request is Cassandra instance name/namespace
	// 4. Cassandras changes on the Manager instance
	// --> reconcile.Request is Manager instance name/namespace
	instance := &v1alpha1.Cassandra{}
	err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
	// if not found we expect it a change in replicaset
	// and get the cassandra instance via label
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

	var managerName string
	ownedByManager := false
	ownerRefList := instance.GetOwnerReferences()
	for _, ownerRef := range ownerRefList {
		if *ownerRef.Controller {
			if ownerRef.Kind == "Manager" {
				managerName = ownerRef.Name
				ownedByManager = true
			}
		}
	}

	if ownedByManager {
		managerInstance := &v1alpha1.Manager{}
		err = r.Client.Get(context.TODO(), types.NamespacedName{Name: managerName, Namespace: request.Namespace}, managerInstance)
		if err == nil {
			if managerInstance.Spec.Services.Cassandras != nil {
				for _, cassandraManagerInstance := range managerInstance.Spec.Services.Cassandras {
					if cassandraManagerInstance.Name == request.Name {
						instance.Spec.CommonConfiguration = utils.MergeCommonConfiguration(
							managerInstance.Spec.CommonConfiguration,
							cassandraManagerInstance.Spec.CommonConfiguration)
						err = r.Client.Update(context.TODO(), instance)
						if err != nil {
							return reconcile.Result{}, err
						}
					}
				}
			}
		}
	}

	configMap := utils.PrepareConfigMap(request,
		instanceType,
		r.Scheme,
		r.Client)
	err = controllerutil.SetControllerReference(instance, configMap, r.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	err = utils.CreateConfigMap(request, instanceType, configMap, r.Scheme, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}

	intendedDeployment := utils.PrepareIntendedDeployment(GetDeployment(),
		&instance.Spec.CommonConfiguration,
		instanceType,
		request,
		r.Scheme)

	err = controllerutil.SetControllerReference(instance, intendedDeployment, r.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	for idx, container := range intendedDeployment.Spec.Template.Spec.Containers {
		for containerName, image := range instance.Spec.ServiceConfiguration.Images {
			if containerName == container.Name {
				(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Image = image
			}
			if containerName == "cassandra" {
				command := []string{"bash", "-c",
					"/docker-entrypoint.sh cassandra -f -Dcassandra.config=file:///mydata/${POD_IP}.yaml"}
				command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
				(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Command = command

				volumeMountList := []corev1.VolumeMount{}
				volumeMount := corev1.VolumeMount{
					Name:      request.Name + "-" + instanceType + "-volume",
					MountPath: "/mydata",
				}
				volumeMountList = append(volumeMountList, volumeMount)

				(&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList
				var jvmOpts string
				if instance.Spec.ServiceConfiguration.MinHeapSize != "" {
					jvmOpts = "-Xms" + instance.Spec.ServiceConfiguration.MinHeapSize
				}
				if instance.Spec.ServiceConfiguration.MaxHeapSize != "" {
					jvmOpts = jvmOpts + " -Xmx" + instance.Spec.ServiceConfiguration.MaxHeapSize
				}
				if jvmOpts != "" {
					envVars := []corev1.EnvVar{{
						Name:  "JVM_OPTS",
						Value: jvmOpts,
					}}
					(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Env = envVars
				}

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

	err = utils.CompareIntendedWithCurrentDeployment(intendedDeployment,
		&instance.Spec.CommonConfiguration,
		instanceType,
		request,
		r.Scheme,
		r.Client,
		instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	podIPList, podIPMap, err := utils.GetPodIPListAndIPMap(instanceType, request, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}
	if len(podIPList.Items) > 0 {
		err = CreateInstanceConfiguration(request,
			podIPList,
			instanceType,
			r.Client,
			instance.Spec.ServiceConfiguration)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = utils.SetPodsToReady(podIPList, r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = ManageNodeStatus(podIPMap, r.Client, instance)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	utils.SetInstanceActive(r.Client, &instance.Status, intendedDeployment)
	err = r.Client.Status().Update(context.TODO(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
