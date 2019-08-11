package vrouter

import (
	contrailv1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_vrouter")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Vrouter Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileVrouter{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("vrouter-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Vrouter
	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Vrouter{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Vrouter

	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Config{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Manager{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &contrailv1alpha1.Control{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &contrailv1alpha1.Manager{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileVrouter implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileVrouter{}

// ReconcileVrouter reconciles a Vrouter object
type ReconcileVrouter struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Vrouter object and makes changes based on the state read
// and what is in the Vrouter.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileVrouter) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	/*
		reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
		reqLogger.Info("Reconciling Vrouter")

		// Fetch the Vrouter instance
		instance := &contrailv1alpha1.Vrouter{}
		err := r.client.Get(context.TODO(), request.NamespacedName, instance)
		if err != nil {
			if errors.IsNotFound(err) {
				// Request object not found, could have been deleted after reconcile request.
				// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
				// Return and don't requeue
				return reconcile.Result{}, nil
			}
			// Error reading the object - requeue the request.
			return reconcile.Result{}, err
		}

		// Get cassandra, zk and rmq status

		controlInstance := &contrailv1alpha1.Control{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, controlInstance)
		if err != nil {
			if errors.IsNotFound(err) {
				reqLogger.Info("No Control Instance")
				return reconcile.Result{}, err
			}
		}
		controlStatus := false
		if controlInstance.Status.Active != nil {
			controlStatus = *controlInstance.Status.Active
		}

		configInstance := &contrailv1alpha1.Config{}
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, configInstance)
		if err != nil {
			if errors.IsNotFound(err) {
				reqLogger.Info("No Config Instance")
				return reconcile.Result{}, err
			}
		}
		configStatus := false
		if configInstance.Status.Active != nil {
			configStatus = *configInstance.Status.Active
		}

		if !controlStatus || !configStatus {
			reqLogger.Info("control or config not ready")
			return reconcile.Result{}, nil
		}

		var configNodes []string
		for _, ip := range configInstance.Status.Nodes {
			configNodes = append(configNodes, ip)
		}
		configNodeList := strings.Join(configNodes, ",")

		var controlNodes []string
		for _, ip := range controlInstance.Status.Nodes {
			controlNodes = append(controlNodes, ip)
		}
		controlNodeList := strings.Join(controlNodes, ",")

		managerInstance := &contrailv1alpha1.Manager{}
		err = r.client.Get(context.TODO(), request.NamespacedName, managerInstance)
		if err != nil {
			if errors.IsNotFound(err) {
				reqLogger.Info("No Manager Instance")
			}
		} else {
			instance.Spec = managerInstance.Spec.Services.Vrouter
			if managerInstance.Spec.Services.Vrouter.Size != nil {
				instance.Spec.Size = managerInstance.Spec.Services.Vrouter.Size
			} else {
				instance.Spec.Size = managerInstance.Spec.Size
			}
			if managerInstance.Spec.HostNetwork != nil {
				instance.Spec.HostNetwork = managerInstance.Spec.HostNetwork
			}
		}

		// Get default Daemonset
		daemonset := GetDaemonset()

		if managerInstance.Spec.ImagePullSecrets != nil {
			var imagePullSecretsList []corev1.LocalObjectReference
			for _, imagePullSecretName := range managerInstance.Spec.ImagePullSecrets {
				imagePullSecret := corev1.LocalObjectReference{
					Name: imagePullSecretName,
				}
				imagePullSecretsList = append(imagePullSecretsList, imagePullSecret)
			}
			daemonset.Spec.Template.Spec.ImagePullSecrets = imagePullSecretsList
		}

		if instance.Spec.Configuration == nil {
			instance.Spec.Configuration = make(map[string]string)
			reqLogger.Info("config map empty, initializing it")
		}

		instance.Spec.Configuration["CONFIG_NODES"] = configNodeList
		instance.Spec.Configuration["ANALYTICS_NODES"] = configNodeList
		instance.Spec.Configuration["CONTROL_NODES"] = controlNodeList

		// Create initial ConfigMap
		configMap := corev1.ConfigMap{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "vrouter-" + instance.Name,
				Namespace: instance.Namespace,
			},
			Data: instance.Spec.Configuration,
		}
		controllerutil.SetControllerReference(instance, &configMap, r.scheme)
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: "vrouter-" + instance.Name, Namespace: instance.Namespace}, &configMap)
		if err != nil && errors.IsNotFound(err) {
			err = r.client.Create(context.TODO(), &configMap)
			if err != nil {
				reqLogger.Error(err, "Failed to create ConfigMap", "Namespace", instance.Namespace, "Name", "vrouter-"+instance.Name)
				return reconcile.Result{}, err
			}
		}

		// Set Daemonset Name & Namespace

		daemonset.ObjectMeta.Name = "vrouter-" + instance.Name
		daemonset.ObjectMeta.Namespace = instance.Namespace

		// Configure Containers
		for idx, container := range daemonset.Spec.Template.Spec.Containers {
			for containerName, image := range instance.Spec.Images {
				if containerName == container.Name {
					(&daemonset.Spec.Template.Spec.Containers[idx]).Image = image
					(&daemonset.Spec.Template.Spec.Containers[idx]).EnvFrom[0].ConfigMapRef.Name = "vrouter-" + instance.Name
				}
			}
		}

		// Configure InitContainers
		for idx, container := range daemonset.Spec.Template.Spec.InitContainers {
			for containerName, image := range instance.Spec.Images {
				if containerName == container.Name {
					(&daemonset.Spec.Template.Spec.InitContainers[idx]).Image = image
					(&daemonset.Spec.Template.Spec.InitContainers[idx]).EnvFrom[0].ConfigMapRef.Name = "vrouter-" + instance.Name
				}
			}
		}

		// Set HostNetwork
		daemonset.Spec.Template.Spec.HostNetwork = *instance.Spec.HostNetwork

		// Set Selector and Label
		daemonset.Spec.Selector.MatchLabels["app"] = "vrouter-" + instance.Name
		daemonset.Spec.Template.ObjectMeta.Labels["app"] = "vrouter-" + instance.Name

		// Create Daemonset
		controllerutil.SetControllerReference(instance, daemonset, r.scheme)
		err = r.client.Get(context.TODO(), types.NamespacedName{Name: "vrouter-" + instance.Name, Namespace: instance.Namespace}, daemonset)
		if err != nil && errors.IsNotFound(err) {
			err = r.client.Create(context.TODO(), daemonset)
			if err != nil {
				reqLogger.Error(err, "Failed to create Daemonset", "Namespace", instance.Namespace, "Name", "vrouter-"+instance.Name)
				return reconcile.Result{}, err
			}
		}
	*/
	return reconcile.Result{}, nil
}
