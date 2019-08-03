package cassandra

import (
	"context"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_cassandra")
var err error

func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCassandra{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller

	c, err := controller.New("cassandra-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}
	// Watch for changes to primary resource Cassandra
	err = c.Watch(&source.Kind{Type: &v1alpha1.Cassandra{}},
		&handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to PODs
	srcPod := &source.Kind{Type: &corev1.Pod{}}
	podHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1.ReplicaSet{},
	}
	predInitStatus := utils.PodInitStatusChange(map[string]string{"contrail_manager": "cassandra"})
	predPodIPChange := utils.PodIPChange(map[string]string{"contrail_manager": "cassandra"})
	err = c.Watch(srcPod, podHandler, predPodIPChange)
	if err != nil {
		return err
	}
	err = c.Watch(srcPod, podHandler, predInitStatus)
	if err != nil {
		return err
	}

	// Watch for changes to Manager
	srcManager := &source.Kind{Type: &v1alpha1.Manager{}}
	managerHandler := &handler.EnqueueRequestForObject{}
	predManagerSizeChange := utils.ManagerSizeChange(utils.CassandraGroupKind())
	// Watch for Manager events.
	err = c.Watch(srcManager, managerHandler, predManagerSizeChange)
	if err != nil {
		return err
	}

	srcDeployment := &source.Kind{Type: &appsv1.Deployment{}}
	deploymentHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Cassandra{},
	}
	deploymentPred := utils.DeploymentStatusChange(utils.CassandraGroupKind())
	err = c.Watch(srcDeployment, deploymentHandler, deploymentPred)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCassandra implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCassandra{}

// ReconcileCassandra reconciles a Cassandra object
type ReconcileCassandra struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile reconciles cassandra
func (r *ReconcileCassandra) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Cassandra")
	instanceType := "cassandra"

	// A reconcile.Request for Cassandra can be triggered by 4 different types:
	// 1. Any changes on the Cassandra instance
	// --> reconcile.Request is Cassandra instance name/namespace
	// 2. IP Status change on the Pods
	// --> reconcile.Request is Replicaset name/namespace
	// --> we need to evaluate the label to get the Cassandra instance
	// 3. Status change on the Deployment
	// --> reconcile.Request is Cassandra instance name/namespace
	// 4. Cassandras changes on the Manager instance
	// --> reconcile.Request is Manager instance name/namespace

	instance := &v1alpha1.Cassandra{}
	var i v1alpha1.Instance = instance
	err := r.Client.Get(context.TODO(), request.NamespacedName, instance)
	// if not found we expect it a change in replicaset
	// and get the cassandra instance via label
	if err != nil && errors.IsNotFound(err) {
		replicaSet := &appsv1.ReplicaSet{}
		err = r.Client.Get(context.TODO(), request.NamespacedName, replicaSet)
		if err != nil {
			return reconcile.Result{}, nil
		}
		request.Name = replicaSet.Labels[instanceType]
		err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
		if err != nil {
			return reconcile.Result{}, nil
		}
	}

	managerInstance, err := i.OwnedByManager(r.Client, request)
	if err != nil {
		return reconcile.Result{}, err
	}
	if managerInstance != nil {
		if managerInstance.Spec.Services.Cassandras != nil {
			for _, cassandraManagerInstance := range managerInstance.Spec.Services.Cassandras {
				if cassandraManagerInstance.Name == request.Name {
					instance.Spec.CommonConfiguration = utils.MergeCommonConfiguration(
						managerInstance.Spec.CommonConfiguration,
						cassandraManagerInstance.Spec.CommonConfiguration)
					err = r.Client.Update(context.TODO(), instance)
					if err != nil {
						return reconcile.Result{}, err
					}
				}
			}
		}
	}

	configMap, err := i.CreateConfigMap(request.Name+"-"+instanceType+"-configmap",
		r.Client,
		r.Scheme,
		request)
	if err != nil {
		return reconcile.Result{}, err
	}

	intendedDeployment, err := i.PrepareIntendedDeployment(GetDeployment(),
		&instance.Spec.CommonConfiguration,
		request,
		r.Scheme)

	if err != nil {
		return reconcile.Result{}, err
	}

	i.AddVolumesToIntendedDeployments(intendedDeployment,
		map[string]string{configMap.Name: request.Name + "-" + instanceType + "-volume"})

	for idx, container := range intendedDeployment.Spec.Template.Spec.Containers {
		for containerName, image := range instance.Spec.ServiceConfiguration.Images {
			if containerName == container.Name {
				(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Image = image
			}
			if containerName == "cassandra" {
				command := []string{"bash", "-c",
					"/docker-entrypoint.sh cassandra -f -Dcassandra.config=file:///mydata/${POD_IP}.yaml"}
				//command = []string{"sh", "-c", "while true; do echo hello; sleep 10;done"}
				(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Command = command

				volumeMountList := []corev1.VolumeMount{}
				volumeMount := corev1.VolumeMount{
					Name:      request.Name + "-" + instanceType + "-volume",
					MountPath: "/mydata",
				}
				volumeMountList = append(volumeMountList, volumeMount)

				(&intendedDeployment.Spec.Template.Spec.Containers[idx]).VolumeMounts = volumeMountList
				var jvmOpts string
				if instance.Spec.ServiceConfiguration.MinHeapSize != "" {
					jvmOpts = "-Xms" + instance.Spec.ServiceConfiguration.MinHeapSize
				}
				if instance.Spec.ServiceConfiguration.MaxHeapSize != "" {
					jvmOpts = jvmOpts + " -Xmx" + instance.Spec.ServiceConfiguration.MaxHeapSize
				}
				if jvmOpts != "" {
					envs := (&intendedDeployment.Spec.Template.Spec.Containers[idx]).Env
					envs = append(envs)
					jvmOptEnvVar := corev1.EnvVar{
						Name:  "JVM_OPTS",
						Value: jvmOpts,
					}
					envVars := (&intendedDeployment.Spec.Template.Spec.Containers[idx]).Env
					envVars = append(envVars, jvmOptEnvVar)
					(&intendedDeployment.Spec.Template.Spec.Containers[idx]).Env = envVars
				}

			}
		}
	}

	// Configure InitContainers
	for idx, container := range intendedDeployment.Spec.Template.Spec.InitContainers {
		for containerName, image := range instance.Spec.ServiceConfiguration.Images {
			if containerName == container.Name {
				(&intendedDeployment.Spec.Template.Spec.InitContainers[idx]).Image = image
			}
		}
	}

	err = i.CompareIntendedWithCurrentDeployment(intendedDeployment,
		&instance.Spec.CommonConfiguration,
		request,
		r.Scheme,
		r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}
	/*
		err = i.ManageActiveStatus(intendedDeployment, request, r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}
	*/
	podIPList, podIPMap, err := i.GetPodIPListAndIPMap(request, r.Client)
	if err != nil {
		return reconcile.Result{}, err
	}
	if len(podIPList.Items) > 0 {
		err = i.CreateInstanceConfiguration(request,
			podIPList,
			r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = i.SetPodsToReady(podIPList, r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}

		err = i.ManageNodeStatus(podIPMap, r.Client)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	err = i.SetInstanceActive(r.Client, &instance.Status, intendedDeployment, request)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
