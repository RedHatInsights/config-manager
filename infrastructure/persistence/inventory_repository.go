package persistence

import (
	"config-manager/domain"
	"config-manager/utils"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
)

type InventoryClient struct {
	PlatformURL string
	Client      utils.HTTPClient
}

func (c *InventoryClient) buildURL(page int) string {
	// queryParams := url.Values{}
	// for k, v := range params {
	// 	queryParams.Add(k, fmt.Sprintf("%v", v))
	// }

	return c.PlatformURL + "/api/inventory/v1/hosts?filter[system_profile][rhc_client_id]=not_nil&page=" + string(page)
}

func (c *InventoryClient) GetConnectedClients(ctx echo.Context, page int) (domain.InventoryResponse, error) {
	var results domain.InventoryResponse

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
