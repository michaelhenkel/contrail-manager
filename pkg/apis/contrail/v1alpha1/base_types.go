package v1alpha1

//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

type ServiceStatus struct {
	Active  *bool `json:"active"`
	Created *bool `json:"created"`
}

type Service struct {
	Activate      *bool             `json:"activate,omitempty"`
	Create        *bool             `json:"create,omitempty"`
	Image         string            `json:"image,omitempty"`
	Size          *int32            `json:"size,omitempty"`
	Configuration map[string]string `json:"configuration,omitempty"`
	Images        map[string]string `json:"images,omitempty"`
}

// +k8s:openapi-gen=true
type Status struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html
	Active *bool             `json:"active"`
	Nodes  map[string]string `json:"nodes"`
}
type Global struct {
}
