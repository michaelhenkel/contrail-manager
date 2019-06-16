// +build !ignore_autogenerated

// Code generated by openapi-gen. DO NOT EDIT.

// This file was autogenerated by openapi-gen. Do not edit it manually!

package v1alpha1

import (
	spec "github.com/go-openapi/spec"
	common "k8s.io/kube-openapi/pkg/common"
)

func GetOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Cassandra":       schema_pkg_apis_contrail_v1alpha1_Cassandra(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.CassandraSpec":   schema_pkg_apis_contrail_v1alpha1_CassandraSpec(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Config":          schema_pkg_apis_contrail_v1alpha1_Config(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ConfigSpec":      schema_pkg_apis_contrail_v1alpha1_ConfigSpec(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ConfigStatus":    schema_pkg_apis_contrail_v1alpha1_ConfigStatus(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Manager":         schema_pkg_apis_contrail_v1alpha1_Manager(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ManagerSpec":     schema_pkg_apis_contrail_v1alpha1_ManagerSpec(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ManagerStatus":   schema_pkg_apis_contrail_v1alpha1_ManagerStatus(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Rabbitmq":        schema_pkg_apis_contrail_v1alpha1_Rabbitmq(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.RabbitmqSpec":    schema_pkg_apis_contrail_v1alpha1_RabbitmqSpec(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.RabbitmqStatus":  schema_pkg_apis_contrail_v1alpha1_RabbitmqStatus(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Status":          schema_pkg_apis_contrail_v1alpha1_Status(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Zookeeper":       schema_pkg_apis_contrail_v1alpha1_Zookeeper(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ZookeeperSpec":   schema_pkg_apis_contrail_v1alpha1_ZookeeperSpec(ref),
		"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ZookeeperStatus": schema_pkg_apis_contrail_v1alpha1_ZookeeperStatus(ref),
	}
}

func schema_pkg_apis_contrail_v1alpha1_Cassandra(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Cassandra is the Schema for the cassandras API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.CassandraSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Status"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.CassandraSpec", "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Status", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_contrail_v1alpha1_CassandraSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "CassandraSpec defines the desired state of Cassandra",
				Properties: map[string]spec.Schema{
					"hostNetwork": {
						SchemaProps: spec.SchemaProps{
							Description: "INSERT ADDITIONAL SPEC FIELDS - desired state of cluster Important: Run \"operator-sdk generate k8s\" to regenerate code after modifying this file Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"service": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Service"),
						},
					},
					"contrailStatusImage": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Service"},
	}
}

func schema_pkg_apis_contrail_v1alpha1_Config(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Config is the Schema for the configs API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ConfigSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Status"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ConfigSpec", "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Status", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_contrail_v1alpha1_ConfigSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ConfigSpec defines the desired state of Config",
				Properties: map[string]spec.Schema{
					"hostNetwork": {
						SchemaProps: spec.SchemaProps{
							Description: "INSERT ADDITIONAL SPEC FIELDS - desired state of cluster Important: Run \"operator-sdk generate k8s\" to regenerate code after modifying this file Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"service": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Service"),
						},
					},
					"contrailStatusImage": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Service"},
	}
}

func schema_pkg_apis_contrail_v1alpha1_ConfigStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ConfigStatus defines the observed state of Config",
				Properties: map[string]spec.Schema{
					"active": {
						SchemaProps: spec.SchemaProps{
							Description: "INSERT ADDITIONAL STATUS FIELD - define observed state of cluster Important: Run \"operator-sdk generate k8s\" to regenerate code after modifying this file Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
				},
				Required: []string{"active"},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_contrail_v1alpha1_Manager(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Manager is the Schema for the managers API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ManagerSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ManagerStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ManagerSpec", "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ManagerStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_contrail_v1alpha1_ManagerSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "type Service struct {\n\tActivate      *bool `json:\"activate,omitempty\"`\n\tImage         string `json:\"image,omitempty\"`\n\tSize          *int `json:\"size,omitempty\"`\n\tConfiguration map[string]string `json:\"configuration,omitempty\"`\n}\n\ntype Global struct {\n\n} // EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN! // NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.\n\nManagerSpec defines the desired state of Manager",
				Properties: map[string]spec.Schema{
					"config": {
						SchemaProps: spec.SchemaProps{
							Description: "INSERT ADDITIONAL SPEC FIELDS - desired state of cluster Important: Run \"operator-sdk generate k8s\" to regenerate code after modifying this file Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html",
							Ref:         ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Service"),
						},
					},
					"cassandra": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Service"),
						},
					},
					"zookeeper": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Service"),
						},
					},
					"size": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"hostNetwork": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"boolean"},
							Format: "",
						},
					},
					"contrailStatusImage": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Service"},
	}
}

