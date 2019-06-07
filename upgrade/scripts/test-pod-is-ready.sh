#!/usr/bin/env bash

for var in SC_NAMESPACE POD_LABEL; do
    if [ -z "${!var}" ] ; then
        echo "ERROR: $var is not set"
        discoverUnsetVar=true
    fi
done

fuse=0
while [ $fuse -le 120 ]; do
    status=$(kubectl get po -n ${SC_NAMESPACE} -l ${POD_LABEL} -ojson | jq -r '.items | .[].status.conditions | .[] | select(.type == "Ready") | .status')
    containerStatus=$(kubectl get po -n ${SC_NAMESPACE} -l ${POD_LABEL} -ojson | jq -r '.items | .[].status.conditions | .[] | select(.type == "ContainersReady") | .status')
    if [[ $status == "True" ]] && [[ $containerStatus == "True" ]]; then
        echo "- Pod with upgrade test is ready"
        break
    fi
    echo "- Pod with upgrade test is not ready. wait..."
    sleep 1

    ((fuse+=1))
done
