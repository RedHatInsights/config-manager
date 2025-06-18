# Development environment

This is docker-compose based dev-setup which helps to deploy dependent services like cloud-connector and host-inventory locally. 

## Installation

Before starting the development make sure below packages/binaries are installed on the system. 

| Package/binary                 | Version  | Documentation                                                                  |
|--------------------------------|----------|--------------------------------------------------------------------------------|
| golang                         | >=1.23   | https://go.dev/doc/install                                                     |
| mqtt cli                       | >=4.24.0 | https://hivemq.github.io/mqtt-cli/docs/installation/                           |
| docker & docker-compose        | latest   | https://docs.docker.com/desktop/install/fedora/                                |
| librdkafka &  librdkafka-devel | latest   | Use package-manager to install it.  dnf install -y librdkafka librdkafka-devel |
| kcat (formaly kafkacat) | latest   | https://docs.confluent.io/platform/current/tools/kafkacat-usage.html |


## Clone the repository
```bash
git clone git@github.com:RedHatInsights/config-manager.git
```

Update /etc/hosts with below. 

```bash
127.0.0.1       kafka
127.0.0.1       minio
```

## Usage

### Run dependent services
Use below command to start kafka, cloud-connector and other dependent services. 

```bash
cd scripts
docker-compose up
```
Note - If you are unable to pull image from quay.io then try to login quay.io using docker login quay.io and run docker-compose up again.

After running `docker-compose up` run the below command give a minute for all the container to come up and then run below command and make sure there are **>=20** containers running. 
```bash
docker ps | wc -l 
```

Now we are ready to start config-manager locally. 

### Running inventory-consumer locally. 

Use the below make command to start inventory-consumer service. 

```bash
make start-inventory-consumer
```

### Sending data to local config-manager. 
```bash
make send_mqtt_msg
```
Above command sends host-registration request to host-inventory and also send the connection status to cloud-connector service. 

### Running config-manager API.

Use the below make command to start api server.

```
make configure-xjoin
make start-httpapi
```