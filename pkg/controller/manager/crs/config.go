package cr

import (
	"github.com/ghodss/yaml"
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
)

var yamlDataConfig = `
apiVersion: contrail.juniper.net/v1alpha1
kind: Config
metadata:
  name: example-config
spec:
  # Add fields here
  size: 3
  service: 
    activate: true
`

func GetConfigCr() *contrailv1alpha1.Config {
	cr := contrailv1alpha1.Config{}
	err := yaml.Unmarshal([]byte(yamlDataConfig), &cr)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDataConfig))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &cr)
	if err != nil {
		panic(err)
	}
	return &cr
}
