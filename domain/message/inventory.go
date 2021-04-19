package message

import "config-manager/domain"

type InventoryEvent struct {
	Type string      `json:"type"`
	Host domain.Host `json:"host"`
}

type InventoryUpdate struct {
	Operation string         `json:"operation"`
	Data      HostUpdateData `json:"data"`
}

type HostUpdateData struct {
	ID            string                  `json:"id"`
	Account       string                  `json:"account"`
	SystemProfile HostUpdateSystemProfile `json:"system_profile"`
}

type HostUpdateSystemProfile struct {
	RHCState string `json:"rhc_config_state"`
}
