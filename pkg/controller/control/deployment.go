package control
	
import(
	appsv1 "k8s.io/api/apps/v1"
	"github.com/ghodss/yaml"
)

var yamlDatacontrol= `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: control
  namespace: default
spec:
  replicas: 1
  selector:
    matchLabels:
      app: control
  template:
    metadata:
      labels:
        app: control
        contrail_manager: control
    spec:
      containers:
      - envFrom:
        - configMapRef:
            name: control
        image: docker.io/michaelhenkel/contrail-controller-control-control:5.2.0-dev1
        imagePullPolicy: Always
        name: control
        volumeMounts:
        - mountPath: /var/log/contrail
          name: control-logs
      - envFrom:
        - configMapRef:
            name: control
        image: docker.io/michaelhenkel/contrail-controller-control-dns:5.2.0-dev1
        imagePullPolicy: Always
        name: dns
        volumeMounts:
        - mountPath: /var/log/contrail
          name: control-logs
        - mountPath: /etc/contrail
          name: etc-contrail
      - envFrom:
        - configMapRef:
            name: control
        image: docker.io/michaelhenkel/contrail-controller-control-named:5.2.0-dev1
        imagePullPolicy: Always
        name: named
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /var/log/contrail
          name: control-logs
        - mountPath: /etc/contrail
          name: etc-contrail
      - env:
        - name: NODE_TYPE
          value: control
        - name: DOCKER_HOST
          value: unix://mnt/docker.sock
        envFrom:
        - configMapRef:
            name: tfcontrolcmv1
        image: docker.io/michaelhenkel/contrail-nodemgr:5.2.0-dev1
        imagePullPolicy: Always
        name: nodemanager
        volumeMounts:
        - mountPath: /var/log/contrail
          name: control-logs
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
        envFrom:
        - configMapRef:
            name: control
        image: busybox
        imagePullPolicy: Always
        name: init
        volumeMounts:
        - mountPath: /tmp/podinfo
          name: status
      - env:
        - name: CONTRAIL_STATUS_IMAGE
          value: docker.io/michaelhenkel/contrail-status:5.2.0-dev1
        envFrom:
        - configMapRef:
            name: control
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
      restartPolicy: Always
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      volumes:
      - hostPath:
          path: /var/log/contrail/control
          type: ""
        name: control-logs
      - hostPath:
          path: /var/run
          type: ""
        name: docker-unix-socket
      - hostPath:
          path: /usr/bin
          type: ""
        name: host-usr-bin
      - emptyDir: {}
        name: etc-contrail
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
	err := yaml.Unmarshal([]byte(yamlDatacontrol), &deployment)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDatacontrol))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &deployment)
	if err != nil {
		panic(err)
	}
	return &deployment
}
	