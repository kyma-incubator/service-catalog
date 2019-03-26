#!/usr/bin/env bash

set -o errexit

eval $(minikube docker-env)
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o service-catalog ./../../cmd/service-catalog/main.go
docker build -t service-catalog .
docker tag service-catalog eu.gcr.io/kyma-project/develop/service-catalog/service-catalog-amd64:crd-0.0.1
kubectl -n kyma-system delete po -l app=catalog-catalog-webhook
