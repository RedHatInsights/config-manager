#!/bin/bash
TAG=`cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 7 | head -n 1`
IMAGE="localhost:5000/config-manager"
podman build -t $IMAGE:$TAG -f Dockerfile
podman push $IMAGE:$TAG `minikube ip`:5000/config-manager:$TAG --tls-verify=false
bonfire deploy -c ./bonfire_config.yaml --get-dependencies --namespace config-manager -p config-manager/config-manager/$IMAGE_TAG=$TAG config-manager -i $IMAGE=$TAG
echo $TAG
