package utils

import (
	"bytes"
	"io/ioutil"
	"net/http"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type MockDoType func(req *http.Request) (*http.Response, error)

type ClientMock struct {
	MockDo MockDoType
}

func (m *ClientMock) Do(req *http.Request) (*http.Response, error) {
	return m.MockDo(req)
}

func SetupMockHTTPClient(expectedResponse string, status int) *ClientMock {
	r := ioutil.NopCloser(bytes.NewReader([]byte(expectedResponse)))

	client := &ClientMock{
		MockDo: func(*http.Request) (*http.Response, error) {
			return &http.Response{
				StatusCode: status,
				Body:       r,
			}, nil
		},
	}

	return client
}
