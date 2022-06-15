# Getting Started

These instructions will result in an (almost) fully function development
environment for config-manager using minikube to run kubernetes and bonfire to
deploy the dependent platform applications into the cluster.

## Prerequisites

### quay.io

Follow the [instructions for generating a quay pull
secret](https://consoledot.pages.redhat.com/docs/dev/getting-started/local/environment.html#_get_your_quay_pull_secret),
but addition to downloading a pull secret file, copy the "podman login"
instructions and log in.

You must have access to the
[quay.io/cloudservices](https://quay.io/organization/cloudservices?tab=teams)
repository. In your app-interface user file, make sure `quay_username` is set to
your quay.io username and that you are a member of the `insights-engineers.yml`
role.

### `minikube`

Follow the [instructions on installing
minikube](https://consoledot.pages.redhat.com/docs/dev/getting-started/local/environment.html#_install_minikube).

### `oc`

Follow the [instructions on installing
oc](https://docs.openshift.com/container-platform/4.2/cli_reference/openshift_cli/getting-started-cli.html#cli-installing-cli_cli-developer-commands).

### `bonfire`

Follow the [instructions on installing
bonfire](https://github.com/RedHatInsights/bonfire#installation).

## Set up a kubernetes cluster using minikube

You only need to do this once.

### Start `minikube`

```sh
minikube start --cpus 8 --disk-size 36GB --memory 16GB --addons=registry --driver=kvm2
```

### Install Clowder Custom Resource Definitions

```sh
curl https://raw.githubusercontent.com/RedHatInsights/clowder/master/build/kube_setup.sh | bash
```

### Install Clowder

```sh
minikube kubectl -- apply --filename $(curl https://api.github.com/repos/RedHatInsights/clowder/releases/latest | jq '.assets[0].browser_download_url' -r)
```

### Create a namespace

Choose a namespace to deploy applications into. The name you choose doesn't
matter, but you'll be typing it a lot. Pick something short. I like 'fog'
(because fog is a cloud on the ground... get it?).

```sh
minikube kubectl -- create namespace fog
```

### Create Pull Secrets

Some deployments are hard-coded to use a secret named `quay-cloudservices-pull`,
so copy the pull secret you downloaded [above](#create-pull-secrets) and rename
it:

```sh
cp $USER-secret.yml quay-cloudservices-pull.yml
sed -ie "s/$USER-pull-secret/quay-cloudservices-pull/" quay-cloudservices-pull.yml
```

Now create the secrets in your new namespace:

```sh
minikube kubectl -- --namespace fog create --filename $USER-secret.yml
minikube kubectl -- --namespace fog create --filename quay-cloudservices-pull.yml
```

### Deploy ClowdEnvironment

```sh
bonfire deploy-env --namespace fog --quay-user $USER
```

### Deploy Applications

```sh
bonfire deploy --namespace fog --local-config-path ./bonfire_config.yaml \
    cloud-connector \
    host-inventory \
    playbook-dispatcher
```

## Run config-manager

There are two ways to run config-manager: directly on localhost or deployed into
the minikube cluster. Both have advantages and drawbacks, so both are presented
equally here.

### Run directly on localhost

Running directly on localhost allows for a more rapid edit/execute/debug
development cycle, but requires forwarding ports from localhost into the cluster
services.

#### Forward ports

The included `kube-port-forward.sh` script will forward ports from localhost to
all the necessary services in the cluster.

```sh
bash scripts/kube-port-forward.sh
```

| service                 | local port | cluster port |
| ----------------------- | ---------- | ------------ |
| mosquitto               | 1883       | 1883         |
| playbook-dispatcher-api | 8001       | 8000         |
| host-inventory-service  | 8002       | 8000         |
| cloud-connector         | 8003       | 8080         |
| kafka-ext-bootstrap     | 9094       | 9094         |

#### Run database

```sh
podman run --env POSTGRES_PASSWORD=insights --env POSTGRES_USER=insights --env POSTGRES_DB=insights --publish 5432:5432 --name config-manager-db --detach postgres
```

#### Run config-manager

```sh
CM_LOG_LEVEL=debug \
CM_KAFKA_BROKERS=localhost:9094 \
CM_DISPATCHER_HOST=http://localhost:8001/ \
CM_DISPATCHER_PSK=$(kubectl -n fog get secrets/psk-playbook-dispatcher -o json | jq '.data.key' -r | base64 -d) \
CM_INVENTORY_HOST=http://localhost:8002/ \
CM_CLOUD_CONNECTOR_HOST=http://localhost:8003/api/cloud-connector/v1/ \
CM_CLOUD_CONNECTOR_PSK=$(kubectl -n fog get secrets/psk-cloud-connector -o json | jq '.data["client-psk"]' -r | base64 -d) \
CM_WEB_PORT=8080 \
CM_DB_USER=insights \
CM_DB_PASS=insights \
CM_DB_NAME=insights \
go run . run
```

### Deployed into the cluster

Running config-manager in the cluster more accurately simulates the environment
in which config-manager runs in the production and stage environments. It more
seamlessly interacts with other applications (such as cloud-connector and
playbook-dispatcher), but makes for a much slower development cycle. Every time
a code change is made, a new image has to be built, pushed and deployed.
Fortunately, there is a convenient script to do this for you. See
`scripts/README.md` for details.

```sh
bash scripts/bonfire-deploy.sh
```

### Forward the port

The included `kube-port-forward.sh` script will forward ports from localhost to
all the necessary services in the cluster, including the config-manager-service
if detected.

```sh
bash scripts/kube-port-forward.sh
```

## Send HTTP requests

It should now be possible to interact with config-manager's HTTP API using
`curl` or `ht`.

```sh
ht GET http://localhost:8080/api/config-manager/v1/states/current x-rh-identity:$(xrhidgen user | base64 -w0)
```

# Debugging

## Monitoring Kafka topics

```sh
# Identify the environment name and export it
export CONFIG_MANAGER_ENV=$(kubectl -n fog get svc -l env=env-fog,app.kubernetes.io/name=kafka -o json | jq '.items[0].metadata.labels["app.kubernetes.io/instance"]' -r)
minikube kubectl -- -n fog run -it --rm --image=edenhill/kcat:1.7.1 kcat -- -b $CONFIG_MANAGER_ENV-kafka-bootstrap.fog.svc.cluster.local:9092 -t platform.inventory.events
```

## Produce an Inventory Event Kafka message

```sh
# Identify the environment name and export it
export CONFIG_MANAGER_ENV=$(kubectl -n fog get svc -l env=env-fog,app.kubernetes.io/name=kafka -o json | jq '.items[0].metadata.labels["app.kubernetes.io/instance"]' -r)
jq --compact-output --null-input --arg id $(uuidgen | tr -d "\n") '{"type":"created","host":{"id":$id,"account":"0000001","reporter":"cloud-connector","system_profile":{"rhc_client_id":$id}}}' | minikube kubectl -- -n fog run -i --rm --image=edenhill/kcat:1.7.1 $(mktemp XXXXXX) -- -b $CONFIG_MANAGER_ENV-kafka-bootstrap.fog.svc.cluster.local:9092 -t platform.inventory.events -P -H event_type=created 
```
