package kubemanager

import (
	"context"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"github.com/michaelhenkel/contrail-manager/pkg/controller/utils"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_kubemanager")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Kubemanager Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileKubemanager{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("kubemanager-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Kubemanager
	err = c.Watch(&source.Kind{Type: &v1alpha1.Kubemanager{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to PODs
	srcPod := &source.Kind{Type: &corev1.Pod{}}
	podHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &appsv1.ReplicaSet{},
	}
	predInitStatus := utils.PodInitStatusChange(map[string]string{"contrail_manager": "kubemanager"})
	predPodIPChange := utils.PodIPChange(map[string]string{"contrail_manager": "kubemanager"})
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
	predManagerSizeChange := utils.ManagerSizeChange(utils.KubemanagerGroupKind())
	// Watch for Manager events.
	err = c.Watch(srcManager, managerHandler, predManagerSizeChange)
	if err != nil {
		return err
	}

	// Watch for changes to Manager
	srcCassandra := &source.Kind{Type: &v1alpha1.Cassandra{}}
	cassandraHandler := &handler.EnqueueRequestForObject{}
	predCassandraSizeChange := utils.CassandraActiveChange()
	// Watch for Manager events.
	err = c.Watch(srcCassandra, cassandraHandler, predCassandraSizeChange)
	if err != nil {
		return err
	}

	// Watch for changes to Manager
	srcConfig := &source.Kind{Type: &v1alpha1.Config{}}
	configHandler := &handler.EnqueueRequestForObject{}
	predConfigSizeChange := utils.ConfigActiveChange()
	// Watch for Manager events.
	err = c.Watch(srcConfig, configHandler, predConfigSizeChange)
	if err != nil {
		return err
	}

	// Watch for changes to Manager
	srcRabbitmq := &source.Kind{Type: &v1alpha1.Rabbitmq{}}
	rabbitmqHandler := &handler.EnqueueRequestForObject{}
	predRabbitmqSizeChange := utils.RabbitmqActiveChange()
	// Watch for Manager events.
	err = c.Watch(srcRabbitmq, rabbitmqHandler, predRabbitmqSizeChange)
	if err != nil {
		return err
	}

	// Watch for changes to Manager
	srcZookeeper := &source.Kind{Type: &v1alpha1.Zookeeper{}}
	zookeeperHandler := &handler.EnqueueRequestForObject{}
	predZookeeperSizeChange := utils.ZookeeperActiveChange()
	// Watch for Manager events.
	err = c.Watch(srcZookeeper, zookeeperHandler, predZookeeperSizeChange)
	if err != nil {
		return err
	}

	srcDeployment := &source.Kind{Type: &appsv1.Deployment{}}
	deploymentHandler := &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &v1alpha1.Kubemanager{},
	}
	deploymentPred := utils.DeploymentStatusChange(utils.KubemanagerGroupKind())
	err = c.Watch(srcDeployment, deploymentHandler, deploymentPred)
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileKubemanager implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileKubemanager{}

// ReconcileKubemanager reconciles a Kubemanager object
type ReconcileKubemanager struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	Client client.Client
	Scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Kubemanager object and makes changes based on the state read
// and what is in the Kubemanager.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileKubemanager) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	var err error
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Kubemanager")
	instanceType := "kubemanager"
	instance := &v1alpha1.Kubemanager{}
	var i v1alpha1.Instance = instance
	err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil && errors.IsNotFound(err) {
		isReplicaset := i.IsReplicaset(&request, instanceType, r.Client)
		isManager := i.IsManager(&request, r.Client)
		isZookeeper := i.IsZookeeper(&request, r.Client)
		isRabbitmq := i.IsRabbitmq(&request, r.Client)
		isCassandra := i.IsCassandra(&request, r.Client)
		isConfig := i.IsConfig(&request, r.Client)
		if !isConfig && !isCassandra && !isRabbitmq && !isZookeeper && !isReplicaset && !isManager {
			return reconcile.Result{}, nil
		}
		err = r.Client.Get(context.TODO(), request.NamespacedName, instance)
		if err != nil {
			return reconcile.Result{}, nil
		}
	}
	cassandraActive := false
	zookeeperActive := false
	rabbitmqActive := false
	configActive := false
	cassandraActive = utils.IsCassandraActive(instance.Spec.ServiceConfiguration.CassandraInstance,
		request.Namespace, r.Client)
	zookeeperActive = utils.IsZookeeperActive(instance.Spec.ServiceConfiguration.ZookeeperInstance,
		request.Namespace, r.Client)
	rabbitmqActive = utils.IsRabbitmqActive(instance.Labels["contrail_cluster"],
		request.Namespace, r.Client)
	configActive = utils.IsConfigActive(instance.Labels["contrail_cluster"],
		request.Namespace, r.Client)
	if !configActive || !cassandraActive || !rabbitmqActive || !zookeeperActive {
		return reconcile.Result{}, nil
	}

	managerInstance, err := i.OwnedByManager(r.Client, request)
	if err != nil {
		return reconcile.Result{}, err
	}
	if managerInstance != nil {
		if managerInstance.Spec.Services.Kubemanagers != nil {
			for _, kubemanagerManagerInstance := range managerInstance.Spec.Services.Kubemanagers {
				if kubemanagerManagerInstance.Name == request.Name {
					instance.Spec.CommonConfiguration = utils.MergeCommonConfiguration(
						managerInstance.Spec.CommonConfiguration,
						kubemanagerManagerInstance.Spec.CommonConfiguration)
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

	var serviceAccountName string
	if instance.Spec.ServiceConfiguration.ServiceAccount != "" {
		serviceAccountName = instance.Spec.ServiceConfiguration.ServiceAccount
	} else {
		serviceAccountName = "contrail-service-account"
	}

	var clusterRoleName string
	if instance.Spec.ServiceConfiguration.ClusterRole != "" {
		clusterRoleName = instance.Spec.ServiceConfiguration.ClusterRole
	} else {
		clusterRoleName = "contrail-cluster-role"
	}

	var clusterRoleBindingName string
	if instance.Spec.ServiceConfiguration.ClusterRoleBinding != "" {
		clusterRoleBindingName = instance.Spec.ServiceConfiguration.ClusterRoleBinding
	} else {
		clusterRoleBindingName = "contrail-cluster-role-binding"
	}

	existingServiceAccount := &corev1.ServiceAccount{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: serviceAccountName, Namespace: instance.Namespace}, existingServiceAccount)
	if err != nil && errors.IsNotFound(err) {
		serviceAccount := &corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "v1",
				Kind:       "ServiceAccount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      serviceAccountName,
				Namespace: instance.Namespace,
			},
		}
		controllerutil.SetControllerReference(instance, serviceAccount, r.Scheme)
		err = r.Client.Create(context.TODO(), serviceAccount)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	existingClusterRole := &rbacv1.ClusterRole{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: clusterRoleName}, existingClusterRole)
	if err != nil && errors.IsNotFound(err) {
		clusterRole := &rbacv1.ClusterRole{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac/v1",
				Kind:       "ClusterRole",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterRoleName,
				Namespace: instance.Namespace,
			},
			Rules: []rbacv1.PolicyRule{{
				Verbs: []string{
					"*",
				},
				APIGroups: []string{
					"*",
				},
				Resources: []string{
					"*",
				},
			}},
		}
		controllerutil.SetControllerReference(instance, clusterRole, r.Scheme)
		err = r.Client.Create(context.TODO(), clusterRole)
		if err != nil {
			return reconcile.Result{}, err
		}
	}

	existingClusterRoleBinding := &rbacv1.ClusterRoleBinding{}
	err = r.Client.Get(context.TODO(), types.NamespacedName{Name: clusterRoleBindingName}, existingClusterRoleBinding)
	if err != nil && errors.IsNotFound(err) {
		clusterRoleBinding := &rbacv1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: "rbac/v1",
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterRoleBindingName,
				Namespace: instance.Namespace,
			},
			Subjects: []rbacv1.Subject{{
				Kind:      "ServiceAccount",
				Name:      serviceAccountName,
				Namespace: instance.Namespace,
			}},
			RoleRef: rbacv1.RoleRef{
				APIGroup: "rbac.authorization.k8s.io",
				Kind:     "ClusterRole",
				Name:     clusterRoleName,
			},
		}
		controllerutil.SetControllerReference(instance, clusterRoleBinding, r.Scheme)
		err = r.Client.Create(context.TODO(), clusterRoleBinding)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	intendedDeployment.Spec.Template.Spec.ServiceAccountName = serviceAccountName
	for idx, container := range intendedDeployment.Spec.Template.Spec.Containers {
		if container.Name == "kubemanager" {
			command := []string{"bash", "-c",
				"/usr/bin/python /usr/bin/contrail-kube-manager -c /etc/mycontrail/kubemanager.${POD_IP}"}
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

	err = i.CompareIntendedWithCurrentDeployment(intendedDeployment,
		&instance.Spec.CommonConfiguration,
		request,
		r.Scheme,
		r.Client,
		false)
	if err != nil {
		return reconcile.Result{}, err
	}

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
