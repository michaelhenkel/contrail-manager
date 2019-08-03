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

// CassandraStatus defines the observed state of Cassandra

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cassandra is the Schema for the cassandras API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraSpec `json:"spec,omitempty"`
	Status Status        `json:"status,omitempty"`
}

// CassandraSpec is the Spec for the cassandras API
// +k8s:openapi-gen=true
type CassandraSpec struct {
	CommonConfiguration  CommonConfiguration    `json:"commonConfiguration"`
	ServiceConfiguration CassandraConfiguration `json:"serviceConfiguration"`
}

// CassandraConfiguration is the Spec for the cassandras API
// +k8s:openapi-gen=true
type CassandraConfiguration struct {
	Images         map[string]string `json:"images"`
	ClusterName    string            `json:"clusterName,omitempty"`
	ListenAddress  string            `json:"listenAddress,omitempty"`
	Port           int               `json:"port,omitempty"`
	CqlPort        int               `json:"cqlPort,omitempty"`
	SslStoragePort int               `json:"sslStoragePort,omitempty"`
	StoragePort    int               `json:"storagePort,omitempty"`
	JmxLocalPort   int               `json:"jmxLocalPort,omitempty"`
	MaxHeapSize    string            `json:"maxHeapSize,omitempty"`
	MinHeapSize    string            `json:"minHeapSize,omitempty"`
	StartRpc       bool              `json:"startRpc,omitempty"`
}

// CassandraList contains a list of Cassandra
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CassandraList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cassandra `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cassandra{}, &CassandraList{})
}

func (c Cassandra) GetCrd() *apiextensionsv1beta1.CustomResourceDefinition {
	return crds.GetCassandraCrd()
}

func (c *Cassandra) CreateInstanceConfiguration(request reconcile.Request,
	podList *corev1.PodList,
	client client.Client) error {
	instanceType := "cassandra"
	instanceConfigMapName := request.Name + "-" + instanceType + "-configmap"
	configMapInstanceDynamicConfig := &corev1.ConfigMap{}
	err := client.Get(context.TODO(),
		types.NamespacedName{Name: instanceConfigMapName, Namespace: request.Namespace},
		configMapInstanceDynamicConfig)
	if err != nil {
		return err
	}
	//currentConfigMap := *configMapInstanceDynamicConfig

	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })

	for idx := range podList.Items {
		var seeds []string
		for idx2 := range podList.Items {
			seeds = append(seeds, podList.Items[idx2].Status.PodIP)
		}
		var cassandraConfigBuffer bytes.Buffer
		configtemplates.CassandraConfig.Execute(&cassandraConfigBuffer, struct {
			ClusterName         string
			Seeds               string
			StoragePort         string
			SslStoragePort      string
			ListenAddress       string
			BroadcastAddress    string
			CQLPort             string
			StartRPC            string
			RPCPort             string
			RPCAddress          string
			RPCBroadcastAddress string
		}{
			ClusterName:         c.Spec.ServiceConfiguration.ClusterName,
			Seeds:               strings.Join(seeds, ","),
			StoragePort:         strconv.Itoa(c.Spec.ServiceConfiguration.StoragePort),
			SslStoragePort:      strconv.Itoa(c.Spec.ServiceConfiguration.SslStoragePort),
			ListenAddress:       podList.Items[idx].Status.PodIP,
			BroadcastAddress:    podList.Items[idx].Status.PodIP,
			CQLPort:             strconv.Itoa(c.Spec.ServiceConfiguration.CqlPort),
			StartRPC:            strconv.FormatBool(c.Spec.ServiceConfiguration.StartRpc),
			RPCPort:             strconv.Itoa(c.Spec.ServiceConfiguration.Port),
			RPCAddress:          podList.Items[idx].Status.PodIP,
			RPCBroadcastAddress: podList.Items[idx].Status.PodIP,
		})
		cassandraConfigString := cassandraConfigBuffer.String()

		if configMapInstanceDynamicConfig.Data == nil {
			data := map[string]string{podList.Items[idx].Status.PodIP + ".yaml": cassandraConfigString}
			configMapInstanceDynamicConfig.Data = data
		} else {
			configMapInstanceDynamicConfig.Data[podList.Items[idx].Status.PodIP+".yaml"] = cassandraConfigString
		}
		err = client.Update(context.TODO(), configMapInstanceDynamicConfig)
		if err != nil {
			return err
		}
		//}

	}
	return nil
}

func (c *Cassandra) CreateConfigMap(configMapName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request) (*corev1.ConfigMap, error) {
	return CreateConfigMap(configMapName,
		client,
		scheme,
		request,
		"cassandra",
		c)
}

func (c *Cassandra) OwnedByManager(client client.Client, request reconcile.Request) (*Manager, error) {
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

func (c *Cassandra) PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	return PrepareIntendedDeployment(instanceDeployment, commonConfiguration, "cassandra", request, scheme, c)
}

func (c *Cassandra) AddVolumesToIntendedDeployments(intendedDeployment *appsv1.Deployment, volumeConfigMapMap map[string]string) {
	AddVolumesToIntendedDeployments(intendedDeployment, volumeConfigMapMap)
}

func (c *Cassandra) CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme, client client.Client) error {
	return CompareIntendedWithCurrentDeployment(intendedDeployment, commonConfiguration, "cassandra", request, scheme, client, c)
}

func (c *Cassandra) GetPodIPListAndIPMap(request reconcile.Request, client client.Client) (*corev1.PodList, map[string]string, error) {
	return GetPodIPListAndIPMap("cassandra", request, client)
}

func (c *Cassandra) SetPodsToReady(podIPList *corev1.PodList, client client.Client) error {
	return SetPodsToReady(podIPList, client)
}

func (c *Cassandra) ManageNodeStatus(podNameIPMap map[string]string,
	client client.Client) error {

	c.Status.Nodes = podNameIPMap
	portMap := map[string]string{"port": strconv.Itoa(c.Spec.ServiceConfiguration.Port),
		"cqlPort": strconv.Itoa(c.Spec.ServiceConfiguration.CqlPort)}

	c.Status.Ports = portMap
	err := client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

func (c *Cassandra) SetInstanceActive(client client.Client, status *Status, deployment *appsv1.Deployment, request reconcile.Request) error {
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
