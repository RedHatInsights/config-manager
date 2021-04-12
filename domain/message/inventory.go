package message

import "config-manager/domain"

type InventoryEvent struct {
	Type string      `json:"type"`
	Host domain.Host `json:"host"`
}

// type InventoryHost struct {
// 	ID            string            `json:"id"`
// 	Account       string            `json:"account"`
// 	Reporter      string            `json:"reporter"`
// 	SystemProfile HostSystemProfile `json:"system_profile"`
// }

// type HostSystemProfile struct {
// 	RHCID    string `json:"rhc_client_id"`
// 	RHCState string `json:"rhc_config_state"`
// }
