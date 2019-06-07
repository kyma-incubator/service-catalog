#!/usr/bin/env bash

set -o errexit

VER=0.1

eval $(minikube docker-env)
env CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o upgrade-test
docker build -t upgrade-test .
docker tag upgrade-test eu.gcr.io/kyma-project/develop/service-catalog/upgrade-test:$VER
docker push eu.gcr.io/kyma-project/develop/service-catalog/upgrade-test:$VER

rm -f upgrade-test
