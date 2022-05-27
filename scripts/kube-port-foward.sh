#!/bin/bash

_term() {
    kill -TERM "${CHILD_PIDS[@]}"
}

trap _term SIGTERM

declare -a CHILD_PIDS

declare -A SERVICES=(
    [mosquitto]=1883
    [playbook-dispatcher-api]=8001:8000
    [host-inventory-service]=8002:8000
    [cloud-connector]=8003:8080
)

for SERVICE in "${!SERVICES[@]}"; do
    PORT_MAP="${SERVICES[$SERVICE]}"
    minikube kubectl -- -n fog port-forward --address 0.0.0.0 svc/"${SERVICE}" "${PORT_MAP}" &
    CHILD_PIDS+=($!)
done

wait "${CHILD_PIDS[@]}"
