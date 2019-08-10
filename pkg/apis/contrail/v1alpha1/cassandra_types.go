package v1alpha1

import (
	"bytes"
	"context"
	"sort"
	"strconv"

	configtemplates "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1/templates"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Cassandra is the Schema for the cassandras API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Cassandra struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CassandraSpec   `json:"spec,omitempty"`
	Status CassandraStatus `json:"status,omitempty"`
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
	Port           *int              `json:"port,omitempty"`
	CqlPort        *int              `json:"cqlPort,omitempty"`
	SslStoragePort *int              `json:"sslStoragePort,omitempty"`
	StoragePort    *int              `json:"storagePort,omitempty"`
	JmxLocalPort   *int              `json:"jmxLocalPort,omitempty"`
	MaxHeapSize    string            `json:"maxHeapSize,omitempty"`
	MinHeapSize    string            `json:"minHeapSize,omitempty"`
	StartRPC       *bool             `json:"startRpc,omitempty"`
}

// CassandraStatus defines the status of the cassandra object
// +k8s:openapi-gen=true
type CassandraStatus struct {
	Active *bool                `json:"active,omitempty"`
	Nodes  map[string]string    `json:"nodes,omitempty"`
	Ports  CassandraStatusPorts `json:"ports,omitempty"`
}

// CassandraStatusPorts defines the status of the ports of the cassandra object
type CassandraStatusPorts struct {
	Port    string `json:"port,omitempty"`
	CqlPort string `json:"cqlPort,omitempty"`
	JmxPort string `json:"jmxPort,omitempty"`
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

// InstanceConfiguration creates the cassandra instance configuration
func (c *Cassandra) InstanceConfiguration(request reconcile.Request,
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
	sort.SliceStable(podList.Items, func(i, j int) bool { return podList.Items[i].Status.PodIP < podList.Items[j].Status.PodIP })
	cassandraConfigInterface := c.ConfigurationParameters()
	cassandraConfig := cassandraConfigInterface.(CassandraConfiguration)
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
			ClusterName:         cassandraConfig.ClusterName,
			Seeds:               seeds[0],
			StoragePort:         strconv.Itoa(*cassandraConfig.StoragePort),
			SslStoragePort:      strconv.Itoa(*cassandraConfig.SslStoragePort),
			ListenAddress:       podList.Items[idx].Status.PodIP,
			BroadcastAddress:    podList.Items[idx].Status.PodIP,
			CQLPort:             strconv.Itoa(*cassandraConfig.CqlPort),
			StartRPC:            "true",
			RPCPort:             strconv.Itoa(*cassandraConfig.Port),
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
	}
	return nil
}

// CreateConfigMap creates a configmap for cassandra service
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

// OwnedByManager checks of the cassandra object is owned by the Manager
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

// PrepareIntendedDeployment prepares the intented deployment for the Cassandra object
func (c *Cassandra) PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme) (*appsv1.Deployment, error) {
	return PrepareIntendedDeployment(instanceDeployment, commonConfiguration, "cassandra", request, scheme, c)
}

// AddVolumesToIntendedDeployments adds volumes to the Cassandra deployment
func (c *Cassandra) AddVolumesToIntendedDeployments(intendedDeployment *appsv1.Deployment, volumeConfigMapMap map[string]string) {
	AddVolumesToIntendedDeployments(intendedDeployment, volumeConfigMapMap)
}

// CompareIntendedWithCurrentDeployment compares inteded vs actual Cassandra deployment
func (c *Cassandra) CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment, commonConfiguration *CommonConfiguration, request reconcile.Request, scheme *runtime.Scheme, client client.Client, increaseVersion bool) error {
	return CompareIntendedWithCurrentDeployment(intendedDeployment, commonConfiguration, "cassandra", request, scheme, client, c, increaseVersion)
}

// PodIPListAndIPMap get pod ips and names for the Cassandra pods
func (c *Cassandra) PodIPListAndIPMap(request reconcile.Request, client client.Client) (*corev1.PodList, map[string]string, error) {
	return PodIPListAndIPMap("cassandra", request, client)
}

