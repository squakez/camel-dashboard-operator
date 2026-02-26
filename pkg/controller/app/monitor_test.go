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

package app

import (
	"testing"
	"time"

	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestGetSLIExchangeSuccessRate(t *testing.T) {
	app := v1alpha1.RuntimeInfo{
		Exchange: &v1alpha1.ExchangeInfo{
			Total:  100,
			Failed: 5,
		},
	}

	target := v1alpha1.RuntimeInfo{
		Exchange: &v1alpha1.ExchangeInfo{
			Total:  200,
			Failed: 15,
		},
	}

	interval := 30 * time.Second

	result := getSLIExchangeSuccessRate(app, target, &interval, 5, 10)

	assert.NotNil(t, result)
	assert.Equal(t, 100, result.SamplingIntervalTotal)
	assert.Equal(t, 10, result.SamplingIntervalFailed)
	assert.Equal(t, "90.00", result.SuccessPercentage)
	assert.Equal(t, v1alpha1.SLIExchangeStatusWarning, result.Status)
}

func TestGetInfo(t *testing.T) {
	now := metav1.NewTime(time.Now())
	later := metav1.NewTime(time.Now().Add(1 * time.Minute))

	tests := []struct {
		name     string
		pods     []v1alpha1.PodInfo
		expected *v1alpha1.RuntimeInfo
	}{
		{
			name: "no runtime info at all -> nil",
			pods: []v1alpha1.PodInfo{
				{Ready: true},
				{Ready: true},
			},
			expected: nil,
		},
		{
			name: "aggregate exchange info",
			pods: []v1alpha1.PodInfo{
				{
					Runtime: &v1alpha1.RuntimeInfo{
						RuntimeProvider: "quarkus",
						RuntimeVersion:  "3.2.0",
						CamelVersion:    "4.0.0",
						Exchange: &v1alpha1.ExchangeInfo{
							Total:         100,
							Failed:        5,
							Pending:       10,
							Succeeded:     85,
							LastTimestamp: &now,
						},
					},
				},
				{
					Runtime: &v1alpha1.RuntimeInfo{
						Exchange: &v1alpha1.ExchangeInfo{
							Total:         50,
							Failed:        2,
							Pending:       5,
							Succeeded:     43,
							LastTimestamp: &later,
						},
					},
				},
			},
			expected: &v1alpha1.RuntimeInfo{
				RuntimeProvider: "quarkus",
				RuntimeVersion:  "3.2.0",
				CamelVersion:    "4.0.0",
				Exchange: &v1alpha1.ExchangeInfo{
					Total:         150,
					Failed:        7,
					Pending:       15,
					Succeeded:     128,
					LastTimestamp: &later, // latest timestamp wins
				},
			},
		},
		{
			name: "runtime present but no exchange data",
			pods: []v1alpha1.PodInfo{
				{
					Runtime: &v1alpha1.RuntimeInfo{
						RuntimeProvider: "springboot",
						RuntimeVersion:  "3.0",
						CamelVersion:    "4.1",
					},
				},
			},
			expected: &v1alpha1.RuntimeInfo{
				RuntimeProvider: "springboot",
				RuntimeVersion:  "3.0",
				CamelVersion:    "4.1",
				Exchange:        &v1alpha1.ExchangeInfo{},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := getInfo(tt.pods)

			if tt.expected == nil {
				assert.Nil(t, result)
				return
			}

			assert.NotNil(t, result)
			assert.Equal(t, tt.expected.RuntimeProvider, result.RuntimeProvider)
			assert.Equal(t, tt.expected.RuntimeVersion, result.RuntimeVersion)
			assert.Equal(t, tt.expected.CamelVersion, result.CamelVersion)

			assert.Equal(t, tt.expected.Exchange.Total, result.Exchange.Total)
			assert.Equal(t, tt.expected.Exchange.Failed, result.Exchange.Failed)
			assert.Equal(t, tt.expected.Exchange.Pending, result.Exchange.Pending)
			assert.Equal(t, tt.expected.Exchange.Succeeded, result.Exchange.Succeeded)

			if tt.expected.Exchange.LastTimestamp != nil {
				assert.True(t,
					tt.expected.Exchange.LastTimestamp.Equal(result.Exchange.LastTimestamp),
				)
			}
		})
	}
}