func schema_pkg_apis_contrail_v1alpha1_ManagerStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ManagerStatus defines the observed state of Manager",
				Properties: map[string]spec.Schema{
					"config": {
						SchemaProps: spec.SchemaProps{
							Description: "INSERT ADDITIONAL STATUS FIELD - define observed state of cluster Important: Run \"operator-sdk generate k8s\" to regenerate code after modifying this file Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html",
							Ref:         ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ServiceStatus"),
						},
					},
					"cassandra": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ServiceStatus"),
						},
					},
					"zookeeper": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ServiceStatus"),
						},
					},
				},
				Required: []string{"config", "cassandra", "zookeeper"},
			},
		},
		Dependencies: []string{
			"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ServiceStatus"},
	}
}

func schema_pkg_apis_contrail_v1alpha1_Rabbitmq(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Rabbitmq is the Schema for the rabbitmqs API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.RabbitmqSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.RabbitmqStatus"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.RabbitmqSpec", "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.RabbitmqStatus", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_contrail_v1alpha1_RabbitmqSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "RabbitmqSpec defines the desired state of Rabbitmq",
				Properties:  map[string]spec.Schema{},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_contrail_v1alpha1_RabbitmqStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "RabbitmqStatus defines the observed state of Rabbitmq",
				Properties:  map[string]spec.Schema{},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_contrail_v1alpha1_Status(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Properties: map[string]spec.Schema{
					"active": {
						SchemaProps: spec.SchemaProps{
							Description: "INSERT ADDITIONAL STATUS FIELD - define observed state of cluster Important: Run \"operator-sdk generate k8s\" to regenerate code after modifying this file Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"nodes": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
				},
				Required: []string{"active", "nodes"},
			},
		},
		Dependencies: []string{},
	}
}

func schema_pkg_apis_contrail_v1alpha1_Zookeeper(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "Zookeeper is the Schema for the zookeepers API",
				Properties: map[string]spec.Schema{
					"kind": {
						SchemaProps: spec.SchemaProps{
							Description: "Kind is a string value representing the REST resource this object represents. Servers may infer this from the endpoint the client submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"apiVersion": {
						SchemaProps: spec.SchemaProps{
							Description: "APIVersion defines the versioned schema of this representation of an object. Servers should convert recognized schemas to the latest internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources",
							Type:        []string{"string"},
							Format:      "",
						},
					},
					"metadata": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"),
						},
					},
					"spec": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ZookeeperSpec"),
						},
					},
					"status": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Status"),
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Status", "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.ZookeeperSpec", "k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMeta"},
	}
}

func schema_pkg_apis_contrail_v1alpha1_ZookeeperSpec(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ZookeeperSpec defines the desired state of Zookeeper",
				Properties: map[string]spec.Schema{
					"hostNetwork": {
						SchemaProps: spec.SchemaProps{
							Description: "INSERT ADDITIONAL SPEC FIELDS - desired state of cluster Important: Run \"operator-sdk generate k8s\" to regenerate code after modifying this file Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"service": {
						SchemaProps: spec.SchemaProps{
							Ref: ref("github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Service"),
						},
					},
					"contrailStatusImage": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
					"clientPort": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"electionPort": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"serverPort": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"integer"},
							Format: "int32",
						},
					},
					"heapSize": {
						SchemaProps: spec.SchemaProps{
							Type:   []string{"string"},
							Format: "",
						},
					},
				},
			},
		},
		Dependencies: []string{
			"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1.Service"},
	}
}

func schema_pkg_apis_contrail_v1alpha1_ZookeeperStatus(ref common.ReferenceCallback) common.OpenAPIDefinition {
	return common.OpenAPIDefinition{
		Schema: spec.Schema{
			SchemaProps: spec.SchemaProps{
				Description: "ZookeeperStatus defines the observed state of Zookeeper",
				Properties: map[string]spec.Schema{
					"active": {
						SchemaProps: spec.SchemaProps{
							Description: "INSERT ADDITIONAL STATUS FIELD - define observed state of cluster Important: Run \"operator-sdk generate k8s\" to regenerate code after modifying this file Add custom validation using kubebuilder tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html",
							Type:        []string{"boolean"},
							Format:      "",
						},
					},
					"nodes": {
						SchemaProps: spec.SchemaProps{
							Type: []string{"object"},
							AdditionalProperties: &spec.SchemaOrBool{
								Schema: &spec.Schema{
									SchemaProps: spec.SchemaProps{
										Type:   []string{"string"},
										Format: "",
									},
								},
							},
						},
					},
				},
				Required: []string{"active", "nodes"},
			},
		},
		Dependencies: []string{},
	}
}
