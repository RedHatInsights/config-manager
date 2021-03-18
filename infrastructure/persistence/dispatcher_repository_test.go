package persistence_test

import (
	"bytes"
	"config-manager/domain"
	"config-manager/infrastructure/persistence"
	"context"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockDoType func(req *http.Request) (*http.Response, error)

type ClientMock struct {
	MockDo MockDoType
}

func (m *ClientMock) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

func setupClient(expectedResponse string) *ClientMock {
	r := ioutil.NopCloser(bytes.NewReader([]byte(expectedResponse)))

	client := &ClientMock{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: 207,
				Body:       r,
			}, nil
		},
	}

	return client
}

var inputs = []domain.DispatcherInput{
	domain.DispatcherInput{
		Recipient: "276c4685-fdfb-4172-930f-4148b8340c2e",
		Account:   "0000001",
		URL:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
		Labels: map[string]string{
			"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
		},
	},
	domain.DispatcherInput{
		Recipient: "9a76b28b-0e09-41c8-bf01-79d1bef72646",
		Account:   "0000001",
		URL:       "https://cloud.redhat.com/api/config-manager/v1/states/e417581a-d649-4cdc-9506-6eb7fdbfd66d/playbook",
		Labels: map[string]string{
			"test": "e417581a-d649-4cdc-9506-6eb7fdbfd66d",
		},
	},
}

func TestDispatchSuccess(t *testing.T) {
	response := `[
		{"code": 200, "id": "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7"},
		{"code": 200, "id": "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4"}
	]`

	dispatcher := &persistence.DispatcherClient{
		DispatcherHost: "test",
		DispatcherPSK:  "test",
		Client:         setupClient(response),
	}

	results, err := dispatcher.Dispatch(context.Background(), inputs)

	assert.Nil(t, err)
	assert.Equal(t, len(results), 2, "there should be two response objects")

	assert.Equal(t, results[0].Code, 200, "the response code should be 200")
	assert.Equal(t, results[0].RunID, "3d711f8b-77d0-4ed5-a5b5-1d282bf930c7", "The id from the response should be included")

	assert.Equal(t, results[1].Code, 200, "the response code should be 200")
	assert.Equal(t, results[1].RunID, "74368f32-4e6d-4ea2-9b8f-22dac89f9ae4", "The id from the response should be included")
}

func TestDispatchNotFound(t *testing.T) {
	response := `[
		{"code": 404},
		{"code": 404}
	]`

	dispatcher := &persistence.DispatcherClient{
		DispatcherHost: "test",
		DispatcherPSK:  "test",
		Client:         setupClient(response),
	}

	results, err := dispatcher.Dispatch(context.Background(), inputs)

	assert.Nil(t, err)
	assert.Equal(t, len(results), 2, "there should be two response objects")

	assert.Equal(t, results[0].Code, 404, "the response code should be 404")
	assert.Equal(t, results[0].RunID, "", "RunID should be empty")

	assert.Equal(t, results[1].Code, 404, "the response code should be 404")
	assert.Equal(t, results[1].RunID, "", "RunID should be empty")
}
