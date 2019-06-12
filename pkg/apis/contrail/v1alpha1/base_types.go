package v1alpha1

import (
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Service struct {
	Activate      *bool `json:"activate,omitempty"`
	Create        *bool `json:"create,omitempty"`
	Image         string `json:"image,omitempty"`
	Size          *int `json:"size,omitempty"`
	Configuration map[string]string `json:"configuration,omitempty"`
}

type Global struct {

}