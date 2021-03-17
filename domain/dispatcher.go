package domain

import "context"

type DispatcherInput struct {
	Recipient string            `json:"recipient"`
	Account   string            `json:"account"`
	URL       string            `json:"url"`
	Labels    map[string]string `json:"labels"`
	Timeout   int               `json:"timeout"`
}

type DispatcherResponse struct {
	Code  int    `json:"code"`
	RunID string `json:"id"`
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

type DispatcherClient interface {
	Dispatch(ctx context.Context, input DispatcherInput) (*DispatcherResponse, error)
	// GetStatus(ctx context.Context, labels string) ()
}
