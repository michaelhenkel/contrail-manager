package contrailtest

import (
	"context"
	"testing"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/kylelemons/godebug/diff"
	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var config = &v1alpha1.Config{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "config1",
		Namespace: "default",
		Labels: map[string]string{
			"contrail_cluster": "cluster1",
		},
	},
}

var control = &v1alpha1.Control{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "control1",
		Namespace: "default",
		Labels: map[string]string{
			"contrail_cluster": "cluster1",
			"control_role":     "master",
		},
	},
}

var kubemanager = &v1alpha1.Kubemanager{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "kubemanager1",
		Namespace: "default",
		Labels: map[string]string{
			"contrail_cluster": "cluster1",
		},
	},
}

var webui = &v1alpha1.Webui{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "webui1",
		Namespace: "default",
		Labels: map[string]string{
			"contrail_cluster": "cluster1",
		},
	},
}

var cassandra = &v1alpha1.Cassandra{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "cassandra1",
		Namespace: "default",
		Labels: map[string]string{
			"contrail_cluster": "cluster1",
		},
	},
}

var zookeeper = &v1alpha1.Zookeeper{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "zookeeper1",
		Namespace: "default",
		Labels: map[string]string{
			"contrail_cluster": "cluster1",
		},
	},
}

var rabbitmq = &v1alpha1.Rabbitmq{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "rabbitmq",
		Namespace: "default",
		Labels: map[string]string{
			"contrail_cluster": "cluster1",
		},
	},
}

var rabbitmqList = &v1alpha1.RabbitmqList{}
var zookeeperList = &v1alpha1.ZookeeperList{}
var cassandraList = &v1alpha1.CassandraList{}
var configList = &v1alpha1.ConfigList{}
var controlList = &v1alpha1.ControlList{}
var kubemanagerList = &v1alpha1.KubemanagerList{}
var webuiList = &v1alpha1.WebuiList{}

var configMap = &corev1.ConfigMap{}

type Environment struct {
	client              *client.Client
	configPodList       corev1.PodList
	rabbitmqPodList     corev1.PodList
	zookeeperPodList    corev1.PodList
	cassandraPodList    corev1.PodList
	controlPodList      corev1.PodList
	kubemanbagerPodList corev1.PodList
	webuiPodList        corev1.PodList
	configInstance      v1alpha1.Instance
	controlInstance     v1alpha1.Instance
	cassandraInstance   v1alpha1.Instance
	zookeeperInstance   v1alpha1.Instance
	rabbitmqInstance    v1alpha1.Instance
	kubemanagerInstance v1alpha1.Instance
	webuiInstance       v1alpha1.Instance
}

