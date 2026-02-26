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
	"net/http"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/client"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// nonManagedCamelDeployment represents a regular Camel application built and deployed outside the operator lifecycle.
type nonManagedCamelDeployment struct {
	deploy     *appsv1.Deployment
	httpClient *http.Client
}

// CamelApp return an CamelApp resource fed by the Camel application adapter.
func (app *nonManagedCamelDeployment) CamelApp(ctx context.Context, c client.Client) *v1alpha1.CamelApp {
	newApp := v1alpha1.NewApp(app.deploy.Namespace, app.deploy.Labels[v1alpha1.AppLabel])
	newApp.SetAnnotations(map[string]string{
		v1alpha1.AppImportedNameLabel: app.deploy.Name,
		v1alpha1.AppImportedKindLabel: "Deployment",
		v1alpha1.AppSyntheticLabel:    "true",
	})
	newApp.ImportCamelAnnotations(app.deploy.Annotations)
	references := []metav1.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       app.deploy.Name,
			UID:        app.deploy.UID,
			Controller: &controller,
		},
	}
	newApp.SetOwnerReferences(references)

	return &newApp
}

// GetAppPhase returns the phase of the backing Camel application.
func (app *nonManagedCamelDeployment) GetAppPhase(ctx context.Context, c client.Client) v1alpha1.CamelAppPhase {
	if app.deploy.Status.AvailableReplicas == app.deploy.Status.Replicas {
		if app.deploy.Status.Replicas == 0 {
			return v1alpha1.CamelAppPhasePaused
		}
		return v1alpha1.CamelAppPhaseRunning
	}

	return v1alpha1.CamelAppPhaseError
}

// GetAppImage returns the container image of the backing Camel application.
func (app *nonManagedCamelDeployment) GetAppImage() string {
	return app.deploy.Spec.Template.Spec.Containers[0].Image
}

// GetReplicas returns the number of desired replicas for the backing Camel application.
func (app *nonManagedCamelDeployment) GetReplicas() *int32 {
	return app.deploy.Spec.Replicas
}

// GetAnnotations returns the backing deployment object annotations.
func (app *nonManagedCamelDeployment) GetAnnotations() map[string]string {
	return app.deploy.Annotations
}

// GetPods returns the pods backing the Camel application.
func (app *nonManagedCamelDeployment) GetPods(ctx context.Context, c client.Client) ([]v1alpha1.PodInfo, error) {
	return getPods(*app.httpClient, ctx, c, app.deploy.GetNamespace(),
		app.deploy.Spec.Selector.MatchLabels, getObservabilityPort(app.GetAnnotations()), true)
}

// SetMonitoringCondition sets the health and monitoring conditions on the target app.
func (app *nonManagedCamelDeployment) SetMonitoringCondition(srcApp, targetApp *v1alpha1.CamelApp, pods []v1alpha1.PodInfo) {
	setMonitoringCondition(srcApp, targetApp, pods)
}
