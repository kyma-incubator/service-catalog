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

set -u
set -o errexit

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )
SC_CHART_NAME="catalog"
export SC_NAMESPACE="catalog"

echo "- Initialize Minikube"
bash ${CURRENT_DIR}/scripts/minikube.sh

echo "- Installing Tiller..."
kubectl apply -f ${CURRENT_DIR}/assets/tiller.yaml

bash ${CURRENT_DIR}/scripts/is-ready.sh kube-system name tiller

echo "- Installing ServiceCatalog"
helm repo add svc-cat https://svc-catalog-charts.storage.googleapis.com
helm install svc-cat/catalog --name ${SC_CHART_NAME} --namespace ${SC_NAMESPACE} --wait

echo "- Installing Test broker"
helm install ${CURRENT_DIR}/charts/test-broker-2-0-1.tar.gz --name test-broker --namespace test-broker --wait

echo "- Prepare test resources"
kubectl apply -f ${CURRENT_DIR}/assets/upgrade-job-rbac.yaml
kubectl apply -n ${SC_NAMESPACE} -f ${CURRENT_DIR}/assets/prepare-upgrade-test-job.yaml

export POD_LABEL="prepare-test-job=true"
bash ${CURRENT_DIR}/scripts/test-pod-is-ready.sh

echo "- Prepare upgrade job logs:"
kubectl logs -n ${SC_NAMESPACE} $(kubectl get po -n ${SC_NAMESPACE} -l prepare-test-job=true -ojson | jq -r '.items | .[].metadata.name') -f

echo "- Upgrade ServiceCatalog"
helm upgrade ${SC_CHART_NAME} ${CURRENT_DIR}/charts/service-catalog-crd-0-3-1.tar.gz --namespace ${SC_NAMESPACE} --wait

echo "- Execute upgrade tests"
kubectl apply -n ${SC_NAMESPACE} -f ${CURRENT_DIR}/assets/execute-upgrade-test-job.yaml

export POD_LABEL="execute-test-job=true"
bash ${CURRENT_DIR}/scripts/test-pod-is-ready.sh

echo "- Execute upgrade job logs:"
kubectl logs -n ${SC_NAMESPACE} $(kubectl get po -n ${SC_NAMESPACE} -l execute-test-job=true -ojson | jq -r '.items | .[].metadata.name') -f

echo "- Cleanup"
kubectl delete serviceaccount upgrade-test-job-account -n ${SC_NAMESPACE}
kubectl delete clusterrole upgrade-test-job-account
kubectl delete clusterrolebinding upgrade-test-job-account
kubectl delete jobs.batch prepare-upgrade-test-job -n ${SC_NAMESPACE}
kubectl delete jobs.batch execute-upgrade-test-job -n ${SC_NAMESPACE}
