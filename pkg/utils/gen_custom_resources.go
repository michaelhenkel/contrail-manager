package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	crdManager    = "crdManager.go"
	crManager     = "crManager.go"
	crdDirectory  = "../../deploy/crds/"
	crDirectory   = "../../deploy/crds/"
	filePrefix    = "contrail_v1alpha1_"
	crdFileSuffix = "_crd.yaml"
	crFileSuffix  = "_cr.yaml"
)

//var crdList = [...]string{"contrail_v1alpha1_config_crd.yaml"}
//var crList = [...]string{"contrail_v1alpha1_config_cr.yaml"}
var serviceMap = map[string]runtime.Object{
	"config":      &contrailv1alpha1.Config{},
	"control":     &contrailv1alpha1.Control{},
	"kubemanager": &contrailv1alpha1.Kubemanager{},
	"webui":       &contrailv1alpha1.Webui{},
	"vrouter":     &contrailv1alpha1.Vrouter{},
	"cassandra":   &contrailv1alpha1.Cassandra{},
	"zookeeper":   &contrailv1alpha1.Zookeeper{},
	"rabbitmq":    &contrailv1alpha1.Rabbitmq{},
}

//go:generate go run gen_custom_resources.go
func main() {
	createCrds()
	createCrs()
}

func createCrs() {
	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	var packageTemplate = template.Must(template.New("").Parse(`package cr
	
import(
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/ghodss/yaml"
)

var yamlData{{ .Kind }}= {{ .YamlData }}

func Get{{ .Kind }}Cr() *contrailv1alpha1.{{ .Kind }}{
	cr := contrailv1alpha1.{{ .Kind }}{}
	err := yaml.Unmarshal([]byte(yamlData{{ .Kind }}), &cr)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlData{{ .Kind }}))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal([]byte(jsonData), &cr)
	if err != nil {
		panic(err)
	}
	return &cr
}
	`))

	var crManagerTemplate = template.Must(template.New("").Funcs(funcMap).Parse(`package manager
	
import(
	"context"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/api/errors"
	"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	cr "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crs"
)

func (r *ReconcileManager) CreateResource(instance *v1alpha1.Manager, obj runtime.Object, name string, namespace string) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Reconciling Manager")
	
	objectKind := obj.GetObjectKind()
	groupVersionKind := objectKind.GroupVersionKind()

	gkv := schema.FromAPIVersionAndKind(groupVersionKind.Group + "/" + groupVersionKind.Version, groupVersionKind.Kind)
	newObj, err := scheme.Scheme.New(gkv)
	if err != nil {
		return err
	}

	newObj = obj

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, newObj)
	if err != nil && errors.IsNotFound(err) {

		switch groupVersionKind.Kind{
		{{- range .KindList }}
		case "{{ . }}":
			typedObject := &v1alpha1.{{ . }}{}
			typedObject = newObj.(*v1alpha1.{{ . }})
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" Created.")
			/*
			err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, typedObject)
			if err != nil {
				reqLogger.Info("Failed to get created CR " + name)
				return err
			}
			
			err = r.client.Update(context.TODO(), typedObject)
			if err != nil{
				reqLogger.Info("Failed to update created CR " + name)
				return err
			}
			reqLogger.Info("CR " + name +" updated.")
			*/
		{{- end }}
		}
	}
	return nil
}
/*
func (r *ReconcileManager) UpdateResource(instance *v1alpha1.Manager, obj runtime.Object, name string, namespace string) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Reconciling Manager")
	
	objectKind := obj.GetObjectKind()
	groupVersionKind := objectKind.GroupVersionKind()

	gkv := schema.FromAPIVersionAndKind(groupVersionKind.Group + "/" + groupVersionKind.Version, groupVersionKind.Kind)
	newObj, err := scheme.Scheme.New(gkv)
	if err != nil {
		return err
	}

	newObj = obj

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, newObj)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("CR " + name +" not found. Creating it.")
		err = r.client.Create(context.TODO(), newObj)	
		if err != nil {
			reqLogger.Error(err, "Failed to create new newObj.", "resource.Namespace", namespace, "resource.Name", name)
			return err
		}
		reqLogger.Info("CR " + namespace +" created.")
	}
	return nil
}
*/

func (r *ReconcileManager) ManageCr(instance *v1alpha1.Manager) error{
	var err error
	{{- range .KindList }}
	var {{ . }}Status v1alpha1.ServiceStatus
	{{- end }}

	{{- range .KindList }}
	{{ . }}Created := true
	if instance.Status.{{ . }} == nil {
		{{ . }}Created = false
		{{ . }}Status = v1alpha1.ServiceStatus{
			Created: instance.Spec.{{ . }}.Create,
		}
	} else if instance.Status.{{ . }}.Created == nil {
		{{ . }}Created = false
		{{ . }}Status = *instance.Status.{{ . }}
		{{ . }}Status.Created = instance.Spec.{{ . }}.Create
	} else if *instance.Status.{{ . }}.Created && !*instance.Spec.{{ . }}.Create {
		cr := cr.Get{{ . }}Cr()
		cr.Name = instance.Name
		cr.Namespace = instance.Namespace
		err := r.client.Delete(context.TODO(), cr)
		if err != nil {
			return err
		}
		instance.Status.{{ . }}.Created = instance.Spec.{{ . }}.Create
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	} else if !*instance.Status.{{ . }}.Created && *instance.Spec.{{ . }}.Create {
		{{ . }}Created = false
	}
	{{- end }}
	{{- range .KindList }}
	if !{{ . }}Created{
		if instance.Spec.{{ . }} != nil{
			{{ . }}Created := instance.Spec.{{ . }}.Create
			if *{{ . }}Created{
				cr := cr.Get{{ . }}Cr()
				cr.Spec.Service = instance.Spec.{{ . }}
				cr.Name = instance.Name
				cr.Namespace = instance.Namespace
				err = r.CreateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				/*
				err = r.UpdateResource(instance, cr, cr.Name, cr.Namespace)
				if err != nil {
					return err
				}
				*/	
			}
			instance.Status.{{ . }} = &{{ . }}Status
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}			
		}
	}
	{{- end }}
	return nil
}
	
	`))
	var kindList []string
	for crTemplate, crResource := range serviceMap {
		crFile := filePrefix + crTemplate + crFileSuffix
		yamlData, err := ioutil.ReadFile(crDirectory + crFile)
		if err != nil {
			panic(err)
		}
		crKind := strings.Split(crTemplate, ".")
		jsonData, err := yaml.YAMLToJSON([]byte(yamlData))
		if err != nil {
			panic(err)
		}
		cr := crResource
		err = yaml.Unmarshal([]byte(jsonData), &cr)
		if err != nil {
			panic(err)
		}
		f, err := os.Create("../controller/manager/crs/" + crKind[0] + ".go")
		if err != nil {
			panic(err)
		}
		schema := cr.GetObjectKind()
		groupVersionKind := schema.GroupVersionKind()
		kind := groupVersionKind.Kind

		yamlDataQuoted := fmt.Sprintf("`\n")
		yamlDataQuoted = yamlDataQuoted + string(yamlData)
		yamlDataQuoted = yamlDataQuoted + "`"
		packageTemplate.Execute(f, struct {
			YamlData string
			Kind     string
		}{
			YamlData: yamlDataQuoted,
			Kind:     kind,
		})
		kindList = append(kindList, kind)
	}
	f, err := os.Create("../controller/manager/" + crManager)
	if err != nil {
		panic(err)
	}
	crManagerTemplate.Execute(f, struct {
		KindList []string
	}{
		KindList: kindList,
	})
}

