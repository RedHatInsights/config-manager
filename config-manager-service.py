from twisted.web import server, resource
from twisted.internet import reactor, endpoints

from afkak.client import KafkaClient
from afkak.consumer import Consumer as KafkaConsumer
from afkak.producer import Producer as KafkaProducer
import afkak

import html
import json


kafka_producer = None
INVENTORY_EVENT_TOPIC = "platform.inventory.events"
OUTPUT_RECEIVED_EVENT_TOPIC = "platform.playbook_dispatcher.events"

def _get_account_from_request(request):
    return request.args[b'account'][0].decode()


class ConfigManager(resource.Resource):
    pass


class State(resource.Resource):
    isLeaf = True

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


class SyncResults(resource.Resource):
    isLeaf = True

    def __init__(self, output_storage):
        self.output_storage = output_storage

    def render_GET(self, request):
        account = _get_account_from_request(request)
        run_id = self._get_run_id_from_request(request)
        json_doc = self.getHighLevelOutput(account, run_id)
        return json_doc.encode()

    def getHighLevelOutput(self, account, run_id):
        print("Looking up sync results for account ", account, " run id ", run_id)
        return json.dumps(self.output_storage[account][run_id])

    def _get_run_id_from_request(self, request):
        return request.args[b'run_id'][0].decode()


class PerformSync(resource.Resource):
    isLeaf = True

    runId = 0

    def __init__(self, config_storage, sync_storage, output_storage):
        self.config_storage = config_storage
        self.sync_storage = sync_storage
        self.output_storage = output_storage

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

        playbook = self._generate_playbook(requested_state)

        connector_message_ids = self._send_jobs_to_connector_service(playbook, connected_client_id_list)

        # ------------------------------- 
        # TESTING HACK!!
        def send_output_received_events(account, connected_hosts):
            print("send_output_received_events was called:", account, connected_hosts)
            kafka_producer.send_messages(
                    OUTPUT_RECEIVED_EVENT_TOPIC,
                    key=connector_message_ids[0].encode(),
                    msgs=[b"compliance finished"])

        reactor.callLater(2.5, send_output_received_events, account, connected_hosts)
        # ------------------------------- 

        request.setResponseCode(201)

        json_doc = json.dumps({'id': run_id})
        return json_doc.encode()

    def _generate_run_id(self):
        self.runId += 1
        return str(self.runId)

    def _get_connected_hosts_per_account(self, account):
        return [ ("inv_id_host_1", "client_id_1"),
                 ("inv_id_host_2", "client_id_2") ]

    def _generate_playbook(self, requested_state):
        print("Building playbook...")
        # FIXME: build playbook
        return "ima playbook"

    def _send_jobs_to_connector_service(self, playbook, connected_client_id_list):
        print("Sending job message for each connected client to the connector-service...")
        # FIXME: 
        connector_message_ids = ["123", "456", "789"]
        return connector_message_ids


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
            print("FIXME:  WHAT NOW??")


class OutputReceivedEventProcessor:

    def __call__(self, consumer, message_list):
        print("Got output received message...")
        for outter_message in message_list:
            key = outter_message.message.key
            msg = outter_message.message.value
            print("key:", key)
            print("msg:", msg)
            # FIXME:  WHAT NOW??
            print("FIXME:  WHAT NOW??")


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

root = ConfigManager()
root.putChild('state'.encode(), State(config_storage))
root.putChild('sync_results'.encode(), SyncResults(output_storage))
root.putChild('sync'.encode(), PerformSync(config_storage, sync_storage, output_storage))

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
                     processor = OutputReceivedEventProcessor())

kakfa_producer = start_kafka_producer(kafka_client=kafka_client)

# FIXME: Send some inventory events from outta the 
# blue...these might need to be connector-service events
host_updated_event = {"account": "010101", "id": "6785", "insights_id": "3234", "connected_client_id": "2352"}
reactor.callLater(10, send_inventory_events, host_updated_event)

reactor.run()
