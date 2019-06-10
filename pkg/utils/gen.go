package main

import(
	"fmt"
	"os"
	"strings"
	"text/template"
	"io/ioutil"

	"github.com/ghodss/yaml"

	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
)

const (
	configCrd = "contrail_v1alpha1_config_crd.yaml"
	crdDirectory = "../../deploy/crds/"
)

//go:generate go run gen.go
func main(){
	crdList := []string{configCrd}
	for _, crdTemplate := range(crdList){
		yamlData, err := ioutil.ReadFile(crdDirectory + crdTemplate)
		if err != nil {
			panic(err)
		}
		crdKind := strings.Split(crdTemplate, ".")
		jsonData, err := yaml.YAMLToJSON([]byte(yamlData))
		if err != nil {
			panic(err)
		}
		crd := apiextensionsv1beta1.CustomResourceDefinition{}
		err = yaml.Unmarshal([]byte(jsonData), &crd)
		if err != nil {
			panic(err)
		}
		f, err := os.Create("../controller/manager/crds/" + crdKind[0] + ".go")
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
			Kind: crd.Spec.Names.Kind,
		})
	}
}

var packageTemplate = template.Must(template.New("").Parse(`package manager

import(
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"github.com/ghodss/yaml"
)

var yamlData = {{ .YamlData }}

func Get{{ .Kind }}Crd() *apiextensionsv1beta1.CustomResourceDefinition{
	crd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := yaml.Unmarshal([]byte(yamlData), &crd)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlData))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &crd)
	if err != nil {
		panic(err)
	}
	return &crd
}
`))