func createCrds() {

	funcMap := template.FuncMap{
		"ToLower": strings.ToLower,
	}

	var packageTemplate = template.Must(template.New("").Parse(`package crd
	
import(
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"github.com/ghodss/yaml"
)

var yamlData{{ .Kind }} = {{ .YamlData }}

func Get{{ .Kind }}Crd() *apiextensionsv1beta1.CustomResourceDefinition{
	crd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := yaml.Unmarshal([]byte(yamlData{{ .Kind }}), &crd)
	if err != nil {
		panic(err)
	}
	jsonData, err := yaml.YAMLToJSON([]byte(yamlData{{ .Kind }}))
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
	var crdManagerTemplate = template.Must(template.New("").Funcs(funcMap).Parse(`package manager
	
import(
	"context"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/api/errors"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"k8s.io/apimachinery/pkg/runtime"
	{{- range .KindList }}
	"github.com/michaelhenkel/contrail-manager/pkg/controller/{{ . | ToLower }}"
	{{- end }}
)

func (r *ReconcileManager) createCrd(instance *contrailv1alpha1.Manager, crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Reconciling Manager")
	newCrd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crd.Spec.Names.Plural + "." + crd.Spec.Group, Namespace: newCrd.Namespace}, &newCrd)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group +" not found. Creating it.")
		controllerutil.SetControllerReference(&newCrd, &newCrd, r.scheme)
		err = r.client.Create(context.TODO(), crd)	
		if err != nil {
			reqLogger.Error(err, "Failed to create new crd.", "crd.Namespace", crd.Namespace, "crd.Name", crd.Name)
			return err
		}
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group +" created.")
	}
	return nil
}

