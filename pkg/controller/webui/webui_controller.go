package webui

import (
	"context"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_webui")

func resourceHandler(myclient client.Client) handler.Funcs {
	appHandler := handler.Funcs{
		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			listOps := &client.ListOptions{Namespace: e.Meta.GetNamespace()}
			list := &v1alpha1.WebuiList{}
			err := myclient.List(context.TODO(), listOps, list)
			if err == nil {
				for _, app := range list.Items {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      app.GetName(),
						Namespace: e.Meta.GetNamespace(),
					}})
				}
			}
		},
		UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
			listOps := &client.ListOptions{Namespace: e.MetaNew.GetNamespace()}
			list := &v1alpha1.WebuiList{}
			err := myclient.List(context.TODO(), listOps, list)
			if err == nil {
				for _, app := range list.Items {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      app.GetName(),
						Namespace: e.MetaNew.GetNamespace(),
					}})
				}
			}
		},
		DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
			listOps := &client.ListOptions{Namespace: e.Meta.GetNamespace()}
			list := &v1alpha1.WebuiList{}
			err := myclient.List(context.TODO(), listOps, list)
			if err == nil {
				for _, app := range list.Items {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      app.GetName(),
						Namespace: e.Meta.GetNamespace(),
					}})
				}
			}
		},
		GenericFunc: func(e event.GenericEvent, q workqueue.RateLimitingInterface) {
			listOps := &client.ListOptions{Namespace: e.Meta.GetNamespace()}
			list := &v1alpha1.WebuiList{}
			err := myclient.List(context.TODO(), listOps, list)
			if err == nil {
				for _, app := range list.Items {
					q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
						Name:      app.GetName(),
						Namespace: e.Meta.GetNamespace(),
					}})
				}
			}
		},
	}
	return appHandler
}

