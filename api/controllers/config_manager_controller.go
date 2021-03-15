package controllers

import (
	"config-manager/application"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/labstack/echo/v4"

	"github.com/redhatinsights/platform-go-middlewares/identity"
)

// ConfigManagerController implements ServerInterface
type ConfigManagerController struct {
	ConfigManagerService *application.ConfigManagerService
	Server               *echo.Echo
	URLBasePath          string
}

// Routes sets up middlewares and registers handlers for each route
func (cmc *ConfigManagerController) Routes() {
	sub := cmc.Server.Group(cmc.URLBasePath)
	sub.Use(echo.WrapMiddleware(identity.EnforceIdentity))
	// TODO: This is a weird way to register the routes. Should probably
	// remove Server from this controller and instead create an api "main.go"
	RegisterHandlers(sub, cmc)
}

// Start starts an http server with addr
func (cmc *ConfigManagerController) Start(addr string) {
	cmc.Server.Start(addr)
}

// TODO: Again I don't like this.. Come up with a better solution for validating params (middleware?)
func translateStatesParams(params GetStatesParams) map[string]interface{} {
	p := map[string]interface{}{
		"limit":  50,
		"offset": 0,
	}

	if params.Limit != nil {
		p["limit"] = int(*params.Limit)
	}
	if params.Offset != nil {
		p["offset"] = int(*params.Offset)
	}

	return p
}

// GetStates get the archive of state changes for requesting account
// (GET /states)
func (cmc *ConfigManagerController) GetStates(ctx echo.Context, params GetStatesParams) error {
	id := identity.Get(ctx.Request().Context())
	fmt.Println("Getting state changes for account: ", id.Identity.AccountNumber)

	p := translateStatesParams(params)

	// Add filter and sort-by
	states, err := cmc.ConfigManagerService.GetStateChanges(
		id.Identity.AccountNumber,
		p["limit"].(int),
		p["offset"].(int),
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, states)
}

// UpdateStates updates the configuration state for requesting account
// (POST /states)
func (cmc *ConfigManagerController) UpdateStates(ctx echo.Context) error {
	id := identity.Get(ctx.Request().Context())
	fmt.Println("Updating and applying state for account: ", id.Identity.AccountNumber)

	payload := &State{}
	bytes, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = json.Unmarshal(bytes, payload)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	acc, err := cmc.ConfigManagerService.UpdateAccountState(id.Identity.AccountNumber, "demo-user", payload.AdditionalProperties)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// This should happen before the call to update state. Perhaps as another api call that responds with a list
	// of clients to be passed into this endpoint - preflight check
	clients, err := cmc.ConfigManagerService.GetClients(id.Identity.AccountNumber)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// TODO: Update ApplyState to return proper response data (dispatcher response code + id per client)

	_, err = cmc.ConfigManagerService.ApplyState(acc, "demo-user", clients.Clients)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, acc)
}

// GetCurrentState gets the current configuration state for requesting account
// (GET /states/current)
func (cmc *ConfigManagerController) GetCurrentState(ctx echo.Context) error {
	id := identity.Get(ctx.Request().Context())
	fmt.Println("Getting current state for account: ", id.Identity.AccountNumber)

	acc, err := cmc.ConfigManagerService.GetAccountState(id.Identity.AccountNumber)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, acc)
}

// GetStateById gets a single configuration state for requesting account
// (GET /states/{id})
func (cmc *ConfigManagerController) GetStateById(ctx echo.Context, stateID StateIDParam) error {
	id := identity.Get(ctx.Request().Context())
	fmt.Printf("Getting state change for account: %s, with id: %s\n", id.Identity.AccountNumber, string(stateID))

	state, err := cmc.ConfigManagerService.GetSingleStateChange(string(stateID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, state)
}

// GetPlaybookById generates and returns a playbook to a requesting client via a state ID
// (GET /states/{id}/playbook)
func (cmc *ConfigManagerController) GetPlaybookById(ctx echo.Context, stateID StateIDParam) error {
	id := identity.Get(ctx.Request().Context())
	fmt.Printf("Getting playbook for account: %s, with id: %s\n", id.Identity.AccountNumber, string(stateID))

	playbook, err := cmc.ConfigManagerService.GetPlaybook(string(stateID))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, playbook)
}
