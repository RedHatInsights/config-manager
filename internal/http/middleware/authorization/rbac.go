package authorization

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type RbacClient interface {
	GetDefaultWorkspaceID(context.Context, string) (string, error)
}

type rbacClient struct {
	baseURL string
	client  http.Client
}

func newRbacClient(baseURL string) RbacClient {
	return &rbacClient{
		baseURL: baseURL,
		client:  http.Client{},
	}
}

var _ RbacClient = &rbacClient{}

type workspace struct {
	ID string `json:"id"`
}

type response struct {
	Data []workspace `json:"data"`
}

// TODO
// - authentication support based on https://issues.redhat.com/browse/RHCLOUD-40287
func (a *rbacClient) GetDefaultWorkspaceID(context context.Context, orgID string) (string, error) {

	url := fmt.Sprintf("%s/api/rbac/v2/workspaces/?type=default", a.baseURL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	var response response
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling response: %v", err)
	}

	if len(response.Data) != 1 {
		return "", fmt.Errorf("unexpected number of default workspaces: %d", len(response.Data))
	}

	return response.Data[0].ID, nil
}
