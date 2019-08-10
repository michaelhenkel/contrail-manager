package v1alpha1

import (
	"context"
	"sort"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// ServiceStatus provides information on the current status of the service
// +k8s:openapi-gen=true
type ServiceStatus struct {
	Name              *string `json:"name,omitempty"`
	Active            *bool   `json:"active,omitempty"`
	Created           *bool   `json:"created,omitempty"`
	ControllerRunning *bool   `json:"controllerRunning,omitempty"`
}

// Status is the status of the service
// +k8s:openapi-gen=true
type Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Active *bool             `json:"active,omitempty"`
	Nodes  map[string]string `json:"nodes,omitempty"`
	Ports  map[string]string `json:"ports,omitempty"`
}

// ServiceInstance is the interface to manage instances
type ServiceInstance interface {
	Get(client.Client, reconcile.Request) error
	Update(client.Client) error
	Create(client.Client) error
	Delete(client.Client) error
}

// CommonConfiguration is the common services struct
// +k8s:openapi-gen=true
type CommonConfiguration struct {
	// Activate defines if the service will be activated by Manager
	Activate *bool `json:"activate,omitempty"`
	// Create defines if the service will be created by Manager
	Create *bool `json:"create,omitempty"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`
	// Host networking requested for this pod. Use the host's network namespace.
	// If this option is set, the ports that will be used must be specified.
	// Default to false.
	// +k8s:conversion-gen=false
	// +optional
	HostNetwork *bool `json:"hostNetwork,omitempty" protobuf:"varint,11,opt,name=hostNetwork"`
	// ImagePullSecrets is an optional list of references to secrets in the same namespace to use for pulling any of the images used by this PodSpec.
	ImagePullSecrets []string `json:"imagePullSecrets,omitempty"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`
}

// ResourceConfiguration is the interface with methods for service configurations
type ResourceConfiguration interface {
	ConfigurationParameters() interface{}
	InstanceConfiguration(reconcile.Request, *corev1.PodList, client.Client) error
	CompareIntendedWithCurrentDeployment(*appsv1.Deployment, *CommonConfiguration, reconcile.Request, *runtime.Scheme, client.Client, bool) error
	PodIPListAndIPMap(reconcile.Request, client.Client) (*corev1.PodList, map[string]string, error)
}

// ResourceIdentification is the interface with the service identification methods
type ResourceIdentification interface {
	OwnedByManager(client.Client, reconcile.Request) (*Manager, error)
}

// ResourceObject is the interface with the object methods
type ResourceObject interface {
	CreateConfigMap(string, client.Client, *runtime.Scheme, reconcile.Request) (*corev1.ConfigMap, error)
	PrepareIntendedDeployment(*appsv1.Deployment, *CommonConfiguration, reconcile.Request, *runtime.Scheme) (*appsv1.Deployment, error)
	AddVolumesToIntendedDeployments(*appsv1.Deployment, map[string]string)
}

// ResourceStatus is the interface with service status methods
type ResourceStatus interface {
	SetPodsToReady(*corev1.PodList, client.Client) error
	ManageNodeStatus(map[string]string, client.Client) error
	SetInstanceActive(client.Client, interface{}, *appsv1.Deployment, reconcile.Request) error
}

//SetPodsToReady sets the status label of a POD to ready
func SetPodsToReady(podList *corev1.PodList, client client.Client) error {
	for _, pod := range podList.Items {
		pod.ObjectMeta.Labels["status"] = "ready"
		err := client.Update(context.TODO(), &pod)
		if err != nil {
			return err
		}
	}
	return nil
}

