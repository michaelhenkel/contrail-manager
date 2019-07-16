package utils

import (
	"context"
	"strings"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/util/workqueue"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// const defines the condsts
const (
	CASSANDRA  = "Cassandra.contrail.juniper.net"
	ZOOKEEPER  = "Zookeeper.contrail.juniper.net"
	MANAGER    = "Manager.contrail.juniper.net"
	REPLICASET = "ReplicaSet.apps"
	DEPLOYMENT = "Deployment.apps"
)

// GetGroupKindFromObject return GK
func GetGroupKindFromObject(object runtime.Object) schema.GroupKind {
	objectKind := object.GetObjectKind()
	objectGroupVersionKind := objectKind.GroupVersionKind()
	return objectGroupVersionKind.GroupKind()
}

// ReconcileNilNotFound reconciles
func ReconcileNilNotFound(err error) (reconcile.Result, error) {
	if err != nil {
		if errors.IsNotFound(err) {
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// ReconcileErr reconciles
func ReconcileErr(err error) (reconcile.Result, error) {
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}

// AppHandler handles
func AppHandler(appGroupKind schema.GroupKind, requestFor string) handler.Funcs {
	appHandler := handler.Funcs{

		CreateFunc: func(e event.CreateEvent, q workqueue.RateLimitingInterface) {
			q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      appGroupKind.String() + "/" + e.Meta.GetName(),
				Namespace: e.Meta.GetNamespace(),
			}})
		},
		UpdateFunc: func(e event.UpdateEvent, q workqueue.RateLimitingInterface) {
			q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      appGroupKind.String() + "/" + e.MetaNew.GetName(),
				Namespace: e.MetaNew.GetNamespace(),
			}})
		},
		DeleteFunc: func(e event.DeleteEvent, q workqueue.RateLimitingInterface) {
			q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      appGroupKind.String() + "/" + e.Meta.GetName(),
				Namespace: e.Meta.GetNamespace(),
			}})
		},
		GenericFunc: func(e event.GenericEvent, q workqueue.RateLimitingInterface) {
			q.Add(reconcile.Request{NamespacedName: types.NamespacedName{
				Name:      appGroupKind.String() + "/" + e.Meta.GetName(),
				Namespace: e.Meta.GetNamespace(),
			}})
		},
	}
	return appHandler
}

// CassandraGroupKind returns group kind
func CassandraGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(CASSANDRA)
}

// ZookeeperGroupKind returns group kind
func ZookeeperGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(ZOOKEEPER)
}

// ReplicaSetGroupKind returns group kind
func ReplicaSetGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(REPLICASET)
}

// ManagerGroupKind returns group kind
func ManagerGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(MANAGER)
}

// DeploymentGroupKind returns group kind
func DeploymentGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(DEPLOYMENT)
}

// AppSizeChange returns
func AppSizeChange(appGroupKind schema.GroupKind) predicate.Funcs {
	pred := predicate.Funcs{}
	switch appGroupKind {
	case CassandraGroupKind():
		pred = predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldInstance := e.ObjectOld.(*v1alpha1.Cassandra)
				newInstance := e.ObjectNew.(*v1alpha1.Cassandra)
				return oldInstance.Spec.Size != newInstance.Spec.Size
			},
		}
	case ZookeeperGroupKind():
		pred = predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldInstance := e.ObjectOld.(*v1alpha1.Zookeeper)
				newInstance := e.ObjectNew.(*v1alpha1.Zookeeper)
				return oldInstance.Spec.Size != newInstance.Spec.Size
			},
		}
	case ManagerGroupKind():

	}
	return pred
}

// ManagerSizeChange monitors per application size change
func ManagerSizeChange(appGroupKind schema.GroupKind) predicate.Funcs {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {

			oldManager := e.ObjectOld.(*v1alpha1.Manager)
			newManager := e.ObjectNew.(*v1alpha1.Manager)
			var oldSize, newSize int32
			switch appGroupKind {
			case CassandraGroupKind():
				if oldManager.Spec.Services.Cassandra.Size != nil {
					oldSize = *oldManager.Spec.Services.Cassandra.Size
				} else {
					oldSize = *oldManager.Spec.Size
				}
				if newManager.Spec.Services.Cassandra.Size != nil {
					newSize = *newManager.Spec.Services.Cassandra.Size
				} else {
					newSize = *newManager.Spec.Size
				}
			}
			if oldSize != newSize {
				return true
			}
			return false
		},
	}
	return pred
}

