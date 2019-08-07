package v1alpha1

import (
	"bytes"
	"context"
	"net"
	"sort"
	"strconv"
	"strings"

	configtemplates "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1/templates"
	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	"gopkg.in/yaml.v2"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
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

	Spec   KubemanagerSpec `json:"spec,omitempty"`
	Status Status          `json:"status,omitempty"`
}

// KubemanagerSpec is the Spec for the kubemanagers API
// +k8s:openapi-gen=true
type KubemanagerSpec struct {
	CommonConfiguration  CommonConfiguration      `json:"commonConfiguration"`
	ServiceConfiguration KubemanagerConfiguration `json:"serviceConfiguration"`
}

// KubemanagerConfiguration is the Spec for the kubemanagers API
// +k8s:openapi-gen=true
type KubemanagerConfiguration struct {
	Images                map[string]string `json:"images"`
	CassandraInstance     string            `json:"cassandraInstance,omitempty"`
	ZookeeperInstance     string            `json:"zookeeperInstance,omitempty"`
	UseKubeadmConfig      bool              `json:"useKubeadmConfig,omitempty"`
	ServiceAccount        string            `json:"serviceAccount,omitempty"`
	ClusterRole           string            `json:"clusterRole,omitempty"`
	ClusterRoleBinding    string            `json:"clusterRoleBinding,omitempty"`
	CloudOrchestrator     string            `json:"cloudOrchestrator,omitempty"`
	KubernetesAPIServer   string            `json:"kubernetesAPIServer,omitempty"`
	KubernetesAPIPort     *int              `json:"kubernetesAPIPort,omitempty"`
	PodSubnet             string            `json:"podSubnet,omitempty"`
	ServiceSubnet         string            `json:"serviceSubnet,omitempty"`
	KubernetesClusterName string            `json:"kubernetesClusterName,omitempty"`
	IPFabricForwarding    bool              `json:"ipFabricForwarding,omitempty"`
	IPFabricSnat          bool              `json:"ipFabricSnat,omitempty"`
	KubernetesTokenFile   string            `json:"kubernetesTokenFile,omitempty"`
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

func (c Kubemanager) GetCrd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetKubemanagerCrd()
}

