package v1alpha1

import (
	"bytes"
	"context"
	"net"
	"sort"
	"strconv"

	configtemplates "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1/templates"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// KubemanagerStatus defines the observed state of Kubemanager

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Kubemanager is the Schema for the kubemanagers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Kubemanager struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KubemanagerSpec   `json:"spec,omitempty"`
	Status KubemanagerStatus `json:"status,omitempty"`
}

// KubemanagerSpec is the Spec for the kubemanagers API
// +k8s:openapi-gen=true
type KubemanagerSpec struct {
	CommonConfiguration  CommonConfiguration      `json:"commonConfiguration"`
	ServiceConfiguration KubemanagerConfiguration `json:"serviceConfiguration"`
}

// +k8s:openapi-gen=true
type KubemanagerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Active *bool             `json:"active,omitempty"`
	Nodes  map[string]string `json:"nodes,omitempty"`
}

// KubemanagerConfiguration is the Spec for the kubemanagers API
// +k8s:openapi-gen=true
type KubemanagerConfiguration struct {
	Images                map[string]string `json:"images"`
	CassandraInstance     string            `json:"cassandraInstance,omitempty"`
	ZookeeperInstance     string            `json:"zookeeperInstance,omitempty"`
	UseKubeadmConfig      *bool             `json:"useKubeadmConfig,omitempty"`
	ServiceAccount        string            `json:"serviceAccount,omitempty"`
	ClusterRole           string            `json:"clusterRole,omitempty"`
	ClusterRoleBinding    string            `json:"clusterRoleBinding,omitempty"`
	CloudOrchestrator     string            `json:"cloudOrchestrator,omitempty"`
	KubernetesAPIServer   string            `json:"kubernetesAPIServer,omitempty"`
	KubernetesAPIPort     *int              `json:"kubernetesAPIPort,omitempty"`
	KubernetesAPISSLPort  *int              `json:"kubernetesAPISSLPort,omitempty"`
	PodSubnets            string            `json:"podSubnets,omitempty"`
	ServiceSubnets        string            `json:"serviceSubnets,omitempty"`
	KubernetesClusterName string            `json:"kubernetesClusterName,omitempty"`
	IPFabricSubnets       string            `json:"ipFabricSubnets,omitempty"`
	IPFabricForwarding    *bool             `json:"ipFabricForwarding,omitempty"`
	IPFabricSnat          *bool             `json:"ipFabricSnat,omitempty"`
	KubernetesTokenFile   string            `json:"kubernetesTokenFile,omitempty"`
	HostNetworkService    *bool             `json:"hostNetworkService,omitempty"`
}

