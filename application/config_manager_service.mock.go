package application

import (
	"config-manager/domain"
	"context"

	"github.com/stretchr/testify/mock"
)

type ConfigManagerServiceMock struct {
	mock.Mock
}

func (m *ConfigManagerServiceMock) GetAccountState(id string) (*domain.AccountState, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.AccountState), args.Error(1)
}

func (m *ConfigManagerServiceMock) ApplyState(ctx context.Context, acc *domain.AccountState, clients []domain.Host) ([]domain.DispatcherResponse, error) {
	args := m.Called(ctx, acc, clients)
	return args.Get(0).([]domain.DispatcherResponse), args.Error(1)
}

func (m *ConfigManagerServiceMock) GetSingleStateChange(id string) (*domain.AccountState, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.AccountState), args.Error(1)
}
