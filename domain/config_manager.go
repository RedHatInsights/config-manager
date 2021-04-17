package domain

import "context"

type ConfigManagerInterface interface {
	GetAccountState(id string) (*AccountState, error)
	ApplyState(ctx context.Context, acc *AccountState, clients []Host) ([]DispatcherResponse, error)
}
