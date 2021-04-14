package message

import "config-manager/domain"

type InventoryEvent struct {
	Type string      `json:"type"`
	Host domain.Host `json:"host"`
}