// SetPodsToReady sets Cassandra PODs to ready
func (c *Cassandra) SetPodsToReady(podIPList *corev1.PodList, client client.Client) error {
	return SetPodsToReady(podIPList, client)
}

// ManageNodeStatus manages the status of the Cassandra nodes
func (c *Cassandra) ManageNodeStatus(podNameIPMap map[string]string,
	client client.Client) error {
	c.Status.Nodes = podNameIPMap
	cassandraConfigInterface := c.ConfigurationParameters()
	cassandraConfig := cassandraConfigInterface.(CassandraConfiguration)
	c.Status.Ports.Port = strconv.Itoa(*cassandraConfig.Port)
	c.Status.Ports.CqlPort = strconv.Itoa(*cassandraConfig.CqlPort)
	c.Status.Ports.JmxPort = strconv.Itoa(*cassandraConfig.JmxLocalPort)
	err := client.Status().Update(context.TODO(), c)
	if err != nil {
		return err
	}
	return nil
}

// SetInstanceActive sets the Cassandra instance to active
func (c *Cassandra) SetInstanceActive(client client.Client, statusInterface interface{}, deployment *appsv1.Deployment, request reconcile.Request) error {
	status := statusInterface.(*CassandraStatus)
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

// IsCassandra returns true if instance is cassandra
func (c *Cassandra) IsCassandra(request *reconcile.Request, client client.Client) bool {
	return true
}

// IsManager returns true if instance is manager
func (c *Cassandra) IsManager(request *reconcile.Request, client client.Client) bool {
	return true
}

// IsZookeeper returns true if instance is zookeeper
func (c *Cassandra) IsZookeeper(request *reconcile.Request, client client.Client) bool {
	return true
}

// IsRabbitmq returns true if instance is rabbitmq
func (c *Cassandra) IsRabbitmq(request *reconcile.Request, client client.Client) bool {
	return true
}

// IsReplicaset returns true if instance is replicaset
func (c *Cassandra) IsReplicaset(request *reconcile.Request, instanceType string, client client.Client) bool {
	return true
}

// IsConfig returns true if instance is config
func (c *Cassandra) IsConfig(request *reconcile.Request, client client.Client) bool {
	return true
}

// IsActive returns true if instance is active
func (c *Cassandra) IsActive(name string, namespace string, client client.Client) bool {
	instance := &Cassandra{}
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

// ConfigurationParameters sets the default for the configuration parameters
func (c *Cassandra) ConfigurationParameters() interface{} {
	cassandraConfiguration := CassandraConfiguration{}
	var port int
	var cqlPort int
	var jmxPort int
	var storagePort int
	var sslStoragePort int
	if c.Spec.ServiceConfiguration.Port != nil {
		port = *c.Spec.ServiceConfiguration.Port
	} else {
		port = CassandraPort
	}
	cassandraConfiguration.Port = &port
	if c.Spec.ServiceConfiguration.CqlPort != nil {
		cqlPort = *c.Spec.ServiceConfiguration.CqlPort
	} else {
		cqlPort = CassandraCqlPort
	}
	cassandraConfiguration.CqlPort = &cqlPort
	if c.Spec.ServiceConfiguration.JmxLocalPort != nil {
		jmxPort = *c.Spec.ServiceConfiguration.JmxLocalPort
	} else {
		jmxPort = CassandraJmxLocalPort
	}
	cassandraConfiguration.JmxLocalPort = &jmxPort
	if c.Spec.ServiceConfiguration.StoragePort != nil {
		storagePort = *c.Spec.ServiceConfiguration.StoragePort
	} else {
		storagePort = CassandraStoragePort
	}
	cassandraConfiguration.StoragePort = &storagePort
	if c.Spec.ServiceConfiguration.SslStoragePort != nil {
		sslStoragePort = *c.Spec.ServiceConfiguration.SslStoragePort
	} else {
		sslStoragePort = CassandraSslStoragePort
	}
	cassandraConfiguration.SslStoragePort = &sslStoragePort
	if cassandraConfiguration.ClusterName == "" {
		cassandraConfiguration.ClusterName = "ContrailConfigDB"
	}
	if cassandraConfiguration.ListenAddress == "" {
		cassandraConfiguration.ListenAddress = "auto"
	}
	return cassandraConfiguration
}
