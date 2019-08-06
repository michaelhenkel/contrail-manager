package configtemplates

import "text/template"

var ControlControlConfig = template.Must(template.New("").Parse(`[DEFAULT]
# bgp_config_file=bgp_config.xml
bgp_port=179
collectors={{ .CollectorServerList }}
# gr_helper_bgp_disable=0
# gr_helper_xmpp_disable=0
hostip={{ .ListenAddress }}
hostname={{ .Hostname }}
http_server_ip=0.0.0.0
http_server_port=8083
log_file=/var/log/contrail/contrail-control.log
log_level=SYS_NOTICE
log_local=1
# log_files_count=10
# log_file_size=10485760 # 10MB
# log_category=
# log_disable=0
xmpp_server_port=5269
xmpp_auth_enable=False
# Sandesh send rate limit can be used to throttle system logs transmitted per
# second. System logs are dropped if the sending rate is exceeded
# sandesh_send_rate_limit=
[CONFIGDB]
config_db_server_list={{ .CassandraServerList }}
# config_db_username=
# config_db_password=
config_db_use_ssl=false
config_db_ca_certs=/etc/contrail/ssl/certs/ca-cert.pem
rabbitmq_server_list={{ .RabbitmqServerList }}
rabbitmq_vhost=/
rabbitmq_user=guest
rabbitmq_password=guest
rabbitmq_use_ssl=False
[SANDESH]
introspect_ssl_enable=False
sandesh_ssl_enable=False`))

var ControlNamedConfig = template.Must(template.New("").Parse(`options {
    directory "/etc/contrail/dns";
    managed-keys-directory "/etc/contrail/dns";
    empty-zones-enable no;
    pid-file "/etc/contrail/dns/contrail-named.pid";
    session-keyfile "/etc/contrail/dns/session.key";
    listen-on port 53 { any; };
    allow-query { any; };
    allow-recursion { any; };
    allow-query-cache { any; };
    max-cache-size 32M;
};
key "rndc-key" {
    algorithm hmac-md5;
    secret "xvysmOR8lnUQRBcunkC6vg==";
};
controls {
    inet 127.0.0.1 port 8094
    allow { 127.0.0.1; }  keys { "rndc-key"; };
};
logging {
    channel debug_log {
        file "/var/log/contrail/contrail-named.log" versions 3 size 5m;
        severity debug;
        print-time yes;
        print-severity yes;
        print-category yes;
    };
    category default {
        debug_log;
    };
    category queries {
        debug_log;
    };
};`))

var ControlDnsConfig = template.Must(template.New("").Parse(`[DEFAULT]
collectors={{ .CollectorServerList }}
named_config_file = /etc/mycontrail/named.{{ .ListenAddress }}
named_config_directory = /etc/contrail/dns
named_log_file = /var/log/contrail/contrail-named.log
rndc_config_file = contrail-rndc.conf
named_max_cache_size=32M # max-cache-size (bytes) per view, can be in K or M
named_max_retransmissions=12
named_retransmission_interval=1000 # msec
hostip={{ .ListenAddress }}
hostname={{ .Hostname }}
http_server_port=8092
http_server_ip=0.0.0.0
dns_server_port=53
log_file=/var/log/contrail/contrail-dns.log
log_level=SYS_NOTICE
log_local=1
# log_files_count=10
# log_file_size=10485760 # 10MB
# log_category=
# log_disable=0
xmpp_dns_auth_enable=False
# Sandesh send rate limit can be used to throttle system logs transmitted per
# second. System logs are dropped if the sending rate is exceeded
# sandesh_send_rate_limit=
[CONFIGDB]
config_db_server_list={{ .CassandraServerList }}
# config_db_username=
# config_db_password=
config_db_use_ssl=false
config_db_ca_certs=/etc/contrail/ssl/certs/ca-cert.pem
rabbitmq_server_list={{ .RabbitmqServerList }}
rabbitmq_vhost=/
rabbitmq_user=guest
rabbitmq_password=guest
rabbitmq_use_ssl=False
[SANDESH]
introspect_ssl_enable=False
sandesh_ssl_enable=False`))

///etc/contrail/contrail-analytics-nodemgr.conf
var ControlNodemanagerConfig = template.Must(template.New("").Parse(`[DEFAULTS]
http_server_ip=0.0.0.0
log_file=/var/log/contrail/contrail-config-nodemgr.log
log_level=SYS_NOTICE
log_local=1
hostip={{ .ListenAddress }}
db_port={{ .CassandraPort }}
db_jmx_port={{ .CassandraJmxPort }}
db_use_ssl=False
db_use_ssl=False
[COLLECTOR]
server_list={{ .CollectorServerList }}
[SANDESH]
introspect_ssl_enable=False
sandesh_ssl_enable=False`))
