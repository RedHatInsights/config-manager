// Package controllers provides primitives to interact the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package controllers

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Get state changes for requesting account
	// (GET /changes)
	GetChanges(ctx echo.Context, params GetChangesParams) error
	// Get single state change for requesting account
	// (GET /changes/{id})
	GetChangesId(ctx echo.Context, id StateIDParam) error
	// Get a list of runs for each state change
	// (GET /runs)
	GetRuns(ctx echo.Context, params GetRunsParams) error
	// Generate new runs by applying a state change
	// (POST /runs)
	PostRuns(ctx echo.Context) error
	// Get a single run
	// (GET /runs/{id})
	GetRunsId(ctx echo.Context, id RunIDParam) error
	// Get configuration state for requesting account
	// (GET /states)
	GetStates(ctx echo.Context) error
	// Update configuration state for requesting account
	// (POST /states)
	PostStates(ctx echo.Context) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// GetChanges converts echo context to params.
func (w *ServerInterfaceWrapper) GetChanges(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetChangesParams
	// ------------- Optional query parameter "limit" -------------

	err = runtime.BindQueryParameter("form", true, false, "limit", ctx.QueryParams(), &params.Limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter limit: %s", err))
	}

	// ------------- Optional query parameter "offset" -------------

	err = runtime.BindQueryParameter("form", true, false, "offset", ctx.QueryParams(), &params.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter offset: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetChanges(ctx, params)
	return err
}

// GetChangesId converts echo context to params.
func (w *ServerInterfaceWrapper) GetChangesId(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id StateIDParam

	err = runtime.BindStyledParameter("simple", false, "id", ctx.Param("id"), &id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetChangesId(ctx, id)
	return err
}

// GetRuns converts echo context to params.
func (w *ServerInterfaceWrapper) GetRuns(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params GetRunsParams
	// ------------- Optional query parameter "filter" -------------

	err = runtime.BindQueryParameter("form", true, false, "filter", ctx.QueryParams(), &params.Filter)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter filter: %s", err))
	}

	// ------------- Optional query parameter "sort_by" -------------

	err = runtime.BindQueryParameter("form", true, false, "sort_by", ctx.QueryParams(), &params.SortBy)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter sort_by: %s", err))
	}

	// ------------- Optional query parameter "limit" -------------

	err = runtime.BindQueryParameter("form", true, false, "limit", ctx.QueryParams(), &params.Limit)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter limit: %s", err))
	}

	// ------------- Optional query parameter "offset" -------------

	err = runtime.BindQueryParameter("form", true, false, "offset", ctx.QueryParams(), &params.Offset)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter offset: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetRuns(ctx, params)
	return err
}

// PostRuns converts echo context to params.
func (w *ServerInterfaceWrapper) PostRuns(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PostRuns(ctx)
	return err
}

// GetRunsId converts echo context to params.
func (w *ServerInterfaceWrapper) GetRunsId(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "id" -------------
	var id RunIDParam

	err = runtime.BindStyledParameter("simple", false, "id", ctx.Param("id"), &id)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter id: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetRunsId(ctx, id)
	return err
}

// GetStates converts echo context to params.
func (w *ServerInterfaceWrapper) GetStates(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetStates(ctx)
	return err
}

