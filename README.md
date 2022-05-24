# Config Manager

Config Manager is a backend service used by the Service Enablement dashboard to enable or disable various Red Hat services on hosts connected through RHC.

Config Manager handles the following actions:

- Retrieves the current configuration (enabled/disabled) of various services supported by Service Enablement 
- Updates the current configuration of services
- Maintains a history of configuration changes
- Ensures that newly connected hosts are kept up to date with the latest configuration
- Updates the host's system profile in Inventory with the latest "rhc_config_state" ID

Config Manager has two mechanisms to update a host:

- Via a change in the Service Enablement dashboard
- Via a new rhc connection event from Inventory

Updating a host (all hosts) via the API:

![Sequence diagram](./docs/config_manager_api.svg)

## REST interface

The REST interface can be used to view and update the current configuration for all hosts connected through RHC. It can also be used to view a history of previous configuration changes, and obtain logs related to those changes. 

See the [OpenAPI Schema](./schema/api.spec.yaml) for details on interacting with the REST interface.

## Event interface

Config-manager consumes and produces kafka messages based on various events.

In topics:
- platform.inventory.events
- platform.playbook-dispatcher.runs

Out topics:
- platform.inventory.system-profile

Event based workflow:
1. Consume new connection event from inventory
2. If connection is reported via cloud-connector check rhc_state_id in host's system profile
3. If rhc_state_id is out of date apply current state to host
4. Consume run events from playbook-dispatcher
5. If run event is successful write new rhc_state_id to host via the system-profile kafka topic.

## Development

### Dependencies

