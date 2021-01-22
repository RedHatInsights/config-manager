package persistence

import (
	"config-manager/domain"
	"fmt"
)

type ClientListRepository struct {
	InventoryURL string // placeholder
}

func (r *ClientListRepository) GetConnectedClients(accountID string) (*domain.ClientList, error) {
	// placeholder - http request clients from external service (inventory)
	fmt.Println("Getting connected clients from somewhere..")
	clientList := &domain.ClientList{AccountID: accountID}
	var clients []string
	clients = append(clients, "1234", "5678")
	clientList.Clients = clients
	return clientList, nil
}
