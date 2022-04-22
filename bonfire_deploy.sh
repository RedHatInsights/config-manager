#!/bin/bash

TAG=$(tr -dc 'a-zA-Z0-9' < /dev/urandom | fold -w 7 | head -n 1)
IMAGE="localhost:5000/config-manager"
IP=$(minikube ip)

podman build -t "${IMAGE}:${TAG}" -f Dockerfile
podman push "${IMAGE}:${TAG}" "${IP}:5000/config-manager:${TAG}" --tls-verify=false
bonfire deploy -c ./bonfire_config.yaml --get-dependencies --namespace config-manager --set-parameter "config-manager/config-manager/IMAGE_TAG=${TAG}" --set-image-tag "${IMAGE}=${TAG}" config-manager
echo "${TAG}"