//CreateConfigMap creates a config map based on the instance type
func CreateConfigMap(configMapName string,
	client client.Client,
	scheme *runtime.Scheme,
	request reconcile.Request,
	instanceType string,
	object v1.Object) (*corev1.ConfigMap, error) {
	configMap := &corev1.ConfigMap{}
	configMap.SetName(configMapName)
	configMap.SetNamespace(request.Namespace)
	configMap.SetLabels(map[string]string{"contrail_manager": instanceType,
		instanceType: request.Name})
	configMap.Data = make(map[string]string)
	err := client.Get(context.TODO(), types.NamespacedName{Name: configMap.Name, Namespace: request.Namespace}, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			err = controllerutil.SetControllerReference(object, configMap, scheme)

			err = client.Create(context.TODO(), configMap)
			if err != nil {
				return nil, err
			}
		}
	}
	return configMap, nil
}

//PrepareIntendedDeployment prepares the intented deployment
func PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment,
	commonConfiguration *CommonConfiguration,
	instanceType string,
	request reconcile.Request,
	scheme *runtime.Scheme,
	object v1.Object) (*appsv1.Deployment, error) {
	instanceDeploymentName := request.Name + "-" + instanceType + "-deployment"
	intendedDeployment := SetDeploymentCommonConfiguration(instanceDeployment, commonConfiguration)
	intendedDeployment.SetName(instanceDeploymentName)
	intendedDeployment.SetNamespace(request.Namespace)
	intendedDeployment.SetLabels(map[string]string{"contrail_manager": instanceType,
		instanceType: request.Name})
	intendedDeployment.Spec.Selector.MatchLabels = map[string]string{"contrail_manager": instanceType,
		instanceType: request.Name}
	intendedDeployment.Spec.Template.SetLabels(map[string]string{"contrail_manager": instanceType,
		instanceType: request.Name})
	intendedDeployment.Spec.Strategy = appsv1.DeploymentStrategy{}
	err := controllerutil.SetControllerReference(object, intendedDeployment, scheme)
	if err != nil {
		return nil, err
	}
	return intendedDeployment, nil
}

//SetDeploymentCommonConfiguration takes common configuration parameters
// and applies it to the deployment
func SetDeploymentCommonConfiguration(deployment *appsv1.Deployment,
	commonConfiguration *CommonConfiguration) *appsv1.Deployment {
	deployment.Spec.Replicas = commonConfiguration.Replicas
	deployment.Spec.Template.Spec.Tolerations = commonConfiguration.Tolerations
	deployment.Spec.Template.Spec.NodeSelector = commonConfiguration.NodeSelector
	if commonConfiguration.HostNetwork != nil {
		deployment.Spec.Template.Spec.HostNetwork = *commonConfiguration.HostNetwork
	} else {
		deployment.Spec.Template.Spec.HostNetwork = false
	}
	if len(commonConfiguration.ImagePullSecrets) > 0 {
		imagePullSecretList := []corev1.LocalObjectReference{}
		for _, imagePullSecretName := range commonConfiguration.ImagePullSecrets {
			imagePullSecret := corev1.LocalObjectReference{
				Name: imagePullSecretName,
			}
			imagePullSecretList = append(imagePullSecretList, imagePullSecret)
		}
		deployment.Spec.Template.Spec.ImagePullSecrets = imagePullSecretList
	}
	return deployment
}

//AddVolumesToIntendedDeployments adds volumes to a deployment
func AddVolumesToIntendedDeployments(intendedDeployment *appsv1.Deployment, volumeConfigMapMap map[string]string) {
	volumeList := intendedDeployment.Spec.Template.Spec.Volumes
	for configMapName, volumeName := range volumeConfigMapMap {
		volume := corev1.Volume{
			Name: volumeName,
			VolumeSource: corev1.VolumeSource{
				ConfigMap: &corev1.ConfigMapVolumeSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: configMapName,
					},
				},
			},
		}
		volumeList = append(volumeList, volume)
	}
	intendedDeployment.Spec.Template.Spec.Volumes = volumeList
}

