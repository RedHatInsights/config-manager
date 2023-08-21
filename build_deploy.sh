#!/bin/bash

set -exv

IMAGE="quay.io/cloudservices/config-manager"
IMAGE_TAG=$(git rev-parse --short=7 HEAD)

if [[ -z "$QUAY_USER" || -z "$QUAY_TOKEN" ]]; then
    echo "QUAY_USER and QUAY_TOKEN must be set"
    exit 1
fi

DOCKER_CONF="$PWD/.docker"    
mkdir -p "$DOCKER_CONF"    
    
docker login -u="$QUAY_USER" -p="$QUAY_TOKEN" quay.io    
docker build -t "${IMAGE}:${IMAGE_TAG}" .    
docker tag "${IMAGE}:${IMAGE_TAG}" "${IMAGE}:latest"    
    
docker push "${IMAGE}:${IMAGE_TAG}"    
docker push "${IMAGE}:latest"
