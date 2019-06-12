package config
	
import(
	appsv1 "k8s.io/api/apps/v1"
	"github.com/ghodss/yaml"
)

var yamlDataconfig= `
apiVersion: extensions/v1beta1
kind: Deployment
metadata:
  name: tfconfig-cluster-1
  namespace: default
spec:
  progressDeadlineSeconds: 600
  replicas: 3
  revisionHistoryLimit: 10
  selector:
    matchLabels:
      app: config
      config_cr: config
  strategy:
    rollingUpdate:
      maxSurge: 25%
      maxUnavailable: 25%
    type: RollingUpdate
  template:
    metadata:
      creationTimestamp: null
      labels:
        app: config
        config_cr: config
    spec:
      containers:
      - envFrom:
        - configMapRef:
            name: tfconfigcmv1
        image: hub.juniper.net/contrail-nightly/contrail-controller-config-api:5.2.0-0.740
        imagePullPolicy: Always
        lifecycle: {}
        name: api
        resources: {}
        securityContext:
          privileged: false
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: tfconfigcmv1
        image: hub.juniper.net/contrail-nightly/contrail-controller-config-devicemgr:5.2.0-0.740
        imagePullPolicy: Always
        lifecycle: {}
        name: devicemanager
        resources: {}
        securityContext:
          privileged: false
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: tfconfigcmv1
        image: hub.juniper.net/contrail-nightly/contrail-controller-config-schema:5.2.0-0.740
        imagePullPolicy: Always
        lifecycle: {}
        name: schematransformer
        resources: {}
        securityContext:
          privileged: false
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: tfconfigcmv1
        image: hub.juniper.net/contrail-nightly/contrail-controller-config-svcmonitor:5.2.0-0.740
        imagePullPolicy: Always
        lifecycle: {}
        name: servicemonitor
        resources: {}
        securityContext:
          privileged: false
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: tfconfigcmv1
        image: hub.juniper.net/contrail-nightly/contrail-analytics-api:5.2.0-0.740
        imagePullPolicy: Always
        lifecycle: {}
        name: analyticsapi
        resources: {}
        securityContext:
          privileged: false
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: tfconfigcmv1
        image: hub.juniper.net/contrail-nightly/contrail-analytics-collector:5.2.0-0.740
        imagePullPolicy: Always
        lifecycle: {}
        name: collector
        resources: {}
        securityContext:
          privileged: false
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: tfconfigcmv1
        image: hub.juniper.net/contrail-nightly/contrail-external-redis:5.2.0-0.740
        imagePullPolicy: Always
        lifecycle: {}
        name: redis
        resources: {}
        securityContext:
          privileged: false
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
        - mountPath: /var/lib/redis
          name: config-data
      - env:
        - name: DOCKER_HOST
          value: unix://mnt/docker.sock
        - name: NODE_TYPE
          value: config
        envFrom:
        - configMapRef:
            name: tfconfigcmv1
        image: hub.juniper.net/contrail-nightly/contrail-nodemgr:5.2.0-0.740
        imagePullPolicy: Always
        lifecycle: {}
        name: nodemanagerconfig
        resources: {}
        securityContext:
          privileged: false
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
        - mountPath: /mnt
          name: docker-unix-socket
      - env:
        - name: DOCKER_HOST
          value: unix://mnt/docker.sock
        - name: NODE_TYPE
          value: analytics
        envFrom:
        - configMapRef:
            name: tfconfigcmv1
        image: hub.juniper.net/contrail-nightly/contrail-nodemgr:5.2.0-0.740
        imagePullPolicy: Always
        lifecycle: {}
        name: nodemanageranalytics
        resources: {}
        securityContext:
          privileged: false
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
        - mountPath: /mnt
          name: docker-unix-socket
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
      - env:
        - name: CONTRAIL_STATUS_IMAGE
          value: hub.juniper.net/contrail-nightly/contrail-status:5.2.0-0.740
        envFrom:
        - configMapRef:
            name: tfconfigcmv1
        image: hub.juniper.net/contrail-nightly/contrail-node-init:5.2.0-0.740
        imagePullPolicy: Always
        name: nodeinit
        resources: {}
        securityContext:
          privileged: true
          procMount: Default
        terminationMessagePath: /dev/termination-log
        terminationMessagePolicy: File
        volumeMounts:
        - mountPath: /host/usr/bin
          name: host-usr-bin
      nodeSelector:
        node-role.kubernetes.io/master: ""
      restartPolicy: Always
      schedulerName: default-scheduler
      securityContext: {}
      terminationGracePeriodSeconds: 30
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      volumes:
      - hostPath:
          path: /var/log/contrail/config
          type: ""
        name: config-logs
      - hostPath:
          path: /var/lib/contrail/config
          type: ""
        name: config-data
      - hostPath:
          path: /var/run
          type: ""
        name: docker-unix-socket
      - hostPath:
          path: /usr/bin
          type: ""
        name: host-usr-bin
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
	err := yaml.Unmarshal([]byte(yamlDataconfig), &deployment)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDataconfig))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &deployment)
	if err != nil {
		panic(err)
	}
	return &deployment
}
	