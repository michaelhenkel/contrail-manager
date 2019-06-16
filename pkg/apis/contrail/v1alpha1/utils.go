package v1alpha1

import (
	"context"
	"errors"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
)

var log = logf.Log.WithName("controller")

func SetServiceStatus(cl client.Client,
	instance metav1.ObjectMeta,
	name string,
	ro runtime.Object,
	deploymentStatus *appsv1.DeploymentStatus,
	status *Status) error {
	err := cl.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, ro)
	if err != nil {
		return err
	}
	if deploymentStatus.Replicas == deploymentStatus.ReadyReplicas {
		active := true
		status.Active = &active
		err = cl.Status().Update(context.TODO(), ro)
		if err != nil {
			return err
		}
	} else {
		return errors.New("Deployment not ready, yet")
	}
	return nil
}

func MarkInitPodsReady(cl client.Client,
	instance metav1.ObjectMeta,
	name string) error {
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"app": name + "-" + instance.Name})
	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	err := cl.List(context.TODO(), listOps, podList)
	if err != nil {
		return err
	}
	for _, pod := range podList.Items {
		pod.ObjectMeta.Labels["status"] = "ready"
		err = cl.Update(context.TODO(), &pod)
		if err != nil {
			return err
		}
	}
	return nil
}

func InitContainerRunning(cl client.Client,
	instance metav1.ObjectMeta,
	name string,
	ro runtime.Object,
	service Service,
	status *Status) (bool, error) {
	initContainerRunning := true
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"app": name + "-" + instance.Name})
	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	err := cl.List(context.TODO(), listOps, podList)
	if err != nil {
		return false, err
	}
	if len(podList.Items) == int(*service.Size) {
		var podIpMap = make(map[string]string)
		for _, pod := range podList.Items {
			if pod.Status.PodIP != "" {
				podIpMap[pod.Name] = pod.Status.PodIP
			}
			for _, initContainer := range pod.Status.InitContainerStatuses {
				if initContainer.Name == "init" {
					if !initContainer.Ready {
						initContainerRunning = false
					}
				}
			}
		}
		if len(podIpMap) == int(*service.Size) {

			if status.Active == nil {
				active := false
				status.Active = &active
			}
			status.Nodes = podIpMap
			err = cl.Status().Update(context.TODO(), ro)
			if err != nil {
				return false, err
			}
		} else {
			return false, errors.New("Not enough Pod IPs")
		}
	} else {
		return false, errors.New("Not enough Pods")
	}
	return initContainerRunning, nil
}

func getPodMap(name string, instance metav1.ObjectMeta, cl client.Client) (map[string]string, error) {
	reqLogger := log.WithValues("Request.Namespace", instance.Namespace, "Request.Name", instance.Name)
	reqLogger.Info("Reconciling" + name)
	initContainerRunning := true
	var podIpMap = make(map[string]string)
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(map[string]string{"app": name + "-" + instance.Name})
	listOps := &client.ListOptions{Namespace: instance.Namespace, LabelSelector: labelSelector}
	err := cl.List(context.TODO(), listOps, podList)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods", instance.Namespace, instance.Namespace, instance.Name, instance.Name)
		return podIpMap, err
	}
	for _, pod := range podList.Items {
		if pod.Status.PodIP != "" {
			podIpMap[pod.Name] = pod.Status.PodIP
		}
		for _, initContainer := range pod.Status.InitContainerStatuses {
			if initContainer.Name == "init" {
				if !initContainer.Ready {
					initContainerRunning = false
				}
			}
		}
	}
	if len(podIpMap) > 0 && initContainerRunning {
		return podIpMap, nil
	}
	return podIpMap, errors.New("Init Containers not running")
}

func getPodIpList(podIpMap map[string]string, service Service, podList corev1.PodList, cl client.Client) ([]string, error) {
	var podIpList []string
	if len(podIpMap) == int(*service.Size) {
		for _, podIp := range podIpMap {
			podIpList = append(podIpList, podIp)
		}
		for _, pod := range podList.Items {
			pod.ObjectMeta.Labels["status"] = "ready"
			err := cl.Update(context.TODO(), &pod)
			if err != nil {
				return podIpList, err
			}
		}
	}
	return podIpList, nil
}

func FinishDeployment(initContainerRunning bool,
	cl client.Client,
	objectMeta metav1.ObjectMeta,
	instance runtime.Object,
	podIpMap map[string]string,
	status *Status,
	deployment appsv1.Deployment) error {
	reqLogger := log.WithValues("Request.Namespace", objectMeta.Namespace, "Request.Name", objectMeta.Name)
	reqLogger.Info("Reconciling")
	podMap, err := getPodMap(objectMeta.Name, objectMeta, cl)
	if err != nil {
		return err
	}
	if len(podMap) > 0 {
		status.Nodes = podIpMap
		err := cl.Status().Update(context.TODO(), instance)
		if err != nil {
			reqLogger.Error(err, "Failed to update status.")
			return err
		}
		err = cl.Get(context.TODO(), types.NamespacedName{Name: objectMeta.Name + "-" + objectMeta.Name, Namespace: objectMeta.Namespace}, &deployment)
		if err != nil {
			reqLogger.Error(err, "Failed to get Deployment")
			return err
		}
		if deployment.Status.Replicas == deployment.Status.ReadyReplicas {
			active := true
			status.Active = &active
			err = cl.Status().Update(context.TODO(), instance)
			if err != nil {
				reqLogger.Error(err, "Failed to update status.")
				return err
			}
		} else {
			reqLogger.Info("Waiting for Deployment to be complete, requeing")
			return err
		}
		if err != nil {
			reqLogger.Error(err, "Failed to update status.")
			return err
		}
	} else {
		reqLogger.Info("Init Pods not ready, requeing")
		return nil
	}
	return errors.New("Init Container not running")
}

func labelsForService(name string) map[string]string {
	return map[string]string{"app": name}
}

func getPodNames(pods []corev1.Pod) []string {
	var podNames []string
	for _, pod := range pods {
		podNames = append(podNames, pod.Name)
	}
	return podNames
}
