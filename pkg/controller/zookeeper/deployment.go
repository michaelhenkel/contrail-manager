package zookeeper
	
import(
	appsv1 "k8s.io/api/apps/v1"
	"github.com/ghodss/yaml"
)

var yamlDatazookeeper= `
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zookeeper
  namespace: default
  labels:
    app: zookeeper
spec:
  replicas: 1
  selector:
    matchLabels:
      app: zookeeper
  template:
    metadata:
      labels:
        app: zookeeper
        contrail_manager: zookeeper
    spec:
      nodeSelector:
        node-role.kubernetes.io/master: ''
      tolerations:
      - operator: Exists
        effect: NoSchedule
      - operator: Exists
        effect: NoExecute
      hostNetwork: true
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
      containers:
      - name: zookeeper
        env:
        - name: POD_IP
          valueFrom:
            fieldRef:
              fieldPath: status.podIP
        image: docker.io/michaelhenkel/contrail-external-zookeeper:5.2.0-dev1
        imagePullPolicy: Always
        resources:
          requests:
            memory: "1Gi"
        readinessProbe:
          exec:
            command:
            - /bin/bash
            - -c
            - "OK=$(echo ruok | nc ${POD_IP} 2181); if [[ ${OK} == \"imok\" ]]; then exit 0; else exit 1;fi"
          initialDelaySeconds: 15
          timeoutSeconds: 5
        livenessProbe:
          exec:
            command:
            - /bin/bash
            - -c
            - "state=$(echo stats |nc ${POD_IP} 2181); if [[ ${state} == \"This ZooKeeper instance is not currently serving requests\" ]]; then exit 1; else exit 0; fi"
          initialDelaySeconds: 30
          timeoutSeconds: 5
        volumeMounts:
        - mountPath: /tmp/conf
          name: conf
        - mountPath: /var/lib/zookeeper
          name: zookeeper-data
        - mountPath: /var/log/zookeeper
          name: zookeeper-logs
      volumes:
      - name: zookeeper-data
        hostPath:
          path: /var/lib/contrail/zookeeper
      - name: zookeeper-logs
        hostPath:
          path: /var/log/contrail/zookeeper
      - downwardAPI:
          defaultMode: 420
          items:
          - fieldRef:
              apiVersion: v1
              fieldPath: metadata.labels
            path: pod_labels
        name: status
      - downwardAPI:
          defaultMode: 420
          items:
          - fieldRef:
              apiVersion: v1
              fieldPath: metadata.labels
            path: zoo.cfg
        name: conf`

func GetDeployment() *appsv1.Deployment{
	deployment := appsv1.Deployment{}
	err := yaml.Unmarshal([]byte(yamlDatazookeeper), &deployment)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDatazookeeper))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &deployment)
	if err != nil {
		panic(err)
	}
	return &deployment
}
	