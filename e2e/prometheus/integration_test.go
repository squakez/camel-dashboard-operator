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
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/camel-tooling/camel-monitor-operator/e2e/support"
	"github.com/camel-tooling/camel-monitor-operator/pkg/apis/camel/v1alpha1"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"

	. "github.com/onsi/gomega/gstruct"
)

func TestVerifyPrometheusScrapeMetrics(t *testing.T) {
	WithNewTestNamespace(t, func(ctx context.Context, g *WithT, ns string) {
		t.Run("Prometheus", func(t *testing.T) {
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

			g.Eventually(
				CamelMonitorStatus(t, ctx, ns, "timer-to-log"),
				TestTimeoutMedium,
			).Should(
				MatchFields(IgnoreExtras, Fields{
					"Phase": Equal(v1alpha1.CamelMonitorPhaseRunning),
				}),
			)

			// We must verify pod monitor exist
			g.Eventually(PodMonitor(t, ctx, ns, "timer-to-log"), TestTimeoutShort).ShouldNot(BeNil())

			// Start port-forward to Prometheus API
			stopPortForward := PortForwardPrometheus(t, ctx, 9090, 9090, "prometheus", "prometheus-operated")
			defer stopPortForward()

			// Test the prometheus has scraped correctly
			g.Eventually(func() int {
				cmd := exec.Command("sh", "-c",
					`curl -s --get --data-urlencode 'query=camel_exchanges_total{routeId="timer-to-log"}' http://localhost:9090/api/v1/query | jq -r '.data.result[0].value[1]'`,
				)

				out, err := cmd.Output()
				if err != nil {
					t.Fatalf("command failed: %v", err)
				}
				if out == nil {
					// Value is not yet available
					return -1
				}
				// Convert the output string to int
				strVal := strings.TrimSpace(string(out))
				if strVal == "" || strVal == "null" {
					// Value is not yet available
					return -1
				}
				v, err := strconv.Atoi(strVal)
				if err != nil {
					t.Fatalf("string to int failed: %v", err)
				}

				return v
			}, TestTimeoutMedium, 15*time.Second).Should(gomega.BeNumerically(">", 5))
		})
	})
}

func TestVerifyGrafanaDashboard(t *testing.T) {
	WithNewTestNamespace(t, func(ctx context.Context, g *WithT, ns string) {
		t.Run("Grafana", func(t *testing.T) {
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

			g.Eventually(
				CamelMonitorStatus(t, ctx, ns, "timer-to-log"),
				TestTimeoutMedium,
			).Should(
				MatchFields(IgnoreExtras, Fields{
					"Phase": Equal(v1alpha1.CamelMonitorPhaseRunning),
				}),
			)

			// We must verify the dashboard exist
			g.Eventually(GrafanaDashboard(t, ctx, ns, "timer-to-log"), TestTimeoutShort).ShouldNot(BeNil())
			// The struct has no conditions, so we check other fields
			Eventually(func() bool {
				gd, err := GrafanaDashboard(t, ctx, ns, "timer-to-log")()
				if err != nil || gd == nil {
					return false
				}

				return !gd.Status.NoMatchingInstances &&
					!gd.Status.LastResync.IsZero() &&
					gd.Status.UID != ""
			}).Should(BeTrue())
		})
	})
}
