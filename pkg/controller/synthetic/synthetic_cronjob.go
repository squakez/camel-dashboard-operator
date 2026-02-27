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
	"time"

	v1alpha1 "github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/client"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
)

// nonManagedCamelCronjob represents a cron Camel application built and deployed outside the operator lifecycle.
type nonManagedCamelCronjob struct {
	cron       *batchv1.CronJob
	httpClient *http.Client
}

// CamelApp return an CamelApp resource fed by the Camel application adapter.
func (app *nonManagedCamelCronjob) CamelApp(ctx context.Context, c client.Client) *v1alpha1.CamelApp {
	newApp := v1alpha1.NewApp(app.cron.Namespace, app.cron.Labels[v1alpha1.AppLabel])
	newApp.SetAnnotations(map[string]string{
		v1alpha1.AppImportedNameLabel: app.cron.Name,
		v1alpha1.AppImportedKindLabel: "CronJob",
		v1alpha1.AppSyntheticLabel:    "true",
	})
	references := []metav1.OwnerReference{
		{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
			Name:       app.cron.Name,
			UID:        app.cron.UID,
			Controller: &controller,
		},
	}
	newApp.SetOwnerReferences(references)
	return &newApp
}

// GetAppPhase returns the phase of the backing Camel application.
func (app *nonManagedCamelCronjob) GetAppPhase(ctx context.Context, c client.Client) v1alpha1.CamelAppPhase {
	if len(app.cron.Status.Active) > 0 {
		return v1alpha1.CamelAppPhaseRunning
	}

	// If none is active, then it means the app is waiting for scheduling execution.
	return v1alpha1.CamelAppPhasePaused
}

// GetReplicas returns the number of desired replicas for the backing Camel application.
func (app *nonManagedCamelCronjob) GetReplicas() *int32 {
	// In the case of a CronJob we use the number of active jobs instead.
	return ptr.To(int32(len(app.cron.Status.Active)))
}

// GetAppImage returns the container image of the backing Camel application.
func (app *nonManagedCamelCronjob) GetAppImage() string {
	return app.cron.Spec.JobTemplate.Spec.Template.Spec.Containers[0].Image
}

// GetPods returns the pods backing the Camel application.
func (app *nonManagedCamelCronjob) GetPods(ctx context.Context, c client.Client) ([]v1alpha1.PodInfo, error) {
	// In the CronJob case we don't want to inspect the Pod as we are not sure we have the Pod live when
	// the monitoring happens.

	return getPods(*app.httpClient, ctx, c, app.cron.GetNamespace(),
		app.cron.Spec.JobTemplate.Spec.Template.Labels, getObservabilityPort(app.GetAnnotations()), false)
}

// GetAnnotations returns the backing deployment object annotations.
func (app *nonManagedCamelCronjob) GetAnnotations() map[string]string {
	return app.cron.Annotations
}

// SetMonitoringCondition sets the health and monitoring conditions on the target app.
func (app *nonManagedCamelCronjob) SetMonitoringCondition(srcApp, targetApp *v1alpha1.CamelApp, pods []v1alpha1.PodInfo) {
	info := ""
	runningPods := countPodsWithStatus(pods, "Running")
	succeededPods := countPodsWithStatus(pods, "Succeeded")
	// We only verify the status of latest executions. If they are all successful, then we consider the workload healthy.
	if len(pods) > 0 {
		info = fmt.Sprintf("%d out of last %d job succeeded", succeededPods, len(pods)-runningPods)
		targetApp.Status.AddCondition(metav1.Condition{
			Type:               "Monitored",
			Status:             metav1.ConditionTrue,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "MonitoringComplete",
			Message:            "At least one scheduled job has run",
		})
		healthCond := metav1.ConditionFalse
		if len(pods) == runningPods+succeededPods {
			healthCond = metav1.ConditionTrue
		}
		targetApp.Status.AddCondition(metav1.Condition{
			Type:               "Healthy",
			Status:             healthCond,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "HealthCheckCompleted",
			Message:            info,
		})
		if app.cron.Status.LastScheduleTime != nil {
			info += "; Last scheduled time: " + app.cron.Status.LastScheduleTime.Format("2006-01-02 15:04:05")
		}
		if app.cron.Status.LastSuccessfulTime != nil {
			info += "; Last successful time: " + app.cron.Status.LastSuccessfulTime.Format("2006-01-02 15:04:05")
		}
		targetApp.Status.Info = info
	} else {
		targetApp.Status.AddCondition(metav1.Condition{
			Type:               "Monitored",
			Status:             metav1.ConditionFalse,
			LastTransitionTime: metav1.NewTime(time.Now()),
			Reason:             "MonitoringComplete",
			Message:            "No scheduled job has run yet",
		})
	}
}
