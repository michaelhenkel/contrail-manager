package config
	
import(
	appsv1 "k8s.io/api/apps/v1"
	"github.com/ghodss/yaml"
)

var yamlDataconfig= `
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
  name: config
  namespace: default
spec:
  selector:
    matchLabels:
      app: config
  template:
    metadata:
      labels:
        app: config
    spec:
      containers:
      - envFrom:
        - configMapRef:
            name: config
        image: docker.io/michaelhenkel/contrail-controller-config-api:5.2.0-dev1
        imagePullPolicy: Always
        name: api
        readinessProbe:
          httpGet:
            path: /
            port: 8082
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: config
        image: docker.io/michaelhenkel/contrail-controller-config-devicemgr:5.2.0-dev1
        imagePullPolicy: Always
        name: devicemanager
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: config
        image: docker.io/michaelhenkel/contrail-controller-config-schema:5.2.0-dev1
        imagePullPolicy: Always
        name: schematransformer
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: config
        image: docker.io/michaelhenkel/contrail-controller-config-svcmonitor:5.2.0-dev1
        imagePullPolicy: Always
        name: servicemonitor
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: config
        image: docker.io/michaelhenkel/contrail-analytics-api:5.2.0-dev1
        imagePullPolicy: Always
        name: analyticsapi
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: config
        image: docker.io/michaelhenkel/contrail-analytics-collector:5.2.0-dev1
        imagePullPolicy: Always
        name: collector
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
      - envFrom:
        - configMapRef:
            name: config
        image: docker.io/michaelhenkel/contrail-external-redis:5.2.0-dev1
        imagePullPolicy: Always
        name: redis
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
            name: config
        image: docker.io/michaelhenkel/contrail-nodemgr:5.2.0-dev1
        imagePullPolicy: Always
        name: nodemanagerconfig
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
            name: config
        image: docker.io/michaelhenkel/contrail-nodemgr:5.2.0-dev1
        imagePullPolicy: Always
        name: nodemanageranalytics
        volumeMounts:
        - mountPath: /var/log/contrail
          name: config-logs
        - mountPath: /mnt
          name: docker-unix-socket
      dnsPolicy: ClusterFirst
      hostNetwork: true
      initContainers:
      - command:
        - sh
        - -c
        - until grep ready /tmp/podinfo/pod_labels > /dev/null 2>&1; do sleep 1; done
        env:
        - name: CONTRAIL_STATUS_IMAGE
          value: docker.io/michaelhenkel/contrail-status:5.2.0-dev1
        image: busybox
        imagePullPolicy: Always
        name: init
        envFrom:
        - configMapRef:
            name: config
        volumeMounts:
        - mountPath: /tmp/podinfo
          name: status
      - env:
        - name: CONTRAIL_STATUS_IMAGE
          value: docker.io/michaelhenkel/contrail-status:5.2.0-dev1
        envFrom:
        - configMapRef:
            name: config
        image: docker.io/michaelhenkel/contrail-node-init:5.2.0-dev1
        imagePullPolicy: Always
        name: nodeinit
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /host/usr/bin
          name: host-usr-bin
      nodeSelector:
        node-role.kubernetes.io/master: ""
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
	