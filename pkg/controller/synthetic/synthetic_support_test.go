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
	"testing"

	v1alpha1 "github.com/camel-tooling/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestAllPodsReady(t *testing.T) {
	pods := []v1alpha1.PodInfo{
		{Ready: true},
		{Ready: true},
	}
	assert.True(t, allPodsReady(pods))

	pods[1].Ready = false
	assert.False(t, allPodsReady(pods))
}

func TestAllPodsUp(t *testing.T) {
	pods := []v1alpha1.PodInfo{
		{Runtime: &v1alpha1.RuntimeInfo{Status: "UP"}},
		{Runtime: &v1alpha1.RuntimeInfo{Status: "UP"}},
	}
	assert.True(t, allPodsUp(pods))

	pods[1].Runtime.Status = "DOWN"
	assert.False(t, allPodsUp(pods))
}
