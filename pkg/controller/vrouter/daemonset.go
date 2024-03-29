package vrouter
	
import(
	appsv1 "k8s.io/api/apps/v1"
	"github.com/ghodss/yaml"
)

var yamlDatavrouter= `
apiVersion: apps/v1
kind: DaemonSet
metadata:
  name: vrouter
  namespace: default
spec:
  selector:
    matchLabels:
      app: vrouter
  template:
    metadata:
      labels:
        app: vrouter
    spec:
      containers:
      - envFrom:
        - configMapRef:
            name: vrouter
        image: docker.io/michaelhenkel/contrail-vrouter-agent:5.2.0-dev1
        imagePullPolicy: Always
        lifecycle:
          preStop:
            exec:
              command:
              - /clean-up.sh
        name: vrouteragent
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /var/log/contrail
          name: vrouter-logs
        - mountPath: /dev
          name: dev
        - mountPath: /etc/sysconfig/network-scripts
          name: network-scripts
        - mountPath: /host/bin
          name: host-bin
        - mountPath: /usr/src
          name: usr-src
        - mountPath: /lib/modules
          name: lib-modules
        - mountPath: /var/lib/contrail
          name: var-lib-contrail
        - mountPath: /var/contrail/crashes
          name: var-crashes
      - env:
        - name: DOCKER_HOST
          value: unix://mnt/docker.sock
        - name: NODE_TYPE
          value: vrouter
        envFrom:
        - configMapRef:
            name: vrouter
        image: docker.io/michaelhenkel/contrail-nodemgr:5.2.0-dev1
        imagePullPolicy: Always
        name: nodemanager
        volumeMounts:
        - mountPath: /var/log/contrail
          name: vrouter-logs
        - mountPath: /mnt
          name: docker-unix-socket
      dnsPolicy: ClusterFirst
      hostNetwork: true
      initContainers:
      - env:
        - name: CONTRAIL_STATUS_IMAGE
          value: docker.io/michaelhenkel/contrail-status:5.2.0-dev1
        envFrom:
        - configMapRef:
            name: vrouter
        image: docker.io/michaelhenkel/contrail-node-init:5.2.0-dev1
        imagePullPolicy: Always
        name: nodeinit
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /host/usr/bin
          name: host-usr-bin
      - env:
        - name: CONTRAIL_STATUS_IMAGE
          value: docker.io/michaelhenkel/contrail-status:5.2.0-dev1
        envFrom:
        - configMapRef:
            name: vrouter
        image: docker.io/michaelhenkel/contrail-vrouter-kernel-init:5.2.0-dev1
        imagePullPolicy: Always
        name: vrouterkernelinit
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /host/usr/bin
          name: host-usr-bin
        - mountPath: /etc/sysconfig/network-scripts
          name: network-scripts
        - mountPath: /host/bin
          name: host-bin
        - mountPath: /usr/src
          name: usr-src
        - mountPath: /lib/modules
          name: lib-modules
      - env:
        - name: CONTRAIL_STATUS_IMAGE
          value: docker.io/michaelhenkel/contrail-status:5.2.0-dev1
        envFrom:
        - configMapRef:
            name: vrouter
        image: docker.io/michaelhenkel/contrail-kubernetes-cni-init:5.2.0-dev1
        imagePullPolicy: Always
        name: vroutercni
        securityContext:
          privileged: true
        volumeMounts:
        - mountPath: /var/log/contrail
          name: vrouter-logs
        - mountPath: /host/usr/bin
          name: host-usr-bin
        - mountPath: /var/lib/contrail
          name: var-lib-contrail
        - mountPath: /host/etc_cni
          name: etc-cni
        - mountPath: /host/opt_cni_bin
          name: opt-cni-bin
        - mountPath: /host/log_cni
          name: var-log-contrail-cni
      restartPolicy: Always
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      volumes:
      - hostPath:
          path: /var/log/contrail/vrouter
          type: ""
        name: vrouter-logs
      - hostPath:
          path: /var/run
          type: ""
        name: docker-unix-socket
      - hostPath:
          path: /usr/bin
          type: ""
        name: host-usr-bin
      - hostPath:
          path: /var/log/contrail/cni
          type: ""
        name: var-log-contrail-cni
      - hostPath:
          path: /etc/cni
          type: ""
        name: etc-cni
      - hostPath:
          path: /var/contrail/crashes
          type: ""
        name: var-crashes
      - hostPath:
          path: /var/lib/contrail
          type: ""
        name: var-lib-contrail
      - hostPath:
          path: /lib/modules
          type: ""
        name: lib-modules
      - hostPath:
          path: /usr/src
          type: ""
        name: usr-src
      - hostPath:
          path: /bin
          type: ""
        name: host-bin
      - hostPath:
          path: /etc/sysconfig/network-scripts
          type: ""
        name: network-scripts
      - hostPath:
          path: /dev
          type: ""
        name: dev
      - hostPath:
          path: /opt/cni/bin
          type: ""
        name: opt-cni-bin`

func GetDaemonset() *appsv1.DaemonSet{
	daemonSet := appsv1.DaemonSet{}
	err := yaml.Unmarshal([]byte(yamlDatavrouter), &daemonSet)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDatavrouter))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &daemonSet)
	if err != nil {
		panic(err)
	}
	return &daemonSet
}
	