// Add creates a new Webui Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileWebui{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("webui-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Webui
	err = c.Watch(&source.Kind{Type: &v1alpha1.Webui{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	srcPod := &source.Kind{Type: &corev1.Pod{}}
	podHandler := resourceHandler(mgr.GetClient())
	predInitStatus := utils.PodInitStatusChange(map[string]string{"contrail_manager": "webui"})
	predPodIPChange := utils.PodIPChange(map[string]string{"contrail_manager": "webui"})
	err = c.Watch(srcPod, podHandler, predPodIPChange)
	if err != nil {
		return err
	}
	err = c.Watch(srcPod, podHandler, predInitStatus)
	if err != nil {
		return err
	}

	srcCassandra := &source.Kind{Type: &v1alpha1.Cassandra{}}
	cassandraHandler := resourceHandler(mgr.GetClient())
	predCassandraSizeChange := utils.CassandraActiveChange()
	err = c.Watch(srcCassandra, cassandraHandler, predCassandraSizeChange)
	if err != nil {
		return err
	}

	srcConfig := &source.Kind{Type: &v1alpha1.Config{}}
	configHandler := resourceHandler(mgr.GetClient())
	predConfigSizeChange := utils.ConfigActiveChange()
	err = c.Watch(srcConfig, configHandler, predConfigSizeChange)
	if err != nil {
		return err
	}

	srcDeployment := &source.Kind{Type: &appsv1.Deployment{}}
	deploymentHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Webui{},
	}
	deploymentPred := utils.DeploymentStatusChange(utils.WebuiGroupKind())
	err = c.Watch(srcDeployment, deploymentHandler, deploymentPred)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileWebui implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileWebui{}

// ReconcileWebui reconciles a Webui object
type ReconcileWebui struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Webui object and makes changes based on the state read
// and what is in the Webui.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileWebui) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	var err error
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Webui")
	instanceType := "webui"
	instance := &v1alpha1.Webui{}
	configInstance := v1alpha1.Config{}

	var resourceObject v1alpha1.ResourceObject = instance
	var resourceConfiguration v1alpha1.ResourceConfiguration = instance
	var resourceStatus v1alpha1.ResourceStatus = instance
	err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil && errors.IsNotFound(err) {
		return reconcile.Result{}, nil
	}
	configActive := configInstance.IsActive(instance.Labels["contrail_cluster"],
		request.Namespace, r.Client)
	if !configActive {
		return reconcile.Result{}, nil
	}

	managerInstance, err := instance.OwnedByManager(r.Client, request)
	if err != nil {
		return reconcile.Result{}, err
	}
	if managerInstance != nil {
		if managerInstance.Spec.Services.Webui != nil {
			webuiManagerInstance := managerInstance.Spec.Services.Webui
			instance.Spec.CommonConfiguration = utils.MergeCommonConfiguration(
				managerInstance.Spec.CommonConfiguration,
				webuiManagerInstance.Spec.CommonConfiguration)
			err = r.Client.Update(context.TODO(), instance)
			if err != nil {
				return reconcile.Result{}, err
			}
		}
	}
	configMap, err := resourceObject.CreateConfigMap(request.Name+"-"+instanceType+"-configmap",
		r.Client,
		r.Scheme,
		request)
	if err != nil {
		return reconcile.Result{}, err
	}

	intendedDeployment, err := resourceObject.PrepareIntendedDeployment(GetDeployment(),
		&instance.Spec.CommonConfiguration,
		request,
		r.Scheme)
	if err != nil {
		return reconcile.Result{}, err
	}

	resourceObject.AddVolumesToIntendedDeployments(intendedDeployment,
		map[string]string{configMap.Name: request.Name + "-" + instanceType + "-volume"})

	for idx, container := range intendedDeployment.Spec.Template.Spec.Containers {
		if container.Name == "webuiweb" {
			envList := (&intendedDeployment.Spec.Template.Spec.Containers[idx]).Env
			env := corev1.EnvVar{
				Name:  "SSL_ENABLE",
				Value: "true",
			}
			envList = append(envList, env)
			env = corev1.EnvVar{
				Name:  "SERVER_CERTFILE",
				Value: "/etc/contrail/webui_ssl/cs-cert.pem",
			}
			envList = append(envList, env)
			env = corev1.EnvVar{
				Name:  "SERVER_KEYFILE",
				Value: "/etc/contrail/webui_ssl/cs-key.pem",
			}
			envList = append(envList, env)
			env = corev1.EnvVar{
				Name:  "SERVER_CA_KEYFILE",
				Value: "",
			}
			envList = append(envList, env)
			env = corev1.EnvVar{
				Name:  "SERVER_CA_CERTFILE",
				Value: "",
			}
			envList = append(envList, env)
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Env = envList
			command := []string{"bash", "-c",
				"/certs-init.sh; sleep 5; /usr/bin/node /usr/src/contrail/contrail-web-core/webServerStart.js --conf_file /etc/mycontrail/config.global.js.${POD_IP}"}
			//command = []string{"bash", "-c",
			//	"SSL_ENABLE=true SERVER_CERTFILE=\"$WEBUI_SSL_CERT_FILE\" SERVER_KEYFILE=\"$WEBUI_SSL_KEY_FILE\" SERVER_CA_KEYFILE='' SERVER_CA_CERTFILE='' /certs-init.sh && while true; do echo hello; sleep 10;done"}
			command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Command = command

			volumeMountList := []corev1.VolumeMount{}
			if len((&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts) > 0 {
				volumeMountList = (&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts
			}
			volumeMount := corev1.VolumeMount{
				Name:      request.Name + "-" + instanceType + "-volume",
				MountPath: "/etc/mycontrail",
			}
			volumeMountList = append(volumeMountList, volumeMount)
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Image = instance.Spec.ServiceConfiguration.Images[container.Name]
		}
		if container.Name == "webuijob" {
			envList := (&intendedDeployment.Spec.Template.Spec.Containers[idx]).Env
			env := corev1.EnvVar{
				Name:  "SSL_ENABLE",
				Value: "true",
			}
			envList = append(envList, env)
			env = corev1.EnvVar{
				Name:  "SERVER_CERTFILE",
				Value: "/etc/contrail/webui_ssl/cs-cert.pem",
			}
			envList = append(envList, env)
			env = corev1.EnvVar{
				Name:  "SERVER_KEYFILE",
				Value: "/etc/contrail/webui_ssl/cs-key.pem",
			}
			envList = append(envList, env)
			env = corev1.EnvVar{
				Name:  "SERVER_CA_KEYFILE",
				Value: "",
			}
			envList = append(envList, env)
			env = corev1.EnvVar{
				Name:  "SERVER_CA_CERTFILE",
				Value: "",
			}
			envList = append(envList, env)
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Env = envList
			command := []string{"bash", "-c",
				"/certs-init.sh; sleep 5;/usr/bin/node /usr/src/contrail/contrail-web-core/jobServerStart.js --conf_file /etc/mycontrail/config.global.js.${POD_IP}"}
			//command = []string{"bash", "-c",
			//	"SSL_ENABLE=true SERVER_CERTFILE=\"$WEBUI_SSL_CERT_FILE\" SERVER_KEYFILE=\"$WEBUI_SSL_KEY_FILE\" SERVER_CA_KEYFILE='' SERVER_CA_CERTFILE='' /certs-init.sh && while true; do echo hello; sleep 10;done"}
			command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Command = command

			volumeMountList := []corev1.VolumeMount{}
			if len((&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts) > 0 {
				volumeMountList = (&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts
			}
			volumeMount := corev1.VolumeMount{
				Name:      request.Name + "-" + instanceType + "-volume",
				MountPath: "/etc/mycontrail",
			}
			volumeMountList = append(volumeMountList, volumeMount)
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Image = instance.Spec.ServiceConfiguration.Images[container.Name]
		}
		if container.Name == "redis" {
			command := []string{"bash", "-c",
				"redis-server --lua-time-limit 15000 --dbfilename '' --bind 127.0.0.1 ${POD_IP} --port 6380"}
			//command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Command = command

			volumeMountList := []corev1.VolumeMount{}
			if len((&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts) > 0 {
				volumeMountList = (&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts
			}
			volumeMount := corev1.VolumeMount{
				Name:      request.Name + "-" + instanceType + "-volume",
				MountPath: "/etc/mycontrail",
			}
			volumeMountList = append(volumeMountList, volumeMount)
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList
			(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Image = instance.Spec.ServiceConfiguration.Images[container.Name]
		}
	}

	for idx, container := range intendedDeployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range instance.Spec.ServiceConfiguration.Images {
			if containerName == container.Name {
				(&intendedDeployment.Spec.Template.Spec.InitContainers[idx]).Image = image
			}
		}
	}

	err = resourceConfiguration.CompareIntendedWithCurrentDeployment(intendedDeployment,
		&instance.Spec.CommonConfiguration,
		request,
		r.Scheme,
		r.Client,
		false)
	if err != nil {
		return reconcile.Result{}, err
	}

	podIPList, podIPMap, err := resourceConfiguration.PodIPListAndIPMap(request, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}
	if len(podIPList.Items) > 0 {
		err = resourceConfiguration.InstanceConfiguration(request,
			podIPList,
			r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = resourceStatus.SetPodsToReady(podIPList, r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = resourceStatus.ManageNodeStatus(podIPMap, r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	err = resourceStatus.SetInstanceActive(r.Client, &instance.Status, intendedDeployment, request)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
