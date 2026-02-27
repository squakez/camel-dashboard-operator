/*
Licensed to the Apache Software Foundation (ASF) under one or more
contributor license agreements.  See the NOTICE file distributed with
this work for additional information regarding copyright ownership.
The ASF licenses this file to You under the Apache License, Version 2.0
(the "License"); you may not use this file except in compliance with
the License.  You may obtain a copy of the License at

   http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package synthetic

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"time"

	v1alpha1 "github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/client"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/platform"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/util/kubernetes"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/util/log"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	clientgocache "k8s.io/client-go/tools/cache"
	"knative.dev/serving/pkg/apis/serving"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	ctrl "sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	controller = true
)

// ManageSyntheticCamelApps is the controller for synthetic Camel Applications. Consider that the lifecycle of the objects are driven
// by the way we are monitoring them. Since we're filtering by some label in the cached client, you must consider an add, update or delete
// accordingly, ie, when the user label the resource, then it is considered as an add, when it removes the label, it is considered as a delete.
func ManageSyntheticCamelApps(ctx context.Context, c client.Client, cache cache.Cache) error {
	informers, err := getInformers(ctx, c, cache)
	if err != nil {
		return err
	}
	for _, informer := range informers {
		_, err := informer.AddEventHandler(clientgocache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				ctrlObj, ok := obj.(ctrl.Object)
				if !ok {
					log.Error(fmt.Errorf("type assertion failed: %v", obj), "Failed to retrieve Object on add event")
					return
				}

				onAdd(ctx, c, ctrlObj)
			},
			DeleteFunc: func(obj interface{}) {
				ctrlObj, ok := obj.(ctrl.Object)
				if !ok {
					log.Errorf(fmt.Errorf("type assertion failed: %v", obj), "Failed to retrieve Object on delete event")
					return
				}

				onDelete(ctx, c, ctrlObj)
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

func onAdd(ctx context.Context, c client.Client, ctrlObj ctrl.Object) {
	log.Infof("Detected a new %s resource named %s in namespace %s",
		ctrlObj.GetObjectKind().GroupVersionKind().Kind, ctrlObj.GetName(), ctrlObj.GetNamespace())
	appName := ctrlObj.GetLabels()[v1alpha1.AppLabel]
	app, err := getSyntheticCamelApp(ctx, c, ctrlObj.GetNamespace(), appName)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			adapter, err := NonManagedCamelApplicationFactory(ctrlObj)
			if err != nil {
				log.Errorf(err, "Some error happened while creating a Camel application adapter for %s", appName)
			}
			if err = createSyntheticCamelApp(ctx, c, adapter.CamelApp(ctx, c)); err != nil {
				log.Errorf(err, "Some error happened while creating a synthetic Camel Application %s", appName)
			}
			log.Infof("Created a synthetic Camel Application %s after %s resource object", app.GetName(), ctrlObj.GetName())
		} else {
			log.Errorf(err, "Some error happened while loading a synthetic Camel Application %s", appName)
		}
	} else {
		log.Infof("Synthetic Camel Application %s (phase %s) already imported. Skipping.", appName, app.Status.Phase)
	}
}

func onDelete(ctx context.Context, c client.Client, ctrlObj ctrl.Object) {
	appName := ctrlObj.GetLabels()[v1alpha1.AppLabel]
	// Importing label removed
	if err := deleteSyntheticCamelApp(ctx, c, ctrlObj.GetNamespace(), appName); err != nil {
		log.Errorf(err, "Some error happened while deleting a synthetic Camel Application %s", appName)
	}
	log.Infof("Deleted synthetic Camel Application %s", appName)
}

func getInformers(ctx context.Context, cl client.Client, c cache.Cache) ([]cache.Informer, error) {
	deploy, err := c.GetInformer(ctx, &appsv1.Deployment{})
	if err != nil {
		return nil, err
	}
	informers := []cache.Informer{deploy}
	// Watch for the CronJob conditionally
	if ok, err := kubernetes.IsAPIResourceInstalled(cl, batchv1.SchemeGroupVersion.String(), reflect.TypeOf(batchv1.CronJob{}).Name()); ok && err == nil {
		cron, err := c.GetInformer(ctx, &batchv1.CronJob{})
		if err != nil {
			return nil, err
		}
		informers = append(informers, cron)
	}
	// Watch for the Knative Services conditionally
	if ok, err := kubernetes.IsAPIResourceInstalled(cl, servingv1.SchemeGroupVersion.String(), reflect.TypeOf(servingv1.Service{}).Name()); ok && err == nil {
		if ok, err := kubernetes.CheckPermission(ctx, cl, serving.GroupName, "services", platform.GetOperatorWatchNamespace(), "", "watch"); ok && err == nil {
			ksvc, err := c.GetInformer(ctx, &servingv1.Service{})
			if err != nil {
				return nil, err
			}
			informers = append(informers, ksvc)
		}
	}

	return informers, nil
}

func getSyntheticCamelApp(ctx context.Context, c client.Client, namespace, name string) (*v1alpha1.CamelApp, error) {
	app := v1alpha1.NewApp(namespace, name)
	err := c.Get(ctx, ctrl.ObjectKeyFromObject(&app), &app)
	return &app, err
}

func createSyntheticCamelApp(ctx context.Context, c client.Client, app *v1alpha1.CamelApp) error {
	return c.Create(ctx, app, ctrl.FieldOwner("camel-dashboard-operator"))
}

func deleteSyntheticCamelApp(ctx context.Context, c client.Client, namespace, name string) error {
	// As the Integration label was removed, we don't know which is the Synthetic Camel Application to remove
	app := v1alpha1.NewApp(namespace, name)
	return c.Delete(ctx, &app)
}

// NonManagedCamelApplicationAdapter represents a Camel application built and deployed outside the operator lifecycle.
type NonManagedCamelApplicationAdapter interface {
	// CamelApp returns a CamelApp resource fed by the Camel application adapter.
	CamelApp(ctx context.Context, c client.Client) *v1alpha1.CamelApp
	// GetAppPhase returns the phase of the backing Camel application.
	GetAppPhase(ctx context.Context, c client.Client) v1alpha1.CamelAppPhase
	// GetAppImage returns the container image of the backing Camel application.
	GetAppImage() string
	// GetReplicas returns the number of desired replicas for the backing Camel application.
	GetReplicas() *int32
	// GetPods returns the actual Pods backing the Camel application.
	GetPods(ctx context.Context, c client.Client) ([]v1alpha1.PodInfo, error)
	// GetAnnotations returns the backing deployment object annotations.
	GetAnnotations() map[string]string
	// SetMonitoringCondition sets the health and monitoring conditions on the target app.
	SetMonitoringCondition(app, targetApp *v1alpha1.CamelApp, pods []v1alpha1.PodInfo)
}

func NonManagedCamelApplicationFactory(obj ctrl.Object) (NonManagedCamelApplicationAdapter, error) {
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}
	deploy, ok := obj.(*appsv1.Deployment)
	if ok {
		return &nonManagedCamelDeployment{deploy: deploy, httpClient: httpClient}, nil
	}
	cronjob, ok := obj.(*batchv1.CronJob)
	if ok {
		return &nonManagedCamelCronjob{cron: cronjob, httpClient: httpClient}, nil
	}
	ksvc, ok := obj.(*servingv1.Service)
	if ok {
		return &nonManagedCamelKnativeService{ksvc: ksvc, httpClient: httpClient}, nil
	}
	return nil, fmt.Errorf("unsupported %s object kind", obj.GetObjectKind().GroupVersionKind().Kind)
}