// KubemanagerList contains a list of Kubemanager
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type KubemanagerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kubemanager `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kubemanager{}, &KubemanagerList{})
}

func (c *Kubemanager) InstanceConfiguration(request reconcile.Request,
	podList *corev1.PodList,
	client client.Client) error {
	instanceConfigMapName := request.Name + "-" + "kubemanager" + "-configmap"
	configMapInstanceDynamicConfig := &corev1.ConfigMap{}
	err := client.Get(context.TODO(),
		types.NamespacedName{Name: instanceConfigMapName, Namespace: request.Namespace},
		configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}

	cassandraNodesInformation, err := NewCassandraClusterConfiguration(c.Spec.ServiceConfiguration.CassandraInstance,
		request.Namespace, client)
	if err != nil {
		return err
	}

	configNodesInformation, err := NewConfigClusterConfiguration(c.Labels["contrail_cluster"],
		request.Namespace, client)
	if err != nil {
		return err
	}

	zookeeperNodesInformation, err := NewZookeeperClusterConfiguration(c.Spec.ServiceConfiguration.ZookeeperInstance,
		request.Namespace, client)
	if err != nil {
		return err
	}

	rabbitmqNodesInformation, err := NewRabbitmqClusterConfiguration(c.Labels["contrail_cluster"],
		request.Namespace, client)
	if err != nil {
		return err
	}

	var podIPList []string
	for _, pod := range podList.Items {
		podIPList = append(podIPList, pod.Status.PodIP)
	}

	kubemanagerConfigInstance := c.ConfigurationParameters()
	kubemanagerConfig := kubemanagerConfigInstance.(KubemanagerConfiguration)

	if *kubemanagerConfig.UseKubeadmConfig {
		controlPlaneEndpoint := ""
		config, err := rest.InClusterConfig()
		if err == nil {
			clientset, err := kubernetes.NewForConfig(config)
			if err == nil {
				kubeadmConfigMapClient := clientset.CoreV1().ConfigMaps("kube-system")
				kcm, _ := kubeadmConfigMapClient.Get("kubeadm-config", metav1.GetOptions{})
				clusterConfig := kcm.Data["ClusterConfiguration"]
				clusterConfigByte := []byte(clusterConfig)
				clusterConfigMap := make(map[interface{}]interface{})
				err = yaml.Unmarshal(clusterConfigByte, &clusterConfigMap)
				if err != nil {
					return err
				}
				controlPlaneEndpoint = clusterConfigMap["controlPlaneEndpoint"].(string)
				kubernetesAPIServer, kubernetesAPISSLPort, _ := net.SplitHostPort(controlPlaneEndpoint)
				kubemanagerConfig.KubernetesAPIServer = kubernetesAPIServer
				kubernetesAPISSLPortInt, err := strconv.Atoi(kubernetesAPISSLPort)
				if err != nil {
					return err
				}
				kubemanagerConfig.KubernetesAPISSLPort = &kubernetesAPISSLPortInt
				kubemanagerConfig.KubernetesClusterName = clusterConfigMap["clusterName"].(string)
				networkConfig := make(map[interface{}]interface{})
				networkConfig = clusterConfigMap["networking"].(map[interface{}]interface{})
				kubemanagerConfig.PodSubnets = networkConfig["podSubnet"].(string)
				kubemanagerConfig.ServiceSubnets = networkConfig["serviceSubnet"].(string)
			}
		}
	}

	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })
	var data = make(map[string]string)
	for idx := range podList.Items {
		var kubemanagerConfigBuffer bytes.Buffer
		configtemplates.KubemanagerConfig.Execute(&kubemanagerConfigBuffer, struct {
			ListenAddress         string
			CloudOrchestrator     string
			KubernetesAPIServer   string
			KubernetesAPIPort     string
			KubernetesAPISSLPort  string
			KubernetesClusterName string
			PodSubnet             string
			IPFabricSubnet        string
			ServiceSubnet         string
			IPFabricForwarding    string
			IPFabricSnat          string
			APIServerList         string
			APIServerPort         string
			CassandraServerList   string
			ZookeeperServerList   string
			RabbitmqServerList    string
			RabbitmqServerPort    string
			CollectorServerList   string
			HostNetworkService    string
		}{
			ListenAddress:         podList.Items[idx].Status.PodIP,
			CloudOrchestrator:     kubemanagerConfig.CloudOrchestrator,
			KubernetesAPIServer:   kubemanagerConfig.KubernetesAPIServer,
			KubernetesAPIPort:     strconv.Itoa(*kubemanagerConfig.KubernetesAPIPort),
			KubernetesAPISSLPort:  strconv.Itoa(*kubemanagerConfig.KubernetesAPISSLPort),
			KubernetesClusterName: kubemanagerConfig.KubernetesClusterName,
			PodSubnet:             kubemanagerConfig.PodSubnets,
			IPFabricSubnet:        kubemanagerConfig.IPFabricSubnets,
			ServiceSubnet:         kubemanagerConfig.ServiceSubnets,
			IPFabricForwarding:    strconv.FormatBool(*kubemanagerConfig.IPFabricForwarding),
			IPFabricSnat:          strconv.FormatBool(*kubemanagerConfig.IPFabricSnat),
			APIServerList:         configNodesInformation.APIServerListCommaSeparated,
			APIServerPort:         configNodesInformation.APIServerPort,
			CassandraServerList:   cassandraNodesInformation.ServerListSpaceSeparated,
			ZookeeperServerList:   zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:    rabbitmqNodesInformation.ServerListCommaSeparatedWithoutPort,
			RabbitmqServerPort:    rabbitmqNodesInformation.Port,
			CollectorServerList:   configNodesInformation.CollectorServerListSpaceSeparated,
			HostNetworkService:    strconv.FormatBool(*kubemanagerConfig.HostNetworkService),
		})
		data["kubemanager."+podList.Items[idx].Status.PodIP] = kubemanagerConfigBuffer.String()
	}
	configMapInstanceDynamicConfig.Data = data
	err = client.Update(context.TODO(), configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}
	return nil
}

func (c *Kubemanager) CreateConfigMap(configMapName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (*corev1.ConfigMap, error) {
	return CreateConfigMap(configMapName,
		client,
		scheme,
		request,
		"kubemanager",
		c)
}

func (c *Kubemanager) OwnedByManager(client client.Client, request reconcile.Request) (*Manager, error) {
	managerName := c.Labels["contrail_cluster"]
	ownerRefList := c.GetOwnerReferences()
	for _, ownerRef := range ownerRefList {
		if *ownerRef.Controller {
			if ownerRef.Kind == "Manager" {
				managerInstance := &Manager{}
				err := client.Get(context.TODO(), types.NamespacedName{Name: managerName, Namespace: request.Namespace}, managerInstance)
				if err != nil {
					return nil, err
				}
				return managerInstance, nil
			}
		}
	}
	return nil, nil
}

// IsActive returns true if instance is active
func (c *Kubemanager) IsActive(name string, namespace string, client client.Client) bool {
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, c)
	if err != nil {
		return false
	}
	if c.Status.Active != nil {
		if *c.Status.Active {
			return true
		}
	}
	return false
}

func (c *Kubemanager) PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	return PrepareIntendedDeployment(instanceDeployment, commonConfiguration, "kubemanager", request, scheme, c)
}

func (c *Kubemanager) AddVolumesToIntendedDeployments(intendedDeployment *appsv1.Deployment, volumeConfigMapMap map[string]string) {
	AddVolumesToIntendedDeployments(intendedDeployment, volumeConfigMapMap)
}

func (c *Kubemanager) CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme, client client.Client, increaseVersion bool) error {
	return CompareIntendedWithCurrentDeployment(intendedDeployment, commonConfiguration, "kubemanager", request, scheme, client, c, increaseVersion)
}

func (c *Kubemanager) PodIPListAndIPMap(request reconcile.Request, client client.Client) (*corev1.PodList, map[string]string, error) {
	return PodIPListAndIPMap("kubemanager", request, client)
}

func (c *Kubemanager) SetPodsToReady(podIPList *corev1.PodList, client client.Client) error {
	return SetPodsToReady(podIPList, client)
}

func (c *Kubemanager) ManageNodeStatus(podNameIPMap map[string]string,
	client client.Client) error {

	c.Status.Nodes = podNameIPMap
	err := client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Kubemanager) SetInstanceActive(client client.Client, statusInterface interface{}, deployment *appsv1.Deployment, request reconcile.Request) error {
	status := statusInterface.(*KubemanagerStatus)
	err := client.Get(context.TODO(), types.NamespacedName{Name: deployment.Name, Namespace: request.Namespace},
		deployment)
	if err != nil {
		return err
	}
	active := false

	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
		active = true
	}

	status.Active = &active
	err = client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Kubemanager) IsCassandra(request *reconcile.Request, myclient client.Client) bool {
	cassandraInstance := &Cassandra{}
	err := myclient.Get(context.TODO(), request.NamespacedName, cassandraInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": cassandraInstance.Labels["contrail_cluster"]})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &KubemanagerList{}
		err = myclient.List(context.TODO(), listOps, list)
		if err == nil {
			if len(list.Items) > 0 {
				request.Name = list.Items[0].Name
				return true
			}
		}
	}
	return false
}

func (c *Kubemanager) IsManager(request *reconcile.Request, myclient client.Client) bool {
	managerInstance := &Manager{}
	err := myclient.Get(context.TODO(), request.NamespacedName, managerInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": managerInstance.GetName()})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &KubemanagerList{}
		err = myclient.List(context.TODO(), listOps, list)
		if err == nil {
			if len(list.Items) > 0 {
				request.Name = list.Items[0].Name
				return true
			}
		}
	}
	return false
}

func (c *Kubemanager) IsZookeeper(request *reconcile.Request, myclient client.Client) bool {
	zookeeperInstance := &Zookeeper{}
	err := myclient.Get(context.TODO(), request.NamespacedName, zookeeperInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": zookeeperInstance.Labels["contrail_cluster"]})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &KubemanagerList{}
		err = myclient.List(context.TODO(), listOps, list)
		if err == nil {
			if len(list.Items) > 0 {
				request.Name = list.Items[0].Name
				return true
			}
		}
	}
	return false
}

func (c *Kubemanager) IsRabbitmq(request *reconcile.Request, myclient client.Client) bool {
	rabbitmqInstance := &Rabbitmq{}
	err := myclient.Get(context.TODO(), request.NamespacedName, rabbitmqInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": rabbitmqInstance.Labels["contrail_cluster"]})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &KubemanagerList{}
		err = myclient.List(context.TODO(), listOps, list)
		if err == nil {
			if len(list.Items) > 0 {
				request.Name = list.Items[0].Name
				return true
			}
		}
	}
	return false
}

func (c *Kubemanager) IsReplicaset(request *reconcile.Request, instanceType string, client client.Client) bool {
	replicaSet := &appsv1.ReplicaSet{}
	err := client.Get(context.TODO(), request.NamespacedName, replicaSet)
	if err == nil {
		request.Name = replicaSet.Labels[instanceType]
		return true
	}
	return false
}

func (c *Kubemanager) IsConfig(request *reconcile.Request, myclient client.Client) bool {
	configInstance := &Config{}
	err := myclient.Get(context.TODO(), request.NamespacedName, configInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": configInstance.Labels["contrail_cluster"]})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &KubemanagerList{}
		err = myclient.List(context.TODO(), listOps, list)
		if err == nil {
			if len(list.Items) > 0 {
				request.Name = list.Items[0].Name
				return true
			}
		}
	}
	return false
}

func (c *Kubemanager) ConfigurationParameters() interface{} {
	kubemanagerConfiguration := KubemanagerConfiguration{}
	var cloudOrchestrator string
	var kubernetesApiServer string
	var kubernetesApiPort int
	var kubernetesApiSSLPort int
	var kubernetesClusterName string
	var podSubnets string
	var ipFabricSubnets string
	var serviceSubnets string
	var ipFabricForwarding bool
	var ipFabricSnat bool
	var hostNetworkService bool
	var useKubeadmConfig bool

	if c.Spec.ServiceConfiguration.CloudOrchestrator != "" {
		cloudOrchestrator = c.Spec.ServiceConfiguration.CloudOrchestrator
	} else {
		cloudOrchestrator = CloudOrchestrator
	}

	if c.Spec.ServiceConfiguration.KubernetesAPIServer != "" {
		kubernetesApiServer = c.Spec.ServiceConfiguration.KubernetesAPIServer
	} else {
		kubernetesApiServer = KubernetesApiServer
	}

	if c.Spec.ServiceConfiguration.KubernetesAPIPort != nil {
		kubernetesApiPort = *c.Spec.ServiceConfiguration.KubernetesAPIPort
	} else {
		kubernetesApiPort = KubernetesApiPort
	}

	if c.Spec.ServiceConfiguration.KubernetesAPISSLPort != nil {
		kubernetesApiSSLPort = *c.Spec.ServiceConfiguration.KubernetesAPISSLPort
	} else {
		kubernetesApiSSLPort = KubernetesApiSSLPort
	}

	if c.Spec.ServiceConfiguration.KubernetesClusterName != "" {
		kubernetesClusterName = c.Spec.ServiceConfiguration.KubernetesClusterName
	} else {
		kubernetesClusterName = KubernetesClusterName
	}

	if c.Spec.ServiceConfiguration.PodSubnets != "" {
		podSubnets = c.Spec.ServiceConfiguration.PodSubnets
	} else {
		podSubnets = KubernetesPodSubnets
	}

	if c.Spec.ServiceConfiguration.IPFabricSubnets != "" {
		ipFabricSubnets = c.Spec.ServiceConfiguration.IPFabricSubnets
	} else {
		ipFabricSubnets = KubernetesIpFabricSubnets
	}

	if c.Spec.ServiceConfiguration.ServiceSubnets != "" {
		serviceSubnets = c.Spec.ServiceConfiguration.ServiceSubnets
	} else {
		serviceSubnets = KubernetesServiceSubnets
	}

	if c.Spec.ServiceConfiguration.IPFabricForwarding != nil {
		ipFabricForwarding = *c.Spec.ServiceConfiguration.IPFabricForwarding
	} else {
		ipFabricForwarding = KubernetesIPFrabricForwarding
	}

	if c.Spec.ServiceConfiguration.HostNetworkService != nil {
		hostNetworkService = *c.Spec.ServiceConfiguration.HostNetworkService
	} else {
		hostNetworkService = KubernetesHostNetworkService
	}

	if c.Spec.ServiceConfiguration.UseKubeadmConfig != nil {
		useKubeadmConfig = *c.Spec.ServiceConfiguration.UseKubeadmConfig
	} else {
		useKubeadmConfig = KubernetesUseKubeadm
	}

	if c.Spec.ServiceConfiguration.IPFabricSnat != nil {
		ipFabricSnat = *c.Spec.ServiceConfiguration.IPFabricSnat
	} else {
		ipFabricSnat = KubernetesIPFrabricSnat
	}

	kubemanagerConfiguration.CloudOrchestrator = cloudOrchestrator
	kubemanagerConfiguration.KubernetesAPIServer = kubernetesApiServer
	kubemanagerConfiguration.KubernetesAPIPort = &kubernetesApiPort
	kubemanagerConfiguration.KubernetesAPISSLPort = &kubernetesApiSSLPort
	kubemanagerConfiguration.KubernetesClusterName = kubernetesClusterName
	kubemanagerConfiguration.PodSubnets = podSubnets
	kubemanagerConfiguration.IPFabricSubnets = ipFabricSubnets
	kubemanagerConfiguration.ServiceSubnets = serviceSubnets
	kubemanagerConfiguration.IPFabricForwarding = &ipFabricForwarding
	kubemanagerConfiguration.HostNetworkService = &hostNetworkService
	kubemanagerConfiguration.UseKubeadmConfig = &useKubeadmConfig
	kubemanagerConfiguration.IPFabricSnat = &ipFabricSnat

	return kubemanagerConfiguration
}
