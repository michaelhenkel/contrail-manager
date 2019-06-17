package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"text/template"

	"github.com/ghodss/yaml"
	//"k8s.io/apimachinery/pkg/runtime"
	//contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	//apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	appsv1 "k8s.io/api/apps/v1"
)

const (
	daemonSetDirectory = "../../deployments/"
)

var daemonSetServiceList = [...]string{"vrouter"}

//go:generate go run gen_daemonset.go
func main() {

	createSts()
}

func createSts() {

	var packageTemplate = template.Must(template.New("").Parse(`package {{ .Kind }}
	
import(
	appsv1 "k8s.io/api/apps/v1"
	"github.com/ghodss/yaml"
)

var yamlData{{ .Kind }}= {{ .YamlData }}

func GetDaemonset() *appsv1.DaemonSet{
	daemonSet := appsv1.DaemonSet{}
	err := yaml.Unmarshal([]byte(yamlData{{ .Kind }}), &daemonSet)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlData{{ .Kind }}))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &daemonSet)
	if err != nil {
		panic(err)
	}
	return &daemonSet
}
	`))

	for _, stsName := range daemonSetServiceList {
		crFile := stsName + ".yaml"
		yamlData, err := ioutil.ReadFile(daemonSetDirectory + crFile)
		if err != nil {
			panic(err)
		}

		jsonData, err := yaml.YAMLToJSON([]byte(yamlData))
		if err != nil {
			panic(err)
		}
		var daemonSet appsv1.DaemonSet
		err = yaml.Unmarshal([]byte(jsonData), &daemonSet)
		if err != nil {
			panic(err)
		}
		f, err := os.Create("../controller/" + stsName + "/daemonset.go")
		if err != nil {
			panic(err)
		}

		yamlDataQuoted := fmt.Sprintf("`\n")
		yamlDataQuoted = yamlDataQuoted + string(yamlData)
		yamlDataQuoted = yamlDataQuoted + "`"
		packageTemplate.Execute(f, struct {
			YamlData string
			Kind     string
		}{
			YamlData: yamlDataQuoted,
			Kind:     stsName,
		})
	}
}