- Golang >= 1.13
- Minikube (see [here](https://consoledot.pages.redhat.com/docs/dev/getting-started/local/environment.html#_install_minikube))
- oc cli (see [here](https://docs.openshift.com/container-platform/4.2/cli_reference/openshift_cli/getting-started-cli.html#cli-installing-cli_cli-developer-commands))
- Bonfire (see [here](https://github.com/RedHatInsights/bonfire#installation))
- Quay Pull Secret (see [here](https://consoledot.pages.redhat.com/docs/dev/getting-started/local/environment.html#_get_your_quay_pull_secret))

### Deploying locally

Config-manager is managed by [Clowder](https://github.com/RedHatInsights/clowder) and can be deployed locally onto a [Minikube](https://minikube.sigs.k8s.io/docs/start/) instance using [Bonfire](https://github.com/RedHatInsights/bonfire).

#### Setting up an ephemeral enviroment

The following steps (detailed
[here](https://consoledot.pages.redhat.com/docs/dev/getting-started/local/environment.html))
should be performed to set up a local ephemeral environment:

1. Start minikube (check [here](https://github.com/RedHatInsights/clowder/blob/master/docs/macos.md) for MacOS instructions)
```sh
minikube start --cpus 8 --disk-size 36GB --memory 16GB --addons=registry --driver=kvm2
```

2. Install Clowder CRDs
```sh
curl https://raw.githubusercontent.com/RedHatInsights/clowder/master/build/kube_setup.sh -o kube_setup.sh && chmod +x kube_setup.sh
./kube_setup.sh
```

3. Install Clowder (replace version with [latest](https://github.com/RedHatInsights/clowder/releases/latest))
```sh
minikube kubectl -- apply -f https://github.com/RedHatInsights/clowder/releases/download/0.15.0/clowder-manifest-0.15.0.yaml --validate=false
```

4. Create a namespace for config-manager
```sh
minikube kubectl -- create ns config-manager
```

5. Download your quay secret if you haven't already, and add it to the
   config-manager namespace
```sh
cp $USER-secret.yml quay-cloudservices-pull.yml
sed -ie "s/$USER-pull-secret/quay-cloudservices-pull/" quay-cloudservices-pull.yml
minikube kubectl -- create $USER-secret.yml --namespace config-manager
minikube kubectl -- create quay-cloudservices-pull.yml --namespace config-manager
```

6. Deploy ClowdEnvironment
```sh
bonfire deploy-env -n config-manager -u $USER
```

7. Deploy apps config-manager depends on
```
bonfire deploy -n config-manager cloud-connector host-inventory playbook-dispatcher
```

#### Deploy config-manager into the ephemeral environment

It is possible to deploy config-manager into the environment as a container
image and run it. This method mostly closely mimics the environment under which
config-manager runs in stage and production, but makes development slightly more
difficult. See [Running config-manager locally](#Running-config-manager-locally)
below for details on how to run config-manager locally while connecting to the
cluster services.

1. Build an image from the current working directory, push it to the minikube
   registry, and deploy it into the local ephemeral environment. Repeat this
   step any time code changes have been made and need to be redeployed.
```sh
./bonfire_deploy.sh
```

2. Forward ports from localhost into the services in the cluster. Change ports
   or services as necessary.
```sh
kubectl -n config-manager port-forward --address 0.0.0.0 svc/mosquitto 1883
kubectl -n config-manager port-forward                   svc/config-manager-db 5432:5432
kubectl -n config-manager port-forward                   svc/host-inventory-db 5433:5432
kubectl -n config-manager port-forward                   svc/cloud-connector-db 5434:5432
kubectl -n config-manager port-forward                   svc/config-manager-service 8000:8000
kubectl -n config-manager port-forward                   svc/playbook-dispatcher-api 8001:8000
kubectl -n config-manager port-forward                   svc/host-inventory-service 8002:8000
kubectl -n config-manager port-forward                   svc/cloud-connector 8003:8080
```

To access Kafka message topics, run a `kcat` container directly in the cluster
is most effective:

```
# Identify the environment name and export it
export CONFIG_MANAGER_ENV=$(kubectl -n config-manager get svc -l env=env-config-manager,app.kubernetes.io/name=kafka -o json | jq '.items[0].metadata.labels["app.kubernetes.io/instance"]' -r)
kubectl -n config-manager run -it --rm --image=edenhill/kcat:1.7.1 kcat -- -b $CONFIG_MANAGER_ENV-kafka-bootstrap.config-manager.svc.cluster.local:9092 -t platform.inventory.events
```

3. Access config-manager
```sh
oc port-forward svc/config-manager-service -n config-manager 8000 &
curl -v -H "x-rh-identity:eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" http://localhost:8000/api/config-manager/v1/states/current
```

#### Running config-manager locally

To more effectively develop config-manager, it is possible to connect it to
running services while running it locally. This can enable visual debugging or
other tracing practices that are made slightly more difficult through the
abstraction layer created by minikube and kubernetes.

This technique is an adaptation of the method described in the [CONTRIBUTING
guide](CONTRIBUTING.md).

Perform all the steps as described in [Setting up an ephemeral enviroment]().
Once the dependent services are running, forward the following ports from
localhost to the appropriate kubernetes service resources:

```
export CONFIG_MANAGER_ENV=$(kubectl -n config-manager get svc -l env=env-config-manager,app.kubernetes.io/name=kafka -o json | jq '.items[0].metadata.labels["app.kubernetes.io/instance"]' -r)
kubectl -n config-manager port-forward --address 0.0.0.0 svc/mosquitto 1883                                  &
kubectl -n config-manager port-forward                   svc/playbook-dispatcher-api 8001:8000               &
kubectl -n config-manager port-forward                   svc/host-inventory-service 8002:8000                &
kubectl -n config-manager port-forward                   svc/cloud-connector 8003:8080                       &
kubectl -n config-manager port-forward                   svc/${CONFIG_MANAGER_ENV}-kafka-bootstrap 9094:9094 &
```

Because we've skipped deploying config-manager into the cluster, we need to run
a PostgreSQL database manually.

```
podman run --env POSTGRES_PASSWORD=insights --env POSTGRES_USER=insights --env POSTGRES_DB=insights --publish 5432:5432 --name config-manager-db --detach postgres
```

Next, run config-manager like so:

```
LOG_LEVEL=debug \
CM_DISPATCHER_HOST=http://localhost:8001/ \
CM_DISPATCHER_PSK=$(kubectl -n config-manager get secrets/psk-playbook-dispatcher -o json | jq '.data.key' -r | base64 -d) \
CM_INVENTORY_HOST=http://localhost:8002/ \
CM_CLOUD_CONNECTOR_HOST=http://localhost:8003/api/cloud-connector/v1/ \
CM_CLOUD_CONNECTOR_PSK=$(kubectl -n config-manager get secrets/psk-cloud-connector -o json | jq '.data["client-psk"]' -r | base64 -d) \
CM_WEB_PORT=8080 \
CM_DB_USER=insights \
CM_DB_PASS=insights \
CM_DB_NAME=insights \
go run . run
```
