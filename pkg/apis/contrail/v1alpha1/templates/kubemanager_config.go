package configtemplates

import "text/template"

//KubemanagerConfig is the template of the Kubemanager service configuration
var KubemanagerConfig = template.Must(template.New("").Parse(`[DEFAULTS]
host_ip={{ .ListenAddress }}
orchestrator={{ .CloudOrchestrator }}
token=eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJpc3MiOiJrdWJlcm5ldGVzL3NlcnZpY2VhY2NvdW50Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9uYW1lc3BhY2UiOiJkZWZhdWx0Iiwia3ViZXJuZXRlcy5pby9zZXJ2aWNlYWNjb3VudC9zZWNyZXQubmFtZSI6ImNvbnRyYWlsLXNlcnZpY2UtYWNjb3VudC10b2tlbi1zYnZidiIsImt1YmVybmV0ZXMuaW8vc2VydmljZWFjY291bnQvc2VydmljZS1hY2NvdW50Lm5hbWUiOiJjb250cmFpbC1zZXJ2aWNlLWFjY291bnQiLCJrdWJlcm5ldGVzLmlvL3NlcnZpY2VhY2NvdW50L3NlcnZpY2UtYWNjb3VudC51aWQiOiI1Mjg5ZjhmNi05ZWJlLTQ5ZDQtYmViOS0wOTExOGEyZTczZTYiLCJzdWIiOiJzeXN0ZW06c2VydmljZWFjY291bnQ6ZGVmYXVsdDpjb250cmFpbC1zZXJ2aWNlLWFjY291bnQifQ.e3mxCPhoy7f34VXcIraYQfbXQOmdTE66_MUyHYLBvbFX-1q92_yg_ojSsHMl9L7PiEITJnd3yZxr1vtNPfZEnEgs2olFYFOSirU9urkIb7MI9sKPf-2wKGfu4rdfQ_ym22SCeIFP4b2hdgvpszsqPUhv1xY4JN9oipGhdThAJQv0wGmX69kO7u5fahwW2e9r4QHR8NPXWFCcRQ2cALkIe85QdkTuzTXi7pKgFAFnWNIxt9HFPtal2l1oVTeLFQQFtEBs7eSZ3p5Aek0YTGXCv0glGXu1saEWjWLn3WEo08ce1NWXWao8T9rEGuZDwjIHVkbGrOrCwSx9DK9yn2eitQ
log_file=/var/log/contrail/contrail-kube-manager.log
log_level=SYS_DEBUG
log_local=1
nested_mode=0
http_server_ip=0.0.0.0
[KUBERNETES]
kubernetes_api_server={{ .KubernetesAPIServer }}
kubernetes_api_port={{ .KubernetesAPIPort }}
kubernetes_api_secure_port={{ .KubernetesAPISSLPort }}
cluster_name={{ .KubernetesClusterName }}
cluster_project={}
cluster_network={}
pod_subnets={{ .PodSubnet }}
ip_fabric_subnets={{ .IPFabricSubnet }}
service_subnets={{ .ServiceSubnet }}
ip_fabric_forwarding={{ .IPFabricForwarding }}
ip_fabric_snat={{ .IPFabricSnat }}
host_network_service={{ .HostNetworkService }}
[VNC]
public_fip_pool={}
vnc_endpoint_ip={{ .APIServerList }}
vnc_endpoint_port={{ .APIServerPort }}
rabbit_server={{ .RabbitmqServerList }}
rabbit_port={{ .RabbitmqServerPort }}
rabbit_vhost=/
rabbit_user=guest
rabbit_password=guest
rabbit_use_ssl=False
rabbit_health_check_interval=10
cassandra_server_list={{ .CassandraServerList }}
cassandra_use_ssl=false
cassandra_ca_certs=/etc/contrail/ssl/certs/ca-cert.pem
collectors={{ .CollectorServerList }}
zk_server_ip={{ .ZookeeperServerList }}
[SANDESH]
introspect_ssl_enable=False
sandesh_ssl_enable=False
`))
