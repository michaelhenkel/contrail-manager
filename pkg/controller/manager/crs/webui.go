package cr
	
import(
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/ghodss/yaml"
)

var yamlDataWebui= `
apiVersion: contrail.juniper.net/v1alpha1
kind: Webui
metadata:
  name: cluster-1
`

func GetWebuiCr() *contrailv1alpha1.Webui{
	cr := contrailv1alpha1.Webui{}
	err := yaml.Unmarshal([]byte(yamlDataWebui), &cr)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDataWebui))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &cr)
	if err != nil {
		panic(err)
	}
	return &cr
}
	