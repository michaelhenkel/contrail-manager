package v1alpha1

//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +k8s:openapi-gen=true
type ServiceStatus struct {
	Active            *bool `json:"active,omitempty"`
	Created           *bool `json:"created,omitempty"`
	ControllerRunning *bool `json:"controllerRunning,omitempty"`
}

// +k8s:openapi-gen=true
type Service struct {
	Activate            *bool             `json:"activate,omitempty"`
	Create              *bool             `json:"create,omitempty"`
	Image               string            `json:"image,omitempty"`
	Size                *int32            `json:"size,omitempty"`
	Configuration       map[string]string `json:"configuration,omitempty"`
	Images              map[string]string `json:"images,omitempty"`
	HostNetwork         *bool             `json:"hostNetwork,omitempty"`
	ContrailStatusImage string            `json:"contrailStatusImage,omitempty"`
}

// +k8s:openapi-gen=true
type Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Active *bool             `json:"active,omitempty"`
	Nodes  map[string]string `json:"nodes,omitempty"`
	Ports  map[string]string `json:"ports,omitempty"`
}
type Global struct {
}