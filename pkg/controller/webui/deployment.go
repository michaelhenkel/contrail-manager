package webui
	
import(
	appsv1 "k8s.io/api/apps/v1"
	"github.com/ghodss/yaml"
)

var yamlDatawebui= `
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webui
  namespace: default
spec:
  selector:
    matchLabels:
      app: webui
  template:
    metadata:
      labels:
        app: webui
    spec:
      containers:
      - envFrom:
        - configMapRef:
            name: webui
        image: docker.io/michaelhenkel/contrail-controller-webui-web:5.2.0-dev1
        imagePullPolicy: Always
        name: webuiweb
        volumeMounts:
        - mountPath: /var/log/contrail
          name: webui-logs
      - envFrom:
        - configMapRef:
            name: webui
        image: docker.io/michaelhenkel/contrail-controller-webui-job:5.2.0-dev1
        imagePullPolicy: Always
        name: webuijob
        volumeMounts:
        - mountPath: /var/log/contrail
          name: webui-logs
      dnsPolicy: ClusterFirst
      hostNetwork: true
      initContainers:
      - env:
        - name: CONTRAIL_STATUS_IMAGE
          value: docker.io/michaelhenkel/contrail-status:5.2.0-dev1
        envFrom:
        - configMapRef:
            name: webui
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
          path: /var/log/contrail/webui
          type: ""
        name: webui-logs
      - hostPath:
          path: /usr/bin
          type: ""
        name: host-usr-bin`

func GetDeployment() *appsv1.Deployment{
	deployment := appsv1.Deployment{}
	err := yaml.Unmarshal([]byte(yamlDatawebui), &deployment)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDatawebui))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &deployment)
	if err != nil {
		panic(err)
	}
	return &deployment
}
	