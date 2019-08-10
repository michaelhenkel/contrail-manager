package v1alpha1

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"

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

// Rabbitmq is the Schema for the rabbitmqs API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Rabbitmq struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RabbitmqSpec   `json:"spec,omitempty"`
	Status RabbitmqStatus `json:"status,omitempty"`
}

// RabbitmqSpec is the Spec for the cassandras API
// +k8s:openapi-gen=true
type RabbitmqSpec struct {
	CommonConfiguration  CommonConfiguration   `json:"commonConfiguration"`
	ServiceConfiguration RabbitmqConfiguration `json:"serviceConfiguration"`
}

// RabbitmqConfiguration is the Spec for the cassandras API
// +k8s:openapi-gen=true
type RabbitmqConfiguration struct {
	Images       map[string]string `json:"images"`
	Port         *int              `json:"port,omitempty"`
	ErlangCookie string            `json:"erlangCookie,omitempty"`
}

// +k8s:openapi-gen=true
type RabbitmqStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Active *bool               `json:"active,omitempty"`
	Nodes  map[string]string   `json:"nodes,omitempty"`
	Ports  RabbitmqStatusPorts `json:"ports,omitempty"`
}

type RabbitmqStatusPorts struct {
	Port string `json:"port,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RabbitmqList contains a list of Rabbitmq
type RabbitmqList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Rabbitmq `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Rabbitmq{}, &RabbitmqList{})
}

func (c *Rabbitmq) InstanceConfiguration(request reconcile.Request,
	podList *corev1.PodList,
	client client.Client) error {
	instanceConfigMapName := request.Name + "-" + "rabbitmq" + "-configmap"
	configMapInstanceDynamicConfig := &corev1.ConfigMap{}
	err := client.Get(context.TODO(),
		types.NamespacedName{Name: instanceConfigMapName, Namespace: request.Namespace},
		configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}
	configMapInstancConfig := &corev1.ConfigMap{}
	err = client.Get(context.TODO(),
		types.NamespacedName{Name: request.Name + "-" + "rabbitmq" + "-configmap-runner", Namespace: request.Namespace},
		configMapInstancConfig)
	if err != nil {
		return err
	}
	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })

	rabbitmqConfigInterface := c.ConfigurationParameters()
	rabbitmqConfig := rabbitmqConfigInterface.(RabbitmqConfiguration)

	rabbitmqConfigString := fmt.Sprintf("listeners.tcp.default = %d\n", *rabbitmqConfig.Port)
	rabbitmqConfigString = rabbitmqConfigString + fmt.Sprintf("loopback_users = none\n")

	data := map[string]string{"rabbitmq.conf": rabbitmqConfigString,
		"RABBITMQ_ERLANG_COOKIE": rabbitmqConfig.ErlangCookie,
		"RABBITMQ_USE_LONGNAME":  "true",
		"RABBITMQ_CONFIG_FILE":   "/etc/rabbitmq/rabbitmq.conf",
		"RABBITMQ_PID_FILE":      "/var/run/rabbitmq.pid",
		"RABBITMQ_CONF_ENV_FILE": "/var/lib/rabbitmq/rabbitmq.env",
	}
	configMapInstanceDynamicConfig.Data = data

	var rabbitmqNodes string
	for idx, pod := range podList.Items {
		configMapInstanceDynamicConfig.Data[strconv.Itoa(idx)] = pod.Status.PodIP
		rabbitmqNodes = rabbitmqNodes + fmt.Sprintf("%s\n", pod.Status.PodIP)
	}
	configMapInstanceDynamicConfig.Data["rabbitmq.nodes"] = rabbitmqNodes
	err = client.Update(context.TODO(), configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}

	var rabbitmqConfigBuffer bytes.Buffer
	configtemplates.RabbitmqConfig.Execute(&rabbitmqConfigBuffer, struct{}{})
	configMapInstancConfig.Data = map[string]string{"run.sh": rabbitmqConfigBuffer.String()}

	err = client.Update(context.TODO(), configMapInstancConfig)
	if err != nil {
		return err
	}

	return nil
}

func (c *Rabbitmq) CreateConfigMap(configMapName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (*corev1.ConfigMap, error) {
	return CreateConfigMap(configMapName,
		client,
		scheme,
		request,
		"rabbitmq",
		c)
}

func (c *Rabbitmq) OwnedByManager(client client.Client, request reconcile.Request) (*Manager, error) {
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
func (c *Rabbitmq) IsActive(name string, namespace string, myclient client.Client) bool {
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": name})
	listOps := &client.ListOptions{Namespace: namespace, LabelSelector: labelSelector}
	rabbitmqList := &RabbitmqList{}
	err := myclient.List(context.TODO(), listOps, rabbitmqList)
	if err != nil {
		return false
	}
	if len(rabbitmqList.Items) > 0 {
		if rabbitmqList.Items[0].Status.Active != nil {
			if *rabbitmqList.Items[0].Status.Active {
				return true
			}
		}
	}
	return false
}

func (c *Rabbitmq) PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	return PrepareIntendedDeployment(instanceDeployment, commonConfiguration, "rabbitmq", request, scheme, c)
}

func (c *Rabbitmq) AddVolumesToIntendedDeployments(intendedDeployment *appsv1.Deployment, volumeConfigMapMap map[string]string) {
	AddVolumesToIntendedDeployments(intendedDeployment, volumeConfigMapMap)
}

func (c *Rabbitmq) CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme, client client.Client, increaseVersion bool) error {
	return CompareIntendedWithCurrentDeployment(intendedDeployment, commonConfiguration, "rabbitmq", request, scheme, client, c, increaseVersion)
}

func (c *Rabbitmq) PodIPListAndIPMap(request reconcile.Request, client client.Client) (*corev1.PodList, map[string]string, error) {
	return PodIPListAndIPMap("rabbitmq", request, client)
}

func (c *Rabbitmq) SetPodsToReady(podIPList *corev1.PodList, client client.Client) error {
	return SetPodsToReady(podIPList, client)
}

func (c *Rabbitmq) ManageNodeStatus(podNameIPMap map[string]string,
	client client.Client) error {
	c.Status.Nodes = podNameIPMap
	rabbitmqConfigInterface := c.ConfigurationParameters()
	rabbitmqConfig := rabbitmqConfigInterface.(RabbitmqConfiguration)
	c.Status.Ports.Port = strconv.Itoa(*rabbitmqConfig.Port)
	err := client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Rabbitmq) SetInstanceActive(client client.Client, statusInterface interface{}, deployment *appsv1.Deployment, request reconcile.Request) error {
	status := statusInterface.(*RabbitmqStatus)
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

func (c *Rabbitmq) ConfigurationParameters() interface{} {
	rabbitmqConfiguration := RabbitmqConfiguration{}
	var port int
	var erlangCookie string
	if c.Spec.ServiceConfiguration.Port != nil {
		port = *c.Spec.ServiceConfiguration.Port
	} else {
		port = RabbitmqNodePort
	}
	if c.Spec.ServiceConfiguration.ErlangCookie != "" {
		erlangCookie = c.Spec.ServiceConfiguration.ErlangCookie
	} else {
		erlangCookie = RabbitmqErlangCookie
	}
	rabbitmqConfiguration.Port = &port
	rabbitmqConfiguration.ErlangCookie = erlangCookie

	return rabbitmqConfiguration
}
