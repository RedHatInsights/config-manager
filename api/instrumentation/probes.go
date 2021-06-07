package instrumentation

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	labelDb                 = "db"
	labelGetAccountState    = "get_account_state"
	labelUpdateAccountState = "update_account_state"
	labelGetStateChanges    = "get_state_changes"
)

var (
	internalErrorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "api_error_total",
		Help: "The total number of errors",
	}, []string{"type", "subtype"})

	requestVerificationErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_request_payload_verification_error",
		Help: "The total number of errors verifying request payloads",
	})

	dispatcherErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_playbook_dispatcher_error_total",
		Help: "The total number of errors talking to playbook dispatcher",
	})

	connectorErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_cloud_connector_error_total",
		Help: "The total number of errors talking to cloud connector",
	})

	inventoryErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_inventory_error_total",
		Help: "The total number of errors talking to inventory",
	})

	playbookRequestOKTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_playbooks_requested_ok_total",
		Help: "The total number of playbooks returned via the api",
	})

	playbookRequestErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "api_playbooks_requested_error_total",
		Help: "The total number of errors when generating playbooks",
	})
)

func GetAccountStateError() {
	internalErrorTotal.WithLabelValues(labelDb, labelGetAccountState).Inc()
}

func UpdateAccountStateError() {
	internalErrorTotal.WithLabelValues(labelDb, labelUpdateAccountState).Inc()
}

func GetStateChangesError() {
	internalErrorTotal.WithLabelValues(labelDb, labelGetStateChanges).Inc()
}

func PayloadVerificationError() {
	requestVerificationErrorTotal.Inc()
}

func CloudConnectorRequestError() {
	connectorErrorTotal.Inc()
}

func PlaybookDispatcherRequestError() {
	dispatcherErrorTotal.Inc()
}

func InventoryRequestError() {
	inventoryErrorTotal.Inc()
}

func PlaybookRequestOK() {
	playbookRequestOKTotal.Inc()
}

func PlaybookRequestError() {
	playbookRequestErrorTotal.Inc()
}

func Start() {
	internalErrorTotal.WithLabelValues(labelDb, labelGetAccountState)
	internalErrorTotal.WithLabelValues(labelDb, labelUpdateAccountState)
	internalErrorTotal.WithLabelValues(labelDb, labelGetStateChanges)
}
