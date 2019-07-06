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

import (
	"context"

	"github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	cr "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crs"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func (r *ReconcileManager) CreateResource(instance *v1alpha1.Manager, obj runtime.Object, name string, namespace string) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Create CR")

	objectKind := obj.GetObjectKind()
	groupVersionKind := objectKind.GroupVersionKind()

	gkv := schema.FromAPIVersionAndKind(groupVersionKind.Group+"/"+groupVersionKind.Version, groupVersionKind.Kind)
	newObj, err := scheme.Scheme.New(gkv)
	if err != nil {
		return err
	}

	newObj = obj

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: name, Namespace: namespace}, newObj)
	if err != nil && errors.IsNotFound(err) {

		switch groupVersionKind.Kind {
		{{- range .KindList }}
		case "{{ . }}":
			typedObject := &v1alpha1.{{ . }}{}
			typedObject = newObj.(*v1alpha1.{{ . }})
			controllerutil.SetControllerReference(instance, typedObject, r.scheme)
			crLabel := map[string]string{"contrail_manager": "{{ . | ToLower }}"}
			typedObject.ObjectMeta.SetLabels(crLabel)
			err = r.client.Create(context.TODO(), typedObject)
			if err != nil {
				reqLogger.Info("Failed to create CR " + name)
				return err
			}
			reqLogger.Info("CR " + name + " Created.")
			return nil
		{{- end }}

		}
	}
	return nil
}

func (r *ReconcileManager) ManageCr(request reconcile.Request) error {
	var err error
	instance := &v1alpha1.Manager{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return err
		}
		return err
	}
	{{- range .KindList }}
	{{ . | ToLower }}CreationStatus := false
	if instance.Status.{{ . }} != nil {
		if instance.Status.{{ . }}.Created != nil {
			{{ . | ToLower }}CreationStatus = *instance.Status.{{ . }}.Created
		}
	}

	{{ . | ToLower }}CreationIntent := false
	if instance.Spec.Services.{{ . }} != nil {
		if instance.Spec.Services.{{ . }}.Create != nil {
			{{ . | ToLower }}CreationIntent = *instance.Spec.Services.{{ . }}.Create
		}
	}
	if {{ . | ToLower }}CreationIntent && !{{ . | ToLower }}CreationStatus {
		//Create {{ . }}
		cr := cr.Get{{ . }}Cr()
		cr.Spec = instance.Spec.Services.{{ . }}
		cr.Name = instance.Name
		cr.Namespace = instance.Namespace
		if instance.Spec.Size != nil && cr.Spec.Size == nil {
			cr.Spec.Size = instance.Spec.Size
		}
		if instance.Spec.HostNetwork != nil && cr.Spec.HostNetwork == nil {
			cr.Spec.HostNetwork = instance.Spec.HostNetwork
		}
		err := r.CreateResource(instance, cr, cr.Name, cr.Namespace)
		if err != nil {
			return err
		}
		created := true
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, instance)
		if err != nil {
			if errors.IsNotFound(err) {
				return err
			}
			return err
		}
		if instance.Status.{{ . }} == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.{{ . }} = status
		} else {
			instance.Status.{{ . }}.Created = &created
		}
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
		err = r.client.Update(context.TODO(), instance)
		if err != nil {
			return err
		}

	}

	if !{{ . | ToLower }}CreationIntent && {{ . | ToLower }}CreationStatus {
		//Delete {{ . }}
		cr := cr.Get{{ . }}Cr()
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, cr)
		if err != nil && errors.IsNotFound(err) {
			return nil
		}
		err = r.client.Delete(context.TODO(), cr)
		if err != nil {
			return err
		}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: request.Name, Namespace: request.Namespace}, instance)
		if err != nil {
			if errors.IsNotFound(err) {
				return err
			}
			return err
		}
		created := false
		if instance.Status.{{ . }} == nil {
			status := &v1alpha1.ServiceStatus{
				Created: &created,
			}
			instance.Status.{{ . }} = status
		} else {
			instance.Status.{{ . }}.Created = &created
		}
		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
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

import (
	"context"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	{{- range .KindList }}
	"github.com/michaelhenkel/contrail-manager/pkg/controller/{{ . | ToLower }}"
	{{- end }}
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func (r *ReconcileManager) createCrd(instance *v1alpha1.Manager, crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Creating CRD")
	newCrd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crd.Spec.Names.Plural + "." + crd.Spec.Group, Namespace: newCrd.Namespace}, &newCrd)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group + " not found. Creating it.")
		//controllerutil.SetControllerReference(&newCrd, crd, r.scheme)
		err = r.client.Create(context.TODO(), crd)
		if err != nil {
			reqLogger.Error(err, "Failed to create new crd.", "crd.Namespace", crd.Namespace, "crd.Name", crd.Name)
			return err
		}
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group + " created.")
	} else if err == nil {
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group + " already exists.")
	}
	return nil
}

func (r *ReconcileManager) ManageCrd(request reconcile.Request) error {
	var err error
	instance := &v1alpha1.Manager{}
	err = r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		return err
	}
	{{- range .KindList }}
	{{ . | ToLower }}ActivationStatus := false
	if instance.Status.{{ . }} != nil {
		if instance.Status.{{ . }}.Active != nil {
			{{ . | ToLower }}ActivationStatus = *instance.Status.{{ . }}.Active
		}
	}

	{{ . | ToLower }}ActivationIntent := false
	if instance.Spec.Services.{{ . }} != nil {
		if instance.Spec.Services.{{ . }}.Activate != nil {
			{{ . | ToLower }}ActivationIntent = *instance.Spec.Services.{{ . }}.Activate
		}
	}
	{{ . | ToLower }}Resource := v1alpha1.{{ . }}{}
	{{ . | ToLower }}Crd := {{ . | ToLower }}Resource.GetCrd()
	{{ . | ToLower }}ControllerActivate := false
	if {{ . | ToLower }}ActivationIntent && !{{ . | ToLower }}ActivationStatus {
		err = r.createCrd(instance, {{ . | ToLower }}Crd)
		if err != nil {
			return err
		}

		controllerRunning := false
		sharedIndexInformer, err := r.cache.GetInformerForKind({{ . | ToLower }}Resource.GroupVersionKind())
		if err == nil {
			controller := sharedIndexInformer.GetController()
			if controller != nil {
				controllerRunning = true
			}
		}
		if !controllerRunning {
			{{ . | ToLower }}ControllerActivate = true
		}

		err = r.controller.Watch(&source.Kind{Type: &v1alpha1.{{ . }}{}}, &handler.EnqueueRequestForOwner{
			IsController: true,
			OwnerType:    &v1alpha1.Manager{},
		})
		if err != nil {
			return err
		}
		active := true
		if instance.Status.{{ . }} == nil {
			status := &v1alpha1.ServiceStatus{
				Active: &active,
			}
			instance.Status.{{ . }} = status
		} else {
			instance.Status.{{ . }}.Active = &active
		}

		err = r.client.Status().Update(context.TODO(), instance)
		if err != nil {
			return err
		}
	}
	{{- end }}
	{{- range .KindList }}
	if {{ . | ToLower }}ControllerActivate{
		err = {{ . | ToLower }}.Add(r.manager)
		if err != nil {
			return err
		}
	}
	{{- end }}

	return nil
}
	
	`))

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
