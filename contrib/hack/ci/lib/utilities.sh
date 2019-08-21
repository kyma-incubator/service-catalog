#!/usr/bin/env bash
# Copyright 2019 The Kubernetes Authors.
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

#
# Library of useful utilities for CI purposes.
#

readonly LIB_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

# Prints first argument as header. Additionally prints current date.
shout() {
    echo -e "
#################################################################################################
# $(date)
# $1
#################################################################################################
"
}

# Installs kind and helm dependencies locally.
# Required envs:
#  - KIND_VERSION
#  - HELM_VERSION
#  - INSTALL_DIR
install::local::kind_and_helm() {
    mkdir -p "${INSTALL_DIR}/bin"
    export PATH="${INSTALL_DIR}/bin:${PATH}"

    pushd "${INSTALL_DIR}"

    shout "- Install helm ${STABLE_HELM_VERSION} locally to a tempdir GOPATH..."
    curl -LO https://git.io/get_helm.sh > ${INSTALL_DIR}/get_helm.sh
    chmod 700 ${INSTALL_DIR}/get_helm.sh
    env HELM_INSTALL_DIR="${INSTALL_DIR}/bin" ./get_helm.sh \
        --version ${STABLE_HELM_VERSION} \
        --no-sudo

    shout "- Install kind ${STABLE_KIND_VERSION} locally to a tempdir GOPATH..."
    env "GOPATH=${INSTALL_DIR}" GO111MODULE="on" go get "sigs.k8s.io/kind@${STABLE_KIND_VERSION}"

    popd
}

# Installs tiller on cluster
install::cluster::tiller() {
    shout "- Installing Tiller..."
    kubectl create -f ${LIB_DIR}/tiller-rbac.yaml
    helm init --service-account tiller --wait
}

# Installs Service Catalog from newest 0.2.x release on k8s cluster.
# Required envs:
#  - SC_CHART_NAME
#  - SC_NAMESPACE
install::cluster::service_catalog_v2() {
    shout "- Installing Service Catalog in version 0.2.x"
    helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
    # TODO: After https://github.com/kyma-project/kyma/issues/5217, change `helm install svc-cat/catalog` to `helm install svc-cat/catalog-apiserver`
    # install always the newest service catalog with apiserver
    helm install svc-cat/catalog --name ${SC_CHART_NAME} --namespace ${SC_NAMESPACE} --wait
}

#
# 'kind'(kubernetes-in-docker) functions
#
readonly KIND_CLUSTER_NAME="kind-ci"

kind::create_cluster() {
    shout "- Create k8s cluster..."
    kind create cluster --name=${KIND_CLUSTER_NAME} --wait=1m
    export KUBECONFIG="$(kind get kubeconfig-path --name=${KIND_CLUSTER_NAME})"
}

kind::delete_cluster() {
    kind delete cluster --name=${KIND_CLUSTER_NAME}
}

# Arguments:
#   $1 - image name to copy into cluster nodes
kind::load_image() {
    kind load docker-image $1 --name=${KIND_CLUSTER_NAME}
}