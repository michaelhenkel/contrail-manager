package cr
	
import(
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/ghodss/yaml"
)

var yamlDataCassandra= `
apiVersion: contrail.juniper.net/v1alpha1
kind: Cassandra
metadata:
  name: example-cassandra
spec:
  # Add fields here
  size: 3
`

func GetCassandraCr() *contrailv1alpha1.Cassandra{
	cr := contrailv1alpha1.Cassandra{}
	err := yaml.Unmarshal([]byte(yamlDataCassandra), &cr)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlDataCassandra))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &cr)
	if err != nil {
		panic(err)
	}
	return &cr
}
	