package domain

import "context"

// InventoryParams represents query parameters to send with HTTP requests.
type InventoryParams struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

// InventoryResponse represents a list of hosts received from the Inventory application.
type InventoryResponse struct {
	Total   int    `json:"total"`
	Count   int    `json:"count"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Results []Host `json:"results"`
}

// Host represents a system record from the Inventory application.
type Host struct {
	ID            string        `json:"id"`
	Account       string        `json:"account"`
	DisplayName   string        `json:"display_name"`
	Reporter      string        `json:"reporter"`
	SystemProfile SystemProfile `json:"system_profile"`
}

// SystemProfile represents the system_profile field of the Host structure.
type SystemProfile struct {
	RHCID    string `json:"rhc_client_id"`
	RHCState string `json:"rhc_config_state"`
}

// InventoryClient is an abstraction of the REST client API methods to interact
// with the platform Inventory application.
type InventoryClient interface {
	GetInventoryClients(ctx context.Context, page int) (InventoryResponse, error)
}
