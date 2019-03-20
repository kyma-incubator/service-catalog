#!/usr/bin/env bash

set -eu

CURRENT_DIR=$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )

echo "- Initialize Minikube"
bash ${CURRENT_DIR}/minikube.sh

echo "- Installing Tiller..."
kubectl apply -f ${CURRENT_DIR}/../assets/tiller.yaml

bash ${CURRENT_DIR}/is-ready.sh kube-system name tiller

echo "- Installing SC"
helm install --name catalog --namespace kyma-system  ${CURRENT_DIR}/../../../../charts/catalog/ --wait

echo "- Installing Pod Preset Helm Chart"
helm install ${CURRENT_DIR}/../assets/pod-preset-chart.tgz  --name podpreset --namespace kyma-system --wait

echo "- Installing Helm Broker Helm Chart"
helm install ${CURRENT_DIR}/../assets/helm-broker-chart.tgz  --name helm-broker --namespace kyma-system --wait
echo "- Installing BUC Helm Chart"
helm install ${CURRENT_DIR}/../assets/buc-chart.tgz  --name buc --namespace kyma-system --wait

echo "- Register Helm Broker in Service Catalog"
kubectl apply -f  ${CURRENT_DIR}/../assets/helm-broker.yaml

echo "- Scale down controller manager"
kubectl -n kyma-system scale deploy --replicas=0 catalog-catalog-controller-manager

echo "- Expose Helm Broker to localhost on port 8081"
export HB_POD_NAME=$(kubectl get po -l app=helm-broker -n kyma-system -o jsonpath='{ .items[*].metadata.name }')
kubectl port-forward -n kyma-system pod/${HB_POD_NAME} 8081:8080
