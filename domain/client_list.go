package domain

type ClientList struct {
	AccountID string
	Clients   []Client
}

type Client struct {
	Hostname string
	ClientID string
}

type ClientListRepository interface {
	GetConnectedClients(accountID string) (*ClientList, error)
}
