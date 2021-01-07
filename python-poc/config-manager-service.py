from twisted.web import server, resource
from twisted.internet import reactor, endpoints

from afkak.client import KafkaClient
from afkak.consumer import Consumer as KafkaConsumer
from afkak.producer import Producer as KafkaProducer
import afkak

import html
import json
import requests


kafka_producer = None
INVENTORY_EVENT_TOPIC = "platform.inventory.events"
OUTPUT_RECEIVED_EVENT_TOPIC = "platform.playbook_dispatcher.events"
MY_HOST_AND_PORT="localhost:8080"
CONNECTOR_SERVICE_URL="http://localhost:8081/job"
AUTH_HEADER_MAP = { "x-rh-identity":
    "eyJpZGVudGl0eSI6IHsiYWNjb3VudF9udW1iZXIiOiAiMDAwMDAwMSIsICJpbnRlcm5hbCI6IHsib3JnX2lkIjogIjAwMDAwMSJ9fX0=" 
}

def _get_account_from_request(request):
    return request.args[b'account'][0].decode()


class ConfigManagerResource(resource.Resource):
    pass


class StateResource(resource.Resource):
    isLeaf = True

    # FIXME: How do we add new apps?
    #          We'll need a new playbook
    #        What about existing clients?
    #        Does adding a new app trigger a sync across all connected clients?
    #        Or default new app to off and let customers enable it??
    default_state = { "insights": "enabled",
                      "compliance": "enabled",
                      "vulnerability": "disabled",
                      "drift": "enabled"}

    def __init__(self, config_storage):
        self.config_storage = config_storage

    def render_GET(self, request):
        account = _get_account_from_request(request)
        json_doc = self._get_latest_state(account)
        return json_doc.encode()

    def _get_default_state(self):
        return self.default_state

    def _get_latest_state(self, account):
        print("Looking up latest requested state for account ", account)

        # FIXME: What is the default state?

        latest_state = self._get_default_state()

        if account not in self.config_storage:
            self.config_storage[account] = latest_state
        else:
            latest_state = self.config_storage[account]

        return json.dumps(latest_state)

    def _update_latest_state(self, account, new_requested_state):
        print("Updating latest requested state for account ", account)
        # Store the state
        # FIXME: Is this a merge or a full replace?  POST is full replace??
        self.config_storage[account] = new_requested_state

    def render_POST(self, request):
        # Does this trigger a push?
        # I don't think so.  We probably need a way to update 
        # the latest requested state and then push
        account = _get_account_from_request(request)
        newdata = request.content.getvalue()
        new_requested_state = json.loads(newdata)

        self._update_latest_state(account, new_requested_state)

        return "".encode()


class SyncResultsResource(resource.Resource):
    isLeaf = True

    def __init__(self, output_storage):
        self.output_storage = output_storage

    def render_GET(self, request):
        account = _get_account_from_request(request)
        run_id = self._get_run_id_from_request(request)
        json_doc = self._get_high_level_output(account, run_id)
        return json_doc.encode()

    def _get_high_level_output(self, account, run_id):
        print("Looking up sync results for account ", account, " run id ", run_id)
        return json.dumps(self.output_storage[account][run_id])

    def _get_run_id_from_request(self, request):
        return request.args[b'run_id'][0].decode()


