package instrumentation

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog/log"
)

const (
	labelDb                 = "db"
	labelGetAccountState    = "get_account_state"
	labelUpdateAccountState = "update_account_state"
	labelGetStateChanges    = "get_state_changes"
	labelGetProfiles        = "get_profiles"
	labelGetProfile         = "get_profile"
	labelCreateProfile      = "create_profile"
	labelPassed             = "ok"
	labelFailed             = "failed"
	labelError              = "error"
)

var (
	internalErrorTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "config_manager_api_error_total",
		Help: "The total number of errors",
	}, []string{"type", "subtype"})

	requestVerificationErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_manager_api_request_payload_verification_error",
		Help: "The total number of errors verifying request payloads",
	})

	dispatcherErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_manager_api_playbook_dispatcher_error_total",
		Help: "The total number of errors talking to playbook dispatcher",
	})

	connectorErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_manager_api_cloud_connector_error_total",
		Help: "The total number of errors talking to cloud connector",
	})

	inventoryErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_manager_api_inventory_error_total",
		Help: "The total number of errors talking to inventory",
	})

	playbookRequestOKTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_manager_api_playbooks_requested_ok_total",
		Help: "The total number of playbooks returned via the api",
	})

	playbookRequestErrorTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "config_manager_api_playbooks_requested_error_total",
		Help: "The total number of errors when generating playbooks",
	})

	kesselRequestTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "config_manager_kessel_requests_total",
		Help: "The total number of Kessel requests",
	}, []string{"status"})
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

func GetProfilesError() {
	internalErrorTotal.WithLabelValues(labelDb, labelGetProfiles).Inc()
}

func GetProfileError() {
	internalErrorTotal.WithLabelValues(labelDb, labelGetProfile).Inc()
}

func CreateProfileError() {
	internalErrorTotal.WithLabelValues(labelDb, labelCreateProfile).Inc()
}

func GetPlaybookError() {
	playbookRequestErrorTotal.Inc()
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

func AuthorizationCheckPassed(principal, org, permission string) {
	kesselRequestTotal.WithLabelValues(labelPassed).Inc()
	log.Debug().Str("principal", principal).Str("org", org).Str("permission", permission).Msg("Authorization check passed")
}

func AuthorizationCheckFailed(principal, org, permission string) {
	kesselRequestTotal.WithLabelValues(labelFailed).Inc()
	log.Debug().Str("principal", principal).Str("org", org).Str("permission", permission).Msg("Authorization check failed")
}

func AuthorizationCheckError(err error) {
	kesselRequestTotal.WithLabelValues(labelError).Inc()
	log.Error().Err(err).Msg("Error performing authorization check")
}

func Start() {
	internalErrorTotal.WithLabelValues(labelDb, labelGetAccountState)
	internalErrorTotal.WithLabelValues(labelDb, labelUpdateAccountState)
	internalErrorTotal.WithLabelValues(labelDb, labelGetStateChanges)
}
