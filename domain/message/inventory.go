package message

import "config-manager/domain"

// InventoryEvent represents a message read from the inventory.events topic.
type InventoryEvent struct {
	Type string      `json:"type"`
	Host domain.Host `json:"host"`
}

// InventoryUpdate represents a message written to the inventory.system-profile
// topic.
type InventoryUpdate struct {
	Operation string           `json:"operation"`
	Metadata  PlatformMetadata `json:"platform_metadata"`
	Data      HostUpdateData   `json:"data"`
}

// PlatformMetadata represents the platform_metadata field of the
// InventoryUpdate.
type PlatformMetadata struct {
	RequestID string `json:"request_id"`
}

// HostUpdateData represents the data field of the InventoryUpdate.
type HostUpdateData struct {
	ID            string                  `json:"id"`
	Account       string                  `json:"account"`
	SystemProfile HostUpdateSystemProfile `json:"system_profile"`
}

// HostUpdateSystemProfile represents the system_profile field of the
// HostUpdateData.
type HostUpdateSystemProfile struct {
	RHCState string `json:"rhc_config_state"`
}
