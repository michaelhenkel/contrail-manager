package configtemplates

import "text/template"

//WebuiWebConfig is the template of the Webui Web service configuration
var WebuiWebConfig = template.Must(template.New("").Parse(`/*
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
config.cnfg.server_ip = [{{ .APIServerList }}];
config.cnfg.server_port = "{{ .APIServerPort }}";
config.cnfg.authProtocol = "http";
config.cnfg.strictSSL = false;
config.cnfg.ca = "/etc/contrail/ssl/certs/ca-cert.pem";
config.cnfg.statusURL = '/global-system-configs';
config.analytics = {};
config.analytics.server_ip = [{{ .AnalyticsServerList }}];
config.analytics.server_port = "{{ .AnalyticsServerPort }}";
config.analytics.authProtocol = "http";
config.analytics.strictSSL = false;
config.analytics.ca = '';
config.analytics.statusURL = '/analytics/uves/bgp-peers';
config.dns = {};
config.dns.server_ip = [{{ .ControlNodeList }}];
config.dns.server_port = '{{ .DnsNodePort }}';
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
config.cassandra.server_ips = [{{ .CassandraServerList }}];
config.cassandra.server_port = '{{ .CassandraPort }}';
config.cassandra.enable_edit = false;
config.cassandra.use_ssl = false;
config.cassandra.ca_certs = '/etc/contrail/ssl/certs/ca-cert.pem';
config.kue = {};
config.kue.ui_port = '3002'
config.webui_addresses = ['{{ .ListenAddress }}'];
config.insecure_access = false;
config.http_port = '8180';
config.https_port = '8143';
config.require_auth = false;
config.node_worker_count = 1;
config.maxActiveJobs = 10;
config.redisDBIndex = 3;
config.CONTRAIL_SERVICE_RETRY_TIME = 300000; //5 minutes
config.redis_server_port = '6380';
config.redis_server_ip = '{{ .RedisServerList }}';
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
`))

//WebuiAuthConfig is the template of the Webui Auth service configuration
var WebuiAuthConfig = template.Must(template.New("").Parse(`/*
* Copyright (c) 2014 Juniper Networks, Inc. All rights reserved.
*/
var auth = {};
auth.admin_user = 'admin';
auth.admin_password = 'contrail123';
auth.admin_token = '';
auth.admin_tenant_name = 'admin';
module.exports = auth;
`))
