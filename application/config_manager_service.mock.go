package application

import (
	"config-manager/domain"
	"config-manager/infrastructure/persistence/dispatcher"
	"config-manager/internal"
	"config-manager/internal/db"
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

func (m *ConfigManagerServiceMock) ApplyState(ctx context.Context, profile db.Profile, clients []internal.Host) ([]dispatcher.RunCreated, error) {
	args := m.Called(ctx, profile, clients)
	return args.Get(0).([]dispatcher.RunCreated), args.Error(1)
}

func (m *ConfigManagerServiceMock) GetSingleStateChange(id string) (*domain.StateArchive, error) {
	args := m.Called(id)
	return args.Get(0).(*domain.StateArchive), args.Error(1)
}

func (m *ConfigManagerServiceMock) SetupHost(ctx context.Context, host internal.Host) (string, error) {
	args := m.Called(ctx, host)
	return args.Get(0).(string), args.Error(1)
}

func (m *ConfigManagerServiceMock) SetApplyState(accountID string, applyState bool) error {
	args := m.Called(accountID, applyState)
	return args.Error(0)
}
