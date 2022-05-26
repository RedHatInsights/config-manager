# scripts

This directory contains scripts useful to developing config-manager.

## `kube-port-forward.sh`

`kube-port-forward.sh` is a simple process management script around the `kubectl
port-forward` command. It allows all the ports necessary for local development
to be forwarded to the needed services.
