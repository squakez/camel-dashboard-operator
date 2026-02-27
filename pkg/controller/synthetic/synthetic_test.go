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
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	servingv1 "knative.dev/serving/pkg/apis/serving/v1"

	v1 "github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNonManagedUnsupported(t *testing.T) {
	pod := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "my-pod",
			Labels: map[string]string{
				v1.AppLabel: "my-imported-it",
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:  "my-cnt",
					Image: "my-img",
				},
			},
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
			Conditions: []corev1.PodCondition{
				{
					Type:   corev1.PodReady,
					Status: corev1.ConditionTrue,
				},
			},
		},
	}

	nilAdapter, err := NonManagedCamelApplicationFactory(pod)
	require.Error(t, err)
	assert.Equal(t, "unsupported Pod object kind", err.Error())
	assert.Nil(t, nilAdapter)
}

func TestNonManagedDeployment(t *testing.T) {
	deploy := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: appsv1.SchemeGroupVersion.String(),
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "my-deploy",
			Labels: map[string]string{
				v1.AppLabel: "my-imported-it",
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptr.To(int32(1)),
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						v1.AppLabel: "my-imported-it",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "my-cnt",
							Image: "my-img",
						},
					},
				},
			},
		},
	}

	expectedIt := v1.NewApp("ns", "my-imported-it")
	expectedIt.SetAnnotations(map[string]string{
		v1.AppImportedNameLabel: "my-deploy",
		v1.AppImportedKindLabel: "Deployment",
		v1.AppSyntheticLabel:    "true",
	})
	references := []metav1.OwnerReference{
		{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
			Name:       deploy.Name,
			UID:        deploy.UID,
			Controller: &controller,
		},
	}
	expectedIt.SetOwnerReferences(references)

	deploymentAdapter, err := NonManagedCamelApplicationFactory(deploy)
	require.NoError(t, err)
	assert.NotNil(t, deploymentAdapter)
	assert.Equal(t, expectedIt.ObjectMeta, deploymentAdapter.CamelApp(context.Background(), nil).ObjectMeta)
}

func TestNonManagedCronJob(t *testing.T) {
	cron := &batchv1.CronJob{
		TypeMeta: metav1.TypeMeta{
			APIVersion: batchv1.SchemeGroupVersion.String(),
			Kind:       "CronJob",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "my-cron",
			Labels: map[string]string{
				v1.AppLabel: "my-imported-it",
			},
		},
		Spec: batchv1.CronJobSpec{
			JobTemplate: batchv1.JobTemplateSpec{
				Spec: batchv1.JobSpec{
					Template: corev1.PodTemplateSpec{
						ObjectMeta: metav1.ObjectMeta{
							Labels: map[string]string{
								v1.AppLabel: "my-imported-it",
							},
						},
						Spec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "my-cnt",
									Image: "my-img",
								},
							},
						},
					},
				},
			},
		},
	}

	expectedIt := v1.NewApp("ns", "my-imported-it")
	expectedIt.SetAnnotations(map[string]string{
		v1.AppImportedNameLabel: "my-cron",
		v1.AppImportedKindLabel: "CronJob",
		v1.AppSyntheticLabel:    "true",
	})
	references := []metav1.OwnerReference{
		{
			APIVersion: "batch/v1",
			Kind:       "CronJob",
			Name:       cron.Name,
			UID:        cron.UID,
			Controller: &controller,
		},
	}
	expectedIt.SetOwnerReferences(references)
	cronJobAdapter, err := NonManagedCamelApplicationFactory(cron)
	require.NoError(t, err)
	assert.NotNil(t, cronJobAdapter)
	assert.Equal(t, expectedIt, *cronJobAdapter.CamelApp(context.Background(), nil))
}

func TestNonManagedKnativeService(t *testing.T) {
	ksvc := &servingv1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: servingv1.SchemeGroupVersion.String(),
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: "ns",
			Name:      "my-ksvc",
			Labels: map[string]string{
				v1.AppLabel: "my-imported-it",
			},
		},
		Spec: servingv1.ServiceSpec{
			ConfigurationSpec: servingv1.ConfigurationSpec{
				Template: servingv1.RevisionTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{
							v1.AppLabel: "my-imported-it",
						},
					},
					Spec: servingv1.RevisionSpec{
						PodSpec: corev1.PodSpec{
							Containers: []corev1.Container{
								{
									Name:  "my-cnt",
									Image: "my-img",
								},
							},
						},
					},
				},
			},
		},
	}

	expectedIt := v1.NewApp("ns", "my-imported-it")
	expectedIt.SetAnnotations(map[string]string{
		v1.AppImportedNameLabel: "my-ksvc",
		v1.AppImportedKindLabel: "KnativeService",
		v1.AppSyntheticLabel:    "true",
	})
	references := []metav1.OwnerReference{
		{
			APIVersion: servingv1.SchemeGroupVersion.String(),
			Kind:       "Service",
			Name:       ksvc.Name,
			UID:        ksvc.UID,
			Controller: &controller,
		},
	}
	expectedIt.SetOwnerReferences(references)

	knativeServiceAdapter, err := NonManagedCamelApplicationFactory(ksvc)
	require.NoError(t, err)
	assert.NotNil(t, knativeServiceAdapter)
	assert.Equal(t, expectedIt, *knativeServiceAdapter.CamelApp(context.Background(), nil))
}