//CompareIntendedWithCurrentDeployment compares the running deployment with the
//deployment
func CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment,
	commonConfiguration *CommonConfiguration,
	instanceType string,
	request reconcile.Request,
	scheme *runtime.Scheme,
	client client.Client,
	object v1.Object,
	increaseVersion bool) error {
	currentDeployment := &appsv1.Deployment{}
	err := client.Get(context.TODO(),
		types.NamespacedName{Name: intendedDeployment.Name, Namespace: request.Namespace},
		currentDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			intendedDeployment.Spec.Template.ObjectMeta.Labels["version"] = "1"
			err = client.Create(context.TODO(), intendedDeployment)
			if err != nil {
				return err
			}
		}
	} else {
		update := false
		if *intendedDeployment.Spec.Replicas != *currentDeployment.Spec.Replicas {
			update = true
		}
		if increaseVersion {
			intendedDeployment.Spec.Strategy = appsv1.DeploymentStrategy{
				Type: "Recreate",
			}
			versionInt, _ := strconv.Atoi(currentDeployment.Spec.Template.ObjectMeta.Labels["version"])
			newVersion := versionInt + 1
			intendedDeployment.Spec.Template.ObjectMeta.Labels["version"] = strconv.Itoa(newVersion)
		} else {
			intendedDeployment.Spec.Template.ObjectMeta.Labels["version"] = currentDeployment.Spec.Template.ObjectMeta.Labels["version"]
		}

		for _, intendedContainer := range intendedDeployment.Spec.Template.Spec.Containers {
			for _, currentContainer := range currentDeployment.Spec.Template.Spec.Containers {
				if intendedContainer.Name == currentContainer.Name {
					if intendedContainer.Image != currentContainer.Image {
						update = true
					}
				}
			}
		}

		if update {
			err = client.Update(context.TODO(), intendedDeployment)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

//PodIPListAndIPMap gets a list with POD IPs and a map of POD names and IPs
func PodIPListAndIPMap(instanceType string,
	request reconcile.Request,
	reconcileClient client.Client) (*corev1.PodList, map[string]string, error) {
	var podNameIPMap = make(map[string]string)

	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": instanceType,
		instanceType: request.Name})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	replicaSetList := &appsv1.ReplicaSetList{}
	err := reconcileClient.List(context.TODO(), listOps, replicaSetList)
	if err != nil {
		return &corev1.PodList{}, map[string]string{}, err
	}

	if len(replicaSetList.Items) > 0 {
		replicaSet := &appsv1.ReplicaSet{}
		for _, rs := range replicaSetList.Items {
			if *rs.Spec.Replicas > 0 {
				replicaSet = &rs
				break
			}
		}
		if replicaSet.Spec.Replicas != nil {
			podList := &corev1.PodList{}
			if podHash, ok := replicaSet.ObjectMeta.Labels["pod-template-hash"]; ok {
				labelSelector := labels.SelectorFromSet(map[string]string{"pod-template-hash": podHash})
				listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
				err := reconcileClient.List(context.TODO(), listOps, podList)
				if err != nil {
					return &corev1.PodList{}, map[string]string{}, err
				}
			}

			if int32(len(podList.Items)) == *replicaSet.Spec.Replicas {
				for _, pod := range podList.Items {
					if pod.Status.PodIP != "" {
						podNameIPMap[pod.Name] = pod.Status.PodIP
					}
				}
			}
			if int32(len(podNameIPMap)) == *replicaSet.Spec.Replicas {
				return podList, podNameIPMap, nil
			}
		}
	}

	return &corev1.PodList{}, map[string]string{}, nil
}

