#!/bin/bash

set -xe

HEAD_UNIX=$(date --date="$(git show --no-patch --format='%cI' HEAD)" +%s)
NOW_UNIX=$(date +%s)
SECDIFF=$(("${NOW_UNIX}" - "${HEAD_UNIX}"))
TAG="$(git rev-parse --short HEAD)-${SECDIFF}"
IMAGE="config-manager"
CLUSTER_IP=$(minikube ip)
TARGET_NAMESPACE=$(minikube kubectl -- get env -o json | jq '.items[0].spec.targetNamespace' -r)
HOST_IP=$(nmcli --terse connection show virbr0 | grep ipv4.addresses | cut -d":" -f2 | cut -d"/" -f1)

podman build -t "${IMAGE}:${TAG}" -f Dockerfile
podman push "${IMAGE}:${TAG}" "${CLUSTER_IP}:5000/${IMAGE}:${TAG}" --tls-verify=false
bonfire deploy \
    --local-config-path "$(dirname ${0})/bonfire.yml" \
    --get-dependencies \
    --namespace "${TARGET_NAMESPACE}" \
    --set-parameter "config-manager/CM_PLAYBOOK_HOST=http://${HOST_IP}:8080" \
    --set-parameter "config-manager/IMAGE=localhost:5000/${IMAGE}" \
    --set-parameter "config-manager/IMAGE_TAG=${TAG}" \
    --set-image-tag "quay.io/cloudservices/config-manager=${TAG}" \
    config-manager
