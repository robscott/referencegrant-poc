#!/usr/bin/env bash

# Copyright 2023 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

echo "Generating v1alpha1 CRDs and deepcopy"
go run sigs.k8s.io/controller-tools/cmd/controller-gen \
        object:headerFile=./hack/boilerplate/boilerplate.generatego.txt \
        crd:crdVersions=v1 \
        output:crd:artifacts:config=config/crd \
        paths=./apis/v1alpha1

readonly APIS_PKG=sigs.k8s.io/referencegrant-poc
readonly VERSION=v1alpha1

echo "Generating ${VERSION} register at ${APIS_PKG}/apis/${VERSION}"
go run k8s.io/code-generator/cmd/register-gen \
    --input-dirs "./apis/v1alpha1" \
    --output-package "./apis/v1alpha1" \
    --go-header-file ./hack/boilerplate/boilerplate.generatego.txt