//NewCassandraClusterConfiguration gets a struct containing various representations of Cassandra nodes string
func NewCassandraClusterConfiguration(name string, namespace string, client client.Client) (*CassandraClusterConfiguration, error) {
	var cassandraNodes []string
	var cassandraCluster CassandraClusterConfiguration
	var port string
	var cqlPort string
	var jmxPort string
	cassandraInstance := &Cassandra{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, cassandraInstance)
	if err != nil {
		return &cassandraCluster, err
	}
	for _, ip := range cassandraInstance.Status.Nodes {
		cassandraNodes = append(cassandraNodes, ip)
	}
	cassandraConfigInterface := cassandraInstance.ConfigurationParameters()
	cassandraConfig := cassandraConfigInterface.(CassandraConfiguration)
	port = strconv.Itoa(*cassandraConfig.Port)
	cqlPort = strconv.Itoa(*cassandraConfig.CqlPort)
	jmxPort = strconv.Itoa(*cassandraConfig.JmxLocalPort)
	sort.SliceStable(cassandraNodes, func(i, j int) bool { return cassandraNodes[i] < cassandraNodes[j] })
	serverListCommaSeparated := strings.Join(cassandraNodes, ":"+port+",")
	serverListCommaSeparated = serverListCommaSeparated + ":" + port
	serverListSpaceSeparated := strings.Join(cassandraNodes, ":"+port+" ")
	serverListSpaceSeparated = serverListSpaceSeparated + ":" + port
	serverListCQLCommaSeparated := strings.Join(cassandraNodes, ":"+cqlPort+",")
	serverListCQLCommaSeparated = serverListCQLCommaSeparated + ":" + cqlPort
	serverListCQLSpaceSeparated := strings.Join(cassandraNodes, ":"+cqlPort+" ")
	serverListCQLSpaceSeparated = serverListCQLSpaceSeparated + ":" + cqlPort
	serverListJMXCommaSeparated := strings.Join(cassandraNodes, ":"+jmxPort+",")
	serverListJMXCommaSeparated = serverListJMXCommaSeparated + ":" + jmxPort
	serverListJMXSpaceSeparated := strings.Join(cassandraNodes, ":"+jmxPort+" ")
	serverListJMXSpaceSeparated = serverListJMXSpaceSeparated + ":" + jmxPort
	serverListCommanSeparatedQuoted := strings.Join(cassandraNodes, "','")
	serverListCommanSeparatedQuoted = "'" + serverListCommanSeparatedQuoted + "'"
	cassandraCluster = CassandraClusterConfiguration{
		Port:                            port,
		CQLPort:                         cqlPort,
		JMXPort:                         jmxPort,
		ServerListCommaSeparated:        serverListCommaSeparated,
		ServerListSpaceSeparated:        serverListSpaceSeparated,
		ServerListCQLCommaSeparated:     serverListCQLCommaSeparated,
		ServerListCQLSpaceSeparated:     serverListCQLSpaceSeparated,
		ServerListJMXCommaSeparated:     serverListJMXCommaSeparated,
		ServerListJMXSpaceSeparated:     serverListJMXSpaceSeparated,
		ServerListCommanSeparatedQuoted: serverListCommanSeparatedQuoted,
	}
	return &cassandraCluster, nil
}

