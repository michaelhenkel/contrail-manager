package crd
	
import(
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"github.com/ghodss/yaml"
)

var yamlDataRabbitmq = `
apiVersion: apiextensions.k8s.io/v1beta1
kind: CustomResourceDefinition
metadata:
  name: rabbitmqs.contrail.juniper.net
spec:
  group: contrail.juniper.net
  names:
    kind: Rabbitmq
    listKind: RabbitmqList
    plural: rabbitmqs
    singular: rabbitmq
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
          type: object
        status:
          type: object
  version: v1alpha1
  versions:
  - name: v1alpha1
    served: true
    storage: true
`

func GetRabbitmqCrd() *apiextensionsv1beta1.CustomResourceDefinition{
	crd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := yaml.Unmarshal([]byte(yamlDataRabbitmq), &crd)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDataRabbitmq))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &crd)
	if err != nil {
		panic(err)
	}
	return &crd
}
	