class PerformSyncResource(resource.Resource):
    isLeaf = True

    runId = 0

    def __init__(self, config_storage, sync_storage, output_storage, message_id_to_run_id_map):
        self.config_storage = config_storage
        self.sync_storage = sync_storage
        self.output_storage = output_storage
        self.message_id_to_run_id_map = message_id_to_run_id_map

    def render_POST(self, request):
        account = _get_account_from_request(request)
        print("Starting sync job for account ", account)
        run_id = self._generate_run_id()
        requested_state = self.config_storage[account]
        connected_hosts = self._get_connected_hosts_per_account(account)

        if account not in self.sync_storage:
            self.sync_storage[account] = {}

        # Record account, run_id, requested state, hosts
        self.sync_storage[account][run_id] = { "requested_state": requested_state,
                                               "connected_hosts": connected_hosts }

        inventory_host_id_list = {host[0]: {} for host in connected_hosts}
        connected_client_id_list = [host[1] for host in connected_hosts]

        # FIXME:  Store the inventory host id or connected client id here??
        self.output_storage[account][run_id] = inventory_host_id_list

        job_message = self._generate_job_request(account)

        connector_message_ids = self._send_jobs_to_connector_service(account, job_message, connected_client_id_list)

        # FIXME: Got to be able to go from a connector messageid, to the account, run_id, host_id
        for i, msg_id in enumerate(connector_message_ids):
            self.message_id_to_run_id_map[msg_id]=(account, run_id, connected_hosts[i][0])

        # ------------------------------- 
        # TESTING HACK!!
        #
        # Trigger some output events in the future
        ansible_output = { "insights": "success",
                           "compliance": "success",
                           "drift": "success" }
        #reactor.callLater(5, send_output_received_events, account, connector_message_ids[0], ansible_output)
        #reactor.callLater(10, send_output_received_events, account, connector_message_ids[1], ansible_output)
        # ------------------------------- 

        request.setResponseCode(201)

        json_doc = json.dumps({'id': run_id})
        return json_doc.encode()

    def _generate_run_id(self):
        self.runId += 1
        return str(self.runId)

    def _generate_job_request(self, account):
        # FIXME:  These urls do not really need to include the account
        job_request = { "payload_url": f"http://{MY_HOST_AND_PORT}/job?account={account}",
                        "return_url": f"http://{MY_HOST_AND_PORT}/playbook_dispatcher?account={account}",
                        "handler": "playbook_runner"}
        return json.dumps(job_request)

    def _get_connected_hosts_per_account(self, account):
        return [ ("inv_id_host_1", "client-0"),
                 ("inv_id_host_2", "client-1") ]

    def _send_jobs_to_connector_service(self, account, job_message, connected_client_id_list):
        print("Sending job message for each connected client to the connector-service...")
        print("job_message: ", job_message)

        connector_message_ids = []
        for connected_client_id in connected_client_id_list:
            connector_message = {"account": account,
                                 "recipient": connected_client_id,
                                 "directive": "NOT_USED",
                                 "payload": job_message }
            print("Calling connector service...")
            # FIXME:  This is bad for twisted
            response = requests.post(CONNECTOR_SERVICE_URL, headers=AUTH_HEADER_MAP, json=connector_message)
            print("connector service response: ", response)
            if response.status_code == 201:
                print("type(response.json()):", type(response.json()))
                connector_message_ids.append(response.json()["id"])
            else:
                # FIXME:
                connector_message_ids.append("BAD NODE!")

        # FIXME: 
        #connector_message_ids = ["123", "456"]
        return connector_message_ids


class JobResource(resource.Resource):
    isLeaf = True

    def __init__(self, config_storage):
        self.config_storage = config_storage

    def render_GET(self, request):
        # FIXME:  Do we only ever return the latest requested state?
        #         For example, user1 kicks off a sync, then user2 changes the state??
        account = _get_account_from_request(request)
        latest_state = self._lookup_latest_state(account)
        json_doc = self._generate_playbook(latest_state)
        return json_doc.encode()

    def _get_run_id_from_request(self, request):
        return request.args[b'run_id'][0].decode()

    def _lookup_latest_state(self, account):
        latest_state = self.config_storage[account]
        return latest_state

    def _generate_playbook(self, requested_state):
        print("Building playbook...")
        print("\there is the requested state:", requested_state)
        # FIXME: Retrieve the playbooks
        #        Need a playbook for enabling and disabling each service
        #        Are the playbooks specific to the rhel version?
        # 
        # GET THE LATEST PLAYBOOK?  or is the playbook specific to a particular run??
        return "ima playbook"

class PlaybookDispatcherResource(resource.Resource):
    isLeaf = True

    def render_POST(self, request):
        print("Recieved playbook run output...")
        account = _get_account_from_request(request)
        connector_message_id = self._get_message_id_from_request(request)
        output = request.content.getvalue().decode()
        print(f"\toutput for message id {connector_message_id} {output}")

        send_output_received_events(account, connector_message_id, output)

        # new_requested_state = json.loads(output)
        return "".encode()

    def _get_message_id_from_request(self, request):
        return request.getHeader("message_id")


# FIXME: If we consume inventory events, then we have to 
# process _ALL_ the events from inventory.  Consider only 
# getting "new connection" / "connection dropped" events 
# from connector-service
class InventoryEventProcessor:

    def __call__(self, consumer, message_list):
        print("Got inventory message...")
        for outter_message in message_list:
            key = outter_message.message.key
            msg = outter_message.message.value
            print("key:", key)
            print("msg:", msg)
            # FIXME:  WHAT NOW??
            print("FIXME:  WHAT NOW??  How do I know if this is a new connectin?")