//NewControlClusterConfiguration gets a struct containing various representations of Control nodes string
func NewControlClusterConfiguration(name string, role string, namespace string, myclient client.Client) (*ControlClusterConfiguration, error) {
	var controlNodes []string
	var controlCluster ControlClusterConfiguration
	var bgpPort string
	var dnsPort string
	var xmppPort string
	var dnsIntrospectPort string
	var controlConfigInterface interface{}
	if name != "" {
		controlInstance := &Control{}
		err := myclient.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, controlInstance)
		if err != nil {
			return &controlCluster, err
		}
		for _, ip := range controlInstance.Status.Nodes {
			controlNodes = append(controlNodes, ip)
		}
		controlConfigInterface = controlInstance.ConfigurationParameters()

	}
	if role != "" {
		labelSelector := labels.SelectorFromSet(map[string]string{"control_role": role})
		listOps := &client.ListOptions{Namespace: namespace, LabelSelector: labelSelector}
		controlList := &ControlList{}
		err := myclient.List(context.TODO(), listOps, controlList)
		if err != nil {
			return &controlCluster, err
		}
		if len(controlList.Items) > 0 {
			for _, ip := range controlList.Items[0].Status.Nodes {
				controlNodes = append(controlNodes, ip)
			}
		}
		controlConfigInterface = controlList.Items[0].ConfigurationParameters()
	}
	controlConfig := controlConfigInterface.(ControlConfiguration)
	bgpPort = strconv.Itoa(*controlConfig.BGPPort)
	dnsPort = strconv.Itoa(*controlConfig.DNSPort)
	xmppPort = strconv.Itoa(*controlConfig.XMPPPort)
	dnsIntrospectPort = strconv.Itoa(*controlConfig.DNSIntrospectPort)
	sort.SliceStable(controlNodes, func(i, j int) bool { return controlNodes[i] < controlNodes[j] })
	serverListXMPPCommaSeparated := strings.Join(controlNodes, ":"+xmppPort+",")
	serverListXMPPCommaSeparated = serverListXMPPCommaSeparated + ":" + xmppPort
	serverListXMPPSpaceSeparated := strings.Join(controlNodes, ":"+xmppPort+" ")
	serverListXMPPSpaceSeparated = serverListXMPPSpaceSeparated + ":" + xmppPort
	serverListDNSCommaSeparated := strings.Join(controlNodes, ":"+dnsPort+",")
	serverListDNSCommaSeparated = serverListDNSCommaSeparated + ":" + dnsPort
	serverListDNSSpaceSeparated := strings.Join(controlNodes, ":"+dnsPort+" ")
	serverListDNSSpaceSeparated = serverListDNSSpaceSeparated + ":" + dnsPort
	serverListCommanSeparatedQuoted := strings.Join(controlNodes, "','")
	serverListCommanSeparatedQuoted = "'" + serverListCommanSeparatedQuoted + "'"
	controlCluster = ControlClusterConfiguration{
		BGPPort:                         bgpPort,
		DNSPort:                         dnsPort,
		DNSIntrospectPort:               dnsIntrospectPort,
		ServerListXMPPCommaSeparated:    serverListXMPPCommaSeparated,
		ServerListXMPPSpaceSeparated:    serverListXMPPSpaceSeparated,
		ServerListDNSCommaSeparated:     serverListDNSCommaSeparated,
		ServerListDNSSpaceSeparated:     serverListDNSSpaceSeparated,
		ServerListCommanSeparatedQuoted: serverListCommanSeparatedQuoted,
	}

	return &controlCluster, nil
}

//NewZookeeperClusterConfiguration gets a struct containing various representations of Zookeeper nodes string
func NewZookeeperClusterConfiguration(name string, namespace string, client client.Client) (*ZookeeperClusterConfiguration, error) {
	var zookeeperNodes []string
	var zookeeperCluster ZookeeperClusterConfiguration
	var port string
	zookeeperInstance := &Zookeeper{}
	err := client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, zookeeperInstance)
	if err != nil {
		return &zookeeperCluster, err
	}
	for _, ip := range zookeeperInstance.Status.Nodes {
		zookeeperNodes = append(zookeeperNodes, ip)

	}
	zookeeperConfigInterface := zookeeperInstance.ConfigurationParameters()
	zookeeperConfig := zookeeperConfigInterface.(ZookeeperConfiguration)
	port = strconv.Itoa(*zookeeperConfig.ClientPort)
	sort.SliceStable(zookeeperNodes, func(i, j int) bool { return zookeeperNodes[i] < zookeeperNodes[j] })
	serverListCommaSeparated := strings.Join(zookeeperNodes, ":"+port+",")
	serverListCommaSeparated = serverListCommaSeparated + ":" + port
	serverListSpaceSeparated := strings.Join(zookeeperNodes, ":"+port+" ")
	serverListSpaceSeparated = serverListSpaceSeparated + ":" + port
	zookeeperCluster = ZookeeperClusterConfiguration{
		ClientPort:               port,
		ServerListCommaSeparated: serverListCommaSeparated,
		ServerListSpaceSeparated: serverListSpaceSeparated,
	}

	return &zookeeperCluster, nil
}

