package kubemanager
	
import(
	appsv1 "k8s.io/api/apps/v1"
	"github.com/ghodss/yaml"
)

var yamlDatakubemanager= `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubemanager
  namespace: default
spec:
  selector:
    matchLabels:
      app: kubemanager
  template:
    metadata:
      labels:
        app: kubemanager
    spec:
      containers:
      - envFrom:
        - configMapRef:
            name: kubemanager
        image: docker.io/michaelhenkel/contrail-kubernetes-kube-manager:5.2.0-dev1
        imagePullPolicy: Always
        name: kubemanager
        volumeMounts:
        - mountPath: /var/log/contrail
          name: kubemanager-logs
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
            name: kubemanager
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
            name: kubemanager
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
      serviceAccount: contrail-service-account-kubemanager
      serviceAccountName: contrail-service-account-kubemanager
      tolerations:
      - effect: NoSchedule
        operator: Exists
      - effect: NoExecute
        operator: Exists
      volumes:
      - hostPath:
          path: /var/log/contrail/kubemanager
          type: ""
        name: kubemanager-logs
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
	err := yaml.Unmarshal([]byte(yamlDatakubemanager), &deployment)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDatakubemanager))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &deployment)
	if err != nil {
		panic(err)
	}
	return &deployment
}
	