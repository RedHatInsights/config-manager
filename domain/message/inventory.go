package message

import "config-manager/domain"

type InventoryEvent struct {
	Type string      `json:"type"`
	Host domain.Host `json:"host"`
}

type InventoryUpdate struct {
	Operation string           `json:"operation"`
	Metadata  PlatformMetadata `json:"platform_metadata"`
	Data      HostUpdateData   `json:"data"`
}

type PlatformMetadata struct {
	RequestID string `json:"request_id"`
}

type HostUpdateData struct {
	ID            string                  `json:"id"`
	Account       string                  `json:"account"`
	SystemProfile HostUpdateSystemProfile `json:"system_profile"`
}

type HostUpdateSystemProfile struct {
	RHCState string `json:"rhc_config_state"`
}