// NewRabbitmqClusterConfiguration gets a struct containing various representations of Rabbitmq nodes string
func NewRabbitmqClusterConfiguration(name string, namespace string, myclient client.Client) (*RabbitmqClusterConfiguration, error) {
	var rabbitmqNodes []string
	var rabbitmqCluster RabbitmqClusterConfiguration
	var port string
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": name})
	listOps := &client.ListOptions{Namespace: namespace, LabelSelector: labelSelector}
	rabbitmqList := &RabbitmqList{}
	err := myclient.List(context.TODO(), listOps, rabbitmqList)
	if err != nil {
		return &rabbitmqCluster, err
	}
	if len(rabbitmqList.Items) > 0 {
		for _, ip := range rabbitmqList.Items[0].Status.Nodes {
			rabbitmqNodes = append(rabbitmqNodes, ip)
		}
		rabbitmqConfigInterface := rabbitmqList.Items[0].ConfigurationParameters()
		rabbitmqConfig := rabbitmqConfigInterface.(RabbitmqConfiguration)
		port = strconv.Itoa(*rabbitmqConfig.Port)
	}
	sort.SliceStable(rabbitmqNodes, func(i, j int) bool { return rabbitmqNodes[i] < rabbitmqNodes[j] })
	serverListCommaSeparated := strings.Join(rabbitmqNodes, ":"+port+",")
	serverListCommaSeparated = serverListCommaSeparated + ":" + port
	serverListSpaceSeparated := strings.Join(rabbitmqNodes, ":"+port+" ")
	serverListSpaceSeparated = serverListSpaceSeparated + ":" + port
	serverListCommaSeparatedWithoutPort := strings.Join(rabbitmqNodes, ",")
	rabbitmqCluster = RabbitmqClusterConfiguration{
		Port:                                port,
		ServerListCommaSeparated:            serverListCommaSeparated,
		ServerListSpaceSeparated:            serverListSpaceSeparated,
		ServerListCommaSeparatedWithoutPort: serverListCommaSeparatedWithoutPort,
	}
	return &rabbitmqCluster, nil
}

