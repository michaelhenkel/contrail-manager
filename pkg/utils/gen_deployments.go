package main

import(
	"fmt"
	"os"
	"text/template"
	"io/ioutil"

	"github.com/ghodss/yaml"
	//"k8s.io/apimachinery/pkg/runtime"
	//contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	//apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	deploymentDirectory = "../../deployments/"
)


var serviceList = [...]string{"cassandra","config"}


//go:generate go run gen_deployments.go
func main(){
	createDeployments()
}

func createDeployments(){

	var packageTemplate = template.Must(template.New("").Parse(`package {{ .Kind }}
	
import(
	appsv1 "k8s.io/api/apps/v1"
	"github.com/ghodss/yaml"
)

var yamlData{{ .Kind }}= {{ .YamlData }}

func GetDeployment() *appsv1.Deployment{
	deployment := appsv1.Deployment{}
	err := yaml.Unmarshal([]byte(yamlData{{ .Kind }}), &deployment)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlData{{ .Kind }}))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &deployment)
	if err != nil {
		panic(err)
	}
	return &deployment
}
	`))

	for _, deploymentName := range(serviceList){
		crFile := deploymentName + ".yaml"
		yamlData, err := ioutil.ReadFile(deploymentDirectory + crFile)
		if err != nil {
			panic(err)
		}

		jsonData, err := yaml.YAMLToJSON([]byte(yamlData))
		if err != nil {
			panic(err)
		}
		var deployment appsv1.Deployment
		err = yaml.Unmarshal([]byte(jsonData), &deployment)
		if err != nil {
			panic(err)
		}
		f, err := os.Create("../controller/"+ deploymentName + "/deployment.go")
		if err != nil {
			panic(err)
		}

		yamlDataQuoted := fmt.Sprintf("`\n")
		yamlDataQuoted = yamlDataQuoted + string(yamlData)
		yamlDataQuoted = yamlDataQuoted + "`"
		packageTemplate.Execute(f, struct{
			YamlData string
			Kind string
		}{
			YamlData: yamlDataQuoted,
			Kind: deploymentName,
		})
	}
}
