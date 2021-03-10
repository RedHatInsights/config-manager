package domain

type DispatcherInput struct {
	Recipient string
	Account   string
	URL       string
	Labels    map[string]string
}

type DispatcherResponse struct {
	Code  int
	RunID string
}

type DispatcherRun struct {
	ClientID  string
	AccountID string
	RunID     string
	URL       string
	Labels    string
	Timeout   int
	Status    string
}

type DispatcherRepository interface {
	Dispatch(clientID string, acc *AccountState) (*DispatcherResponse, error)
	GetStatus(label string) ([]DispatcherRun, error)
}
