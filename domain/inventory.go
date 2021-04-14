package domain

import "github.com/labstack/echo/v4"

type InventoryParams struct {
	Page    int `json:"page"`
	PerPage int `json:"per_page"`
}

type InventoryResponse struct {
	Total   int    `json:"total"`
	Count   int    `json:"count"`
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Results []Host `json:"results"`
}

type Host struct {
	ID            string        `json:"id"`
	Account       string        `json:"account"`
	DisplayName   string        `json:"display_name"`
	Reporter      string        `json:"reporter"`
	SystemProfile SystemProfile `json:"system_profile"`
}

type SystemProfile struct {
	RHCID    string `json:"rhc_client_id"`
	RHCState string `json:"rhc_config_state"`
}

type InventoryClient interface {
	GetInventoryClients(ctx echo.Context, page int) (InventoryResponse, error)
}