func SetupEnv() Environment {
	logf.SetLogger(logf.ZapLogger(true))
	configConfigMap := *configMap
	rabbitmqConfigMap := *configMap
	zookeeperConfigMap := *configMap
	cassandraConfigMap := *configMap
	controlConfigMap := *configMap
	kubemanagerConfigMap := *configMap
	webuiConfigMap := *configMap

	configConfigMap.Name = "config1-config-configmap"
	configConfigMap.Namespace = "default"

	rabbitmqConfigMap.Name = "rabbitmq1-rabbitmq-configmap"
	rabbitmqConfigMap.Namespace = "default"

	zookeeperConfigMap.Name = "zookeeper1-zookeeper-configmap"
	zookeeperConfigMap.Namespace = "default"

	cassandraConfigMap.Name = "cassandra1-cassandra-configmap"
	cassandraConfigMap.Namespace = "default"

	controlConfigMap.Name = "control1-control-configmap"
	controlConfigMap.Namespace = "default"

	kubemanagerConfigMap.Name = "kubemanager1-kubemanager-configmap"
	kubemanagerConfigMap.Namespace = "default"

	webuiConfigMap.Name = "webui1-webui-configmap"
	webuiConfigMap.Namespace = "default"

	s := scheme.Scheme
	s.AddKnownTypes(v1alpha1.SchemeGroupVersion,
		config,
		cassandra,
		zookeeper,
		rabbitmq,
		control,
		kubemanager,
		webui,
		rabbitmqList,
		zookeeperList,
		cassandraList,
		configList,
		controlList,
		kubemanagerList,
		webuiList)

	objs := []runtime.Object{config,
		cassandra,
		zookeeper,
		rabbitmq,
		control,
		kubemanager,
		webui,
		&configConfigMap,
		&controlConfigMap,
		&cassandraConfigMap,
		&zookeeperConfigMap,
		&rabbitmqConfigMap,
		&kubemanagerConfigMap,
		&webuiConfigMap}

	cl := fake.NewFakeClient(objs...)

	var configInstance, rabbitmqInstance, cassandraInstance, zookeeperInstance, controlInstance, kubemanagerInstance, webuiInstance v1alpha1.Instance

	configInstance = config
	rabbitmqInstance = rabbitmq
	cassandraInstance = cassandra
	zookeeperInstance = zookeeper
	controlInstance = control
	kubemanagerInstance = kubemanager
	webuiInstance = webui

	var podServiceMap = make(map[string]map[string]string)
	podServiceMap["configPods"] = map[string]string{"pod1": "1.1.1.1", "pod2": "1.1.1.2", "pod3": "1.1.1.3"}
	podServiceMap["rabbitmqPods"] = map[string]string{"pod1": "1.1.4.1", "pod2": "1.1.4.2", "pod3": "1.1.4.3"}
	podServiceMap["cassandraPods"] = map[string]string{"pod1": "1.1.2.1", "pod2": "1.1.2.2", "pod3": "1.1.2.3"}
	podServiceMap["zookeeperPods"] = map[string]string{"pod1": "1.1.3.1", "pod2": "1.1.3.2", "pod3": "1.1.3.3"}
	podServiceMap["controlPods"] = map[string]string{"pod1": "1.1.5.1", "pod2": "1.1.5.2", "pod3": "1.1.5.3"}
	podServiceMap["kubemanagerPods"] = map[string]string{"pod1": "1.1.6.1", "pod2": "1.1.6.2", "pod3": "1.1.6.3"}
	podServiceMap["webuiPods"] = map[string]string{"pod1": "1.1.7.1", "pod2": "1.1.7.2", "pod3": "1.1.7.3"}

	type PodMap struct {
		configPods      map[string]string
		rabbitmqPods    map[string]string
		cassandraPods   map[string]string
		zookeeperPods   map[string]string
		controlPods     map[string]string
		kubemanagerPods map[string]string
		webuiPods       map[string]string
	}
	podMap := PodMap{
		configPods:      map[string]string{"pod1": "1.1.1.1", "pod2": "1.1.1.2", "pod3": "1.1.1.3"},
		rabbitmqPods:    map[string]string{"pod1": "1.1.4.1", "pod2": "1.1.4.2", "pod3": "1.1.4.3"},
		cassandraPods:   map[string]string{"pod1": "1.1.2.1", "pod2": "1.1.2.2", "pod3": "1.1.2.3"},
		zookeeperPods:   map[string]string{"pod1": "1.1.3.1", "pod2": "1.1.3.2", "pod3": "1.1.3.3"},
		controlPods:     map[string]string{"pod1": "1.1.5.1", "pod2": "1.1.5.2", "pod3": "1.1.5.3"},
		kubemanagerPods: map[string]string{"pod1": "1.1.6.1", "pod2": "1.1.6.2", "pod3": "1.1.6.3"},
		webuiPods:       map[string]string{"pod1": "1.1.7.1", "pod2": "1.1.7.2", "pod3": "1.1.7.3"},
	}

	podTemplate := corev1.Pod{}

	configPodItems := []corev1.Pod{}
	for pod, ip := range podMap.configPods {
		podTemplate.Name = pod
		podTemplate.Namespace = "default"
		podTemplate.Status.PodIP = ip
		configPodItems = append(configPodItems, podTemplate)
	}
	configPodList := corev1.PodList{
		Items: configPodItems,
	}
	rabbitmqPodItems := []corev1.Pod{}
	for pod, ip := range podMap.rabbitmqPods {
		podTemplate.Name = pod
		podTemplate.Namespace = "default"
		podTemplate.Status.PodIP = ip
		rabbitmqPodItems = append(rabbitmqPodItems, podTemplate)
	}
	rabbitmqPodList := corev1.PodList{
		Items: rabbitmqPodItems,
	}
	cassandraPodItems := []corev1.Pod{}
	for pod, ip := range podMap.cassandraPods {
		podTemplate.Name = pod
		podTemplate.Namespace = "default"
		podTemplate.Status.PodIP = ip
		cassandraPodItems = append(cassandraPodItems, podTemplate)
	}
	cassandraPodList := corev1.PodList{
		Items: cassandraPodItems,
	}
	zookeeperPodItems := []corev1.Pod{}
	for pod, ip := range podMap.zookeeperPods {
		podTemplate.Name = pod
		podTemplate.Namespace = "default"
		podTemplate.Status.PodIP = ip
		zookeeperPodItems = append(zookeeperPodItems, podTemplate)
	}
	zookeeperPodList := corev1.PodList{
		Items: zookeeperPodItems,
	}
	controlPodItems := []corev1.Pod{}
	for pod, ip := range podMap.controlPods {
		podTemplate.Name = pod
		podTemplate.Namespace = "default"
		podTemplate.Status.PodIP = ip
		controlPodItems = append(controlPodItems, podTemplate)
	}
	controlPodList := corev1.PodList{
		Items: controlPodItems,
	}
	kubemanagerPodItems := []corev1.Pod{}
	for pod, ip := range podMap.kubemanagerPods {
		podTemplate.Name = pod
		podTemplate.Namespace = "default"
		podTemplate.Status.PodIP = ip
		kubemanagerPodItems = append(kubemanagerPodItems, podTemplate)
	}
	kubemanagerPodList := corev1.PodList{
		Items: kubemanagerPodItems,
	}
	webuiPodItems := []corev1.Pod{}
	for pod, ip := range podMap.webuiPods {
		podTemplate.Name = pod
		podTemplate.Namespace = "default"
		podTemplate.Status.PodIP = ip
		webuiPodItems = append(webuiPodItems, podTemplate)
	}
	webuiPodList := corev1.PodList{
		Items: webuiPodItems,
	}

	configInstance.ManageNodeStatus(podMap.configPods, cl)
	rabbitmqInstance.ManageNodeStatus(podMap.rabbitmqPods, cl)
	cassandraInstance.ManageNodeStatus(podMap.cassandraPods, cl)
	zookeeperInstance.ManageNodeStatus(podMap.zookeeperPods, cl)
	controlInstance.ManageNodeStatus(podMap.controlPods, cl)
	kubemanagerInstance.ManageNodeStatus(podMap.kubemanagerPods, cl)
	webuiInstance.ManageNodeStatus(podMap.webuiPods, cl)

	environment := Environment{
		client:              &cl,
		configPodList:       configPodList,
		cassandraPodList:    cassandraPodList,
		zookeeperPodList:    zookeeperPodList,
		rabbitmqPodList:     rabbitmqPodList,
		controlPodList:      controlPodList,
		kubemanbagerPodList: kubemanagerPodList,
		webuiPodList:        webuiPodList,
		configInstance:      configInstance,
		controlInstance:     controlInstance,
		cassandraInstance:   cassandraInstance,
		zookeeperInstance:   zookeeperInstance,
		rabbitmqInstance:    rabbitmqInstance,
		kubemanagerInstance: kubemanagerInstance,
		webuiInstance:       webuiInstance,
	}
	return environment
}
func TestConfigConfig(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	environment := SetupEnv()
	cl := *environment.client
	err := environment.configInstance.CreateInstanceConfiguration(reconcile.Request{types.NamespacedName{Name: "config1", Namespace: "default"}}, &environment.configPodList, cl)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}
	err = cl.Get(context.TODO(),
		types.NamespacedName{Name: "config1-config-configmap", Namespace: "default"},
		configMap)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}
	if configMap.Data["api.1.1.1.1"] != configConfigHa {
		configDiff := diff.Diff(configMap.Data["api.1.1.1.1"], configConfigHa)
		t.Fatalf("get api config: \n%v\n", configDiff)
	}
	/*
		if configMap.Data["devicemanager.1.1.1.1"] != configDevicemanagerHa {
			devicemanagerDiff := diff.Diff(configMap.Data["devicemanager.1.1.1.1"], configDevicemanagerHa)
			t.Fatalf("get devicemanager config: \n%v\n", devicemanagerDiff)
		}
	*/
}

