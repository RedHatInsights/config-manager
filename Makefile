host_create_msg='{ "operation": "add_host", "data": { "insights_id": "a319580a-2321-47c7-b061-886548bea067", "machine_id": "044afe91-3b12-4439-9e8f-f531462d26b9", "bios_uuid": "ec2c79ab-02ff-d09b-31a9-583a902dd74b", "subscription_manager_id": "4b1efbbe-a447-48d0-98c3-3594aae1d2c5", "ip_addresses": [ "172.31.28.69" ], "mac_addresses": [ "0a:94:1d:36:ff:21", "00:00:00:00:00:00" ], "fqdn": "ip-172-31-28-69.ec2.internal", "provider_id": "i-0bc1b5c57d123ece9", "provider_type": "aws", "system_profile": { "is_marketplace": false, "arch": "x86_64", "bios_release_date": "08/24/2006", "bios_vendor": "Xen", "bios_version": "4.2.amazon", "cpu_flags": [ "abm" ], "cpu_model": "Intel(R) Xeon(R) CPU E5-2686 v4 @ 2.30GHz", "number_of_cpus": 1, "number_of_sockets": 1, "cores_per_socket": 1, "tuned_profile": "virtual-guest", "selinux_current_mode": "enforcing", "selinux_config_file": "enforcing", "enabled_services": [ "user" ], "installed_services": [ "user@" ], "infrastructure_type": "virtual", "infrastructure_vendor": "xen", "installed_packages": [ "zlib-0:1.2.11-17.el8.x86_64" ], "gpg_pubkeys": [ "gpg-pubkey-d4082792-5b32db75", "gpg-pubkey-fd431d51-4ae0493b" ], "kernel_modules": [ "xfs" ], "captured_date": "2021-11-18T06:22:44+00:00", "last_boot_time": "2021-11-10T08:35:44+00:00", "network_interfaces": [ { "ipv4_addresses": [ "172.31.28.69" ], "ipv6_addresses": [ "fe80::894:1dff:fe36:ff21" ], "mac_address": "0a:94:1d:36:ff:21", "mtu": 9001, "name": "eth0", "state": "UP", "type": "ether" }, { "ipv4_addresses": [ "127.0.0.1" ], "ipv6_addresses": [ "::1" ], "mac_address": "00:00:00:00:00:00", "mtu": 65536, "name": "lo", "state": "UNKNOWN", "type": "loopback" } ], "os_kernel_version": "4.18.0", "os_kernel_release": "305.el8", "system_update_method": "dnf", "os_release": "8.4", "operating_system": { "major": 8, "minor": 4, "name": "RHEL" }, "rhsm": { "version": "" }, "running_processes": [ "(sd-pam)" ], "system_memory_bytes": 845565952, "yum_repos": [ { "id": "3scale-amp-2-for-rhel-8-ppc64le-debug-rpms", "name": "Red Hat 3scale API Management Platform 2 for RHEL 8 for Power (Debug RPMs)", "base_url": "https://cdn.redhat.com/content/dist/layered/rhel8/ppc64le/3scale-amp/2/debug", "enabled": false, "gpgcheck": true } ], "dnf_modules": [ { "name": "virt", "stream": "rhel" } ], "cloud_provider": "aws", "installed_products": [ { "id": "408" }, { "id": "479" } ], "is_ros": true }, "tags": { "insights-client": {} }, "stale_timestamp": "2024-01-30T07:31:42.530745+00:00", "reporter": "puptoo", "account": "111000", "org_id": "0002" }, "platform_metadata": { "account": "111000", "category": "collection", "content_type": "application/vnd.redhat.advisor.collection+tgz", "metadata": { "reporter": "", "stale_timestamp": "0001-01-01T00:00:00Z" }, "request_id": "4c114785f7b7/T6DDSKHvSW-000001", "principal": "000001", "org_id": "0002", "service": "advisor", "size": 2549760, "url": "http://minio:9000/insights-upload-perma/4c114785f7b7/T6DDSKHvSW-000001?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=BQA2GEXO711FVBVXDWKM%2F20240129%2Fus-east-1%2Fs3%2Faws4_request&X-Amz-Date=20240129T023136Z&X-Amz-Expires=86400&X-Amz-SignedHeaders=host&X-Amz-Signature=c2dbb4e038987f6536d91a50d6ebc8a33770e5f112ed586642f88a05f40963de", "b64_identity": "eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMTExMDAwIiwgIm9yZ19pZCI6ICIwMDAyIiwgImF1dGhfdHlwZSI6ICJqd3QtYXV0aCIsICJ0eXBlIjogIlVzZXIiLCJ1c2VyIjogeyJ1c2VybmFtZSI6ICJ0dXNlckByZWRoYXQuY29tIiwiZW1haWwiOiAidHVzZXJAcmVkaGF0LmNvbSIsImZpcnN0X25hbWUiOiAidGVzdCIsImxhc3RfbmFtZSI6ICJ1c2VyIiwiaXNfYWN0aXZlIjogdHJ1ZSwiaXNfb3JnX2FkbWluIjogZmFsc2UsICJsb2NhbGUiOiAiZW5fVVMifX19Cg==", "timestamp": "2024-01-29T02:31:36.689823844Z", "elapsed_time": 1706495496.7537477, "is_ros": true } }'
register_host_inventory:
	echo ${host_create_msg} | docker-compose -f scripts/docker-compose.yml exec -T kafka kafka-console-producer --topic platform.inventory.host-ingress --broker-list localhost:29092

mqtt_msg='{ "type": "connection-status", "message_id": "3a57b1ad-5163-47ee-9e57-3bb6d90bdfff", "version": 1, "sent": "2023-12-04T17:22:24+00:00", "content": { "canonical_facts": { "insights_id": "a319580a-2321-47c7-b061-886548bea067", "machine_id": "044afe91-3b12-4439-9e8f-f531462d26b9", "bios_uuid": "ec2c79ab-02ff-d09b-31a9-583a902dd74b", "subscription_manager_id": "4b1efbbe-a447-48d0-98c3-3594aae1d2c5", "ip_addresses": ["172.31.28.69"], "mac_addresses": ["0a:94:1d:36:ff:21", "00:00:00:00:00:00"], "fqdn": "ip-172-31-28-69.ec2.internal" }, "dispatchers": { "playbook": { "ansible-runner-version": "1.2.3" }, "package-manager": null, "rhc-worker-playbook":null }, "state": "online" } }'
create_host: register_host_inventory
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
		--cloud-connector-host=http://127.0.0.1:8084/api/cloud-connector/v1/ \
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