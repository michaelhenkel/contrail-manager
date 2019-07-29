package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// ServiceStatus provides information on the current status of the service
// +k8s:openapi-gen=true
type ServiceStatus struct {
	Active            *bool `json:"active,omitempty"`
	Created           *bool `json:"created,omitempty"`
	ControllerRunning *bool `json:"controllerRunning,omitempty"`
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
