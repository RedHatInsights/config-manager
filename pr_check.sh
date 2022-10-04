#!/bin/bash
# --------------------------------------------
# Export vars for helper scripts to use
# --------------------------------------------
# name of app-sre "application" folder this component lives in; needs to match for the push to quay.
export COMPONENT="config-manager"
# Needs to match the quay repo name set by app.yaml in app-interface
export IMAGE="quay.io/cloudservices/config-manager"
export WORKSPACE=${WORKSPACE:-$APP_ROOT}  # if running in jenkins, use the build's workspace
export APP_ROOT=$(pwd)
export NODE_BUILD_VERSION=16
COMMON_BUILDER=https://raw.githubusercontent.com/RedHatInsights/insights-frontend-builder-common/master

# --------------------------------------------
# Options that must be configured by app owner
# --------------------------------------------
IQE_PLUGINS="config-manager"
IQE_MARKER_EXPRESSION="smoke"
IQE_FILTER_EXPRESSION=""
#
# Derived from:
# https://github.com/RedHatInsights/bonfire/blob/master/cicd/examples/pr_check_template.sh

# It should be possible to execute the following from the app-interface repo:
#
#   yq ".resourceTemplates[] | select(.name == '${COMPONENT_NAME}')" \
#     < "data/services/insights/${APP_NAME}/deploy.yml"
#
readonly APP_NAME="config-manager"
readonly COMPONENT_NAME="config-manager"
readonly REF_ENV="insights-stage"

# The image built from this repository
readonly IMAGE="quay.io/cloudservices/config-manager"

# See ${CICD_ROOT}/cji_smoke_test.sh
readonly IQE_CJI_TIMEOUT="10m"
readonly IQE_PLUGINS="config-manager"

# Install bonfire into a venv and clone the bonfire repo. Export APP_ROOT, BONFIRE_ROOT, etc.
source <(curl -s https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd/bootstrap.sh)

# Build an image and push it to quay. Export no variables.
source "${CICD_ROOT}/build.sh"

EXTRA_DEPLOY_ARGS=""

# Include cloud-connector (not just cloud-connector-api) as a deployment
# argument. This pulls in the 'cloud-connector' resource, which in turn pulls in
# the mosquitto service we need to stand up a live environment.
EXTRA_DEPLOY_ARGS="${EXTRA_DEPLOY_ARGS} cloud-connector"

# Set additional deploy arguments to use live implementations of services rather
# than mocks. This should eventually be removed in favor of changing the static
# parameters in the app-interface ephemeral-base target:
# https://gitlab.cee.redhat.com/service/app-interface/-/blob/master/data/services/insights/config-manager/deploy.yml#L67-69
EXTRA_DEPLOY_ARGS="${EXTRA_DEPLOY_ARGS} --set-parameter config-manager/CM_LOG_LEVEL=debug"
EXTRA_DEPLOY_ARGS="${EXTRA_DEPLOY_ARGS} --set-parameter config-manager/CM_CLOUD_CONNECTOR_IMPL=impl"
EXTRA_DEPLOY_ARGS="${EXTRA_DEPLOY_ARGS} --set-parameter config-manager/CM_DISPATCHER_IMPL=impl"
EXTRA_DEPLOY_ARGS="${EXTRA_DEPLOY_ARGS} --set-parameter config-manager/CM_INVENTORY_IMPL=impl"
EXTRA_DEPLOY_ARGS="${EXTRA_DEPLOY_ARGS} --set-parameter config-manager/CM_DISPATCHER_HOST=http://playbook-dispatcher-api:8000"
EXTRA_DEPLOY_ARGS="${EXTRA_DEPLOY_ARGS} --set-parameter config-manager/CM_INVENTORY_HOST=http://host-inventory-service:8000"
EXTRA_DEPLOY_ARGS="${EXTRA_DEPLOY_ARGS} --set-parameter config-manager/CM_CLOUD_CONNECTOR_HOST=http://cloud-connector:8080"

# Deploy an ephemeral env with the just-created image. Export NAMESPACE.
source "${CICD_ROOT}/deploy_ephemeral_env.sh"

# Run IQE tests in the just-created ephemeral env. Export MINIO_*.
source "${CICD_ROOT}/cji_smoke_test.sh"
