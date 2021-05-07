package dispatcher

import (
	"context"
	"encoding/json"
)

type dispatcherClientMock struct {
}

func NewDispatcherClientMock() DispatcherClient {
	return &dispatcherClientMock{}
}

func (dc *dispatcherClientMock) Dispatch(
	ctx context.Context,
	inputs []RunInput,
) ([]RunCreated, error) {
	bRes := []byte(`[
		{"code": 200, "id": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7"},
		{"code": 200, "id": "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"}
	]`)

	var Results RunsCreated
	err := json.Unmarshal(bRes, &Results)
	return Results, err
}
