package domain

type DispatcherInput struct {
	ClientID  string
	AccountID string
	URL       string
	Labels    string
	Timeout   int
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
	Dispatch(clientID string) (*DispatcherResponse, error)
	GetStatus(label string) ([]DispatcherRun, error)
}
