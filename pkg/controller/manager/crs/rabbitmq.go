package cr
	
import(
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/ghodss/yaml"
)

var yamlDataRabbitmq= `
apiVersion: contrail.juniper.net/v1alpha1
kind: Rabbitmq
metadata:
  name: example-rabbitmq
spec:
  # Add fields here
  size: 3
`

func GetRabbitmqCr() *contrailv1alpha1.Rabbitmq{
	cr := contrailv1alpha1.Rabbitmq{}
	err := yaml.Unmarshal([]byte(yamlDataRabbitmq), &cr)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDataRabbitmq))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &cr)
	if err != nil {
		panic(err)
	}
	return &cr
}
	