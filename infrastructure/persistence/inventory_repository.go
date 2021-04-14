package persistence

import (
	"config-manager/domain"
	"config-manager/utils"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
)

type InventoryClient struct {
	InventoryHost string
	InventoryImpl string
	Client        utils.HTTPClient
}

// TODO this function should accept a map of params instead of hard coding them.
func (c *InventoryClient) buildURL(page int) string {
	Url, err := url.Parse(c.InventoryHost)
	if err != nil {
		fmt.Println("Couldn't parse inventory host")
		return ""
	}
	Url.Path += "/api/inventory/v1/hosts"
	params := url.Values{}
	params.Add("filter[system_profile][rhc_client_id]", "not_nil")
	params.Add("fields[system_profile]", "rhc_client_id,rhc_config_state")
	params.Add("page", fmt.Sprintf("%d", page))
	Url.RawQuery = params.Encode()

	return Url.String()
}

func (c *InventoryClient) GetInventoryClients(ctx echo.Context, page int) (domain.InventoryResponse, error) {
	var results domain.InventoryResponse

	if c.InventoryImpl == "mock" {
		expectedResponse := []byte(`{
			"total": "0",
			"count": "0",
			"page": "1",
			"per_page": "50",
			"results": []
		}`)

		err := json.Unmarshal(expectedResponse, &results)
		return results, err
	}

	req, err := http.NewRequestWithContext(ctx.Request().Context(), "GET", c.buildURL(page), nil)
	if err != nil {
		fmt.Println("Error constructing request to inventory: ", err)
		return results, err
	}
	req.Header.Add("X-Rh-Identity", ctx.Request().Header["X-Rh-Identity"][0]) //TODO: Re-evaluate header forwarding

	res, err := c.Client.Do(req)
	if err != nil {
		fmt.Println("Error during request to inventory: ", err)
		return results, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&results)
	return results, nil
}
