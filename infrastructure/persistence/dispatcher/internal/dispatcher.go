package internal

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type DispatcherClient interface {
	Dispatch(ctx context.Context, inputs []RunInput) (*RunsCreated, error)
}

type dispatcherClientImpl struct {
	client ClientWithResponsesInterface
}

func NewDispatcherClientWithDoer(cfg *viper.Viper, doer HttpRequestDoer) DispatcherClient {
	client := &ClientWithResponses{
		ClientInterface: &Client{
			Server: cfg.GetString("Dispatcher_Host"),
			Client: doer,
			RequestEditors: []RequestEditorFn{
				func(ctx context.Context, req *http.Request) error {
					req.Header.Set("Authorization", fmt.Sprintf("PSK %s", cfg.GetString("Dispatcher_PSK")))
					req.Header.Set("Content-Type", "application/json")
					return nil
				},
			},
		},
	}

	return &dispatcherClientImpl{
		client: client,
	}
}

func NewDispatcherClient(cfg *viper.Viper) DispatcherClient {
	client := &http.Client{
		Timeout: time.Duration(int(time.Second) * cfg.GetInt("Dispatcher_Timeout")),
	}

	return NewDispatcherClientWithDoer(cfg, client)
}

func (dc *dispatcherClientImpl) Dispatch(ctx context.Context, inputs []RunInput) (*RunsCreated, error) {
	res, err := dc.client.ApiInternalRunsCreateWithResponse(ctx, inputs)
	if err != nil {
		return nil, err
	}

	if res.HTTPResponse.StatusCode != 207 {
		return nil, fmt.Errorf("Unexpected error code %d received", res.HTTPResponse.StatusCode)
	}

	return res.JSON207, nil
}

// func (dc *dispatcherClientImpl) GetRuns(ctx context.Context) (*Runs, error) {
// 	service := ServiceNullable("config_manager")
// 	params := &ApiRunsListParams{
// 		Filter: &RunsFilter{
// 			Service: &service,
// 		},
// 	}

// 	res, err := dc.client.ApiRunsListWithResponse(ctx, params)
// 	if err != nil {
// 		return nil, err
// 	}

// 	switch res.HTTPResponse.StatusCode {
// 	case 400:
// 		return nil, fmt.Errorf("400: %s", res.JSON400.Message)
// 	case 403:
// 		return nil, fmt.Errorf("403: %s", res.JSON403.Message)
// 	case 200:
// 		return res.JSON200, nil
// 	default:
// 		return nil, fmt.Errorf("unexpected status code %d", res.HTTPResponse.StatusCode)
// 	}
// }

// func (dc *dispatcherClientImpl) GetRunHosts(ctx context.Context) {

// }