// DeploymentStatusChange monitors per application size change
func DeploymentStatusChange(appGroupKind schema.GroupKind) predicate.Funcs {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldDeployment := e.ObjectOld.(*appsv1.Deployment)
			newDeployment := e.ObjectNew.(*appsv1.Deployment)
			isOwner := false
			for _, owner := range newDeployment.ObjectMeta.OwnerReferences {
				if *owner.Controller {
					groupVersionKind := schema.FromAPIVersionAndKind(owner.APIVersion, owner.Name)
					if appGroupKind == groupVersionKind.GroupKind() {
						isOwner = true
					}
				}
			}
			if (oldDeployment.Status.ReadyReplicas != newDeployment.Status.ReadyReplicas) && isOwner {
				return true
			}
			return false
		},
	}
	return pred
}

// PodIPChange returns predicate function based on group kind
func PodIPChange(appLabel map[string]string) predicate.Funcs {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			for key, value := range e.MetaOld.GetLabels() {
				if appLabel[key] == value {
					oldPod := e.ObjectOld.(*corev1.Pod)
					newPod := e.ObjectNew.(*corev1.Pod)
					return oldPod.Status.PodIP != newPod.Status.PodIP
				}
			}
			return false
		},
	}
	return pred
}

// PodInitStatusChange returns predicate function based on group kind
func PodInitStatusChange(appLabel map[string]string) predicate.Funcs {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			for key, value := range e.MetaOld.GetLabels() {
				if appLabel[key] == value {
					oldPod := e.ObjectOld.(*corev1.Pod)
					newPod := e.ObjectNew.(*corev1.Pod)
					newPodReady := true
					oldPodReady := true
					if newPod.Status.InitContainerStatuses == nil {
						newPodReady = false
					}
					if oldPod.Status.InitContainerStatuses == nil {
						newPodReady = false
					}
					for _, initContainerStatus := range newPod.Status.InitContainerStatuses {
						if initContainerStatus.Name == "init" {
							if !initContainerStatus.Ready {
								newPodReady = false
							}
						}
					}
					for _, initContainerStatus := range oldPod.Status.InitContainerStatuses {
						if initContainerStatus.Name == "init" {
							if !initContainerStatus.Ready {
								oldPodReady = false
							}
						}
					}
					return newPodReady != oldPodReady
				}
			}
			return false
		},
	}
	return pred
}

//GetObjectAndGroupKindFromRequest returns Object and Kind
func GetObjectAndGroupKindFromRequest(request *reconcile.Request, client client.Client) (runtime.Object, *schema.GroupKind, error) {
	appGroupKind := schema.ParseGroupKind(strings.Split(request.Name, "/")[0])
	appName := strings.Split(request.Name, "/")[1]
	var instance runtime.Object
	switch appGroupKind {
	case CassandraGroupKind():
		instance = &v1alpha1.Cassandra{}
	case ZookeeperGroupKind():
		instance = &v1alpha1.Zookeeper{}
	case ManagerGroupKind():
		instance = &v1alpha1.Manager{}
	case ReplicaSetGroupKind():
		instance = &appsv1.ReplicaSet{}
	case DeploymentGroupKind():
		instance = &appsv1.Deployment{}
	}
	request.Name = appName
	err := client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			return nil, nil, err
		}
		return nil, nil, err
	}
	return instance, &appGroupKind, nil
}

//SetRequestName returns GroupKind
func SetRequestName(request *reconcile.Request) *schema.GroupKind {
	groupKind := schema.ParseGroupKind(strings.Split(request.Name, "/")[0])
	name := strings.Split(request.Name, "/")[1]
	request.Name = name
	return &groupKind
}
