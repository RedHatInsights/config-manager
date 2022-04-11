package persistence

import (
	"config-manager/domain"
	"config-manager/utils"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

// InventoryClient provides REST client API methods to interact with the
// platform Inventory application.
type InventoryClient struct {
	InventoryHost string
	InventoryImpl string
	Client        utils.HTTPClient
}

// TODO this function should accept a map of params instead of hard coding them.
func (c *InventoryClient) buildURL(page int) string {
	Url, err := url.Parse(c.InventoryHost)
	if err != nil {
		log.Println("Couldn't parse inventory host")
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

// GetInventoryClients sends an HTTP GET request to the Inventory service,
// marshals the response into a domain.InventoryResponse structure and returns
// it.
func (c *InventoryClient) GetInventoryClients(ctx context.Context, page int) (domain.InventoryResponse, error) {
	var results domain.InventoryResponse

	if c.InventoryImpl == "mock" {
		expectedResponse := []byte(`{
			"total": 0,
			"count": 0,
			"page": 1,
			"per_page": 50,
			"results": []
		}`)

		err := json.Unmarshal(expectedResponse, &results)
		return results, err
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.buildURL(page), nil)
	if err != nil {
		log.Println("Error constructing request to inventory: ", err)
		return results, err
	}
	req.Header.Add("X-Rh-Identity", ctx.Value("X-Rh-Identity").(string)) //TODO: Re-evaluate header forwarding

	res, err := c.Client.Do(req)
	if err != nil {
		log.Println("Error during request to inventory: ", err)
		return results, err
	}
	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&results)
	if err != nil {
		body, _ := ioutil.ReadAll(res.Body)
		log.Println("Error decoding inventory response: ", string(body))
	}
	return results, nil
}
