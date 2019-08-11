// Copyright 2018 The Operator-SDK Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package e2e

import (
	goctx "context"
	"fmt"
	"testing"
	"time"

	apis "github.com/michaelhenkel/contrail-manager/pkg/apis"
	v1alpha1 "github.com/michaelhenkel/contrail-manager/pkg/apis/contrail/v1alpha1"

	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 60
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

/*
func TestRabbitmq(t *testing.T) {
	rabbitmqList := &v1alpha1.RabbitmqList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Rabbitmq",
			APIVersion: "contrail.juniper.net/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, rabbitmqList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("rabbitmq-group", func(t *testing.T) {
		t.Run("Cluster", RabbitmqCluster)
	})
}
*/

func TestManager(t *testing.T) {
	managerList := &v1alpha1.ManagerList{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Manager",
			APIVersion: "contrail.juniper.net/v1alpha1",
		},
	}
	err := framework.AddToFrameworkScheme(apis.AddToScheme, managerList)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}
	// run subtests
	t.Run("manager-group", func(t *testing.T) {
		t.Run("Cluster", ManagerCluster)
	})
}

func ManagerCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	// get global framework variables
	f := framework.Global
	var create = true
	var replicas int32 = 1
	var hostNetwork = false
	var zkReplicas int32 = 1
	manager := &v1alpha1.Manager{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Manager",
			APIVersion: "contrail.juniper.net/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "cluster1",
			Namespace: namespace,
		},
		Spec: v1alpha1.ManagerSpec{
			CommonConfiguration: v1alpha1.CommonConfiguration{
				Replicas:         &replicas,
				HostNetwork:      &hostNetwork,
				ImagePullSecrets: []string{"contrail-nightly"},
			},
			Services: v1alpha1.Services{
				Rabbitmq: &v1alpha1.Rabbitmq{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "rabbitmq1",
						Namespace: namespace,
						Labels:    map[string]string{"contrail_cluster": "cluster1"},
					},
					Spec: v1alpha1.RabbitmqSpec{
						CommonConfiguration: v1alpha1.CommonConfiguration{
							Create: &create,
						},
						ServiceConfiguration: v1alpha1.RabbitmqConfiguration{
							Images: map[string]string{"rabbitmq": "rabbitmq:3.7",
								"init": "busybox"},
						},
					},
				},
				Zookeepers: []*v1alpha1.Zookeeper{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "zookeeper1",
						Namespace: namespace,
						Labels:    map[string]string{"contrail_cluster": "cluster1"},
					},
					Spec: v1alpha1.ZookeeperSpec{
						CommonConfiguration: v1alpha1.CommonConfiguration{
							Create:   &create,
							Replicas: &zkReplicas,
						},
						ServiceConfiguration: v1alpha1.ZookeeperConfiguration{
							Images: map[string]string{"zookeeper": "docker.io/zookeeper:3.5.5",
								"init": "busybox"},
						},
					},
				}},
				Cassandras: []*v1alpha1.Cassandra{{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "cassandra1",
						Namespace: namespace,
						Labels:    map[string]string{"contrail_cluster": "cluster1"},
					},
					Spec: v1alpha1.CassandraSpec{
						CommonConfiguration: v1alpha1.CommonConfiguration{
							Create: &create,
						},
						ServiceConfiguration: v1alpha1.CassandraConfiguration{
							Images: map[string]string{"cassandra": "cassandra:3.11.4",
								"init": "busybox"},
						},
					},
				}},
				Config: &v1alpha1.Config{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "config1",
						Namespace: namespace,
						Labels:    map[string]string{"contrail_cluster": "cluster1"},
					},
					Spec: v1alpha1.ConfigSpec{
						CommonConfiguration: v1alpha1.CommonConfiguration{
							Create: &create,
						},
						ServiceConfiguration: v1alpha1.ConfigConfiguration{
							CassandraInstance: "cassandra1",
							ZookeeperInstance: "zookeeper1",
							Images: map[string]string{"api": "hub.juniper.net/contrail-nightly/contrail-controller-config-api:5.2.0-0.740",
								"devicemanager":        "hub.juniper.net/contrail-nightly/contrail-controller-config-devicemgr:5.2.0-0.740",
								"schematransformer":    "hub.juniper.net/contrail-nightly/contrail-controller-config-schema:5.2.0-0.740",
								"servicemonitor":       "hub.juniper.net/contrail-nightly/contrail-controller-config-svcmonitor:5.2.0-0.740",
								"analyticsapi":         "hub.juniper.net/contrail-nightly/contrail-analytics-api:5.2.0-0.740",
								"collector":            "hub.juniper.net/contrail-nightly/contrail-analytics-collector:5.2.0-0.740",
								"redis":                "redis:4.0.2",
								"nodemanagerconfig":    "hub.juniper.net/contrail-nightly/contrail-nodemgr:5.2.0-0.740",
								"nodemanageranalytics": "hub.juniper.net/contrail-nightly/contrail-nodemgr:5.2.0-0.740",
								"nodeinit":             "hub.juniper.net/contrail-nightly/contrail-node-init:5.2.0-0.740",
								"init":                 "busybox"},
						},
					},
				},
			},
		},
	}

	err = f.Client.Create(goctx.TODO(), manager, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "rabbitmq1-rabbitmq-deployment", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "zookeeper1-zookeeper-deployment", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "cassandra1-cassandra-deployment", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	timeout = time.Second * 400
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "config1-config-deployment", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
	/*
		if err = managerScaleTest(t, f, ctx); err != nil {
			t.Fatal(err)
		}
	*/
}

func RabbitmqCluster(t *testing.T) {
	t.Parallel()
	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err := ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}

	f := framework.Global
	var replicas int32 = 1
	rabbitmq := &v1alpha1.Rabbitmq{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "rabbitmq1",
			Namespace: namespace,
			Labels:    map[string]string{"contrail_cluster": "cluster1"},
		},
		Spec: v1alpha1.RabbitmqSpec{
			CommonConfiguration: v1alpha1.CommonConfiguration{
				Replicas: &replicas,
			},
			ServiceConfiguration: v1alpha1.RabbitmqConfiguration{
				Images: map[string]string{"rabbitmq": "rabbitmq:3.7",
					"init": "busybox"},
			},
		},
	}

	err = f.Client.Create(goctx.TODO(), rabbitmq, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "rabbitmq1-rabbitmq-deployment", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}
}

func managerScaleTest(t *testing.T, f *framework.Framework, ctx *framework.TestCtx) error {
	namespace, err := ctx.GetNamespace()
	if err != nil {
		return fmt.Errorf("could not get namespace: %v", err)
	}
	manager := &v1alpha1.Manager{}
	f.Client.Get(goctx.TODO(), types.NamespacedName{Name: "cluster1", Namespace: namespace}, manager)
	if err != nil {
		return fmt.Errorf("could not get manager: %v", err)
	}
	var replicas int32 = 3
	manager.Spec.CommonConfiguration.Replicas = &replicas
	err = f.Client.Update(goctx.TODO(), manager)
	if err != nil {
		return fmt.Errorf("could not update manager: %v", err)
	}
	timeout = time.Second * 120
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "rabbitmq1-rabbitmq-deployment", 3, retryInterval, timeout)
	if err != nil {
		return fmt.Errorf("rabbitmq deployment is wrong: %v", err)
	}
	err = e2eutil.WaitForDeployment(t, f.KubeClient, namespace, "zookeeper1-zookeeper-deployment", 3, retryInterval, timeout)
	if err != nil {
		return fmt.Errorf("zookeeper deployment is wrong: %v", err)
	}
	return nil
}
