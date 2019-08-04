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
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Config is the Schema for the configs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Config struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ConfigSpec `json:"spec,omitempty"`
	Status Status     `json:"status,omitempty"`
}

// ConfigSpec is the Spec for the cassandras API
// +k8s:openapi-gen=true
type ConfigSpec struct {
	CommonConfiguration  CommonConfiguration `json:"commonConfiguration"`
	ServiceConfiguration ConfigConfiguration `json:"serviceConfiguration"`
}

// ConfigConfiguration is the Spec for the cassandras API
// +k8s:openapi-gen=true
type ConfigConfiguration struct {
	Images            map[string]string `json:"images"`
	Port              int               `json:"port,omitempty"`
	CassandraInstance string            `json:"cassandraInstance,omitempty"`
	ZookeeperInstance string            `json:"zookeeperInstance,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ConfigList contains a list of Config
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func (c Config) GetCrd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetConfigCrd()
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}

func (c *Config) CreateInstanceConfiguration(request reconcile.Request,
	podList *corev1.PodList,
	client client.Client) error {
	instanceConfigMapName := request.Name + "-" + "config" + "-configmap"
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

	var collectorServerList, analyticsServerList, apiServerList, analyticsServerSpaceSeparatedList, apiServerSpaceSeparatedList, redisServerSpaceSeparatedList string
	var podIPList []string
	for _, pod := range podList.Items {
		podIPList = append(podIPList, pod.Status.PodIP)
	}
	collectorServerList = strings.Join(podIPList, ":8086 ")
	collectorServerList = collectorServerList + ":8086"
	analyticsServerList = strings.Join(podIPList, ",")
	apiServerList = strings.Join(podIPList, ",")
	analyticsServerSpaceSeparatedList = strings.Join(podIPList, ":8081 ")
	analyticsServerSpaceSeparatedList = analyticsServerSpaceSeparatedList + ":8081"
	apiServerSpaceSeparatedList = strings.Join(podIPList, ":8082 ")
	apiServerSpaceSeparatedList = apiServerSpaceSeparatedList + ":8082"
	redisServerSpaceSeparatedList = strings.Join(podIPList, ":6379 ")
	redisServerSpaceSeparatedList = redisServerSpaceSeparatedList + ":6379"

	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })
	var data = make(map[string]string)
	for idx := range podList.Items {
		var configApiConfigBuffer bytes.Buffer
		configtemplates.ConfigApiConfig.Execute(&configApiConfigBuffer, struct {
			ListenAddress       string
			ListenPort          string
			CassandraServerList string
			ZookeeperServerList string
			RabbitmqServerList  string
			CollectorServerList string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			ListenPort:          "8082",
			CassandraServerList: cassandraServerSpaceSeparatedList,
			ZookeeperServerList: zookeeperServerCommaSeparateList,
			RabbitmqServerList:  rabbitmqServerCommaSeparatedList,
			CollectorServerList: collectorServerList,
		})
		data["api."+podList.Items[idx].Status.PodIP] = configApiConfigBuffer.String()

		var configDevicemanagerConfigBuffer bytes.Buffer
		configtemplates.ConfigDeviceManagerConfig.Execute(&configDevicemanagerConfigBuffer, struct {
			ListenAddress       string
			ApiServerList       string
			AnalyticsServerList string
			CassandraServerList string
			ZookeeperServerList string
			RabbitmqServerList  string
			CollectorServerList string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			ApiServerList:       apiServerList,
			AnalyticsServerList: analyticsServerList,
			CassandraServerList: cassandraServerSpaceSeparatedList,
			ZookeeperServerList: zookeeperServerCommaSeparateList,
			RabbitmqServerList:  rabbitmqServerCommaSeparatedList,
			CollectorServerList: collectorServerList,
		})
		data["devicemanager."+podList.Items[idx].Status.PodIP] = configDevicemanagerConfigBuffer.String()

		var configSchematransformerConfigBuffer bytes.Buffer
		configtemplates.ConfigSchematransformerConfig.Execute(&configSchematransformerConfigBuffer, struct {
			ListenAddress       string
			ApiServerList       string
			AnalyticsServerList string
			CassandraServerList string
			ZookeeperServerList string
			RabbitmqServerList  string
			CollectorServerList string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			ApiServerList:       apiServerList,
			AnalyticsServerList: analyticsServerList,
			CassandraServerList: cassandraServerSpaceSeparatedList,
			ZookeeperServerList: zookeeperServerCommaSeparateList,
			RabbitmqServerList:  rabbitmqServerCommaSeparatedList,
			CollectorServerList: collectorServerList,
		})
		data["schematransformer."+podList.Items[idx].Status.PodIP] = configSchematransformerConfigBuffer.String()

		var configServicemonitorConfigBuffer bytes.Buffer
		configtemplates.ConfigServicemonitorConfig.Execute(&configServicemonitorConfigBuffer, struct {
			ListenAddress       string
			ApiServerList       string
			AnalyticsServerList string
			CassandraServerList string
			ZookeeperServerList string
			RabbitmqServerList  string
			CollectorServerList string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			ApiServerList:       apiServerList,
			AnalyticsServerList: analyticsServerSpaceSeparatedList,
			CassandraServerList: cassandraServerSpaceSeparatedList,
			ZookeeperServerList: zookeeperServerCommaSeparateList,
			RabbitmqServerList:  rabbitmqServerCommaSeparatedList,
			CollectorServerList: collectorServerList,
		})
		data["servicemonitor."+podList.Items[idx].Status.PodIP] = configServicemonitorConfigBuffer.String()

		var configAnalyticsapiConfigBuffer bytes.Buffer
		configtemplates.ConfigAnalyticsapiConfig.Execute(&configAnalyticsapiConfigBuffer, struct {
			ListenAddress       string
			ApiServerList       string
			AnalyticsServerList string
			CassandraServerList string
			ZookeeperServerList string
			RabbitmqServerList  string
			CollectorServerList string
			RedisServerList     string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			ApiServerList:       apiServerSpaceSeparatedList,
			AnalyticsServerList: analyticsServerSpaceSeparatedList,
			CassandraServerList: cassandraServerSpaceSeparatedList,
			ZookeeperServerList: zookeeperServerSpaceSeparateList,
			RabbitmqServerList:  rabbitmqServerCommaSeparatedList,
			CollectorServerList: collectorServerList,
			RedisServerList:     redisServerSpaceSeparatedList,
		})
		data["analyticsapi."+podList.Items[idx].Status.PodIP] = configAnalyticsapiConfigBuffer.String()

		var configCollectorConfigBuffer bytes.Buffer
		configtemplates.ConfigCollectorConfig.Execute(&configCollectorConfigBuffer, struct {
			ListenAddress       string
			ApiServerList       string
			CassandraServerList string
			ZookeeperServerList string
			RabbitmqServerList  string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			ApiServerList:       apiServerSpaceSeparatedList,
			CassandraServerList: cassandraServerSpaceSeparatedList,
			ZookeeperServerList: zookeeperServerCommaSeparateList,
			RabbitmqServerList:  rabbitmqServerSpaceSeparatedList,
		})
		data["collector."+podList.Items[idx].Status.PodIP] = configCollectorConfigBuffer.String()

		var configNodemanagerconfigConfigBuffer bytes.Buffer
		configtemplates.ConfigNodemanagerConfigConfig.Execute(&configNodemanagerconfigConfigBuffer, struct {
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
		data["nodemanagerconfig."+podList.Items[idx].Status.PodIP] = configNodemanagerconfigConfigBuffer.String()

		var configNodemanageranalyticsConfigBuffer bytes.Buffer
		configtemplates.ConfigNodemanagerAnalyticsConfig.Execute(&configNodemanageranalyticsConfigBuffer, struct {
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
		data["nodemanageranalytics."+podList.Items[idx].Status.PodIP] = configNodemanageranalyticsConfigBuffer.String()
	}
	configMapInstanceDynamicConfig.Data = data
	err = client.Update(context.TODO(), configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) CreateConfigMap(configMapName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (*corev1.ConfigMap, error) {
	return CreateConfigMap(configMapName,
		client,
		scheme,
		request,
		"config",
		c)
}

func (c *Config) OwnedByManager(client client.Client, request reconcile.Request) (*Manager, error) {
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

func (c *Config) PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	return PrepareIntendedDeployment(instanceDeployment, commonConfiguration, "config", request, scheme, c)
}

func (c *Config) AddVolumesToIntendedDeployments(intendedDeployment *appsv1.Deployment, volumeConfigMapMap map[string]string) {
	AddVolumesToIntendedDeployments(intendedDeployment, volumeConfigMapMap)
}

func (c *Config) CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme, client client.Client, increaseVersion bool) error {
	return CompareIntendedWithCurrentDeployment(intendedDeployment, commonConfiguration, "config", request, scheme, client, c, increaseVersion)
}

func (c *Config) GetPodIPListAndIPMap(request reconcile.Request, client client.Client) (*corev1.PodList, map[string]string, error) {
	return GetPodIPListAndIPMap("config", request, client)
}

func (c *Config) SetPodsToReady(podIPList *corev1.PodList, client client.Client) error {
	return SetPodsToReady(podIPList, client)
}

func (c *Config) ManageNodeStatus(podNameIPMap map[string]string,
	client client.Client) error {
	c.Status.Nodes = podNameIPMap
	portMap := map[string]string{"port": strconv.Itoa(c.Spec.ServiceConfiguration.Port)}
	c.Status.Ports = portMap
	err := client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) SetInstanceActive(client client.Client, status *Status, deployment *appsv1.Deployment, request reconcile.Request) error {
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
