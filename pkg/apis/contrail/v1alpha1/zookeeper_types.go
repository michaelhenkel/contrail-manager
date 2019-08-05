package v1alpha1

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"

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

// ZookeeperSpec is the Spec for the zookeepers API
// +k8s:openapi-gen=true
type ZookeeperSpec struct {
	CommonConfiguration  CommonConfiguration    `json:"commonConfiguration"`
	ServiceConfiguration ZookeeperConfiguration `json:"serviceConfiguration"`
}

// ZookeeperConfiguration is the Spec for the zookeepers API
// +k8s:openapi-gen=true
type ZookeeperConfiguration struct {
	Images       map[string]string `json:"images"`
	ClientPort   int               `json:"clientPort,omitempty"`
	ElectionPort int               `json:"electionPort,omitempty"`
	ServerPort   int               `json:"serverPort,omitempty"`
}

// ZookeeperStatus defines the observed state of Zookeeper
// +k8s:openapi-gen=true
type ZookeeperStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Active *bool             `json:"active"`
	Nodes  map[string]string `json:"nodes"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Zookeeper is the Schema for the zookeepers API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Zookeeper struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ZookeeperSpec `json:"spec,omitempty"`
	Status Status        `json:"status,omitempty"`
}

func (z Zookeeper) GetCrd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetZookeeperCrd()
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// ZookeeperList contains a list of Zookeeper
type ZookeeperList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Zookeeper `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Zookeeper{}, &ZookeeperList{})
}

func (c *Zookeeper) CreateInstanceConfiguration(request reconcile.Request,
	podList *corev1.PodList,
	client client.Client) error {
	instanceConfigMapName := request.Name + "-" + "zookeeper" + "-configmap"
	configMapInstanceDynamicConfig := &corev1.ConfigMap{}
	err := client.Get(context.TODO(),
		types.NamespacedName{Name: instanceConfigMapName, Namespace: request.Namespace},
		configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}
	configMapInstancConfig := &corev1.ConfigMap{}
	err = client.Get(context.TODO(),
		types.NamespacedName{Name: instanceConfigMapName + "-1", Namespace: request.Namespace},
		configMapInstancConfig)
	if err != nil {
		return err
	}
	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })
	for idx := range podList.Items {
		if configMapInstanceDynamicConfig.Data == nil {
			data := map[string]string{podList.Items[idx].Status.PodIP: strconv.Itoa(idx + 1)}
			configMapInstanceDynamicConfig.Data = data
		} else {
			configMapInstanceDynamicConfig.Data[podList.Items[idx].Status.PodIP] = strconv.Itoa(idx + 1)
		}
		var zkServerString string
		for idx2 := range podList.Items {
			zkServerString = zkServerString + fmt.Sprintf("server.%d=%s:%s:participant\n",
				idx2+1, podList.Items[idx2].Status.PodIP,
				strconv.Itoa(c.Spec.ServiceConfiguration.ServerPort)+":"+strconv.Itoa(c.Spec.ServiceConfiguration.ElectionPort))
		}
		configMapInstanceDynamicConfig.Data["zoo.cfg.dynamic.100000000"] = zkServerString
		err = client.Update(context.TODO(), configMapInstanceDynamicConfig)
		if err != nil {
			return err
		}
		var zookeeperConfigBuffer, zookeeperLogBuffer, zookeeperXslBuffer, zookeeperAuthBuffer bytes.Buffer

		configtemplates.ZookeeperConfig.Execute(&zookeeperConfigBuffer, struct {
			ClientPort string
		}{
			ClientPort: strconv.Itoa(c.Spec.ServiceConfiguration.ClientPort),
		})
		configtemplates.ZookeeperAuthConfig.Execute(&zookeeperAuthBuffer, struct{}{})
		configtemplates.ZookeeperLogConfig.Execute(&zookeeperLogBuffer, struct{}{})
		configtemplates.ZookeeperXslConfig.Execute(&zookeeperXslBuffer, struct{}{})
		data := map[string]string{"zoo.cfg": zookeeperConfigBuffer.String(),
			"log4j.properties":  zookeeperLogBuffer.String(),
			"configuration.xsl": zookeeperXslBuffer.String(),
			"jaas.conf":         zookeeperAuthBuffer.String()}
		configMapInstancConfig.Data = data

		err = client.Update(context.TODO(), configMapInstancConfig)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *Zookeeper) CreateConfigMap(configMapName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (*corev1.ConfigMap, error) {
	return CreateConfigMap(configMapName,
		client,
		scheme,
		request,
		"zookeeper",
		c)
}

func (c *Zookeeper) OwnedByManager(client client.Client, request reconcile.Request) (*Manager, error) {
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

func (c *Zookeeper) PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	return PrepareIntendedDeployment(instanceDeployment, commonConfiguration, "zookeeper", request, scheme, c)
}

func (c *Zookeeper) AddVolumesToIntendedDeployments(intendedDeployment *appsv1.Deployment, volumeConfigMapMap map[string]string) {
	AddVolumesToIntendedDeployments(intendedDeployment, volumeConfigMapMap)
}

func (c *Zookeeper) CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme, client client.Client, increaseVersion bool) error {
	return CompareIntendedWithCurrentDeployment(intendedDeployment, commonConfiguration, "zookeeper", request, scheme, client, c, increaseVersion)
}

func (c *Zookeeper) GetPodIPListAndIPMap(request reconcile.Request, client client.Client) (*corev1.PodList, map[string]string, error) {
	return GetPodIPListAndIPMap("zookeeper", request, client)
}

func (c *Zookeeper) SetPodsToReady(podIPList *corev1.PodList, client client.Client) error {
	return SetPodsToReady(podIPList, client)
}

func (c *Zookeeper) ManageNodeStatus(podNameIPMap map[string]string,
	client client.Client) error {
	c.Status.Nodes = podNameIPMap
	portMap := map[string]string{"port": strconv.Itoa(c.Spec.ServiceConfiguration.ClientPort)}
	c.Status.Ports = portMap
	err := client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Zookeeper) SetInstanceActive(client client.Client, status *Status, deployment *appsv1.Deployment, request reconcile.Request) error {
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

func (c *Zookeeper) IsCassandra(request *reconcile.Request, client client.Client) bool {
	return true
}

func (c *Zookeeper) IsManager(request *reconcile.Request, client client.Client) bool {
	return true
}

func (c *Zookeeper) IsZookeeper(request *reconcile.Request, client client.Client) bool {
	return true
}

func (c *Zookeeper) IsRabbitmq(request *reconcile.Request, client client.Client) bool {
	return true
}

func (c *Zookeeper) IsReplicaset(request *reconcile.Request, instanceType string, client client.Client) bool {
	return true
}

func (c *Zookeeper) IsConfig(request *reconcile.Request, client client.Client) bool {
	return true
}
