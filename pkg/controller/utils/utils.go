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
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// const defines the Group constants
const (
	CASSANDRA   = "Cassandra.contrail.juniper.net"
	ZOOKEEPER   = "Zookeeper.contrail.juniper.net"
	RABBITMQ    = "Rabbitmq.contrail.juniper.net"
	CONFIG      = "Config.contrail.juniper.net"
	CONTROL     = "Control.contrail.juniper.net"
	WEBUI       = "Webui.contrail.juniper.net"
	VROUTER     = "Vrouter.contrail.juniper.net"
	KUBEMANAGER = "Kubemanager.contrail.juniper.net"
	MANAGER     = "Manager.contrail.juniper.net"
	REPLICASET  = "ReplicaSet.apps"
	DEPLOYMENT  = "Deployment.apps"
)

var err error

// GetGroupKindFromObject return GK
func GetGroupKindFromObject(object runtime.Object) schema.GroupKind {
	objectKind := object.GetObjectKind()
	objectGroupVersionKind := objectKind.GroupVersionKind()
	return objectGroupVersionKind.GroupKind()
}

// WebuiGroupKind returns group kind
func WebuiGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(WEBUI)
}

// VrouterGroupKind returns group kind
func VrouterGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(VROUTER)
}

// ControlGroupKind returns group kind
func ControlGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(CONTROL)
}

// ConfigGroupKind returns group kind
func ConfigGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(CONFIG)
}

// KubemanagerGroupKind returns group kind
func KubemanagerGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(KUBEMANAGER)
}

// CassandraGroupKind returns group kind
func CassandraGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(CASSANDRA)
}

// ZookeeperGroupKind returns group kind
func ZookeeperGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(ZOOKEEPER)
}

// RabbitmqGroupKind returns group kind
func RabbitmqGroupKind() schema.GroupKind {
	return schema.ParseGroupKind(RABBITMQ)
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

// ManagerSizeChange monitors per application size change
func ManagerSizeChange(appGroupKind schema.GroupKind) predicate.Funcs {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {

			oldManager := e.ObjectOld.(*v1alpha1.Manager)
			newManager := e.ObjectNew.(*v1alpha1.Manager)
			var oldSize, newSize int32
			switch appGroupKind {
			case CassandraGroupKind():
				for _, oldInstance := range oldManager.Spec.Services.Cassandras {
					if oldInstance.Spec.CommonConfiguration.Replicas != nil {
						oldSize = *oldInstance.Spec.CommonConfiguration.Replicas
					} else {
						oldSize = *oldManager.Spec.CommonConfiguration.Replicas
					}
					for _, newInstance := range newManager.Spec.Services.Cassandras {
						if oldInstance.Name == newInstance.Name {
							if newInstance.Spec.CommonConfiguration.Replicas != nil {
								newSize = *newInstance.Spec.CommonConfiguration.Replicas
							} else {
								newSize = *newManager.Spec.CommonConfiguration.Replicas
							}
						}
					}
				}
			case ZookeeperGroupKind():
				for _, oldInstance := range oldManager.Spec.Services.Zookeepers {
					if oldInstance.Spec.CommonConfiguration.Replicas != nil {
						oldSize = *oldInstance.Spec.CommonConfiguration.Replicas
					} else {
						oldSize = *oldManager.Spec.CommonConfiguration.Replicas
					}
					for _, newInstance := range newManager.Spec.Services.Zookeepers {
						if oldInstance.Name == newInstance.Name {
							if newInstance.Spec.CommonConfiguration.Replicas != nil {
								newSize = *newInstance.Spec.CommonConfiguration.Replicas
							} else {
								newSize = *newManager.Spec.CommonConfiguration.Replicas
							}
						}
					}
				}

			case RabbitmqGroupKind():
				oldInstance := oldManager.Spec.Services.Rabbitmq
				if oldInstance.Spec.CommonConfiguration.Replicas != nil {
					oldSize = *oldInstance.Spec.CommonConfiguration.Replicas
				} else {
					oldSize = *oldManager.Spec.CommonConfiguration.Replicas
				}
				newInstance := newManager.Spec.Services.Rabbitmq
				if oldInstance.Name == newInstance.Name {
					if newInstance.Spec.CommonConfiguration.Replicas != nil {
						newSize = *newInstance.Spec.CommonConfiguration.Replicas
					} else {
						newSize = *newManager.Spec.CommonConfiguration.Replicas
					}
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
					groupVersionKind := schema.FromAPIVersionAndKind(owner.APIVersion, owner.Kind)
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

// CassandraActiveChange returns predicate function based on group kind
func CassandraActiveChange() predicate.Funcs {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldCassandra := e.ObjectOld.(*v1alpha1.Cassandra)
			newCassandra := e.ObjectNew.(*v1alpha1.Cassandra)
			newCassandraActive := false
			oldCassandraActive := false
			if newCassandra.Status.Active != nil {
				newCassandraActive = *newCassandra.Status.Active
			}
			if oldCassandra.Status.Active != nil {
				oldCassandraActive = *oldCassandra.Status.Active
			}
			if !oldCassandraActive && newCassandraActive {
				return true
			}
			return false

		},
	}
	return pred
}

// ConfigActiveChange returns predicate function based on group kind
func ConfigActiveChange() predicate.Funcs {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldConfig := e.ObjectOld.(*v1alpha1.Config)
			newConfig := e.ObjectNew.(*v1alpha1.Config)
			newConfigActive := false
			oldConfigActive := false
			if newConfig.Status.Active != nil {
				newConfigActive = *newConfig.Status.Active
			}
			if oldConfig.Status.Active != nil {
				oldConfigActive = *oldConfig.Status.Active
			}
			if !oldConfigActive && newConfigActive {
				return true
			}
			return false

		},
	}
	return pred
}

// RabbitmqActiveChange returns predicate function based on group kind
func RabbitmqActiveChange() predicate.Funcs {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldRabbitmq := e.ObjectOld.(*v1alpha1.Rabbitmq)
			newRabbitmq := e.ObjectNew.(*v1alpha1.Rabbitmq)
			newRabbitmqActive := false
			oldRabbitmqActive := false
			if newRabbitmq.Status.Active != nil {
				newRabbitmqActive = *newRabbitmq.Status.Active
			}
			if oldRabbitmq.Status.Active != nil {
				oldRabbitmqActive = *oldRabbitmq.Status.Active
			}
			if !oldRabbitmqActive && newRabbitmqActive {
				return true
			}
			return false

		},
	}
	return pred
}

