package cassandra

import (
	"context"
	"fmt"
	"sort"
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

	// Create initial ConfigMaps
	volumeList := deployment.Spec.Template.Spec.Volumes

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

	volume := corev1.Volume{
		Name: "cassandra-" + cassandraInstance.Name,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: "cassandra-" + cassandraInstance.Name,
				},
			},
		},
	}
	volumeList = append(volumeList, volume)

	deployment.Spec.Template.Spec.Volumes = volumeList

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
				command := []string{"bash", "-c", "/docker-entrypoint.sh cassandra -f -Dcassandra.config=file:///mydata/${POD_IP}.yaml"}
				//command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
				(&deployment.Spec.Template.Spec.Containers[idx]).Command = command
				volumeMountList := []corev1.VolumeMount{}

				volumeMount := corev1.VolumeMount{
					Name:      "cassandra-" + cassandraInstance.Name,
					MountPath: "/mydata",
				}
				volumeMountList = append(volumeMountList, volumeMount)

				(&deployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList

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
	fmt.Println("Creating or Updating Cassandra")
	controllerutil.SetControllerReference(cassandraInstance, deployment, r.Scheme)
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "cassandra-" + cassandraInstance.Name, Namespace: cassandraInstance.Namespace}, deployment)
	if err != nil && errors.IsNotFound(err) {
		fmt.Println("Creating Cassandra")
		deployment.Spec.Template.ObjectMeta.Labels["version"] = "1"
		err = r.Client.Create(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create Deployment", "Namespace", cassandraInstance.Namespace, "Name", "cassandra-"+cassandraInstance.Name)
			return reconcile.Result{}, err
		}
	} else if err == nil && *deployment.Spec.Replicas != *cassandraInstance.Spec.Size {
		fmt.Println("Updating Cassandra")
		deployment.Spec.Replicas = cassandraInstance.Spec.Size
		err = r.Client.Update(context.TODO(), deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to update Deployment", "Namespace", cassandraInstance.Namespace, "Name", "cassandra-"+cassandraInstance.Name)
			return reconcile.Result{}, err
		}
		active := false
		cassandraInstance.Status.Active = &active
		err = r.Client.Status().Update(context.TODO(), cassandraInstance)
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
		fmt.Println("Ready Replicas: ", deployment.Status.ReadyReplicas)
		fmt.Println("Spec Replicas: ", *deployment.Spec.Replicas)
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
		cassandraInstanceList := &v1alpha1.CassandraList{}
		err := r.Client.List(context.TODO(), listOps, cassandraInstanceList)
		if err != nil {
			return reconcile.Result{}, err
		}
		cassandraInstance := cassandraInstanceList.Items[0]
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
			err = r.Client.Get(context.TODO(), types.NamespacedName{Name: "cassandra-" + cassandraInstance.Name, Namespace: request.Namespace}, configMapInstanceDynamicConfig)
			if err != nil {
				return reconcile.Result{}, err
			}

			sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })

			for idx := range podList.Items {
				fmt.Println("############## POD NAME ############# ", podList.Items[idx].Status.PodIP)

				var seeds []string
				for idx2 := range podList.Items {
					seeds = append(seeds, podList.Items[idx2].Status.PodIP)
				}
				cassandraConfigString := `cluster_name: ` + cassandraInstance.Spec.Configuration["CASSANDRA_CLUSTER_NAME"] + `
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
storage_port: ` + cassandraInstance.Spec.Configuration["CASSANDRA_STORAGE_PORT"] + `
ssl_storage_port: ` + cassandraInstance.Spec.Configuration["CASSANDRA_SSL_STORAGE_PORT"] + `
listen_address: ` + podList.Items[idx].Status.PodIP + `
broadcast_address: ` + podList.Items[idx].Status.PodIP + `
start_native_transport: true
native_transport_port: ` + cassandraInstance.Spec.Configuration["CASSANDRA_CQL_PORT"] + `
start_rpc: true
rpc_address: 0.0.0.0
rpc_port: ` + cassandraInstance.Spec.Configuration["CASSANDRA_PORT"] + `
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
			fmt.Println("LABELING PODS ############################# podList length ", len(podList.Items))
			fmt.Println("LABELING PODS ############################# replicas  ", *replicaSet.Spec.Replicas)
			fmt.Println("LABELING PODS ############################# 0 ", podNameIpMap)
			for _, pod := range podList.Items {
				pod.ObjectMeta.Labels["status"] = "ready"
				fmt.Println("LABELING PODS ############################# 1 ")
				err = r.Client.Update(context.TODO(), &pod)
				if err != nil {
					return reconcile.Result{}, err
				}
			}
			fmt.Println("LABELING PODS ############################# 2 ")
			cassandraList := &v1alpha1.CassandraList{}
			cassandraListOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
			err = r.Client.List(context.TODO(), cassandraListOps, cassandraList)
			if err != nil {
				return reconcile.Result{}, err
			}
			cassandra := cassandraList.Items[0]
			cassandra.Status.Nodes = podNameIpMap
			portMap := map[string]string{"port": cassandra.Spec.Configuration["ZOOKEEPER_PORT"]}
			cassandra.Status.Ports = portMap
			err = r.Client.Status().Update(context.TODO(), &cassandra)
			if err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("All POD IPs available ZOOKEEPER: " + replicaSet.ObjectMeta.Labels["contrail_manager"])

		}
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
