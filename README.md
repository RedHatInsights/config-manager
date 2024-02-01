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

See the [OpenAPI Schema](./internal/http/v1/openapi.yaml) for details on interacting with the REST interface.

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

See the
[CONTRIBUTING](https://github.com/RedHatInsights/config-manager/blob/master/CONTRIBUTING.md)
guide for details on getting started contributing to config-manager.
