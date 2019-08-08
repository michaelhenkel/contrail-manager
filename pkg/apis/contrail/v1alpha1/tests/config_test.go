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

var pod1 = &corev1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "pod1",
		Namespace: "default",
	},
	Status: corev1.PodStatus{
		PodIP: "1.1.1.1",
	},
}

var pod2 = &corev1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "pod2",
		Namespace: "default",
	},
	Status: corev1.PodStatus{
		PodIP: "1.1.1.2",
	},
}

var pod3 = &corev1.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "pod3",
		Namespace: "default",
	},
	Status: corev1.PodStatus{
		PodIP: "1.1.1.3",
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
	podList             corev1.PodList
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

	var configPods = map[string]string{"pod1": "1.1.1.1", "pod2": "1.1.1.2", "pod3": "1.1.1.3"}
	var rabbitmqPods = map[string]string{"pod1": "1.1.4.1", "pod2": "1.1.4.2", "pod3": "1.1.4.3"}
	var cassandraPods = map[string]string{"pod1": "1.1.2.1", "pod2": "1.1.2.2", "pod3": "1.1.2.3"}
	var zookeeperPods = map[string]string{"pod1": "1.1.3.1", "pod2": "1.1.3.2", "pod3": "1.1.3.3"}
	var controlPods = map[string]string{"pod1": "1.1.5.1", "pod2": "1.1.5.2", "pod3": "1.1.5.3"}
	var kubemanagerPods = map[string]string{"pod1": "1.1.6.1", "pod2": "1.1.6.2", "pod3": "1.1.6.3"}
	var webuiPods = map[string]string{"pod1": "1.1.7.1", "pod2": "1.1.7.2", "pod3": "1.1.7.3"}

	configInstance.ManageNodeStatus(configPods, cl)
	rabbitmqInstance.ManageNodeStatus(rabbitmqPods, cl)
	cassandraInstance.ManageNodeStatus(cassandraPods, cl)
	zookeeperInstance.ManageNodeStatus(zookeeperPods, cl)
	controlInstance.ManageNodeStatus(controlPods, cl)
	kubemanagerInstance.ManageNodeStatus(kubemanagerPods, cl)
	webuiInstance.ManageNodeStatus(webuiPods, cl)

	podItems := []corev1.Pod{}
	podItems = append(podItems, *pod1)
	podItems = append(podItems, *pod2)
	podItems = append(podItems, *pod3)
	podList := corev1.PodList{
		Items: podItems,
	}
	environment := Environment{
		client:              &cl,
		podList:             podList,
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
	err := environment.configInstance.CreateInstanceConfiguration(reconcile.Request{types.NamespacedName{Name: "config1", Namespace: "default"}}, &environment.podList, cl)
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
		t.Fatalf("get devicemanager config: \n%v\n", configDiff)
	}
	if configMap.Data["devicemanager.1.1.1.1"] != configDevicemanagerHa {
		devicemanagerDiff := diff.Diff(configMap.Data["devicemanager.1.1.1.1"], configDevicemanagerHa)
		t.Fatalf("get devicemanager config: \n%v\n", devicemanagerDiff)
	}
}

func TestWebuiConfig(t *testing.T) {
	logf.SetLogger(logf.ZapLogger(true))

	environment := SetupEnv()
	cl := *environment.client
	err := environment.webuiInstance.CreateInstanceConfiguration(reconcile.Request{types.NamespacedName{Name: "webui1", Namespace: "default"}}, &environment.podList, cl)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}
	err = cl.Get(context.TODO(),
		types.NamespacedName{Name: "webui1-webui-configmap", Namespace: "default"},
		configMap)
	if err != nil {
		t.Fatalf("get configmap: (%v)", err)
	}
	if configMap.Data["config.global.js.1.1.1.1"] != webuiConfigHa {
		configDiff := diff.Diff(configMap.Data["config.global.js.1.1.1.1"], webuiConfigHa)
		t.Fatalf("get webui config: \n%v\n", configDiff)
	}
}

var webuiConfigSingle = `/*
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
config.cnfg.server_ip = ['1.1.1.1'];
config.cnfg.server_port = "8082";
config.cnfg.authProtocol = "http";
config.cnfg.strictSSL = false;
config.cnfg.ca = "/etc/contrail/ssl/certs/ca-cert.pem";
config.cnfg.statusURL = '/global-system-configs';
config.analytics = {};
config.analytics.server_ip = ['1.1.1.1'];
config.analytics.server_port = "8086";
config.analytics.authProtocol = "http";
config.analytics.strictSSL = false;
config.analytics.ca = '';
config.analytics.statusURL = '/analytics/uves/bgp-peers';
config.dns = {};
config.dns.server_ip = ['1.1.1.1'];
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
config.cassandra.server_ips = ['1.1.1.1'];
config.cassandra.server_port = '9042';
config.cassandra.enable_edit = false;
config.cassandra.use_ssl = false;
config.cassandra.ca_certs = '/etc/contrail/ssl/certs/ca-cert.pem';
config.kue = {};
config.kue.ui_port = '3002'
config.webui_addresses = ['1.1.1.1'];
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
config.dns.server_ip = ['1.1.1.1','1.1.1.2','1.1.1.3'];
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
config.cassandra.server_ips = ['1.1.1.1','1.1.1.2','1.1.1.3'];
config.cassandra.server_port = '9042';
config.cassandra.enable_edit = false;
config.cassandra.use_ssl = false;
config.cassandra.ca_certs = '/etc/contrail/ssl/certs/ca-cert.pem';
config.kue = {};
config.kue.ui_port = '3002'
config.webui_addresses = ['1.1.1.1'];
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

var configDevicemanagerHa = `[DEFAULTS]
host_ip=1.1.1.1
http_server_ip=0.0.0.0
api_server_ip=1.1.1.1,1.1.1.2,1.1.1.3
api_server_port=8082
api_server_use_ssl=False
analytics_server_ip=1.1.1.1,1.1.1.2,1.1.1.3
analytics_server_port=8081
push_mode=1
log_file=/var/log/contrail/contrail-device-manager.log
log_level=SYS_NOTICE
log_local=1
cassandra_server_list=1.1.2.1:9160 1.1.2.2:9160 1.1.2.3:9160
cassandra_use_ssl=false
cassandra_ca_certs=/etc/contrail/ssl/certs/ca-cert.pem
zk_server_ip=1.1.3.1:2181,1.1.3.2:2181,1.1.3.3:2181
# configure directories for job manager
# the same directories must be mounted to dnsmasq and DM container
dnsmasq_conf_dir=/etc/dnsmasq
tftp_dir=/etc/tftp
dhcp_leases_file=/var/lib/dnsmasq/dnsmasq.leases
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
