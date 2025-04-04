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

package cmd

import (
	"github.com/spf13/cobra"
	"github.com/squakez/camel-dashboard-operator/pkg/cmd/operator"
	"github.com/squakez/camel-dashboard-operator/pkg/platform"
	"github.com/squakez/camel-dashboard-operator/pkg/util/defaults"
)

const (
	operatorCommand       = "operator"
	defaultHealthPort     = 8081
	defaultMonitoringPort = 8080
)

func newCmdOperator(rootCmdOptions *RootCmdOptions) (*cobra.Command, *operatorCmdOptions) {
	options := operatorCmdOptions{}

	cmd := cobra.Command{
		Use:     "operator",
		Short:   "Run the Camel Dashboard operator",
		Long:    `Run the Camel Dashboard operator`,
		Hidden:  true,
		PreRunE: decode(&options, rootCmdOptions.Flags),
		Run:     options.run,
	}

	cmd.Flags().Int32("health-port", defaultHealthPort, "The port of the health endpoint")
	cmd.Flags().Int32("monitoring-port", defaultMonitoringPort, "The port of the metrics endpoint")
	cmd.Flags().Bool("leader-election", true, "Use leader election")
	cmd.Flags().String("leader-election-id", "", "Use the given ID as the leader election Lease name")

	return &cmd, &options
}

type operatorCmdOptions struct {
	HealthPort       int32  `mapstructure:"health-port"`
	MonitoringPort   int32  `mapstructure:"monitoring-port"`
	LeaderElection   bool   `mapstructure:"leader-election"`
	LeaderElectionID string `mapstructure:"leader-election-id"`
}

func (o *operatorCmdOptions) run(_ *cobra.Command, _ []string) {
	leaderElectionID := o.LeaderElectionID
	if leaderElectionID == "" {
		if defaults.OperatorID() != "" {
			leaderElectionID = platform.GetOperatorLockName(defaults.OperatorID())
		} else {
			leaderElectionID = platform.OperatorLockName
		}
	}

	// TODO fix this
	// operator.Run(o.HealthPort, o.MonitoringPort, o.LeaderElection, leaderElectionID)
	operator.Run(defaultHealthPort, defaultMonitoringPort, true, leaderElectionID)
}
