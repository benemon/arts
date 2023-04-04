#!/bin/bash

export GOOS=linux 
export GOARCH=amd64

go build .
mv arts build/
until oc start-build arts --from-dir=build/
do
    echo ...
    sleep 1
done
oc get builds | grep 'Failed' | awk '{print $1}' | xargs oc delete build
oc get pods | grep 'Error\|Completed' | awk '{print $1}' | xargs oc delete pod
