package domain

type ClientList struct {
	AccountID string
	Clients   []string
}

type ClientListRepository interface {
	GetConnectedClients(accountID string) (*ClientList, error)
}
