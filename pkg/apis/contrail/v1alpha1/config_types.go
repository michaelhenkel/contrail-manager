package v1alpha1

import (
	"bytes"
	"context"
	"sort"
	"strconv"
	"strings"

	configtemplates "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1/templates"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
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

	Spec   ConfigSpec   `json:"spec,omitempty"`
	Status ConfigStatus `json:"status,omitempty"`
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
	APIPort           *int              `json:"apiPort,omitempty"`
	AnalyticsPort     *int              `json:"analyticsPort,omitempty"`
	CollectorPort     *int              `json:"collectorPort,omitempty"`
	RedisPort         *int              `json:"redisPort,omitempty"`
	CassandraInstance string            `json:"cassandraInstance,omitempty"`
	ZookeeperInstance string            `json:"zookeeperInstance,omitempty"`
}

// +k8s:openapi-gen=true
type ConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Active *bool             `json:"active,omitempty"`
	Nodes  map[string]string `json:"nodes,omitempty"`
	Ports  ConfigStatusPorts `json:"ports,omitempty"`
}

type ConfigStatusPorts struct {
	APIPort       string `json:"apiPort,omitempty"`
	AnalyticsPort string `json:"analyticsPort,omitempty"`
	CollectorPort string `json:"collectorPort,omitempty"`
	RedisPort     string `json:"redisPort,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// ConfigList contains a list of Config
type ConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Config `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Config{}, &ConfigList{})
}

