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
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/platform"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNonManagedCamelDeploymentStatic(t *testing.T) {
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        "my-app",
			Namespace:   "default",
			UID:         "1234",
			Annotations: map[string]string{"foo": "bar"},
			Labels:      map[string]string{v1alpha1.AppLabel: "my-app"},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(3),
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{Name: "main", Image: "my-image:v1"},
					},
				},
			},
		},
		Status: appsv1.DeploymentStatus{
			Replicas:          3,
			AvailableReplicas: 2,
		},
	}

	app := nonManagedCamelDeployment{deploy: deploy}

	// Test GetAppImage
	image := app.GetAppImage()
	assert.Equal(t, "my-image:v1", image)

	// Test GetReplicas
	replicas := app.GetReplicas()
	assert.NotNil(t, replicas)
	assert.Equal(t, int32(3), *replicas)

	// Test GetAnnotations
	ann := app.GetAnnotations()
	assert.Equal(t, "bar", ann["foo"])

	// Test GetAppPhase when not all replicas available
	phase := app.GetAppPhase(context.TODO(), nil)
	assert.Equal(t, v1alpha1.CamelAppPhaseError, phase)

	// Test GetAppPhase when all replicas available
	deploy.Status.AvailableReplicas = 3
	phase = app.GetAppPhase(context.TODO(), nil)
	assert.Equal(t, v1alpha1.CamelAppPhaseRunning, phase)

	// Test GetAppPhase when replicas = 0
	deploy.Status.Replicas = 0
	deploy.Status.AvailableReplicas = 0
	phase = app.GetAppPhase(context.TODO(), nil)
	assert.Equal(t, v1alpha1.CamelAppPhasePaused, phase)

	// Test CamelApp static fields
	deploy.Status.Replicas = 2
	deploy.Status.AvailableReplicas = 2
	camelApp := app.CamelApp(t.Context(), nil)
	assert.Equal(t, "my-app", camelApp.Name)
	assert.Equal(t, "default", camelApp.Namespace)
	assert.Equal(t, "my-app", camelApp.Annotations[v1alpha1.AppImportedNameLabel])
	assert.Equal(t, "Deployment", camelApp.Annotations[v1alpha1.AppImportedKindLabel])
	assert.Equal(t, "true", camelApp.Annotations[v1alpha1.AppSyntheticLabel])
	assert.Len(t, camelApp.OwnerReferences, 1)
	assert.Equal(t, "1234", string(camelApp.OwnerReferences[0].UID))
}

func int32Ptr(i int32) *int32 {
	return &i
}

func TestSetHealthHttpError(t *testing.T) {
	podInfo := &v1alpha1.PodInfo{}
	err := setHealth(podInfo, "127.0.0.1", 0)
	require.Error(t, err)
}

func TestSetHealthStatusOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"Healthy"}`))
	}))
	defer server.Close()

	podInfo := &v1alpha1.PodInfo{
		ObservabilityService: &v1alpha1.ObservabilityServiceInfo{},
	}

	host, portStr, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	require.NoError(t, err)

	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	err = setHealth(podInfo, host, port)
	require.NoError(t, err)

	require.NotNil(t, podInfo.Runtime)
	require.Equal(t, "Healthy", podInfo.Runtime.Status)
}

func TestSetHealthStatus503(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"status":"Degraded"}`))
	}))
	defer server.Close()

	podInfo := &v1alpha1.PodInfo{
		ObservabilityService: &v1alpha1.ObservabilityServiceInfo{},
	}

	host, portStr, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	require.NoError(t, err)

	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	err = setHealth(podInfo, host, port)
	require.NoError(t, err)

	require.Equal(t, "Degraded", podInfo.Runtime.Status)
}

func TestSetHealthStatusUnknown(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"status":"Not found"}`))
	}))
	defer server.Close()

	podInfo := &v1alpha1.PodInfo{
		ObservabilityService: &v1alpha1.ObservabilityServiceInfo{},
	}

	host, portStr, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	require.NoError(t, err)

	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	err = setHealth(podInfo, host, port)
	require.NoError(t, err)

	require.Equal(t, "Unknown", podInfo.Runtime.Status)
}

func TestSetMetricsStatusOK(t *testing.T) {
	metricsPayload := `
# HELP app_info Application info
# TYPE app_info gauge
app_info{runtime="quarkus",version="1.0.0"} 1

# TYPE camel_exchanges_total counter
camel_exchanges_total 5

# TYPE camel_exchanges_failed_total counter
camel_exchanges_failed_total 1

# TYPE camel_exchanges_succeeded_total counter
camel_exchanges_succeeded_total 4

# TYPE camel_exchanges_inflight gauge
camel_exchanges_inflight 2

# TYPE camel_exchanges_last_timestamp gauge
camel_exchanges_last_timestamp 123456
`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Contains(t, r.Header.Get("Accept"), "text/plain")

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(metricsPayload))
	}))
	defer server.Close()

	podInfo := &v1alpha1.PodInfo{
		ObservabilityService: &v1alpha1.ObservabilityServiceInfo{},
	}

	host, portStr, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	require.NoError(t, err)

	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	err = setMetrics(*server.Client(), podInfo, host, port)
	require.NoError(t, err)

	// Verify endpoint + port set
	require.Equal(t, platform.DefaultObservabilityMetrics, podInfo.ObservabilityService.MetricsEndpoint)
	require.Equal(t, port, podInfo.ObservabilityService.MetricsPort)

	// Verify runtime + exchange initialized
	require.NotNil(t, podInfo.Runtime)
	require.NotNil(t, podInfo.Runtime.Exchange)

	require.Equal(t, 5, podInfo.Runtime.Exchange.Total)
}

func TestSetMetricsStatusNotOK(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	podInfo := &v1alpha1.PodInfo{
		ObservabilityService: &v1alpha1.ObservabilityServiceInfo{},
	}

	host, portStr, err := net.SplitHostPort(strings.TrimPrefix(server.URL, "http://"))
	require.NoError(t, err)

	port, err := strconv.Atoi(portStr)
	require.NoError(t, err)

	err = setMetrics(*server.Client(), podInfo, host, port)
	require.Error(t, err)
	require.Contains(t, err.Error(), "HTTP status not OK")
}
