package v1alpha1

import (
	"context"
	"strconv"

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

//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
	// If specified, these secrets will be passed to individual puller implementations for them to use. For example,
	// in the case of docker, only DockerConfig type secrets are honored.
	// More info: https://kubernetes.io/docs/concepts/containers/images#specifying-imagepullsecrets-on-a-pod
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	ImagePullSecrets []corev1.LocalObjectReference `json:"imagePullSecrets,omitempty" patchStrategy:"merge" patchMergeKey:"name" protobuf:"bytes,15,rep,name=imagePullSecrets"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
	// Number of desired pods. This is a pointer to distinguish between explicit
	// zero and not specified. Defaults to 1.
	// +optional
	Replicas *int32 `json:"replicas,omitempty" protobuf:"varint,1,opt,name=replicas"`
}

type Instance interface {
	//ManageActiveStatus(*appsv1.Deployment, reconcile.Request, client.Client) error
	CreateConfigMap(string, client.Client, *runtime.Scheme, reconcile.Request) (*corev1.ConfigMap, error)
	OwnedByManager(client.Client, reconcile.Request) (*Manager, error)
	PrepareIntendedDeployment(*appsv1.Deployment, *CommonConfiguration, reconcile.Request, *runtime.Scheme) (*appsv1.Deployment, error)
	AddVolumesToIntendedDeployments(*appsv1.Deployment, map[string]string)
	CompareIntendedWithCurrentDeployment(*appsv1.Deployment, *CommonConfiguration, reconcile.Request, *runtime.Scheme, client.Client, bool) error
	GetPodIPListAndIPMap(reconcile.Request, client.Client) (*corev1.PodList, map[string]string, error)
	CreateInstanceConfiguration(reconcile.Request, *corev1.PodList, client.Client) error
	SetPodsToReady(*corev1.PodList, client.Client) error
	ManageNodeStatus(map[string]string, client.Client) error
	SetInstanceActive(client.Client, *Status, *appsv1.Deployment, reconcile.Request) error
}

func SetInstanceActive(client client.Client, status *Status, deployment *appsv1.Deployment, request reconcile.Request) error {
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
	return nil
}

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

func SetDeploymentCommonConfiguration(deployment *appsv1.Deployment,
	commonConfiguration *CommonConfiguration) *appsv1.Deployment {
	deployment.Spec.Replicas = commonConfiguration.Replicas
	deployment.Spec.Template.Spec.Tolerations = commonConfiguration.Tolerations
	deployment.Spec.Template.Spec.NodeSelector = commonConfiguration.NodeSelector
	deployment.Spec.Template.Spec.HostNetwork = *commonConfiguration.HostNetwork
	deployment.Spec.Template.Spec.ImagePullSecrets = commonConfiguration.ImagePullSecrets
	return deployment
}

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
			/*
				labels := intendedDeployment.Spec.Selector.MatchLabels
				labels["version"] = "1"
				intendedDeployment.Spec.Template.SetLabels(labels)
				intendedDeployment.Spec.Selector.MatchLabels = labels
			*/
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

func GetPodIPListAndIPMap(instanceType string,
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
	/*
		if len(replicaSetList.Items) > 0 {
			fmt.Println("len(replicaSetList.Items)", len(replicaSetList.Items))
			replicaSet := &appsv1.ReplicaSet{}
			for _, rs := range replicaSetList.Items {
				fmt.Printf("%+v\n", rs)
				if *rs.Spec.Replicas > 0 {
					replicaSet = &rs
				} else {
					replicaSet = &rs
				}
			}
			podList := &corev1.PodList{}
			if podHash, ok := replicaSet.ObjectMeta.Labels["pod-template-hash"]; ok {
				labelSelector := labels.SelectorFromSet(map[string]string{"pod-template-hash": podHash})
				listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
				err := reconcileClient.List(context.TODO(), listOps, podList)
				if err != nil {
					return &corev1.PodList{}, map[string]string{}, err
				}
			}

			if replicaSet.Spec.Replicas != nil {
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
	*/

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
