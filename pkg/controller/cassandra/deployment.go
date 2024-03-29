package cassandra
	
import(
	appsv1 "k8s.io/api/apps/v1"
	"github.com/ghodss/yaml"
)

var yamlDatacassandra= `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: tfcassandra-cluster-1
  namespace: default
spec:
  replicas: 3
  selector:
    matchLabels:
      app: cassandra
      cassandra_cr: cassandra
  template:
    metadata:
      labels:
        app: cassandra
        cassandra_cr: cassandra
        contrail_manager: cassandra
    spec:
      terminationGracePeriodSeconds: 1800
      containers:
      - image: hub.juniper.net/contrail-nightly/contrail-external-cassandra:5.2.0-0.740
        imagePullPolicy: Always
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        lifecycle:
          preStop:
            exec:
              command: 
              - /bin/sh
              - -c
              - nodetool drain
        readinessProbe:
          exec:
            command:
            - /bin/bash
            - -c
            - "seeds=$(grep -r '  - seeds:' /mydata/${POD_IP}.yaml |awk -F'  - seeds: ' '{print $2}'|tr  ',' ' ') &&  for seed in $(echo $seeds); do if [[ $(nodetool status | grep $seed |awk '{print $1}') != 'UN' ]]; then exit -1; fi; done"
          initialDelaySeconds: 15
          timeoutSeconds: 5
        name: cassandra
        securityContext:
          capabilities:
            add:
              - IPC_LOCK
          privileged: false
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/log/cassandra
          name: cassandra-logs
        - mountPath: /var/lib/cassandra
          name: cassandra-data
      dnsPolicy: ClusterFirst
      hostNetwork: true
      imagePullSecrets:
      - name: contrail
      initContainers:
      - command:
        - sh
        - -c
        - until grep ready /tmp/podinfo/pod_labels > /dev/null 2>&1; do sleep 1; done
        env:
        - name: CONTRAIL_STATUS_IMAGE
          value: hub.juniper.net/contrail-nightly/contrail-status:5.2.0-0.740
        image: busybox
        imagePullPolicy: Always
        name: init
        resources: {}
        securityContext:
          privileged: false
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /tmp/podinfo
          name: status
      nodeSelector:
        node-role.kubernetes.io/master: ""
      restartPolicy: Always
      schedulerName: default-scheduler
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      volumes:
      - hostPath:
          path: /var/log/contrail/cassandra
          type: ""
        name: cassandra-logs
      - hostPath:
          path: /var/lib/contrail/cassandra
          type: ""
        name: cassandra-data
      - downwardAPI:
          defaultMode: 420
          items:
          - fieldRef:
              apiVersion: v1
              fieldPath: metadata.labels
            path: pod_labels
          - fieldRef:
              apiVersion: v1
              fieldPath: metadata.labels
            path: pod_labelsx
        name: status`

func GetDeployment() *appsv1.Deployment{
	deployment := appsv1.Deployment{}
	err := yaml.Unmarshal([]byte(yamlDatacassandra), &deployment)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDatacassandra))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &deployment)
	if err != nil {
		panic(err)
	}
	return &deployment
}
	