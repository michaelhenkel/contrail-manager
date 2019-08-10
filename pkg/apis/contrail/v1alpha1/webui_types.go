package v1alpha1

import (
	"bytes"
	"context"
	"sort"

	configtemplates "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1/templates"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Webui is the Schema for the webuis API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Webui struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   WebuiSpec   `json:"spec,omitempty"`
	Status WebuiStatus `json:"status,omitempty"`
}

// WebuiSpec is the Spec for the cassandras API
// +k8s:openapi-gen=true
type WebuiSpec struct {
	CommonConfiguration  CommonConfiguration `json:"commonConfiguration"`
	ServiceConfiguration WebuiConfiguration  `json:"serviceConfiguration"`
}

// WebuiConfiguration is the Spec for the cassandras API
// +k8s:openapi-gen=true
type WebuiConfiguration struct {
	Images            map[string]string `json:"images"`
	CassandraInstance string            `json:"cassandraInstance,omitempty"`
}

// +k8s:openapi-gen=true
type WebuiStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Active *bool             `json:"active,omitempty"`
	Nodes  map[string]string `json:"nodes,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// WebuiList contains a list of Webui
type WebuiList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Webui `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Webui{}, &WebuiList{})
}

func (c *Webui) InstanceConfiguration(request reconcile.Request,
	podList *corev1.PodList,
	client client.Client) error {
	instanceConfigMapName := request.Name + "-" + "webui" + "-configmap"
	configMapInstanceDynamicConfig := &corev1.ConfigMap{}
	err := client.Get(context.TODO(),
		types.NamespacedName{Name: instanceConfigMapName, Namespace: request.Namespace},
		configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}

	controlNodesInformation, err := NewControlClusterConfiguration("", "master",
		request.Namespace, client)
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
	var podIPList []string
	for _, pod := range podList.Items {
		podIPList = append(podIPList, pod.Status.PodIP)
	}
	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })
	var data = make(map[string]string)
	for idx := range podList.Items {
		var webuiWebConfigBuffer bytes.Buffer
		configtemplates.WebuiWebConfig.Execute(&webuiWebConfigBuffer, struct {
			ListenAddress       string
			Hostname            string
			APIServerList       string
			APIServerPort       string
			AnalyticsServerList string
			AnalyticsServerPort string
			ControlNodeList     string
			DnsNodePort         string
			CassandraServerList string
			CassandraPort       string
			RedisServerList     string
		}{
			ListenAddress:       podList.Items[idx].Status.PodIP,
			Hostname:            podList.Items[idx].Name,
			APIServerList:       configNodesInformation.APIServerListQuotedCommaSeparated,
			APIServerPort:       configNodesInformation.APIServerPort,
			AnalyticsServerList: configNodesInformation.AnalyticsServerListQuotedCommaSeparated,
			AnalyticsServerPort: configNodesInformation.CollectorPort,
			ControlNodeList:     controlNodesInformation.ServerListCommanSeparatedQuoted,
			DnsNodePort:         controlNodesInformation.DNSIntrospectPort,
			CassandraServerList: cassandraNodesInformation.ServerListCommanSeparatedQuoted,
			CassandraPort:       cassandraNodesInformation.CQLPort,
			RedisServerList:     "127.0.0.1",
		})
		data["config.global.js."+podList.Items[idx].Status.PodIP] = webuiWebConfigBuffer.String()
		//fmt.Println("DATA ", data)
		var webuiAuthConfigBuffer bytes.Buffer
		configtemplates.WebuiAuthConfig.Execute(&webuiAuthConfigBuffer, struct {
		}{})
		data["contrail-webui-userauth.js"] = webuiAuthConfigBuffer.String()
	}
	configMapInstanceDynamicConfig.Data = data
	err = client.Update(context.TODO(), configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}
	return nil
}

func (c *Webui) CreateConfigMap(configMapName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (*corev1.ConfigMap, error) {
	return CreateConfigMap(configMapName,
		client,
		scheme,
		request,
		"webui",
		c)
}

func (c *Webui) OwnedByManager(client client.Client, request reconcile.Request) (*Manager, error) {
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
func (c *Webui) IsActive(name string, namespace string, client client.Client) bool {
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

func (c *Webui) PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	return PrepareIntendedDeployment(instanceDeployment, commonConfiguration, "webui", request, scheme, c)
}

func (c *Webui) AddVolumesToIntendedDeployments(intendedDeployment *appsv1.Deployment, volumeConfigMapMap map[string]string) {
	AddVolumesToIntendedDeployments(intendedDeployment, volumeConfigMapMap)
}

func (c *Webui) CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme, client client.Client, increaseVersion bool) error {
	return CompareIntendedWithCurrentDeployment(intendedDeployment, commonConfiguration, "webui", request, scheme, client, c, increaseVersion)
}

func (c *Webui) PodIPListAndIPMap(request reconcile.Request, client client.Client) (*corev1.PodList, map[string]string, error) {
	return PodIPListAndIPMap("webui", request, client)
}

func (c *Webui) SetPodsToReady(podIPList *corev1.PodList, client client.Client) error {
	return SetPodsToReady(podIPList, client)
}

func (c *Webui) ManageNodeStatus(podNameIPMap map[string]string,
	client client.Client) error {
	c.Status.Nodes = podNameIPMap
	err := client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Webui) SetInstanceActive(client client.Client, statusInterface interface{}, deployment *appsv1.Deployment, request reconcile.Request) error {
	status := statusInterface.(*WebuiStatus)
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

func (c *Webui) IsCassandra(request *reconcile.Request, client client.Client) bool {
	return true
}

func (c *Webui) IsManager(request *reconcile.Request, myclient client.Client) bool {
	managerInstance := &Manager{}
	err := myclient.Get(context.TODO(), request.NamespacedName, managerInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": managerInstance.GetName()})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &WebuiList{}
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

func (c *Webui) IsZookeeper(request *reconcile.Request, client client.Client) bool {
	return true
}

func (c *Webui) IsReplicaset(request *reconcile.Request, instanceType string, client client.Client) bool {
	replicaSet := &appsv1.ReplicaSet{}
	err := client.Get(context.TODO(), request.NamespacedName, replicaSet)
	if err == nil {
		request.Name = replicaSet.Labels[instanceType]
		return true
	}
	return false
}

func (c *Webui) IsConfig(request *reconcile.Request, myclient client.Client) bool {
	configInstance := &Config{}
	err := myclient.Get(context.TODO(), request.NamespacedName, configInstance)
	if err == nil {
		labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": configInstance.Labels["contrail_cluster"]})
		listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
		list := &WebuiList{}
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

func (c *Webui) IsRabbitmq(request *reconcile.Request, client client.Client) bool {
	return true
}

func (c *Webui) ConfigurationParameters() interface{} {
	var configurationMap = make(map[string]string)
	return configurationMap
}
