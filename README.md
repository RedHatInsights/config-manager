# Config Manager

Config Manager is a backend service used by the Service Enablement dashboard to enable or disable various Red Hat services on hosts connected through RHC.

Config Manager handles the following actions:

- Retrieves the current configuration (enabled/disabled) of various services supported by Service Enablement 
- Updates the current configuration of services
- Maintains a history of configuration changes
- Ensures that newly connected hosts are kept up to date with the latest configuration
- Updates the host record in Inventory with the latest "rhc_config_state" ID

Config Manager has two mechanisms to update a host:

- Via a change in the Service Enablement dashboard
- Via a new rhc connection event from Inventory

Updating a host (all hosts) via the API:

![Sequence diagram](./docs/config_manager_api.svg)

Automatic updating of a newly connected host via kafka:

Todo

## REST interface

The REST interface can be used to view and update the current configuration for all hosts connected through RHC. It can also be used to view a history of previous configuration changes, and obtain logs related to those changes. 

See the [OpenAPI Schema](./schema/api.spec.yaml) for details on interacting with the REST interface.

## Development

### Dependencies

- Golang >= 1.13

### Deploying locally

Config-manager is managed by [Clowder](https://github.com/RedHatInsights/clowder) and can be deployed locally onto a [Minikube](https://minikube.sigs.k8s.io/docs/start/) instance using [Bonfire](https://github.com/RedHatInsights/bonfire).

The following steps (detailed [here](https://clouddot.pages.redhat.com/docs/dev/getting-started/backend-local.html)) should be performed before attempting to build and deploy a new instance of config-manager:

1. Install minikube
```sh
$ curl -LO https://storage.googleapis.com/minikube/releases/latest/minikube-latest.x86_64.rpm
$ sudo rpm -ivh minikube-latest.x86_64.rpm
```

2. Start minikube (check [here](https://github.com/RedHatInsights/clowder/blob/master/docs/macos.md) for MacOS instructions)
```sh
minikube start --cpus 4 --disk-size 36GB --memory 8000MB --addons=registry --driver=kvm2
```

3. Install Clowder CRDs
```sh
curl  https://raw.githubusercontent.com/RedHatInsights/clowder/master/build/kube_setup.sh -o kube_setup.sh  && chmod +x kube_setup.sh
./kube_setup.sh
```

4. Install Clowder (replace version with [latest](https://github.com/RedHatInsights/clowder/releases/latest))
```sh
minikube kubectl -- apply -f https://github.com/RedHatInsights/clowder/releases/download/0.15.0/clowder-manifest-0.15.0.yaml --validate=false
```

5. Create a namespace for config-manager
```sh
kubectl create ns config-manager
```

6. Deploy config-manager
```sh
./bonfire_deploy.sh
```

7. Access config-manager
```sh
oc port-forward svc/config-manager-service -n config-manager 8000 &
curl -v -H "x-rh-identity:eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" http://localhost:8000/api/config-manager/v1/states/current
```

Once prerequisite steps 1-5 have been completed steps 6-7 can be repeated as needed to deploy new changes to the local environment. 