func TestWebuiConfig(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	environment := SetupEnv()
	cl := *environment.client
	err := environment.webuiInstance.CreateInstanceConfiguration(reconcile.Request{types.NamespacedName{Name: "webui1", Namespace: "default"}}, &environment.webuiPodList, cl)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}
	err = cl.Get(context.TODO(),
		types.NamespacedName{Name: "webui1-webui-configmap", Namespace: "default"},
		configMap)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}
	if configMap.Data["config.global.js.1.1.7.1"] != webuiConfigHa {
		configDiff := diff.Diff(configMap.Data["config.global.js.1.1.7.1"], webuiConfigHa)
		t.Fatalf("get webui config: \n%v\n", configDiff)
	}
}

func TestCassandraConfig(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	environment := SetupEnv()
	cl := *environment.client
	err := environment.cassandraInstance.CreateInstanceConfiguration(reconcile.Request{types.NamespacedName{Name: "cassandra1", Namespace: "default"}}, &environment.cassandraPodList, cl)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}
	err = cl.Get(context.TODO(),
		types.NamespacedName{Name: "cassandra1-cassandra-configmap", Namespace: "default"},
		configMap)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}
	if configMap.Data["1.1.2.1.yaml"] != cassandraConfig {
		configDiff := diff.Diff(configMap.Data["1.1.2.1.yaml"], cassandraConfig)
		t.Fatalf("get cassandra config: \n%v\n", configDiff)
	}
}