class OutputReceivedEventProcessor:

    def __init__(self, output_storage, message_id_to_run_id_map):
        self.output_storage = output_storage
        self.message_id_to_run_id_map = message_id_to_run_id_map

    def __call__(self, consumer, message_list):
        print("Got output received message...")
        for outter_message in message_list:
            key = outter_message.message.key.decode()
            msg = json.loads(outter_message.message.value)
            print("\tmsg:", msg)

            # FIXME: Got to be able to go from a connector messageid, to the account, run_id, host_id
            (account, run_id, host_id) = self._get_message_id_details(key)
            print("account:", account)

            if account != msg["account"]:
                print("Error ...invalid data")
                return

            try:
                self.output_storage[account][run_id][host_id] = msg["ansible_output"]
            except KeyError as ke:
                print("KeyError while processing output response:", ke)

    def _get_message_id_details(self, key):
        print("message_id_to_run_id_map:", self.message_id_to_run_id_map)
        print("key:", key)
        return self.message_id_to_run_id_map[key]


def send_output_received_events(account, connector_message_id, output):
    print("send_output_received_events was called:", account, connector_message_id)
    print("\ttype(output):", type(output))
    msg = { "account": account,
            "message_id": connector_message_id,
            "ansible_output": output,
    }
    json_msg = json.dumps(msg)
    #json_msg = '{"compliance": "success", "drift": "failure", "insights": "success"}'
    kafka_producer.send_messages(
            OUTPUT_RECEIVED_EVENT_TOPIC,
            key=connector_message_id.encode(),
            msgs=[json_msg.encode()])


def start_kafka_consumer(kafka_client=None, topic=None, consumer_group=None, processor=None):
    partition = 0
    kafka_consumer = KafkaConsumer(kafka_client,
                              topic,
                              partition,
                              processor,
                              consumer_group=consumer_group)
    kafka_consumer.start(afkak.OFFSET_LATEST)


def send_inventory_events(host_updated_event):
    print("send_inventory_events was called...")
    host_id = host_updated_event["id"]
    print("inventory host id: ", host_id)
    print("host update event: ", host_updated_event)
    json_doc = json.dumps(host_updated_event)
    kafka_producer.send_messages(
            INVENTORY_EVENT_TOPIC,
            key=host_id.encode(),
            msgs=[json_doc.encode()])


def start_kafka_producer(kafka_client=None):
    kafka_client = KafkaClient("localhost:29092", reactor=reactor)
    global kafka_producer
    kafka_producer = KafkaProducer(kafka_client)
    print("kafka_producer: ", kafka_producer)
    return kafka_producer


config_storage = {}
output_storage = { "010101":
        { "1234":
            { "inv_id_host_123": {"insights": "failure",
                           "compliance": "success",
                                "drift": "success"},
               "inv_id_host_321": {"insights": "success",
                            "compliance": "success",
                            "drift": "success"},
               "inv_id_host_x1": {"no response received": "N/A"},  # FIXME:  no response was received
               "inv_id_host_x2": {"disconnected at connector service":"N/A"},  # FIXME:
        }
    }
}
sync_storage = {}
message_id_to_run_id_map = {}

root = ConfigManagerResource()
root.putChild('state'.encode(), StateResource(config_storage))
root.putChild('sync_results'.encode(), SyncResultsResource(output_storage))
root.putChild('sync'.encode(), PerformSyncResource(config_storage, sync_storage, output_storage, message_id_to_run_id_map))
root.putChild('job'.encode(), JobResource(config_storage))
root.putChild('playbook_dispatcher'.encode(), PlaybookDispatcherResource())

site = server.Site(root)
endpoint = endpoints.TCP4ServerEndpoint(reactor, 8080)
endpoint.listen(site)

kafka_client = KafkaClient("localhost:29092", reactor=reactor)

# FIXME: Move this into its own pod
start_kafka_consumer(kafka_client=kafka_client,
                     topic=INVENTORY_EVENT_TOPIC,
                     consumer_group="config-manager-inventory-consumer",
                     processor = InventoryEventProcessor())

# FIXME: Move this into its own pod
start_kafka_consumer(kafka_client=kafka_client,
                     topic=OUTPUT_RECEIVED_EVENT_TOPIC,
                     consumer_group="config-manager-output-received-consumer",
                     processor = OutputReceivedEventProcessor(output_storage, message_id_to_run_id_map))

kakfa_producer = start_kafka_producer(kafka_client=kafka_client)

# FIXME: Send some inventory events from outta the 
# blue...these might need to be connector-service events
host_updated_event = {"account": "010101", "id": "6785", "insights_id": "3234", "connected_client_id": "2352"}
#reactor.callLater(10, send_inventory_events, host_updated_event)

reactor.run()
