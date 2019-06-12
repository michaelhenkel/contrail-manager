package v1alpha1

import (
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ServiceStatus struct {
	Active *bool `json:"active"`
	Created *bool `json:"created"`
}


type Service struct {
	Activate      *bool `json:"activate,omitempty"`
	Create        *bool `json:"create,omitempty"`
	Image         string `json:"image,omitempty"`
	Size          *int32 `json:"size,omitempty"`
	Configuration map[string]string `json:"configuration,omitempty"`
	Images        map[string]string `json:"images,omitempty"`
}

type Global struct {

}