func (c *Config) InstanceConfiguration(request reconcile.Request,
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

	cassandraNodesInformation, err := NewCassandraClusterConfiguration(c.Spec.ServiceConfiguration.CassandraInstance,
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

	configConfigInterface := c.ConfigurationParameters()
	configConfig := configConfigInterface.(ConfigConfiguration)

	var collectorServerList, analyticsServerList, apiServerList, analyticsServerSpaceSeparatedList, apiServerSpaceSeparatedList, redisServerSpaceSeparatedList string
	var podIPList []string
	for _, pod := range podList.Items {
		podIPList = append(podIPList, pod.Status.PodIP)
	}
	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })
	sort.SliceStable(podIPList, func(i, j int) bool { return podIPList[i] < podIPList[j] })

	collectorServerList = strings.Join(podIPList, ":"+strconv.Itoa(*configConfig.CollectorPort)+" ")
	collectorServerList = collectorServerList + ":" + strconv.Itoa(*configConfig.CollectorPort)
	analyticsServerList = strings.Join(podIPList, ",")
	apiServerList = strings.Join(podIPList, ",")
	analyticsServerSpaceSeparatedList = strings.Join(podIPList, ":"+strconv.Itoa(*configConfig.AnalyticsPort)+" ")
	analyticsServerSpaceSeparatedList = analyticsServerSpaceSeparatedList + ":" + strconv.Itoa(*configConfig.AnalyticsPort)
	apiServerSpaceSeparatedList = strings.Join(podIPList, ":"+strconv.Itoa(*configConfig.APIPort)+" ")
	apiServerSpaceSeparatedList = apiServerSpaceSeparatedList + ":" + strconv.Itoa(*configConfig.APIPort)
	redisServerSpaceSeparatedList = strings.Join(podIPList, ":"+strconv.Itoa(*configConfig.RedisPort)+" ")
	redisServerSpaceSeparatedList = redisServerSpaceSeparatedList + ":" + strconv.Itoa(*configConfig.RedisPort)

	var data = make(map[string]string)
	for idx := range podList.Items {
		var configApiConfigBuffer bytes.Buffer
		configtemplates.ConfigAPIConfig.Execute(&configApiConfigBuffer, struct {
			ListenAddress       string
			ListenPort          string
			CassandraServerList string
			ZookeeperServerList string
			RabbitmqServerList  string
			CollectorServerList string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			ListenPort:          strconv.Itoa(*configConfig.APIPort),
			CassandraServerList: cassandraNodesInformation.ServerListSpaceSeparated,
			ZookeeperServerList: zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:  rabbitmqNodesInformation.ServerListCommaSeparated,
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
			CassandraServerList: cassandraNodesInformation.ServerListSpaceSeparated,
			ZookeeperServerList: zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:  rabbitmqNodesInformation.ServerListCommaSeparated,
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
			CassandraServerList: cassandraNodesInformation.ServerListSpaceSeparated,
			ZookeeperServerList: zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:  rabbitmqNodesInformation.ServerListCommaSeparated,
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
			CassandraServerList: cassandraNodesInformation.ServerListSpaceSeparated,
			ZookeeperServerList: zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:  rabbitmqNodesInformation.ServerListCommaSeparated,
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
			CassandraServerList: cassandraNodesInformation.ServerListSpaceSeparated,
			ZookeeperServerList: zookeeperNodesInformation.ServerListSpaceSeparated,
			RabbitmqServerList:  rabbitmqNodesInformation.ServerListCommaSeparated,
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
			CassandraServerList: cassandraNodesInformation.ServerListCQLSpaceSeparated,
			ZookeeperServerList: zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:  rabbitmqNodesInformation.ServerListSpaceSeparated,
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
			CassandraPort:       cassandraNodesInformation.CQLPort,
			CassandraJmxPort:    cassandraNodesInformation.JMXPort,
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
			CassandraPort:       cassandraNodesInformation.CQLPort,
			CassandraJmxPort:    cassandraNodesInformation.JMXPort,
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

func (c *Config) PodIPListAndIPMap(request reconcile.Request, client client.Client) (*corev1.PodList, map[string]string, error) {
	return PodIPListAndIPMap("config", request, client)
}

func (c *Config) SetPodsToReady(podIPList *corev1.PodList, client client.Client) error {
	return SetPodsToReady(podIPList, client)
}

func (c *Config) ManageNodeStatus(podNameIPMap map[string]string,
	client client.Client) error {
	c.Status.Nodes = podNameIPMap
	configConfigInterface := c.ConfigurationParameters()
	configConfig := configConfigInterface.(ConfigConfiguration)
	c.Status.Ports.APIPort = strconv.Itoa(*configConfig.APIPort)
	c.Status.Ports.AnalyticsPort = strconv.Itoa(*configConfig.AnalyticsPort)
	c.Status.Ports.CollectorPort = strconv.Itoa(*configConfig.CollectorPort)
	c.Status.Ports.RedisPort = strconv.Itoa(*configConfig.RedisPort)
	err := client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) SetInstanceActive(client client.Client, statusInterface interface{}, deployment *appsv1.Deployment, request reconcile.Request) error {
	status := statusInterface.(*ConfigStatus)
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

// IsActive returns true if instance is active
func (c *Config) IsActive(name string, namespace string, client client.Client) bool {
	instance := &Config{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, instance)
	if err != nil {
		return false
	}
	if instance.Status.Active != nil {
		if *instance.Status.Active {
			return true
		}
	}
	return false
}

func (c *Config) ConfigurationParameters() interface{} {
	configConfiguration := ConfigConfiguration{}
	var apiPort int
	var analyticsPort int
	var collectorPort int
	var redisPort int
	if c.Spec.ServiceConfiguration.APIPort != nil {
		apiPort = *c.Spec.ServiceConfiguration.APIPort
	} else {
		apiPort = ConfigApiPort
	}
	configConfiguration.APIPort = &apiPort

	if c.Spec.ServiceConfiguration.AnalyticsPort != nil {
		analyticsPort = *c.Spec.ServiceConfiguration.AnalyticsPort
	} else {
		analyticsPort = AnalyticsApiPort
	}
	configConfiguration.AnalyticsPort = &analyticsPort

	if c.Spec.ServiceConfiguration.CollectorPort != nil {
		collectorPort = *c.Spec.ServiceConfiguration.CollectorPort
	} else {
		collectorPort = CollectorPort
	}
	configConfiguration.CollectorPort = &collectorPort

	if c.Spec.ServiceConfiguration.RedisPort != nil {
		redisPort = *c.Spec.ServiceConfiguration.RedisPort
	} else {
		redisPort = RedisServerPort
	}
	configConfiguration.RedisPort = &redisPort

	return configConfiguration

}
