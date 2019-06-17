package cr
	
import(
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/ghodss/yaml"
)

var yamlDataKubemanager= `
apiVersion: contrail.juniper.net/v1alpha1
kind: Kubemanager
metadata:
  name: cluster-1
`

func GetKubemanagerCr() *contrailv1alpha1.Kubemanager{
	cr := contrailv1alpha1.Kubemanager{}
	err := yaml.Unmarshal([]byte(yamlDataKubemanager), &cr)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDataKubemanager))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &cr)
	if err != nil {
		panic(err)
	}
	return &cr
}
	