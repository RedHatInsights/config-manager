#!/bin/bash
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

# The image built from this repository
readonly IMAGE="quay.io/cloudservices/config-manager"

# See ${CICD_ROOT}/cji_smoke_test.sh
readonly IQE_CJI_TIMEOUT="10m"
IQE_MARKER_EXPRESSION="config_manager"  # writable because cji_smoke_test.sh uses var=${var:=val}
readonly IQE_PLUGINS="rhc"

# Install bonfire into a venv and clone the bonfire repo. Export APP_ROOT, BONFIRE_ROOT, etc.
source <(curl -s https://raw.githubusercontent.com/RedHatInsights/bonfire/master/cicd/bootstrap.sh)

# Build an image and push it to quay. Export no variables.
source "${CICD_ROOT}/build.sh"

# Deploy an ephemeral env with the just-created image. Export NAMESPACE.
source "${CICD_ROOT}/deploy_ephemeral_env.sh"

# Run IQE tests in the just-created ephemeral env. Export MINIO_*.
source "${CICD_ROOT}/cji_smoke_test.sh"
