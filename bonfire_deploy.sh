#!/bin/bash

set -xe

HEAD_UNIX=$(date --date="$(git show --no-patch --format='%cI' HEAD)" +%s)
NOW_UNIX=$(date +%s)
SECDIFF=$(("${NOW_UNIX}" - "${HEAD_UNIX}"))
TAG="$(git rev-parse --short HEAD)-${SECDIFF}"
IMAGE="config-manager"
IP=$(minikube ip)

podman build -t "${IMAGE}:${TAG}" -f Dockerfile
podman push "${IMAGE}:${TAG}" "${IP}:5000/${IMAGE}:${TAG}" --tls-verify=false
bonfire deploy \
    --local-config-path ./bonfire_config.yaml \
    --get-dependencies \
    --namespace config-manager \
    --set-parameter "config-manager/CM_PLAYBOOK_HOST=http://${IP}/" \
    --set-parameter "config-manager/IMAGE=localhost:5000/${IMAGE}" \
    --set-parameter "config-manager/IMAGE_TAG=${TAG}" \
    --set-image-tag "quay.io/cloudservices/config-manager=${TAG}" \
    config-manager
