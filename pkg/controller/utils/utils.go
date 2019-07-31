package utils

import (
	"context"
	"strings"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
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

var err error

// SetPodsToReady sets the status field of the Pods to ready
func SetPodsToReady(podList *corev1.PodList, client client.Client) error {
	for _, pod := range podList.Items {
		pod.ObjectMeta.Labels["status"] = "ready"
		err = client.Update(context.TODO(), &pod)
		if err != nil {
			return err
		}
	}
	return nil
}

// GetPodIPListAndIPMap gets a list and a map of Pod IPs
func GetPodIPListAndIPMap(instanceType string,
	request reconcile.Request,
	reconcileClient client.Client) (*corev1.PodList, map[string]string, error) {
	var podNameIPMap = make(map[string]string)
	//instanceDeploymentName := request.Name + "-" + instanceType + "-deployment"

	labelSelector := labels.SelectorFromSet(map[string]string{"contrail_manager": instanceType,
		instanceType: request.Name})
	listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
	replicaSetList := &appsv1.ReplicaSetList{}
	err = reconcileClient.List(context.TODO(), listOps, replicaSetList)
	if err != nil {
		return &corev1.PodList{}, map[string]string{}, err
	}

	if len(replicaSetList.Items) > 0 {
		replicaSet := &appsv1.ReplicaSet{}
		for _, rs := range replicaSetList.Items {
			if *rs.Spec.Replicas > 0 {
				replicaSet = &rs
			}
		}
		podList := &corev1.PodList{}
		if podHash, ok := replicaSet.ObjectMeta.Labels["pod-template-hash"]; ok {
			labelSelector := labels.SelectorFromSet(map[string]string{"pod-template-hash": podHash})
			listOps := &client.ListOptions{Namespace: request.Namespace, LabelSelector: labelSelector}
			err := reconcileClient.List(context.TODO(), listOps, podList)
			if err != nil {
				return &corev1.PodList{}, map[string]string{}, err
			}
		}

		if int32(len(podList.Items)) == *replicaSet.Spec.Replicas {
			for _, pod := range podList.Items {
				if pod.Status.PodIP != "" {
					podNameIPMap[pod.Name] = pod.Status.PodIP
				}
			}
		}

		if int32(len(podNameIPMap)) == *replicaSet.Spec.Replicas {
			return podList, podNameIPMap, nil
		}
	}
	return &corev1.PodList{}, map[string]string{}, nil
}

// CompareIntendedWithCurrentDeployment compares intended with current deployment
// and updates the status
func CompareIntendedWithCurrentDeployment(intendedDeployment *appsv1.Deployment,
	commonConfiguration *v1alpha1.CommonConfiguration,
	instanceType string,
	request reconcile.Request,
	scheme *runtime.Scheme,
	client client.Client,
	instanceInterface interface{}) error {
	currentDeployment := &appsv1.Deployment{}
	err = client.Get(context.TODO(),
		types.NamespacedName{Name: intendedDeployment.Name, Namespace: request.Namespace},
		currentDeployment)
	if err != nil {
		if errors.IsNotFound(err) {
			err = client.Create(context.TODO(), intendedDeployment)
			if err != nil {
				return err
			}
		}
	} else {
		update := false
		if intendedDeployment.Spec.Replicas != currentDeployment.Spec.Replicas {
			update = true
		}
		for _, intendedContainer := range intendedDeployment.Spec.Template.Spec.Containers {
			for _, currentContainer := range currentDeployment.Spec.Template.Spec.Containers {
				if intendedContainer.Name == currentContainer.Name {
					if intendedContainer.Image != currentContainer.Image {
						update = true
					}
				}
			}
		}
		if update {
			err = client.Update(context.TODO(), intendedDeployment)
			if err != nil {
				return err
			}
		}
	}
	err = ManageActiveStatus(&currentDeployment.Status.ReadyReplicas,
		intendedDeployment.Spec.Replicas,
		instanceInterface,
		client)

	if err != nil {
		return err
	}
	return nil
}

// ManageActiveStatus manages the Active status field
func ManageActiveStatus(currentDeploymentReadyReplicas *int32,
	intendedDeploymentReplicas *int32,
	instanceInterface interface{},
	client client.Client) error {
	active := false
	if currentDeploymentReadyReplicas == intendedDeploymentReplicas {
		active = true
		switch instanceInterface.(type) {
		case v1alpha1.Cassandra:
			instance := instanceInterface.(*v1alpha1.Cassandra)
			instance.Status.Active = &active
		}
	}
	switch instanceInterface.(type) {
	case v1alpha1.Cassandra:
		instance := instanceInterface.(*v1alpha1.Cassandra)
		instance.Status.Active = &active
		err = client.Status().Update(context.TODO(), instance)
	}
	if err != nil {
		return err
	}
	return nil

}

func SetInstanceActive(client client.Client, status *v1alpha1.Status, deployment *appsv1.Deployment) {
	active := false
	if deployment.Status.ReadyReplicas == *deployment.Spec.Replicas {
		active = true
	}
	status.Active = &active
}

// PrepareConfigMap prepares the initial ConfigMap
func PrepareConfigMap(request reconcile.Request,
	instanceType string,
	scheme *runtime.Scheme,
	client client.Client) *corev1.ConfigMap {
	instanceConfigMapName := request.Name + "-" + instanceType + "-configmap"
	configMap := &corev1.ConfigMap{}
	configMap.SetName(instanceConfigMapName)
	configMap.SetNamespace(request.Namespace)
	configMap.SetLabels(map[string]string{"contrail_manager": instanceType,
		instanceType: request.Name})
	configMap.Data = make(map[string]string)
	return configMap
}

func CreateConfigMap(request reconcile.Request,
	instanceType string,
	configMap *corev1.ConfigMap,
	scheme *runtime.Scheme,
	client client.Client) error {
	instanceConfigMapName := request.Name + "-" + instanceType + "-configmap"
	err = client.Get(context.TODO(), types.NamespacedName{Name: instanceConfigMapName, Namespace: request.Namespace}, configMap)
	if err != nil {
		if errors.IsNotFound(err) {
			err = client.Create(context.TODO(), configMap)
			if err != nil {
				return err
			}
		}
	} else {
		err = client.Update(context.TODO(), configMap)
		if err != nil {
			return err
		}
	}
	return nil
}

// PrepareIntendedDeployment prepares the intended Deployment
func PrepareIntendedDeployment(instanceDeployment *appsv1.Deployment,
	commonConfiguration *v1alpha1.CommonConfiguration,
	instanceType string,
	request reconcile.Request,
	scheme *runtime.Scheme) *appsv1.Deployment {
	instanceDeploymentName := request.Name + "-" + instanceType + "-deployment"
	instanceConfigMapName := request.Name + "-" + instanceType + "-configmap"
	instanceVolumeName := request.Name + "-" + instanceType + "-volume"
	intendedDeployment := SetDeploymentCommonConfiguration(instanceDeployment, commonConfiguration)
	intendedDeployment.SetName(instanceDeploymentName)
	intendedDeployment.SetNamespace(request.Namespace)
	intendedDeployment.SetLabels(map[string]string{"contrail_manager": instanceType,
		instanceType: request.Name})
	intendedDeployment.Spec.Selector.MatchLabels = map[string]string{"contrail_manager": instanceType,
		instanceType: request.Name}
	intendedDeployment.Spec.Template.SetLabels(map[string]string{"contrail_manager": instanceType,
		instanceType: request.Name})

	volumeList := intendedDeployment.Spec.Template.Spec.Volumes
	volume := corev1.Volume{
		Name: instanceVolumeName,
		VolumeSource: corev1.VolumeSource{
			ConfigMap: &corev1.ConfigMapVolumeSource{
				LocalObjectReference: corev1.LocalObjectReference{
					Name: instanceConfigMapName,
				},
			},
		},
	}
	volumeList = append(volumeList, volume)

	intendedDeployment.Spec.Template.Spec.Volumes = volumeList
	return intendedDeployment
}

// GetInstanceFromList return object from list of objects
func GetInstanceFromList(objectList []*interface{}, request reconcile.Request) interface{} {

	for _, object := range objectList {
		ob := *object
		switch ob.(type) {
		case v1alpha1.Cassandra:
			cassandraObject := ob.(v1alpha1.Cassandra)
			if cassandraObject.GetName() == request.Name {
				return cassandraObject
			}
		}
	}
	return nil
}

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
				return oldInstance.Spec.CommonConfiguration.Replicas != newInstance.Spec.CommonConfiguration.Replicas
			},
		}
	case ZookeeperGroupKind():
		pred = predicate.Funcs{
			UpdateFunc: func(e event.UpdateEvent) bool {
				oldInstance := e.ObjectOld.(*v1alpha1.Zookeeper)
				newInstance := e.ObjectNew.(*v1alpha1.Zookeeper)
				return oldInstance.Spec.CommonConfiguration.Replicas != newInstance.Spec.CommonConfiguration.Replicas
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

func SetDeploymentCommonConfiguration(deployment *appsv1.Deployment,
	commonConfiguration *v1alpha1.CommonConfiguration) *appsv1.Deployment {
	deployment.Spec.Replicas = commonConfiguration.Replicas
	deployment.Spec.Template.Spec.Tolerations = commonConfiguration.Tolerations
	deployment.Spec.Template.Spec.NodeSelector = commonConfiguration.NodeSelector
	deployment.Spec.Template.Spec.HostNetwork = *commonConfiguration.HostNetwork
	deployment.Spec.Template.Spec.ImagePullSecrets = commonConfiguration.ImagePullSecrets
	return deployment
}
