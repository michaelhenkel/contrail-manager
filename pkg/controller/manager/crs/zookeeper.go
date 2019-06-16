package cr
	
import(
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/ghodss/yaml"
)

var yamlDataZookeeper= `
apiVersion: contrail.juniper.net/v1alpha1
kind: Zookeeper
metadata:
  name: cluster-1
`

func GetZookeeperCr() *contrailv1alpha1.Zookeeper{
	cr := contrailv1alpha1.Zookeeper{}
	err := yaml.Unmarshal([]byte(yamlDataZookeeper), &cr)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDataZookeeper))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &cr)
	if err != nil {
		panic(err)
	}
	return &cr
}
	