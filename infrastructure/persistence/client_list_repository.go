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
	var clients []domain.Client
	clients = append(clients,
		domain.Client{Hostname: "localhost", ClientID: "276c4685-fdfb-4172-930f-4148b8340c2e"},
		domain.Client{Hostname: "demo.example.lab", ClientID: "9a76b28b-0e09-41c8-bf01-79d1bef72646"})
	clientList.Clients = clients
	return clientList, nil
}