func (r *ReconcileManager) updateCrd(instance *contrailv1alpha1.Manager, crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Reconciling Manager")
	newCrd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crd.Spec.Names.Plural + "." + crd.Spec.Group, Namespace: newCrd.Namespace}, &newCrd)
	if err != nil{
		reqLogger.Info("Failed to get CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group)
		return err
	}
	controllerutil.SetControllerReference(instance, &newCrd, r.scheme)
	err = r.client.Update(context.TODO(), &newCrd)
	if err != nil{
		reqLogger.Info("Resource version: " + crd.ObjectMeta.ResourceVersion)
		reqLogger.Info("Failed to update CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group)
		return err
	}
	reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group +" updated.")
	return nil
}


func (r *ReconcileManager) ActivateResource(instance *contrailv1alpha1.Manager,
	ro runtime.Object,
	crd *apiextensionsv1beta1.CustomResourceDefinition) error {
		err := r.createCrd(instance, crd)
		if err != nil {
			return err
		}
		/*
		err = r.updateCrd(instance, crd)
		if err != nil {
			return err
		}
		*/
		err = r.addWatch(ro)
		if err != nil {
			return err
		}	
		return nil
}


func (r *ReconcileManager) ManageCrd(instance *contrailv1alpha1.Manager) error{
	var err error
	{{- range .KindList }}
	var {{ . }}Status contrailv1alpha1.ServiceStatus
	{{- end }}	

	{{- range .KindList }}
	{{ . }}Active := true
	if instance.Status.{{ . }} == nil {
		{{ . }}Active = false
		active := true
		{{ . }}Status = contrailv1alpha1.ServiceStatus{
			Active: &active,
		}
	} else if instance.Status.{{ . }}.Active == nil {
		{{ . }}Active = false
		active := true
		{{ . }}Status = *instance.Status.{{ . }}
		{{ . }}Status.Active = &active

	}
	{{- end }}
	{{- range .KindList }}
	if !{{ . }}Active{
		if instance.Spec.{{ . }} != nil {
			{{ . }}Activated := instance.Spec.{{ . }}.Activate
			if *{{ . }}Activated{
				resource := contrailv1alpha1.{{ . }}{}
				err = r.ActivateResource(instance, &resource, crds.Get{{ . }}Crd())
				if err != nil {
					return err
				}
			}
		}
	}
	{{- end }}
	{{- range .KindList }}
	if !{{ . }}Active{
		if instance.Spec.{{ . }} != nil {
			{{ . }}Activated := instance.Spec.{{ . }}.Activate
			if *{{ . }}Activated{
				err = {{ . | ToLower }}.Add(r.manager)
				if err != nil {
					return err
				}
			}
			instance.Status.{{ . }} = &{{ . }}Status
			err := r.client.Status().Update(context.TODO(), instance)
			if err != nil {
				return err
			}
		}
	}
	{{- end }}
	return nil
}
	
	`))

	//crdList := []string{configCrd}
	var kindList []string
	for crdTemplate, _ := range serviceMap {
		crdFile := filePrefix + crdTemplate + crdFileSuffix
		yamlData, err := ioutil.ReadFile(crdDirectory + crdFile)
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
		packageTemplate.Execute(f, struct {
			YamlData string
			Kind     string
		}{
			YamlData: yamlDataQuoted,
			Kind:     crd.Spec.Names.Kind,
		})
		kindList = append(kindList, crd.Spec.Names.Kind)
	}
	f, err := os.Create("../controller/manager/" + crdManager)
	if err != nil {
		panic(err)
	}
	crdManagerTemplate.Execute(f, struct {
		KindList []string
	}{
		KindList: kindList,
	})
}