func (c *Kubemanager) CreateInstanceConfiguration(request reconcile.Request,
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

	cassandraNodes, cassandraPort, cassandraCqlPort, _, err := GetCassandraNodes(c.Spec.ServiceConfiguration.CassandraInstance,
		request.Namespace, client)
	if err != nil {
		return err
	}
	var cassandraServerSpaceSeparatedList, cassandraCqlServerSpaceSeparateList string
	cassandraServerSpaceSeparatedList = strings.Join(cassandraNodes, ":"+strconv.Itoa(cassandraPort)+" ")
	cassandraServerSpaceSeparatedList = cassandraServerSpaceSeparatedList + ":" + strconv.Itoa(cassandraPort)
	cassandraCqlServerSpaceSeparateList = strings.Join(cassandraNodes, ":"+strconv.Itoa(cassandraCqlPort)+" ")
	cassandraCqlServerSpaceSeparateList = cassandraCqlServerSpaceSeparateList + ":" + strconv.Itoa(cassandraCqlPort)

	zookeeperNodes, zookeeperPort, err := GetZookeeperNodes(c.Spec.ServiceConfiguration.ZookeeperInstance,
		request.Namespace, client)
	if err != nil {
		return err
	}
	var zookeeperServerSpaceSeparateList, zookeeperServerCommaSeparateList string
	zookeeperServerSpaceSeparateList = strings.Join(zookeeperNodes, ":"+strconv.Itoa(zookeeperPort)+" ")
	zookeeperServerSpaceSeparateList = zookeeperServerSpaceSeparateList + ":" + strconv.Itoa(zookeeperPort)
	zookeeperServerCommaSeparateList = strings.Join(zookeeperNodes, ":"+strconv.Itoa(zookeeperPort)+",")
	zookeeperServerCommaSeparateList = zookeeperServerCommaSeparateList + ":" + strconv.Itoa(zookeeperPort)

	rabbitmqNodes, rabbitmqPort, err := GetRabbitmqNodes(c.Labels["contrail_cluster"],
		request.Namespace, client)
	if err != nil {
		return err
	}
	var rabbitmqServerCommaSeparatedList, rabbitmqServerSpaceSeparatedList, rabbitmqServerList string
	rabbitmqServerCommaSeparatedList = strings.Join(rabbitmqNodes, ":"+strconv.Itoa(rabbitmqPort)+",")
	rabbitmqServerCommaSeparatedList = rabbitmqServerCommaSeparatedList + ":" + strconv.Itoa(rabbitmqPort)
	rabbitmqServerSpaceSeparatedList = strings.Join(rabbitmqNodes, ":"+strconv.Itoa(rabbitmqPort)+" ")
	rabbitmqServerSpaceSeparatedList = rabbitmqServerSpaceSeparatedList + ":" + strconv.Itoa(rabbitmqPort)
	rabbitmqServerList = strings.Join(rabbitmqNodes, ",")

	var collectorServerList, apiServerList, analyticsServerSpaceSeparatedList, apiServerSpaceSeparatedList, redisServerSpaceSeparatedList string
	var podIPList []string
	for _, pod := range podList.Items {
		podIPList = append(podIPList, pod.Status.PodIP)
	}
	collectorServerList = strings.Join(podIPList, ":8086 ")
	collectorServerList = collectorServerList + ":8086"
	apiServerList = strings.Join(podIPList, ",")
	analyticsServerSpaceSeparatedList = strings.Join(podIPList, ":8081 ")
	analyticsServerSpaceSeparatedList = analyticsServerSpaceSeparatedList + ":8081"
	apiServerSpaceSeparatedList = strings.Join(podIPList, ":8082 ")
	apiServerSpaceSeparatedList = apiServerSpaceSeparatedList + ":8082"
	apiServerList = strings.Join(podIPList, ",")
	redisServerSpaceSeparatedList = strings.Join(podIPList, ":6379 ")
	redisServerSpaceSeparatedList = redisServerSpaceSeparatedList + ":6379"

	var PodSubnet string
	var KubernetesClusterName string
	var IPFabricSubnet string
	var ServiceSubnet string
	var KubernetesAPIServer string
	var KubernetesAPIPort string

	if c.Spec.ServiceConfiguration.UseKubeadmConfig {
		controlPlaneEndpoint := ""
		KubernetesClusterName = "kubernetes"
		PodSubnet = "10.32.0.0/12"
		ServiceSubnet = "10.96.0.0/12"
		KubernetesAPIServer = "10.96.0.1"
		KubernetesAPIPort = "443"

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
				KubernetesAPIServer, KubernetesAPIPort, _ = net.SplitHostPort(controlPlaneEndpoint)
				KubernetesClusterName = clusterConfigMap["clusterName"].(string)
				networkConfig := make(map[interface{}]interface{})
				networkConfig = clusterConfigMap["networking"].(map[interface{}]interface{})
				PodSubnet = networkConfig["podSubnet"].(string)
				ServiceSubnet = networkConfig["serviceSubnet"].(string)
			}
		}
	}
	if c.Spec.ServiceConfiguration.KubernetesAPIServer != "" {
		KubernetesAPIServer = c.Spec.ServiceConfiguration.KubernetesAPIServer
	}
	if c.Spec.ServiceConfiguration.KubernetesAPIPort != nil {
		KubernetesAPIPort = strconv.Itoa(*c.Spec.ServiceConfiguration.KubernetesAPIPort)
	}
	if c.Spec.ServiceConfiguration.PodSubnet != "" {
		PodSubnet = c.Spec.ServiceConfiguration.PodSubnet
	}
	if c.Spec.ServiceConfiguration.ServiceSubnet != "" {
		ServiceSubnet = c.Spec.ServiceConfiguration.ServiceSubnet
	}
	if c.Spec.ServiceConfiguration.KubernetesClusterName != "" {
		KubernetesClusterName = c.Spec.ServiceConfiguration.KubernetesClusterName
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
		}{
			ListenAddress:         podList.Items[idx].Status.PodIP,
			CloudOrchestrator:     c.Spec.ServiceConfiguration.CloudOrchestrator,
			KubernetesAPIServer:   KubernetesAPIServer,
			KubernetesAPIPort:     KubernetesAPIPort,
			KubernetesClusterName: KubernetesClusterName,
			PodSubnet:             PodSubnet,
			IPFabricSubnet:        IPFabricSubnet,
			ServiceSubnet:         ServiceSubnet,
			IPFabricForwarding:    strconv.FormatBool(c.Spec.ServiceConfiguration.IPFabricForwarding),
			IPFabricSnat:          strconv.FormatBool(c.Spec.ServiceConfiguration.IPFabricSnat),
			APIServerList:         apiServerList,
			APIServerPort:         "8082",
			CassandraServerList:   cassandraServerSpaceSeparatedList,
			ZookeeperServerList:   zookeeperServerCommaSeparateList,
			RabbitmqServerList:    rabbitmqServerList,
			RabbitmqServerPort:    strconv.Itoa(rabbitmqPort),
			CollectorServerList:   collectorServerList,
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

func (c *Kubemanager) PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	return PrepareIntendedDeployment(instanceDeployment, commonConfiguration, "kubemanager", request, scheme, c)
}

func (c *Kubemanager) AddVolumesToIntendedDeployments(intendedDeployment *appsv1.Deployment, volumeConfigMapMap map[string]string) {
	AddVolumesToIntendedDeployments(intendedDeployment, volumeConfigMapMap)
}

func (c *Kubemanager) CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme, client client.Client, increaseVersion bool) error {
	return CompareIntendedWithCurrentDeployment(intendedDeployment, commonConfiguration, "kubemanager", request, scheme, client, c, increaseVersion)
}

func (c *Kubemanager) GetPodIPListAndIPMap(request reconcile.Request, client client.Client) (*corev1.PodList, map[string]string, error) {
	return GetPodIPListAndIPMap("kubemanager", request, client)
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

func (c *Kubemanager) SetInstanceActive(client client.Client, status *Status, deployment *appsv1.Deployment, request reconcile.Request) error {
	err := SetInstanceActive(client, status, deployment, request)
	if err != nil {
		return err
	}
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

func (c *Kubemanager) GetConfigurationParameters() map[string]string {
	var configurationMap = make(map[string]string)
	return configurationMap
}
