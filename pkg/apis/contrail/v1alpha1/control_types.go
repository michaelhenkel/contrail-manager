package v1alpha1

import (
	"bytes"
	"context"
	"sort"
	"strconv"

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

	Spec   ControlSpec   `json:"spec,omitempty"`
	Status ControlStatus `json:"status,omitempty"`
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
	BGPPort           *int              `json:"bgpPort,omitempty"`
	ASNNumber         *int              `json:"asnNumber,omitempty"`
	XMPPPort          *int              `json:"xmppPort,omitempty"`
	DNSPort           *int              `json:"dnsPort,omitempty"`
	DNSIntrospectPort *int              `json:"dnsIntrospectPort,omitempty"`
}

// +k8s:openapi-gen=true
type ControlStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Active *bool              `json:"active,omitempty"`
	Nodes  map[string]string  `json:"nodes,omitempty"`
	Ports  ControlStatusPorts `json:"ports,omitempty"`
}

type ControlStatusPorts struct {
	BGPPort           string `json:"bgpPort,omitempty"`
	ASNNumber         string `json:"asnNumber,omitempty"`
	XMPPPort          string `json:"xmppPort,omitempty"`
	DNSPort           string `json:"dnsPort,omitempty"`
	DNSIntrospectPort string `json:"dnsIntrospectPort,omitempty"`
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

	cassandraNodesInformation, err := GetCassandraNodes(c.Spec.ServiceConfiguration.CassandraInstance,
		request.Namespace, client)
	if err != nil {
		return err
	}

	zookeeperNodesInformation, err := GetZookeeperNodes(c.Spec.ServiceConfiguration.ZookeeperInstance,
		request.Namespace, client)
	if err != nil {
		return err
	}

	rabbitmqNodesInformation, err := GetRabbitmqNodes(c.Labels["contrail_cluster"],
		request.Namespace, client)
	if err != nil {
		return err
	}

	configNodesInformation, err := GetConfigNodesStatus(c.Labels["contrail_cluster"],
		request.Namespace, client)
	if err != nil {
		return err
	}

	var podIPList []string
	for _, pod := range podList.Items {
		podIPList = append(podIPList, pod.Status.PodIP)
	}

	controlConfigInterface := c.GetConfigurationParameters()
	controlConfig := controlConfigInterface.(ControlConfiguration)

	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })
	var data = make(map[string]string)
	for idx := range podList.Items {
		var controlControlConfigBuffer bytes.Buffer
		configtemplates.ControlControlConfig.Execute(&controlControlConfigBuffer, struct {
			ListenAddress       string
			Hostname            string
			BGPPort             string
			ASNNumber           string
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
			BGPPort:             strconv.Itoa(*controlConfig.BGPPort),
			ASNNumber:           strconv.Itoa(*controlConfig.ASNNumber),
			APIServerList:       configNodesInformation.APIServerListSpaceSeparated,
			APIServerPort:       configNodesInformation.APIServerPort,
			CassandraServerList: cassandraNodesInformation.ServerListCQLSpaceSeparated,
			ZookeeperServerList: zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:  rabbitmqNodesInformation.ServerListSpaceSeparated,
			RabbitmqServerPort:  rabbitmqNodesInformation.Port,
			CollectorServerList: configNodesInformation.CollectorServerListSpaceSeparated,
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
			APIServerList:       configNodesInformation.APIServerListSpaceSeparated,
			APIServerPort:       configNodesInformation.APIServerPort,
			CassandraServerList: cassandraNodesInformation.ServerListCQLSpaceSeparated,
			ZookeeperServerList: zookeeperNodesInformation.ServerListCommaSeparated,
			RabbitmqServerList:  rabbitmqNodesInformation.ServerListSpaceSeparated,
			RabbitmqServerPort:  rabbitmqNodesInformation.Port,
			CollectorServerList: configNodesInformation.CollectorServerListSpaceSeparated,
		})
		data["dns."+podList.Items[idx].Status.PodIP] = controlDnsConfigBuffer.String()

		var controlNodemanagerBuffer bytes.Buffer
		configtemplates.ControlNodemanagerConfig.Execute(&controlNodemanagerBuffer, struct {
			ListenAddress       string
			CollectorServerList string
			CassandraPort       string
			CassandraJmxPort    string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			CollectorServerList: configNodesInformation.CollectorServerListSpaceSeparated,
			CassandraPort:       cassandraNodesInformation.CQLPort,
			CassandraJmxPort:    cassandraNodesInformation.JMXPort,
		})
		data["nodemanager."+podList.Items[idx].Status.PodIP] = controlNodemanagerBuffer.String()

		var controlProvisionBuffer bytes.Buffer
		configtemplates.ControlProvisionConfig.Execute(&controlProvisionBuffer, struct {
			ListenAddress string
			APIServerList string
			ASNNumber     string
			BGPPort       string
			APIServerPort string
		}{
			ListenAddress: podList.Items[idx].Status.PodIP,
			APIServerList: configNodesInformation.APIServerListCommaSeparated,
			ASNNumber:     strconv.Itoa(*controlConfig.ASNNumber),
			BGPPort:       strconv.Itoa(*controlConfig.BGPPort),
			APIServerPort: configNodesInformation.APIServerPort,
		})
		data["provision.sh."+podList.Items[idx].Status.PodIP] = controlProvisionBuffer.String()

		var controlDeProvisionBuffer bytes.Buffer
		configtemplates.ControlDeProvisionConfig.Execute(&controlDeProvisionBuffer, struct {
			User          string
			Password      string
			Tenant        string
			APIServerList string
			APIServerPort string
		}{
			User:          KeystoneAuthAdminUser,
			Password:      KeystoneAuthAdminPassword,
			Tenant:        KeystoneAuthAdminTenant,
			APIServerList: configNodesInformation.APIServerListQuotedCommaSeparated,
			APIServerPort: configNodesInformation.APIServerPort,
		})
		data["deprovision.py."+podList.Items[idx].Status.PodIP] = controlDeProvisionBuffer.String()
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
	controlConfigInterface := c.GetConfigurationParameters()
	controlConfig := controlConfigInterface.(ControlConfiguration)
	c.Status.Ports.BGPPort = strconv.Itoa(*controlConfig.BGPPort)
	c.Status.Ports.ASNNumber = strconv.Itoa(*controlConfig.ASNNumber)
	c.Status.Ports.XMPPPort = strconv.Itoa(*controlConfig.XMPPPort)
	c.Status.Ports.DNSPort = strconv.Itoa(*controlConfig.DNSPort)
	c.Status.Ports.DNSIntrospectPort = strconv.Itoa(*controlConfig.DNSIntrospectPort)
	err := client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Control) SetInstanceActive(client client.Client, statusInterface interface{}, deployment *appsv1.Deployment, request reconcile.Request) error {
	status := statusInterface.(*ControlStatus)
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

func (c *Control) GetConfigurationParameters() interface{} {
	controlConfiguration := ControlConfiguration{}
	var bgpPort int
	var asnNumber int
	var xmppPort int
	var dnsPort int
	if c.Spec.ServiceConfiguration.BGPPort != nil {
		bgpPort = *c.Spec.ServiceConfiguration.BGPPort
	} else {
		bgpPort = BgpPort
	}

	if c.Spec.ServiceConfiguration.ASNNumber != nil {
		asnNumber = *c.Spec.ServiceConfiguration.ASNNumber
	} else {
		asnNumber = BgpAsn
	}

	if c.Spec.ServiceConfiguration.XMPPPort != nil {
		xmppPort = *c.Spec.ServiceConfiguration.XMPPPort
	} else {
		xmppPort = XmppServerPort
	}

	if c.Spec.ServiceConfiguration.DNSPort != nil {
		dnsPort = *c.Spec.ServiceConfiguration.DNSPort
	} else {
		dnsPort = DnsServerPort
	}

	if c.Spec.ServiceConfiguration.DNSIntrospectPort != nil {
		dnsPort = *c.Spec.ServiceConfiguration.DNSIntrospectPort
	} else {
		dnsPort = DnsIntrospectPort
	}
	controlConfiguration.BGPPort = &bgpPort
	controlConfiguration.ASNNumber = &asnNumber
	controlConfiguration.XMPPPort = &xmppPort
	controlConfiguration.DNSPort = &dnsPort
	controlConfiguration.DNSIntrospectPort = &dnsPort

	return controlConfiguration
}