// NewConfigClusterConfiguration gets a struct containing various representations of Config nodes string
func NewConfigClusterConfiguration(name string, namespace string, myclient client.Client) (*ConfigClusterConfiguration, error) {
	var configNodes []string
	var configCluster ConfigClusterConfiguration
	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_cluster": name})
	listOps := &client.ListOptions{Namespace: namespace, LabelSelector: labelSelector}
	configList := &ConfigList{}
	err := myclient.List(context.TODO(), listOps, configList)
	if err != nil {
		return &configCluster, err
	}

	var apiServerPort string
	var collectorServerPort string
	var analyticsServerPort string
	var redisServerPort string

	if len(configList.Items) > 0 {
		for _, ip := range configList.Items[0].Status.Nodes {
			configNodes = append(configNodes, ip)
		}
		configConfigInterface := configList.Items[0].ConfigurationParameters()
		configConfig := configConfigInterface.(ConfigConfiguration)
		apiServerPort = strconv.Itoa(*configConfig.APIPort)
		analyticsServerPort = strconv.Itoa(*configConfig.AnalyticsPort)
		collectorServerPort = strconv.Itoa(*configConfig.CollectorPort)
		redisServerPort = strconv.Itoa(*configConfig.RedisPort)
	}
	sort.SliceStable(configNodes, func(i, j int) bool { return configNodes[i] < configNodes[j] })
	apiServerListQuotedCommaSeparated := strings.Join(configNodes, "','")
	apiServerListQuotedCommaSeparated = "'" + apiServerListQuotedCommaSeparated + "'"
	analyticsServerListQuotedCommaSeparated := strings.Join(configNodes, "','")
	analyticsServerListQuotedCommaSeparated = "'" + analyticsServerListQuotedCommaSeparated + "'"
	apiServerListCommaSeparated := strings.Join(configNodes, ",")
	apiServerListSpaceSeparated := strings.Join(configNodes, ":"+apiServerPort+" ")
	apiServerListSpaceSeparated = apiServerListSpaceSeparated + ":" + apiServerPort
	analyticsServerListSpaceSeparated := strings.Join(configNodes, ":"+analyticsServerPort+" ")
	analyticsServerListSpaceSeparated = analyticsServerListSpaceSeparated + ":" + analyticsServerPort
	collectorServerListSpaceSeparated := strings.Join(configNodes, ":"+collectorServerPort+" ")
	collectorServerListSpaceSeparated = collectorServerListSpaceSeparated + ":" + collectorServerPort
	configCluster = ConfigClusterConfiguration{
		APIServerPort:                           apiServerPort,
		APIServerListQuotedCommaSeparated:       apiServerListQuotedCommaSeparated,
		APIServerListCommaSeparated:             apiServerListCommaSeparated,
		APIServerListSpaceSeparated:             apiServerListSpaceSeparated,
		AnalyticsServerPort:                     analyticsServerPort,
		AnalyticsServerListSpaceSeparated:       analyticsServerListSpaceSeparated,
		AnalyticsServerListQuotedCommaSeparated: analyticsServerListQuotedCommaSeparated,
		CollectorPort:                           collectorServerPort,
		CollectorServerListSpaceSeparated:       collectorServerListSpaceSeparated,
		RedisPort:                               redisServerPort,
	}
	return &configCluster, nil
}

//ConfigClusterConfiguration defines all configuration knobs used to write the config file
type ConfigClusterConfiguration struct {
	APIServerPort                           string
	APIServerListSpaceSeparated             string
	APIServerListQuotedCommaSeparated       string
	APIServerListCommaSeparated             string
	AnalyticsServerPort                     string
	AnalyticsServerListSpaceSeparated       string
	AnalyticsServerListQuotedCommaSeparated string
	CollectorServerListSpaceSeparated       string
	CollectorPort                           string
	RedisPort                               string
}

//ControlClusterConfiguration defines all configuration knobs used to write the config file
type ControlClusterConfiguration struct {
	BGPPort                         string
	DNSPort                         string
	DNSIntrospectPort               string
	ServerListXMPPCommaSeparated    string
	ServerListXMPPSpaceSeparated    string
	ServerListDNSCommaSeparated     string
	ServerListDNSSpaceSeparated     string
	ServerListCommanSeparatedQuoted string
}

//ZookeeperClusterConfiguration defines all configuration knobs used to write the config file
type ZookeeperClusterConfiguration struct {
	ClientPort               string
	ServerPort               string
	ElectionPort             string
	ServerListCommaSeparated string
	ServerListSpaceSeparated string
}

//RabbitmqClusterConfiguration defines all configuration knobs used to write the config file
type RabbitmqClusterConfiguration struct {
	Port                                string
	ServerListCommaSeparated            string
	ServerListSpaceSeparated            string
	ServerListCommaSeparatedWithoutPort string
}

//CassandraClusterConfiguration defines all configuration knobs used to write the config file
type CassandraClusterConfiguration struct {
	Port                            string
	CQLPort                         string
	JMXPort                         string
	ServerListCommaSeparated        string
	ServerListSpaceSeparated        string
	ServerListCQLCommaSeparated     string
	ServerListCQLSpaceSeparated     string
	ServerListJMXCommaSeparated     string
	ServerListJMXSpaceSeparated     string
	ServerListCommanSeparatedQuoted string
}