// ZookeeperActiveChange returns predicate function based on group kind
func ZookeeperActiveChange() predicate.Funcs {
	pred := predicate.Funcs{
		UpdateFunc: func(e event.UpdateEvent) bool {
			oldZookeeper := e.ObjectOld.(*v1alpha1.Zookeeper)
			newZookeeper := e.ObjectNew.(*v1alpha1.Zookeeper)
			newZookeeperActive := false
			oldZookeeperActive := false
			if newZookeeper.Status.Active != nil {
				newZookeeperActive = *newZookeeper.Status.Active
			}
			if oldZookeeper.Status.Active != nil {
				oldZookeeperActive = *oldZookeeper.Status.Active
			}
			if !oldZookeeperActive && newZookeeperActive {
				return true
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

// MergeCommonConfiguration combines common configuration of manager and service
func MergeCommonConfiguration(manager v1alpha1.CommonConfiguration,
	instance v1alpha1.CommonConfiguration) v1alpha1.CommonConfiguration {
	if len(instance.NodeSelector) == 0 && len(manager.NodeSelector) > 0 {
		instance.NodeSelector = manager.NodeSelector
	}
	//hostNetworkDefault := true
	//instance.HostNetwork = &hostNetworkDefault
	if instance.HostNetwork == nil && manager.HostNetwork != nil {
		instance.HostNetwork = manager.HostNetwork
	}
	if len(instance.ImagePullSecrets) == 0 && len(manager.ImagePullSecrets) > 0 {
		instance.ImagePullSecrets = manager.ImagePullSecrets
	}
	if len(instance.Tolerations) == 0 && len(manager.Tolerations) > 0 {
		instance.Tolerations = manager.Tolerations
	}
	//var defaultReplicas int32 = 1
	//instance.Replicas = &defaultReplicas
	if instance.Replicas == nil && manager.Replicas != nil {
		instance.Replicas = manager.Replicas
	}
	return instance
}