// PostStates converts echo context to params.
func (w *ServerInterfaceWrapper) PostStates(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PostStates(ctx)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/changes", wrapper.GetChanges)
	router.GET(baseURL+"/changes/:id", wrapper.GetChangesId)
	router.GET(baseURL+"/runs", wrapper.GetRuns)
	router.POST(baseURL+"/runs", wrapper.PostRuns)
	router.GET(baseURL+"/runs/:id", wrapper.GetRunsId)
	router.GET(baseURL+"/states", wrapper.GetStates)
	router.POST(baseURL+"/states", wrapper.PostStates)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"H4sIAAAAAAAC/9SYX2/bNhDAvwrB7VGNnG590Vub7o+xDA3i7akIClo6WSwkUiGPWY1A3304UrKlSHbk",
	"NGm3N8vkne5+94dH3fNUV7VWoNDy5J7XwogKEIx/upSVRPqRgU2NrFFqxRP+p/giK1cx5ao1GKZzZsC6",
	"Ei1DzQygM4pHXNLWWwdmyyOuRAU84aVXGHGbFlCJoDkXrkSevFlEvAqKefJ6QU9ShafziOO2JnmpEDZg",
	"eNNE/EOeW5iwbqkymQoEy7AAZlEYlGrDam0l7SBzacFbxgyUAuUdkOX0L9EoAYFZQNopESpSJJBVAtNi",
	"L3rAQx2smnSx79Ni0qdrp5bvrygGY78sCgQmTFqQvTIDhTKXYDpDaoHF3g6Z8YgbuHXSQMYTNA76Nv1o",
	"IOcJ/yHehz8Oqzb2RnTm2F9liWBIZsrdPKzOVX0p1lDuVK+0wXfbQ6qtNvhpvR3oBkXsPvLUgEDIPgki",
	"vX9IhE2HfxBCfrNjbdFItfEGrAjn94fdmsEbsqn9k2Tepql2yqd3Jb5cgtpgwZPzkEK7x5FjUSfo9fqK",
	"NroGgxK8WrFXe8yq7u1NRL7NdSHipY/vvCSIAuRZ2n3EWlf1+jOk+NDVpaodPoe/X2vU79piSIv7cXCW",
	"SqIUqM3k6mVHb7Ry7dRz+NYrm0eE/pIVWBRVTWJFz6VjQjvXZ6VN22aooHpQjkns6T0p1ZydYdIqbGwi",
	"7ursdFhTGXHt1EUAP45hqjOY7j7OMloMp+utA0ut7uGR8ZTYHDBx+Z5U5NpU5DB3zje1qUxc7Vh27dg4",
	"pWg94talKVjLI54LWToDE803NH/f9ul0nREUvrdZGCO2nY4e1rmqOpEJjbuWKbLMjwqivBoEa+THiKNX",
	"8TacGt+xYE/q2d+q/L6mqfa5zs+cQTQOBXxm3u/59ndTh3iFsoKxSOPR5npc3Rda5XLDKqHEBgyzYO5k",
	"6jVILGG0gUf8DowNsouzxdk52aNrUKKWPOE/+b8iP5J4JHFaCLUJnDZhQqYsFPT6ZcYT/hvgRbslGkz8",
	"H6d57rfE4UbQRI9ubKfz5oYGI1trZYNBrxeL0PcUQigFUdclDexSq/iz1f6gO2F22mWFJz4k/eEPQvVz",
	"eOVw6Z3I2HXbV/3o5apKmG2gw8L413JkuTZdD6abRFfGJNaxju9l1swAvsxORj4YVL8Zz+fFKdWmhAHV",
	"o1BNez4cgunPj1M59i4zM/K3dz+Zsfs/VRaeznOGT7BSWn8dpsD4yIFIi0E8SXWt7US8rrTtAtbG+53O",
	"ts/m7eC+03i3X5TsboCYBvwAniISwBT8E9itt4xev/Up/wBgl/mP9hIy4wmNpPdx4aXzbyYdSq22Nxgv",
	"FPHYMzla/auw4wU9eJhTT60k2nM+3vPWYQEKW9tYJS1BYLnR1W7WbyL+Zkr/UiEYJUq2AnMHhv1ijDYT",
	"YFM/Q7hArk20Qy33WOX2YL9s7YYL/ARsv8oqUYfPcAog8x8a18DaK9ro40vzv8iNJ8f3b+/2SSH2CrzG",
	"0CqcKXnCC8TaJnGcltplZwayQuBZqqtY1DIO6l+1U2h8RzPm0Nba6Myl/oGOvqBzpuzuw6v/ktbcNP8G",
	"AAD//2rRgYCKFgAA",
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file.
func GetSwagger() (*openapi3.Swagger, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	swagger, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("error loading Swagger: %s", err)
	}
	return swagger, nil
}
