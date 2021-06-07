package controllers

import (
	"config-manager/api/instrumentation"
	"config-manager/application"
	"config-manager/domain"
	"config-manager/utils"
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/getkin/kin-openapi/openapi3"

	oapiMiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
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
func (cmc *ConfigManagerController) Routes(spec *openapi3.Swagger) {
	openapi3.DefineStringFormat("uuid", `^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89aAbB][a-f0-9]{3}-[a-f0-9]{12}$`)
	sub := cmc.Server.Group(cmc.URLBasePath)
	sub.Use(echo.WrapMiddleware(identity.EnforceIdentity))
	sub.Use(oapiMiddleware.OapiRequestValidator(spec))
	RegisterHandlers(sub, cmc)
}

// TODO: Again I don't like this.. Come up with a better solution for validating params (middleware?)
func translateStatesParams(params GetStatesParams) map[string]interface{} {
	p := map[string]interface{}{
		"limit":   50,
		"offset":  0,
		"sort_by": "created_at:desc",
	}

	if params.Limit != nil {
		p["limit"] = int(*params.Limit)
	}
	if params.Offset != nil {
		p["offset"] = int(*params.Offset)
	}
	if params.SortBy != nil {
		p["sort_by"] = string(*params.SortBy)
	}

	return p
}

func (cmc *ConfigManagerController) getClients(ctx echo.Context, currentState domain.AccountState) ([]domain.Host, error) {
	//TODO There's probably a better way to do this
	ctxWithID := context.WithValue(ctx.Request().Context(), "X-Rh-Identity", ctx.Request().Header["X-Rh-Identity"][0])
	var clients []domain.Host
	inventoryRHCIDs := make(map[string]bool)
	var err error

	// workaround: If insights is disabled - get clients from cloud-connector instead of inventory
	// This should be a temporary fix.
	if currentState.State["insights"] == "enabled" {
		res, err := cmc.ConfigManagerService.GetInventoryClients(ctxWithID, 1)
		if err != nil {
			instrumentation.InventoryRequestError()
			return nil, err
		}
		clients = append(clients, res.Results...)
		for _, client := range res.Results {
			inventoryRHCIDs[client.SystemProfile.RHCID] = true
		}

		for len(clients) < res.Total {
			page := res.Page + 1
			res, err = cmc.ConfigManagerService.GetInventoryClients(ctxWithID, page)
			if err != nil {
				instrumentation.InventoryRequestError()
				return nil, err
			}
			clients = append(clients, res.Results...)
			for _, client := range res.Results {
				inventoryRHCIDs[client.SystemProfile.RHCID] = true
			}
		}
	}

	res, err := cmc.ConfigManagerService.GetConnectedClients(ctxWithID, currentState.AccountID)
	if err != nil {
		instrumentation.CloudConnectorRequestError()
		return nil, err
	}

	for clientID := range res {
		if !inventoryRHCIDs[clientID] {
			clients = append(clients, domain.Host{
				Account: currentState.AccountID,
				SystemProfile: domain.SystemProfile{
					RHCID: clientID,
				},
			})
		}
	}

	return clients, err
}

// GetStates get the archive of state changes for requesting account
// (GET /states)
func (cmc *ConfigManagerController) GetStates(ctx echo.Context, params GetStatesParams) error {
	id := identity.Get(ctx.Request().Context())
	log.Println("Getting state changes for account: ", id.Identity.AccountNumber)

	p := translateStatesParams(params)

	// Add filter and sort-by
	states, err := cmc.ConfigManagerService.GetStateChanges(
		id.Identity.AccountNumber,
		p["sort_by"].(string),
		p["limit"].(int),
		p["offset"].(int),
	)
	if err != nil {
		instrumentation.GetStateChangesError()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, states)
}

// UpdateStates updates the configuration state for requesting account
// (POST /states)
func (cmc *ConfigManagerController) UpdateStates(ctx echo.Context) error {
	id := identity.Get(ctx.Request().Context())
	log.Println("Updating and applying state for account: ", id.Identity.AccountNumber)

	payload := &domain.StateMap{}
	bytes, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = json.Unmarshal(bytes, payload)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	currentState, err := cmc.ConfigManagerService.GetAccountState(id.Identity.AccountNumber)

	err = utils.VerifyStatePayload(currentState.State, *payload)
	if err != nil {
		log.Printf("Payload verification error: %s", err.Error())
		instrumentation.PayloadVerificationError()
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	acc, err := cmc.ConfigManagerService.UpdateAccountState(id.Identity.AccountNumber, "demo-user", *payload)
	if err != nil {
		instrumentation.UpdateAccountStateError()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	clients, err := cmc.getClients(ctx, *currentState)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	// TODO: Update ApplyState to return proper response data (dispatcher response code + id per client)

	results, err := cmc.ConfigManagerService.ApplyState(ctx.Request().Context(), acc, clients)
	if err != nil {
		instrumentation.PlaybookDispatcherRequestError()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	log.Println("Dispatcher results: ", results)

	return ctx.JSON(http.StatusOK, acc)
}

// GetCurrentState gets the current configuration state for requesting account
// (GET /states/current)
func (cmc *ConfigManagerController) GetCurrentState(ctx echo.Context) error {
	id := identity.Get(ctx.Request().Context())
	log.Println("Getting current state for account: ", id.Identity.AccountNumber)

	acc, err := cmc.ConfigManagerService.GetAccountState(id.Identity.AccountNumber)
	if err != nil {
		instrumentation.GetAccountStateError()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.JSON(http.StatusOK, acc)
}

// GetStateById gets a single configuration state for requesting account
// (GET /states/{id})
func (cmc *ConfigManagerController) GetStateById(ctx echo.Context, stateID StateIDParam) error {
	id := identity.Get(ctx.Request().Context())
	log.Printf("Getting state change for account: %s, with id: %s\n", id.Identity.AccountNumber, string(stateID))

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
	log.Printf("Getting playbook for account: %s, with id: %s\n", id.Identity.AccountNumber, string(stateID))

	playbook, err := cmc.ConfigManagerService.GetPlaybook(string(stateID))
	if err != nil {
		instrumentation.PlaybookRequestError()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	instrumentation.PlaybookRequestOK()
	return ctx.String(http.StatusOK, playbook)
}

// GetPlaybookPreview generates and returns a playbook preview to a requesting client
// (GET /states/preview)
func (cmc *ConfigManagerController) GetPlaybookPreview(ctx echo.Context) error {
	id := identity.Get(ctx.Request().Context())
	log.Printf("Getting playbook preview for account: %s\n", id.Identity.AccountNumber)

	payload := &domain.StateMap{}
	bytes, err := ioutil.ReadAll(ctx.Request().Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	err = json.Unmarshal(bytes, payload)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	playbook, err := cmc.ConfigManagerService.PlaybookGenerator.GeneratePlaybook(*payload)
	if err != nil {
		instrumentation.PlaybookRequestError()
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return ctx.String(http.StatusOK, playbook)
}
