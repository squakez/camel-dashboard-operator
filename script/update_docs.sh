#!/bin/bash

# Licensed to the Apache Software Foundation (ASF) under one or more
# contributor license agreements.  See the NOTICE file distributed with
# this work for additional information regarding copyright ownership.
# The ASF licenses this file to You under the Apache License, Version 2.0
# (the "License"); you may not use this file except in compliance with
# the License.  You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

location=$(dirname $0)

echo "Scraping information from Makefile"
LAST_RELEASED_VERSION=$(grep '^LAST_RELEASED_VERSION ?= ' Makefile | sed 's/^.* \?= //')
RUNTIME_VERSION=$(grep '^DEFAULT_RUNTIME_VERSION := ' Makefile | sed 's/^.* \?= //')

CATALOG="$location/../pkg/resources/resources/camel-catalog-$RUNTIME_VERSION.yaml"
# This script requires the catalog to be available (via make build-resources for instance)
if [ ! -f $CATALOG ]; then
    echo "❗ catalog not available. Make sure to download it before calling this script."
    exit 1
fi

KUSTOMIZE_VERSION=$(grep '^KUSTOMIZE_VERSION := ' Makefile | sed 's/^.* \?= //' | sed 's/^.//')

echo "Camel K Runtime version: $RUNTIME_VERSION"
echo "Kustomize version: $KUSTOMIZE_VERSION"

yq -i ".asciidoc.attributes.kustomize-version = \"$KUSTOMIZE_VERSION\"" $location/../docs/antora.yml

echo "Scraping information from catalog available at: $CATALOG"
RUNTIME_VERSION=$(yq '.spec.runtime.version' $CATALOG)
CAMEL_VERSION=$(yq '.spec.runtime.metadata."camel.version"' $CATALOG)
re="^([[:digit:]]+)\.([[:digit:]]+)\.([[:digit:]]+)$"
if ! [[ $CAMEL_VERSION =~ $re ]]; then
    echo "❗ argument must match semantic version: $CAMEL_VERSION"
    exit 1
fi
CAMEL_DOCS_VERSION="${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.x"
CAMEL_QUARKUS_VERSION=$(yq '.spec.runtime.metadata."camel-quarkus.version"' $CATALOG)
re="^([[:digit:]]+)\.([[:digit:]]+)\.([[:digit:]]+)$"
if ! [[ $CAMEL_QUARKUS_VERSION =~ $re ]]; then
    echo "❗ argument must match semantic version: $CAMEL_QUARKUS_VERSION"
    exit 1
fi
CAMEL_QUARKUS_DOCS_VERSION="${BASH_REMATCH[1]}.${BASH_REMATCH[2]}.x"
QUARKUS_VERSION=$(yq '.spec.runtime.metadata."quarkus.version"' $CATALOG)

echo "Camel K latest version: $LAST_RELEASED_VERSION"
echo "Camel K Runtime version: $RUNTIME_VERSION"
echo "Camel version: $CAMEL_VERSION"
echo "Camel Quarkus version: $CAMEL_QUARKUS_VERSION"
echo "Quarkus version: $QUARKUS_VERSION"

yq -i ".asciidoc.attributes.last-released-version = \"$LAST_RELEASED_VERSION\"" $location/../docs/antora.yml
yq -i ".asciidoc.attributes.camel-dashboard-runtime-version = \"$RUNTIME_VERSION\"" $location/../docs/antora.yml
yq -i ".asciidoc.attributes.camel-version = \"$CAMEL_VERSION\"" $location/../docs/antora.yml
yq -i ".asciidoc.attributes.camel-docs-version = \"$CAMEL_DOCS_VERSION\"" $location/../docs/antora.yml
yq -i ".asciidoc.attributes.camel-quarkus-version = \"$CAMEL_QUARKUS_VERSION\"" $location/../docs/antora.yml
yq -i ".asciidoc.attributes.camel-quarkus-docs-version = \"$CAMEL_QUARKUS_DOCS_VERSION\"" $location/../docs/antora.yml
yq -i ".asciidoc.attributes.quarkus-version = \"$QUARKUS_VERSION\"" $location/../docs/antora.yml

echo "Scraping information from go.mod"
KNATIVE_API_VERSION=$(grep '^.*knative.dev/eventing ' $location/../go.mod | sed 's/^.* //' | sed 's/^.//')
KUBE_API_VERSION=$(grep '^.*k8s.io/api ' $location/../go.mod | sed 's/^.* //' | sed 's/^.//')
OPERATOR_FWK_API_VERSION=$(grep '^.*github.com/operator-framework/api ' $location/../go.mod | sed 's/^.* //' | sed 's/^.//')
SERVICE_BINDING_OP_VERSION=$(grep '^.*github.com/redhat-developer/service-binding-operator ' $location/../go.mod | sed 's/^.* //' | sed 's/^.//')
PROMETHEUS_OP_VERSION=$(grep '^.*github.com/prometheus-operator/prometheus-operator/pkg/apis/monitoring ' $location/../go.mod | sed 's/^.* //' | sed 's/^.//')

echo "Kubernetes API version: $KUBE_API_VERSION"
echo "Operator Framework API version: $OPERATOR_FWK_API_VERSION"
echo "Knative API version: $KNATIVE_API_VERSION"
echo "Service Binding Operator version: $SERVICE_BINDING_OP_VERSION"
echo "Prometheus Operator version: $PROMETHEUS_OP_VERSION"

yq -i ".asciidoc.attributes.kubernetes-api-version = \"$KUBE_API_VERSION\"" $location/../docs/antora.yml
yq -i ".asciidoc.attributes.operator-fwk-api-version = \"$OPERATOR_FWK_API_VERSION\"" $location/../docs/antora.yml
yq -i ".asciidoc.attributes.knative-api-version = \"$KNATIVE_API_VERSION\"" $location/../docs/antora.yml
yq -i ".asciidoc.attributes.service-binding-op-version = \"$SERVICE_BINDING_OP_VERSION\"" $location/../docs/antora.yml
yq -i ".asciidoc.attributes.prometheus-op-version = \"$PROMETHEUS_OP_VERSION\"" $location/../docs/antora.yml
