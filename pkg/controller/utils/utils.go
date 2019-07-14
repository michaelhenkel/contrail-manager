package utils

import (
	"context"
	"strings"

	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"
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
	CASSANDRA = "Cassandra.contrail.juniper.net"
	ZOOKEEPER = "Zookeeper.contrail.juniper.net"
)

// GetGroupKindFromObject return GK
func GetGroupKindFromObject(object runtime.Object) schema.GroupKind {
	objectKind := object.GetObjectKind()
	objectGroupVersionKind := objectKind.GroupVersionKind()
	return objectGroupVersionKind.GroupKind()
}

// AppHandler handles
func AppHandler(appGroupKind schema.GroupKind) handler.Funcs {
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

	}
	return pred
}

//GetObjectAndGroupKindFromRequest returns Object and Kind
func GetObjectAndGroupKindFromRequest(request *reconcile.Request, client client.Client) (runtime.Object, *schema.GroupKind, error) {
	appGroupKind := schema.ParseGroupKind(strings.Split(request.Name, "/")[0])
	appName := strings.Split(request.Name, "/")[1]
	var instance runtime.Object
	switch appGroupKind {
	case AppAGroupKind():
		instance = &appv1alpha1.AppA{}
	case AppBGroupKind():
		instance = &appv1alpha1.AppB{}
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