var webuiConfigHa = `/*
* Copyright (c) 2014 Juniper Networks, Inc. All rights reserved.
*/
var config = {};
config.orchestration = {};
config.orchestration.Manager = "none";
config.orchestrationModuleEndPointFromConfig = false;
config.contrailEndPointFromConfig = true;
config.regionsFromConfig = false;
config.endpoints = {};
config.endpoints.apiServiceType = "ApiServer";
config.endpoints.opServiceType = "OpServer";
config.regions = {};
config.regions.RegionOne = "http://127.0.0.1:5000/v2.0";
config.serviceEndPointTakePublicURL = true;
config.networkManager = {};
config.networkManager.ip = "127.0.0.1";
config.networkManager.port = "9696";
config.networkManager.authProtocol = "http";
config.networkManager.apiVersion = [];
config.networkManager.strictSSL = false;
config.networkManager.ca = "";
config.imageManager = {};
config.imageManager.ip = "127.0.0.1";
config.imageManager.port = "9292";
config.imageManager.authProtocol = "http";
config.imageManager.apiVersion = ['v1', 'v2'];
config.imageManager.strictSSL = false;
config.imageManager.ca = "";
config.computeManager = {};
config.computeManager.ip = "127.0.0.1";
config.computeManager.port = "8774";
config.computeManager.authProtocol = "http";
config.computeManager.apiVersion = ['v1.1', 'v2'];
config.computeManager.strictSSL = false;
config.computeManager.ca = "";
config.identityManager = {};
config.identityManager.ip = "127.0.0.1";
config.identityManager.port = "5000";
config.identityManager.authProtocol = "http";
config.identityManager.apiVersion = ['v3'];
config.identityManager.strictSSL = false;
config.identityManager.ca = "";
config.storageManager = {};
config.storageManager.ip = "127.0.0.1";
config.storageManager.port = "8776";
config.storageManager.authProtocol = "http";
config.storageManager.apiVersion = ['v1'];
config.storageManager.strictSSL = false;
config.storageManager.ca = "";
config.cnfg = {};
config.cnfg.server_ip = ['1.1.1.1','1.1.1.2','1.1.1.3'];
config.cnfg.server_port = "8082";
config.cnfg.authProtocol = "http";
config.cnfg.strictSSL = false;
config.cnfg.ca = "/etc/contrail/ssl/certs/ca-cert.pem";
config.cnfg.statusURL = '/global-system-configs';
config.analytics = {};
config.analytics.server_ip = ['1.1.1.1','1.1.1.2','1.1.1.3'];
config.analytics.server_port = "8086";
config.analytics.authProtocol = "http";
config.analytics.strictSSL = false;
config.analytics.ca = '';
config.analytics.statusURL = '/analytics/uves/bgp-peers';
config.dns = {};
config.dns.server_ip = ['1.1.5.1','1.1.5.2','1.1.5.3'];
config.dns.server_port = '8092';
config.dns.statusURL = '/Snh_PageReq?x=AllEntries%20VdnsServersReq';
config.vcenter = {};
config.vcenter.server_ip = "127.0.0.1";         //vCenter IP
config.vcenter.server_port = "443";                                //Port
config.vcenter.authProtocol = "https";   //http or https
config.vcenter.datacenter = "vcenter";      //datacenter name
config.vcenter.dvsswitch = "vswitch";         //dvsswitch name
config.vcenter.strictSSL = false;                                  //Validate the certificate or ignore
config.vcenter.ca = '';                                            //specify the certificate key file
config.vcenter.wsdl = "/usr/src/contrail/contrail-web-core/webroot/js/vim.wsdl";
config.introspect = {};
config.introspect.ssl = {};
config.introspect.ssl.enabled = false;
config.introspect.ssl.key = '/etc/contrail/ssl/private/server-privkey.pem';
config.introspect.ssl.cert = '/etc/contrail/ssl/certs/server.pem';
config.introspect.ssl.ca = '/etc/contrail/ssl/certs/ca-cert.pem';
config.introspect.ssl.strictSSL = false;
config.jobServer = {};
config.jobServer.server_ip = '127.0.0.1';
config.jobServer.server_port = '3000';
config.files = {};
config.files.download_path = '/tmp';
config.cassandra = {};
config.cassandra.server_ips = ['1.1.2.1','1.1.2.2','1.1.2.3'];
config.cassandra.server_port = '9042';
config.cassandra.enable_edit = false;
config.cassandra.use_ssl = false;
config.cassandra.ca_certs = '/etc/contrail/ssl/certs/ca-cert.pem';
config.kue = {};
config.kue.ui_port = '3002'
config.webui_addresses = ['1.1.7.1'];
config.insecure_access = false;
config.http_port = '8180';
config.https_port = '8143';
config.require_auth = false;
config.node_worker_count = 1;
config.maxActiveJobs = 10;
config.redisDBIndex = 3;
config.CONTRAIL_SERVICE_RETRY_TIME = 300000; //5 minutes
config.redis_server_port = '6380';
config.redis_server_ip = '127.0.0.1';
config.redis_dump_file = '/var/lib/redis/dump-webui.rdb';
config.redis_password = '';
config.logo_file = '/opt/contrail/images/logo.png';
config.favicon_file = '/opt/contrail/images/favicon.ico';
config.featurePkg = {};
config.featurePkg.webController = {};
config.featurePkg.webController.path = '/usr/src/contrail/contrail-web-controller';
config.featurePkg.webController.enable = true;
config.qe = {};
config.qe.enable_stat_queries = false;
config.logs = {};
config.logs.level = 'debug';
config.getDomainProjectsFromApiServer = false;
config.network = {};
config.network.L2_enable = false;
config.getDomainsFromApiServer = false;
config.jsonSchemaPath = "/usr/src/contrail/contrail-web-core/src/serverroot/configJsonSchemas";
config.server_options = {};
config.server_options.key_file = '/etc/contrail/webui_ssl/cs-key.pem';
config.server_options.cert_file = '/etc/contrail/webui_ssl/cs-cert.pem';
config.server_options.ciphers = 'ECDHE-ECDSA-AES256-GCM-SHA384:ECDHE-RSA-AES256-GCM-SHA384:ECDHE-ECDSA-CHACHA20-POLY1305:ECDHE-RSA-CHACHA20-POLY1305:ECDHE-ECDSA-AES128-GCM-SHA256:ECDHE-RSA-AES128-GCM-SHA256:ECDHE-ECDSA-AES256-SHA384:ECDHE-RSA-AES256-SHA384:ECDHE-ECDSA-AES128-SHA256:ECDHE-RSA-AES128-SHA256:AES256-SHA';
module.exports = config;
config.staticAuth = [];
config.staticAuth[0] = {};
config.staticAuth[0].username = 'admin';
config.staticAuth[0].password = 'contrail123';
config.staticAuth[0].roles = ['cloudAdmin'];
`

