package crd
	
import(
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"github.com/ghodss/yaml"
)

var yamlDataZookeeper = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: zookeepers.contrail.juniper.net
spec:
  group: contrail.juniper.net
  names:
    kind: Zookeeper
    listKind: ZookeeperList
    plural: zookeepers
    singular: zookeeper
  scope: Namespaced
  subresources:
    status: {}
  validation:
    openAPIV3Schema:
      properties:
        apiVersion:
          description: 'APIVersion defines the versioned schema of this representation
            of an object. Servers should convert recognized schemas to the latest
            internal value, and may reject unrecognized values. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#resources'
          type: string
        kind:
          description: 'Kind is a string value representing the REST resource this
            object represents. Servers may infer this from the endpoint the client
            submits requests to. Cannot be updated. In CamelCase. More info: https://git.k8s.io/community/contributors/devel/api-conventions.md#types-kinds'
          type: string
        metadata:
          type: object
        spec:
          properties:
            clientPort:
              format: int64
              type: integer
            contrailStatusImage:
              type: string
            electionPort:
              format: int64
              type: integer
            heapSize:
              type: string
            hostNetwork:
              description: 'INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
                Important: Run "operator-sdk generate k8s" to regenerate code after
                modifying this file Add custom validation using kubebuilder tags:
                https://book.kubebuilder.io/beyond_basics/generating_crd.html'
              type: boolean
            serverPort:
              format: int64
              type: integer
            service:
              properties:
                activate:
                  type: boolean
                configuration:
                  additionalProperties:
                    type: string
                  type: object
                create:
                  type: boolean
                image:
                  type: string
                images:
                  additionalProperties:
                    type: string
                  type: object
                size:
                  format: int32
                  type: integer
              type: object
          type: object
        status:
          properties:
            active:
              description: 'INSERT ADDITIONAL STATUS FIELD - define observed state
                of cluster Important: Run "operator-sdk generate k8s" to regenerate
                code after modifying this file Add custom validation using kubebuilder
                tags: https://book.kubebuilder.io/beyond_basics/generating_crd.html'
              type: boolean
            nodes:
              additionalProperties:
                type: string
              type: object
            ports:
              additionalProperties:
                type: string
              type: object
          type: object
  version: v1alpha1
  versions:
  - name: v1alpha1
    served: true
    storage: true
`

func GetZookeeperCrd() *apiextensionsv1beta1.CustomResourceDefinition{
	crd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := yaml.Unmarshal([]byte(yamlDataZookeeper), &crd)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDataZookeeper))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &crd)
	if err != nil {
		panic(err)
	}
	return &crd
}
	