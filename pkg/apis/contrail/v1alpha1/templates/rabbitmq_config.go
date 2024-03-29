package configtemplates

import "text/template"

//RabbitmqConfig is the template of the Rabbitmq service configuration
var RabbitmqConfig = template.Must(template.New("").Parse(`#!/bin/bash
echo $RABBITMQ_ERLANG_COOKIE > /var/lib/rabbitmq/.erlang.cookie
chmod 0600 /var/lib/rabbitmq/.erlang.cookie
export RABBITMQ_NODENAME=rabbit@${POD_IP}
if [[ $(grep $POD_IP /etc/rabbitmq/0) ]] ; then
  rabbitmq-server
else
  rabbitmqctl --node rabbit@$(cat /etc/rabbitmq/0) ping
  while [[ $? -ne 0 ]]; do
	rabbitmqctl --node rabbit@$(cat /etc/rabbitmq/0) ping
  done
  rabbitmq-server -detached
  rabbitmqctl --node rabbit@$(cat /etc/rabbitmq/0) node_health_check
  while [[ $? -ne 0 ]]; do
	rabbitmqctl --node rabbit@$(cat /etc/rabbitmq/0) node_health_check
  done
  rabbitmqctl stop_app
  sleep 2
  rabbitmqctl join_cluster rabbit@$(cat /etc/rabbitmq/0)
  rabbitmqctl shutdown
  rabbitmq-server
fi
`))