var configConfigHa = `[DEFAULTS]
listen_ip_addr=1.1.1.1
listen_port=8082
http_server_port=8084
http_server_ip=0.0.0.0
log_file=/var/log/contrail/contrail-api.log
log_level=SYS_NOTICE
log_local=1
list_optimization_enabled=True
auth=noauth
aaa_mode=no-auth
cloud_admin_role=admin
global_read_only_role=
cassandra_server_list=1.1.2.1:9160 1.1.2.2:9160 1.1.2.3:9160
cassandra_use_ssl=false
cassandra_ca_certs=/etc/contrail/ssl/certs/ca-cert.pem
zk_server_ip=1.1.3.1:2181,1.1.3.2:2181,1.1.3.3:2181
rabbit_server=1.1.4.1:5673,1.1.4.2:5673,1.1.4.3:5673
rabbit_vhost=/
rabbit_user=guest
rabbit_password=guest
rabbit_use_ssl=False
rabbit_health_check_interval=10
collectors=1.1.1.1:8086 1.1.1.2:8086 1.1.1.3:8086
[SANDESH]
introspect_ssl_enable=False
sandesh_ssl_enable=False`

var cassandraConfig = `cluster_name: ContrailConfigDB
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
  - seeds: "1.1.2.1"
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
storage_port: 7000
ssl_storage_port: 7001
listen_address: 1.1.2.1
broadcast_address: 1.1.2.1
start_native_transport: true
native_transport_port: 9042
start_rpc: true
rpc_address: 1.1.2.1
rpc_port: 9160
broadcast_rpc_address: 1.1.2.1
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
