package manager
	
import(
	"context"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/api/errors"
	apiextensionsv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	crds "github.com/michaelhenkel/contrail-manager/pkg/controller/manager/crds"
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"k8s.io/apimachinery/pkg/runtime"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/config"
)

func (r *ReconcileManager) createCrd(instance *contrailv1alpha1.Manager, crd *apiextensionsv1beta1.CustomResourceDefinition) error {
	reqLogger := log.WithValues("Instance.Namespace", instance.Namespace, "Instance.Name", instance.Name)
	reqLogger.Info("Reconciling Manager")
	newCrd := apiextensionsv1beta1.CustomResourceDefinition{}
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: crd.Spec.Names.Plural + "." + crd.Spec.Group, Namespace: newCrd.Namespace}, &newCrd)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("CRD " + crd.Spec.Names.Plural + "." + crd.Spec.Group +" not found. Creating it.")
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
		err = r.updateCrd(instance, crd)
		if err != nil {
			return err
		}
		err = r.addWatch(ro)
		if err != nil {
			return err
		}	
		return nil
}


func (r *ReconcileManager) ActivateService(instance *contrailv1alpha1.Manager) error{
	var err error
	ConfigActivated := instance.Spec.Config.Activate
	if *ConfigActivated{

		resource := contrailv1alpha1.Config{}

		err = r.ActivateResource(instance, &resource, crds.GetConfigCrd())
		if err != nil {
			return err
		}

		err = config.Add(r.manager)
		if err != nil {
			return err
		}
	}
	return nil
}
	
	