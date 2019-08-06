package v1alpha1

import (
	"bytes"
	"context"
	"sort"
	"strconv"
	"strings"

	configtemplates "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1/templates"
	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ControlStatus defines the observed state of Control

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Control is the Schema for the controls API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Control struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ControlSpec `json:"spec,omitempty"`
	Status Status      `json:"status,omitempty"`
}

// ControlSpec is the Spec for the controls API
// +k8s:openapi-gen=true
type ControlSpec struct {
	CommonConfiguration  CommonConfiguration  `json:"commonConfiguration"`
	ServiceConfiguration ControlConfiguration `json:"serviceConfiguration"`
}

// ControlConfiguration is the Spec for the controls API
// +k8s:openapi-gen=true
type ControlConfiguration struct {
	Images            map[string]string `json:"images"`
	CassandraInstance string            `json:"cassandraInstance,omitempty"`
	ZookeeperInstance string            `json:"zookeeperInstance,omitempty"`
}

// ControlList contains a list of Control
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ControlList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Control `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Control{}, &ControlList{})
}

func (c Control) GetCrd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetControlCrd()
}

func (c *Control) CreateInstanceConfiguration(request reconcile.Request,
	podList *corev1.PodList,
	client client.Client) error {
	instanceConfigMapName := request.Name + "-" + "control" + "-configmap"
	configMapInstanceDynamicConfig := &corev1.ConfigMap{}
	err := client.Get(context.TODO(),
		types.NamespacedName{Name: instanceConfigMapName, Namespace: request.Namespace},
		configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}

	cassandraNodes, cassandraPort, cassandraCqlPort, cassandraJmxPort, err := GetCassandraNodes(c.Spec.ServiceConfiguration.CassandraInstance,
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
	var rabbitmqServerCommaSeparatedList, rabbitmqServerSpaceSeparatedList string
	rabbitmqServerCommaSeparatedList = strings.Join(rabbitmqNodes, ":"+strconv.Itoa(rabbitmqPort)+",")
	rabbitmqServerCommaSeparatedList = rabbitmqServerCommaSeparatedList + ":" + strconv.Itoa(rabbitmqPort)
	rabbitmqServerSpaceSeparatedList = strings.Join(rabbitmqNodes, ":"+strconv.Itoa(rabbitmqPort)+" ")
	rabbitmqServerSpaceSeparatedList = rabbitmqServerSpaceSeparatedList + ":" + strconv.Itoa(rabbitmqPort)

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

	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })
	var data = make(map[string]string)
	for idx := range podList.Items {
		var controlControlConfigBuffer bytes.Buffer
		configtemplates.ControlControlConfig.Execute(&controlControlConfigBuffer, struct {
			ListenAddress       string
			Hostname            string
			APIServerList       string
			APIServerPort       string
			CassandraServerList string
			ZookeeperServerList string
			RabbitmqServerList  string
			RabbitmqServerPort  string
			CollectorServerList string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			Hostname:            podList.Items[idx].Name,
			APIServerList:       apiServerList,
			APIServerPort:       "8082",
			CassandraServerList: cassandraCqlServerSpaceSeparateList,
			ZookeeperServerList: zookeeperServerCommaSeparateList,
			RabbitmqServerList:  rabbitmqServerSpaceSeparatedList,
			RabbitmqServerPort:  strconv.Itoa(rabbitmqPort),
			CollectorServerList: collectorServerList,
		})
		data["control."+podList.Items[idx].Status.PodIP] = controlControlConfigBuffer.String()

		var controlNamedConfigBuffer bytes.Buffer
		configtemplates.ControlNamedConfig.Execute(&controlNamedConfigBuffer, struct {
		}{})
		data["named."+podList.Items[idx].Status.PodIP] = controlNamedConfigBuffer.String()

		var controlDnsConfigBuffer bytes.Buffer
		configtemplates.ControlDnsConfig.Execute(&controlDnsConfigBuffer, struct {
			ListenAddress       string
			Hostname            string
			APIServerList       string
			APIServerPort       string
			CassandraServerList string
			ZookeeperServerList string
			RabbitmqServerList  string
			RabbitmqServerPort  string
			CollectorServerList string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			Hostname:            podList.Items[idx].Name,
			APIServerList:       apiServerList,
			APIServerPort:       "8082",
			CassandraServerList: cassandraCqlServerSpaceSeparateList,
			ZookeeperServerList: zookeeperServerCommaSeparateList,
			RabbitmqServerList:  rabbitmqServerSpaceSeparatedList,
			RabbitmqServerPort:  strconv.Itoa(rabbitmqPort),
			CollectorServerList: collectorServerList,
		})
		data["dns."+podList.Items[idx].Status.PodIP] = controlDnsConfigBuffer.String()

		var configNodemanagerBuffer bytes.Buffer
		configtemplates.ControlNodemanagerConfig.Execute(&configNodemanagerBuffer, struct {
			ListenAddress       string
			CollectorServerList string
			CassandraPort       string
			CassandraJmxPort    string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			CollectorServerList: collectorServerList,
			CassandraPort:       strconv.Itoa(cassandraPort),
			CassandraJmxPort:    strconv.Itoa(cassandraJmxPort),
		})
		data["nodemanager."+podList.Items[idx].Status.PodIP] = configNodemanagerBuffer.String()
	}
	configMapInstanceDynamicConfig.Data = data
	err = client.Update(context.TODO(), configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}
	return nil
}

func (c *Control) CreateConfigMap(configMapName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (*corev1.ConfigMap, error) {
	return CreateConfigMap(configMapName,
		client,
		scheme,
		request,
		"control",
		c)
}

func (c *Control) OwnedByManager(client client.Client, request reconcile.Request) (*Manager, error) {
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

func (c *Control) PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	return PrepareIntendedDeployment(instanceDeployment, commonConfiguration, "control", request, scheme, c)
}

func (c *Control) AddVolumesToIntendedDeployments(intendedDeployment *appsv1.Deployment, volumeConfigMapMap map[string]string) {
	AddVolumesToIntendedDeployments(intendedDeployment, volumeConfigMapMap)
}

func (c *Control) CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme, client client.Client, increaseVersion bool) error {
	return CompareIntendedWithCurrentDeployment(intendedDeployment, commonConfiguration, "control", request, scheme, client, c, increaseVersion)
}

func (c *Control) GetPodIPListAndIPMap(request reconcile.Request, client client.Client) (*corev1.PodList, map[string]string, error) {
	return GetPodIPListAndIPMap("control", request, client)
}

func (c *Control) SetPodsToReady(podIPList *corev1.PodList, client client.Client) error {
	return SetPodsToReady(podIPList, client)
}

func (c *Control) ManageNodeStatus(podNameIPMap map[string]string,
	client client.Client) error {

	c.Status.Nodes = podNameIPMap
	err := client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Control) SetInstanceActive(client client.Client, status *Status, deployment *appsv1.Deployment, request reconcile.Request) error {
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

func (c *Control) IsCassandra(request *reconcile.Request, myclient client.Client) bool {
	cassandraInstance := &Cassandra{}
	err := myclient.Get(context.TODO(), request.NamespacedName, cassandraInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": cassandraInstance.Labels["contrail_cluster"]})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &ControlList{}
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

func (c *Control) IsManager(request *reconcile.Request, myclient client.Client) bool {
	managerInstance := &Manager{}
	err := myclient.Get(context.TODO(), request.NamespacedName, managerInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": managerInstance.GetName()})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &ControlList{}
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

func (c *Control) IsZookeeper(request *reconcile.Request, myclient client.Client) bool {
	zookeeperInstance := &Zookeeper{}
	err := myclient.Get(context.TODO(), request.NamespacedName, zookeeperInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": zookeeperInstance.Labels["contrail_cluster"]})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &ControlList{}
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

func (c *Control) IsRabbitmq(request *reconcile.Request, myclient client.Client) bool {
	rabbitmqInstance := &Rabbitmq{}
	err := myclient.Get(context.TODO(), request.NamespacedName, rabbitmqInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": rabbitmqInstance.Labels["contrail_cluster"]})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &ControlList{}
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

func (c *Control) IsReplicaset(request *reconcile.Request, instanceType string, client client.Client) bool {
	replicaSet := &appsv1.ReplicaSet{}
	err := client.Get(context.TODO(), request.NamespacedName, replicaSet)
	if err == nil {
		request.Name = replicaSet.Labels[instanceType]
		return true
	}
	return false
}

func (c *Control) IsConfig(request *reconcile.Request, myclient client.Client) bool {
	configInstance := &Config{}
	err := myclient.Get(context.TODO(), request.NamespacedName, configInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": configInstance.Labels["contrail_cluster"]})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &ControlList{}
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
