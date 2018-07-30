/*
Copyright 2018 The Kubernetes Authors.
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at
    http://www.apache.org/licenses/LICENSE-2.0
    Unless required by applicable law or agreed to in writing, software
    distributed under the License is distributed on an "AS IS" BASIS,
    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
    See the License for the specific language governing permissions and
    limitations under the License.
*/

package main

import (
	"context"
	"flag"
	"log"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/runtime/signals"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

func main() {
	flag.Parse()
	logf.SetLogger(logf.ZapLogger(false))

	// Setup a Manager
	mrg, err := manager.New(config.GetConfigOrDie(), manager.Options{})
	if err != nil {
		log.Fatal(err)
	}

	// Setup a new controller to Reconciler ReplicaSets
	c, err := controller.New("foo-controller", mrg, controller.Options{
		Reconciler: &reconcileReplicaSet{client: mrg.GetClient()},
	})
	if err != nil {
		log.Fatal(err)
	}

	u := unstructured.Unstructured{}

	u.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "ReplicaSet",
	})

	// Watch ReplicaSets and enqueue ReplicaSet object key
	if err := c.Watch(&source.Kind{Type: &u}, &handler.EnqueueRequestForObject{}); err != nil {
		log.Fatal(err)
	}

	p := unstructured.Unstructured{}
	p.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	})

	// Watch Pods and enqueue owning ReplicaSet key
	if err := c.Watch(&source.Kind{Type: &p},
		&handler.EnqueueRequestForOwner{OwnerType: &u, IsController: true}); err != nil {
		log.Fatal(err)
	}

	log.Fatal(mrg.Start(signals.SetupSignalHandler()))
}

// reconcileReplicaSet reconciles ReplicaSets
type reconcileReplicaSet struct {
	client client.Client
}

// Implement reconcile.Reconciler so the controller can reconcile objects
var _ reconcile.Reconciler = &reconcileReplicaSet{}

func (r *reconcileReplicaSet) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the ReplicaSet from the cache
	rs := &unstructured.Unstructured{}

	rs.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   "apps",
		Version: "v1",
		Kind:    "ReplicaSet",
	})
	err := r.client.Get(context.TODO(), request.NamespacedName, rs)
	if errors.IsNotFound(err) {
		log.Printf("Could not find ReplicaSet %v.\n", request)
		return reconcile.Result{}, nil
	}

	if err != nil {
		log.Printf("Could not fetch ReplicaSet %v for %+v\n", err, request)
		return reconcile.Result{}, err
	}

	// Print the ReplicaSet
	log.Printf("ReplicaSet Name %s Namespace %s, Pod Name: %s\n",
		rs.GetName(), rs.GetNamespace(), rs.Object["spec"])

	// Set the label if it is missing
	if rs.GetLabels() == nil {
		rs.SetLabels(map[string]string{})
	}
	if rs.GetLabels()["hello"] == "world" {
		return reconcile.Result{}, nil
	}

	// Update the ReplicaSet
	rs.GetLabels()["hello"] = "world"
	err = r.client.Update(context.TODO(), rs)
	if err != nil {
		log.Printf("Could not write ReplicaSet %v\n", err)
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}
