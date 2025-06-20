package internal

// Host represents a system record from the Inventory application.
type Host struct {
	ID                    string                 `json:"id"`
	Account               string                 `json:"account"`
	OrgID                 string                 `json:"org_id"`
	DisplayName           string                 `json:"display_name"`
	Reporter              string                 `json:"reporter"`
	PerReporterStaleness  map[string]interface{} `json:"per_reporter_staleness"`
	SubscriptionManagerID string                 `json:"subscription_manager_id"`
	SystemProfile         struct {
		RHCID    string `json:"rhc_client_id"`
		RHCState string `json:"rhc_config_state"`
	} `json:"system_profile"`
}
