from twisted.web import server, resource
from twisted.internet import reactor, endpoints

import html
import json


def getAccountFromRequest(request):
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
        print("request.args:", request.args)
        #print("request.__dict__:", request.__dict__)

        account = getAccountFromRequest(request)
        json_doc = self.getLatestState(account)
        return json_doc.encode()


    def getDefaultState(self):
        return self.default_state

    def getLatestState(self, account):
        print("Looking up latest requested state for account ", account)

        # What is the default state?

        latest_state = self.getDefaultState()

        if account not in self.config_storage:
            self.config_storage[account] = latest_state
        else:
            latest_state = self.config_storage[account]

        return json.dumps(latest_state)

    def render_POST(self, request):
        # Does this trigger a push?
        # I don't think so.  We probably need a way to update 
        # the latest requested state and then push
        account = getAccountFromRequest(request)
        print("account:", account)
        newdata = request.content.getvalue()
        print("newdata:", newdata)
        new_requested_state = json.loads(newdata)

        # Store the state
        # FIXME: Is this a merge or a full replace?  POST is full replace??
        self.config_storage[account] = new_requested_state

        return "".encode()


class SyncResults(resource.Resource):
    isLeaf = True

    def __init__(self, output_storage):
        self.output_storage = output_storage

    def render_GET(self, request):
        account = getAccountFromRequest(request)
        run_id = self.getRunIdFromRequest(request)
        json_doc = self.getHighLevelOutput(account, run_id)
        return json_doc.encode()

    def getHighLevelOutput(self, account, run_id):
        print("Looking up sync results for account ", account, " run id ", run_id)
        print(self.output_storage[account][run_id])
        print(type(self.output_storage[account][run_id]))
        return json.dumps(self.output_storage[account][run_id])

    def getRunIdFromRequest(self, request):
        return request.args[b'run_id'][0].decode()


class PerformSync(resource.Resource):
    isLeaf = True

    runId = 0

    def __init__(self, config_storage, sync_storage):
        self.config_storage = config_storage
        self.sync_storage = sync_storage

    def render_POST(self, request):
        account = getAccountFromRequest(request)
        run_id = self.generateRunId()
        requested_state = self.config_storage[account]
        connected_hosts = self.getConnectedHostsPerAccount(account)

        if account not in self.sync_storage:
            self.sync_storage[account] = {}

        # Record account, run_id, requested state, hosts
        self.sync_storage[account][run_id] = { "requested_state": requested_state,
                                               "connected_hosts": connected_hosts }

        json_doc = json.dumps({'id': run_id})
        return json_doc.encode()

    def generateRunId(self):
        self.runId += 1
        return str(self.runId)

    def getConnectedHostsPerAccount(self, account):
        return [ ("inv_id_host_1", "client_id_1"),
                 ("inv_id_host_2", "client_id_2") ]


config_storage = {}
output_storage = { "010101":
        { "1234":
            { "host_123": {"insights": "failure",
                           "compliance": "success",
                                "drift": "success"},
               "host_321": {"insights": "success",
                            "compliance": "success",
                            "drift": "success"},
               "host_x1": {"no response received": "N/A"},  # FIXME:  no response was received
               "host_x2": {"disconnected at connector service":"N/A"},  # FIXME:
        }
    }
}
sync_storage = {}

root = ConfigManager()
root.putChild('state'.encode(), State(config_storage))
root.putChild('sync_results'.encode(), SyncResults(output_storage))
root.putChild('sync'.encode(), PerformSync(config_storage, sync_storage))

site = server.Site(root)
endpoint = endpoints.TCP4ServerEndpoint(reactor, 8080)
endpoint.listen(site)
reactor.run()
