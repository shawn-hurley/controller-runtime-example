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

	sdk "github.com/operator-framework/operator-sdk/pkg/sdk"
	"github.com/shawn-hurley/controller-runtime-example/pkg/apis/app/v1alpha1"
	"github.com/shawn-hurley/controller-runtime-example/pkg/stub"
	"k8s.io/apimachinery/pkg/runtime"
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
	v1alpha1.AddToScheme(mrg.GetScheme())

	app := v1alpha1.App{}
	// Setup a new controller to Reconciler ReplicaSets
	c, err := controller.New("foo-controller", mrg, controller.Options{
		Reconciler: &sdkHandlerReconciler{
			client:  mrg.GetClient(),
			handler: &stub.Handler{},
			Object:  &app,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	// Watch apps and enqueue CRD object key
	if err := c.Watch(&source.Kind{Type: &app}, &handler.EnqueueRequestForObject{}); err != nil {
		log.Fatal(err)
	}
	log.Fatal(mrg.Start(signals.SetupSignalHandler()))
}

type sdkHandlerReconciler struct {
	handler sdk.Handler
	client  client.Client
	Object  runtime.Object
}

func (s *sdkHandlerReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	o := s.Object.DeepCopyObject()
	s.client.Get(context.TODO(), request.NamespacedName, o)
	e := sdk.Event{Object: o}
	err := s.handler.Handle(context.TODO(), e)
	if err != nil {
		return reconcile.Result{}, err
	}
	return reconcile.Result{}, nil
}
