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

// Code generated by applyconfiguration-gen. DO NOT EDIT.

package v1alpha1

import (
	camelv1alpha1 "github.com/squakez/camel-dashboard-operator/pkg/apis/camel/v1alpha1"
)

// AppStatusApplyConfiguration represents a declarative configuration of the AppStatus type for use
// with apply.
type AppStatusApplyConfiguration struct {
	Phase    *camelv1alpha1.AppPhase     `json:"phase,omitempty"`
	Image    *string                     `json:"image,omitempty"`
	Pods     []PodInfoApplyConfiguration `json:"pods,omitempty"`
	Replicas *int32                      `json:"replicas,omitempty"`
	Info     *string                     `json:"info,omitempty"`
}

// AppStatusApplyConfiguration constructs a declarative configuration of the AppStatus type for use with
// apply.
func AppStatus() *AppStatusApplyConfiguration {
	return &AppStatusApplyConfiguration{}
}

// WithPhase sets the Phase field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Phase field is set to the value of the last call.
func (b *AppStatusApplyConfiguration) WithPhase(value camelv1alpha1.AppPhase) *AppStatusApplyConfiguration {
	b.Phase = &value
	return b
}

// WithImage sets the Image field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Image field is set to the value of the last call.
func (b *AppStatusApplyConfiguration) WithImage(value string) *AppStatusApplyConfiguration {
	b.Image = &value
	return b
}

// WithPods adds the given value to the Pods field in the declarative configuration
// and returns the receiver, so that objects can be build by chaining "With" function invocations.
// If called multiple times, values provided by each call will be appended to the Pods field.
func (b *AppStatusApplyConfiguration) WithPods(values ...*PodInfoApplyConfiguration) *AppStatusApplyConfiguration {
	for i := range values {
		if values[i] == nil {
			panic("nil value passed to WithPods")
		}
		b.Pods = append(b.Pods, *values[i])
	}
	return b
}

// WithReplicas sets the Replicas field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Replicas field is set to the value of the last call.
func (b *AppStatusApplyConfiguration) WithReplicas(value int32) *AppStatusApplyConfiguration {
	b.Replicas = &value
	return b
}

// WithInfo sets the Info field in the declarative configuration to the given value
// and returns the receiver, so that objects can be built by chaining "With" function invocations.
// If called multiple times, the Info field is set to the value of the last call.
func (b *AppStatusApplyConfiguration) WithInfo(value string) *AppStatusApplyConfiguration {
	b.Info = &value
	return b
}
