mqtt_msg='{ "type": "connection-status", "message_id": "3a57b1ad-5163-47ee-9e57-3bb6d90bdfff", "version": 1, "sent": "2023-12-04T17:22:24+00:00", "content": { "canonical_facts": { "insights_id": "a319580a-2321-47c7-b061-886548bea067", "machine_id": "044afe91-3b12-4439-9e8f-f531462d26b9", "bios_uuid": "ec2c79ab-02ff-d09b-31a9-583a902dd74b", "subscription_manager_id": "4b1efbbe-a447-48d0-98c3-3594aae1d2c5", "ip_addresses": ["172.31.28.69"], "mac_addresses": ["0a:94:1d:36:ff:21", "00:00:00:00:00:00"], "fqdn": "ip-172-31-28-69.ec2.internal" }, "dispatchers": { "playbook": { "ansible-runner-version": "1.2.3" }, "package-manager": null, "rhc-worker-playbook":null }, "state": "online" } }'
send_mqtt_msg:
	mqtt pub -V 3 -t redhat/insights/4b1efbbe-a447-48d0-98c3-3594aae1d2c5/control/out -h 127.0.0.1 -p 8883 -d -v -m ${mqtt_msg}


LOCALBIN ?= $(shell pwd)/bin
$(LOCALBIN):
ifeq (,$(wildcard $(LOCALBIN)))
	@echo "🤖 Ensuring $(LOCALBIN) is available"
	mkdir -p $(LOCALBIN)
	@echo "✅ Done"
endif

.PHONY: golangci-lint
GOLANGCILINT := $(LOCALBIN)/golangci-lint
GOLANGCI_URL := https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh
golangci-lint: $(LOCALBIN)
ifeq (,$(wildcard $(GOLANGCILINT)))
	@ echo "📥 Downloading golangci-lint"
	curl -sSfL $(GOLANGCI_URL) | sh -s -- -b $(LOCALBIN) $(GOLANGCI_VERSION)
	@ echo "✅ Done"
endif

.PHONY: lint
lint: golangci-lint
	$(GOLANGCILINT) run --timeout=3m ./...

.PHONY: test
test:
	go test -v ./...

start-httpapi:
	go run main.go \
		--log-level trace \
		--metrics-port 9007 \
		--inventory-host http://127.0.0.1:8001 \
		http-api

start-inventory-consumer:
	go run main.go \
		--log-level=trace \
		--kafka-brokers=localhost:29092 \
		--cloud-connector-host=http://127.0.0.1:8084/api/cloud-connector/ \
		--cloud-connector-client-id=suraj \
		--cloud-connector-psk=surajskey \
		--dispatcher-psk=surajskey \
		--playbook-host=https://random-host \
		--metrics-port=9008 \
		--dispatcher-host=http://127.0.0.1:8002 \
		inventory-consumer


configure-xjoin:
	@./scripts/xjoin-config/configure-xjoin.sh

get_host_from_inventory: 
	curl http://127.0.0.1:8001/api/inventory/v1/hosts -H "x-rh-identity: eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMTExMDAwIiwgIm9yZ19pZCI6ICIwMDAyIiwgImF1dGhfdHlwZSI6ICJqd3QtYXV0aCIsICJ0eXBlIjogIlVzZXIiLCJ1c2VyIjogeyJ1c2VybmFtZSI6ICJ0dXNlckByZWRoYXQuY29tIiwiZW1haWwiOiAidHVzZXJAcmVkaGF0LmNvbSIsImZpcnN0X25hbWUiOiAidGVzdCIsImxhc3RfbmFtZSI6ICJ1c2VyIiwiaXNfYWN0aXZlIjogdHJ1ZSwiaXNfb3JnX2FkbWluIjogZmFsc2UsICJsb2NhbGUiOiAiZW5fVVMifX19Cg==" | jq