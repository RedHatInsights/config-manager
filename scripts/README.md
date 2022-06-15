# scripts

This directory contains scripts useful to developing config-manager.

## `kube-port-forward.sh`

`kube-port-forward.sh` is a simple process management script around the `kubectl
port-forward` command. It allows all the ports necessary for local development
to be forwarded to the needed services.

## `bonfire-deploy.sh`

`bonfire-deploy.sh` builds a container image using `Dockerfile`, pushes it to a
register that is assumed to exist inside a minikube cluster, and then deploys it
using bonfire, overriding image tags and references using a combination of
`bonfire.yml` and command-line parameters.
