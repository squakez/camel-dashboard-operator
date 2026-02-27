//go:build integration
// +build integration

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

package common

import (
	"context"
	"os/exec"
	"testing"
	"time"

	. "github.com/camel-tooling/camel-dashboard-operator/e2e/support"
	"github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	. "github.com/onsi/gomega"

	. "github.com/onsi/gomega/gstruct"
	corev1 "k8s.io/api/core/v1"
)

func TestVerifyCamelKIntegrationCountMessages(t *testing.T) {
	WithNewTestNamespace(t, func(ctx context.Context, g *WithT, ns string) {
		// Test a simple route which sends 5 messages to log
		t.Run("5 messages it", func(t *testing.T) {
			ExpectExecSucceed(t, g,
				exec.Command(
					"kubectl",
					"apply",
					"-f",
					"files/sample-it.yaml",
					"-n",
					ns,
				),
			)
			// The name of the selector, "camel.apache.org/app: camel-sample"
			g.Eventually(PodStatusPhase(t, ctx, ns, "camel.apache.org/app=camel-sample"), TestTimeoutMedium).Should(Equal(corev1.PodRunning))

			// The first time the number of messages is 5
			g.Eventually(
				CamelAppStatus(t, ctx, ns, "camel-sample"),
				TestTimeoutMedium,
			).Should(
				MatchFields(IgnoreExtras, Fields{
					"Phase": Equal(v1alpha1.CamelAppPhaseRunning),
					"Pods": And(
						HaveLen(1),
						ContainElement(
							MatchFields(IgnoreExtras, Fields{
								"Status": Equal("Running"),
								"Runtime": PointTo(MatchFields(IgnoreExtras, Fields{
									"Exchange": PointTo(MatchFields(IgnoreExtras, Fields{
										"Succeeded": Equal(5),
										"Total":     Equal(5),
									})),
								})),
							}),
						),
					),
				}),
			)
			// With this integration the number of exchanges has to stick to 5 consistently
			g.Consistently(
				CamelAppStatus(t, ctx, ns, "camel-sample"),
			).Should(
				MatchFields(IgnoreExtras, Fields{
					"Phase": Equal(v1alpha1.CamelAppPhaseRunning),
					"Pods": And(
						HaveLen(1),
						ContainElement(
							MatchFields(IgnoreExtras, Fields{
								"Status": Equal("Running"),
								"Runtime": PointTo(MatchFields(IgnoreExtras, Fields{
									"Exchange": PointTo(MatchFields(IgnoreExtras, Fields{
										"Succeeded": Equal(5),
										"Total":     Equal(5),
									})),
								})),
							}),
						),
					),
				}),
			)
		})
	})
}

func TestVerifyCamelKIntegrationTimerToLog(t *testing.T) {
	WithNewTestNamespace(t, func(ctx context.Context, g *WithT, ns string) {
		// Test a simple route with indefinite message delivery (all success)
		t.Run("all success messages", func(t *testing.T) {
			ExpectExecSucceed(t, g,
				exec.Command(
					"kubectl",
					"apply",
					"-f",
					"files/timer-to-log.yaml",
					"-n",
					ns,
				),
			)
			// The name of the selector, "camel.apache.org/app: timer-to-log"
			g.Eventually(PodStatusPhase(t, ctx, ns, "camel.apache.org/app=timer-to-log"), TestTimeoutMedium).Should(Equal(corev1.PodRunning))

			// We check the success rate is not reporting weird results
			g.Eventually(
				CamelAppStatus(t, ctx, ns, "timer-to-log"),
				TestTimeoutMedium,
			).Should(
				WithTransform(
					func(s v1alpha1.CamelAppStatus) bool {
						sr := s.SuccessRate
						if sr == nil {
							return false
						}
						// We need tolerance as the check may be slightly higher than 1 minute,
						//ie, time to reconcile.
						tolerance := *sr.SamplingIntervalDuration / 10
						elapsed := time.Since(sr.LastTimestamp.Time)
						if elapsed >= *sr.SamplingIntervalDuration+tolerance {
							return false
						}
						if sr.SamplingIntervalFailed != 0 {
							return false
						}
						if sr.SamplingIntervalTotal <= 0 {
							return false
						}
						if sr.Status != "Success" {
							return false
						}
						if sr.SuccessPercentage != "100.00" {
							return false
						}

						return true
					},
					BeTrue(),
				),
			)
		})
	})
}
