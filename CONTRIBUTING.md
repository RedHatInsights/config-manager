# Getting Started

If you're not logged into quay.io already, make sure you follow the
[instructions for generating a quay pull
secret](https://consoledot.pages.redhat.com/docs/dev/getting-started/local/environment.html#_get_your_quay_pull_secret),
but instead of downloading a pull secret file, copy the "docker login"
insructions.
The file `podman-compose.yml` has been written to start all services that
config-manager requires in containers. `config-manager` itself is intentionally
excluded. It is expected that you will run config-manager directly, to make
attaching a debugger easier.

The services started and managed by podman-compose are:

* kafka & zookeeper
* cloud-connector
* mosquitto
* postgres
* inventory
* playbook-dispatcher

Running `podman-compose up --detach` will start all the services above, and
export ports on the local system so the services can be interacted with.

| service             | port  |
| ------------------- | ----- |
| kafka               | 29092 |
| cloud-connector-api | 8081  |
| mosquitto           | 1883  |
| postgres            | 5432  |
| inventory           | 8888  |
| playbook-dispatcher | 8000  |

Next, run  config-manager:

```
LOG_LEVEL=debug \
CM_DISPATCHER_HOST=http://localhost:8000/api/playbook-dispatcher/v1/ \
CM_DISPATCHER_PSK=swordfish \
CM_INVENTORY_HOST=http://localhost:8888/api/inventory/v1/ \
CM_CLOUD_CONNECTOR_HOST=http://localhost:8081/api/cloud-connector/v1/ \
CM_CLOUD_CONNECTOR_PSK=swordfish \
CM_WEB_PORT=8080 \
go run . run
```

Now you can interact with config-manager on port 8080:

```
curl -v -H "x-rh-identity:eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" http://localhost:8080/api/config-manager/v1/states/current
```

# Simulate a Kafka Inventory event

Start a kafka consumer if you want to monitor traffic on a topic:

```
kcat -C -b localhost:29092 -t platform.inventory.events
```

Then produce a message on the topic:

```
jq --compact-output --null-input --arg id $(uuidgen | tr -d "\n") '{"type":"created","host":{"id":$id,"account":"0000001","reporter":"cloud-connector","system_profile":{"rhc_client_id":$id}}}' | kcat -b localhost:29092 -P -t platform.inventory.events -H event_type=created
```
