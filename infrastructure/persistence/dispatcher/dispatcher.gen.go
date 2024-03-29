// Package dispatcher provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen DO NOT EDIT.
package dispatcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// RequestEditorFn  is the function signature for the RequestEditor callback function
type RequestEditorFn func(ctx context.Context, req *http.Request) error

// Doer performs HTTP requests.
//
// The standard http.Client implements this interface.
type HttpRequestDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

// Client which conforms to the OpenAPI3 specification for this service.
type Client struct {
	// The endpoint of the server conforming to this interface, with scheme,
	// https://api.deepmap.com for example. This can contain a path relative
	// to the server, such as https://api.deepmap.com/dev-test, and all the
	// paths in the swagger spec will be appended to the server.
	Server string

	// Doer for performing requests, typically a *http.Client with any
	// customized settings, such as certificate chains.
	Client HttpRequestDoer

	// A list of callbacks for modifying requests which are generated before sending over
	// the network.
	RequestEditors []RequestEditorFn
}

// ClientOption allows setting custom parameters during construction
type ClientOption func(*Client) error

// Creates a new Client, with reasonable defaults
func NewClient(server string, opts ...ClientOption) (*Client, error) {
	// create a client with sane default values
	client := Client{
		Server: server,
	}
	// mutate client and add all optional params
	for _, o := range opts {
		if err := o(&client); err != nil {
			return nil, err
		}
	}
	// ensure the server URL always has a trailing slash
	if !strings.HasSuffix(client.Server, "/") {
		client.Server += "/"
	}
	// create httpClient, if not already present
	if client.Client == nil {
		client.Client = &http.Client{}
	}
	return &client, nil
}

// WithHTTPClient allows overriding the default Doer, which is
// automatically created using http.Client. This is useful for tests.
func WithHTTPClient(doer HttpRequestDoer) ClientOption {
	return func(c *Client) error {
		c.Client = doer
		return nil
	}
}

// WithRequestEditorFn allows setting up a callback function, which will be
// called right before sending the request. This can be used to mutate the request.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return func(c *Client) error {
		c.RequestEditors = append(c.RequestEditors, fn)
		return nil
	}
}

// The interface specification for the client above.
type ClientInterface interface {
	// ApiInternalRunsCreate request with any body
	ApiInternalRunsCreateWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ApiInternalRunsCreate(ctx context.Context, body ApiInternalRunsCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ApiInternalV2RunsCancel request with any body
	ApiInternalV2RunsCancelWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ApiInternalV2RunsCancel(ctx context.Context, body ApiInternalV2RunsCancelJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ApiInternalHighlevelConnectionStatus request with any body
	ApiInternalHighlevelConnectionStatusWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ApiInternalHighlevelConnectionStatus(ctx context.Context, body ApiInternalHighlevelConnectionStatusJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ApiInternalV2RunsCreate request with any body
	ApiInternalV2RunsCreateWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ApiInternalV2RunsCreate(ctx context.Context, body ApiInternalV2RunsCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ApiInternalV2RecipientsStatus request with any body
	ApiInternalV2RecipientsStatusWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	ApiInternalV2RecipientsStatus(ctx context.Context, body ApiInternalV2RecipientsStatusJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// ApiInternalVersion request
	ApiInternalVersion(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) ApiInternalRunsCreateWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewApiInternalRunsCreateRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ApiInternalRunsCreate(ctx context.Context, body ApiInternalRunsCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewApiInternalRunsCreateRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ApiInternalV2RunsCancelWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewApiInternalV2RunsCancelRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ApiInternalV2RunsCancel(ctx context.Context, body ApiInternalV2RunsCancelJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewApiInternalV2RunsCancelRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ApiInternalHighlevelConnectionStatusWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewApiInternalHighlevelConnectionStatusRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ApiInternalHighlevelConnectionStatus(ctx context.Context, body ApiInternalHighlevelConnectionStatusJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewApiInternalHighlevelConnectionStatusRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ApiInternalV2RunsCreateWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewApiInternalV2RunsCreateRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ApiInternalV2RunsCreate(ctx context.Context, body ApiInternalV2RunsCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewApiInternalV2RunsCreateRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ApiInternalV2RecipientsStatusWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewApiInternalV2RecipientsStatusRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ApiInternalV2RecipientsStatus(ctx context.Context, body ApiInternalV2RecipientsStatusJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewApiInternalV2RecipientsStatusRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) ApiInternalVersion(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewApiInternalVersionRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewApiInternalRunsCreateRequest calls the generic ApiInternalRunsCreate builder with application/json body
func NewApiInternalRunsCreateRequest(server string, body ApiInternalRunsCreateJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewApiInternalRunsCreateRequestWithBody(server, "application/json", bodyReader)
}

// NewApiInternalRunsCreateRequestWithBody generates requests for ApiInternalRunsCreate with any type of body
func NewApiInternalRunsCreateRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/dispatch")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewApiInternalV2RunsCancelRequest calls the generic ApiInternalV2RunsCancel builder with application/json body
func NewApiInternalV2RunsCancelRequest(server string, body ApiInternalV2RunsCancelJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewApiInternalV2RunsCancelRequestWithBody(server, "application/json", bodyReader)
}

// NewApiInternalV2RunsCancelRequestWithBody generates requests for ApiInternalV2RunsCancel with any type of body
func NewApiInternalV2RunsCancelRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/v2/cancel")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewApiInternalHighlevelConnectionStatusRequest calls the generic ApiInternalHighlevelConnectionStatus builder with application/json body
func NewApiInternalHighlevelConnectionStatusRequest(server string, body ApiInternalHighlevelConnectionStatusJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewApiInternalHighlevelConnectionStatusRequestWithBody(server, "application/json", bodyReader)
}

// NewApiInternalHighlevelConnectionStatusRequestWithBody generates requests for ApiInternalHighlevelConnectionStatus with any type of body
func NewApiInternalHighlevelConnectionStatusRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/v2/connection_status")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewApiInternalV2RunsCreateRequest calls the generic ApiInternalV2RunsCreate builder with application/json body
func NewApiInternalV2RunsCreateRequest(server string, body ApiInternalV2RunsCreateJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewApiInternalV2RunsCreateRequestWithBody(server, "application/json", bodyReader)
}

// NewApiInternalV2RunsCreateRequestWithBody generates requests for ApiInternalV2RunsCreate with any type of body
func NewApiInternalV2RunsCreateRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/v2/dispatch")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewApiInternalV2RecipientsStatusRequest calls the generic ApiInternalV2RecipientsStatus builder with application/json body
func NewApiInternalV2RecipientsStatusRequest(server string, body ApiInternalV2RecipientsStatusJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewApiInternalV2RecipientsStatusRequestWithBody(server, "application/json", bodyReader)
}

// NewApiInternalV2RecipientsStatusRequestWithBody generates requests for ApiInternalV2RecipientsStatus with any type of body
func NewApiInternalV2RecipientsStatusRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/v2/recipients/status")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", queryURL.String(), body)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Content-Type", contentType)

	return req, nil
}

// NewApiInternalVersionRequest generates requests for ApiInternalVersion
func NewApiInternalVersionRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/internal/version")
	if operationPath[0] == '/' {
		operationPath = "." + operationPath
	}

	queryURL, err := serverURL.Parse(operationPath)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", queryURL.String(), nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func (c *Client) applyEditors(ctx context.Context, req *http.Request, additionalEditors []RequestEditorFn) error {
	for _, r := range c.RequestEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	for _, r := range additionalEditors {
		if err := r(ctx, req); err != nil {
			return err
		}
	}
	return nil
}

// ClientWithResponses builds on ClientInterface to offer response payloads
type ClientWithResponses struct {
	ClientInterface
}

// NewClientWithResponses creates a new ClientWithResponses, which wraps
// Client with return type handling
func NewClientWithResponses(server string, opts ...ClientOption) (*ClientWithResponses, error) {
	client, err := NewClient(server, opts...)
	if err != nil {
		return nil, err
	}
	return &ClientWithResponses{client}, nil
}

// WithBaseURL overrides the baseURL.
func WithBaseURL(baseURL string) ClientOption {
	return func(c *Client) error {
		newBaseURL, err := url.Parse(baseURL)
		if err != nil {
			return err
		}
		c.Server = newBaseURL.String()
		return nil
	}
}

// ClientWithResponsesInterface is the interface specification for the client with responses above.
type ClientWithResponsesInterface interface {
	// ApiInternalRunsCreate request with any body
	ApiInternalRunsCreateWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ApiInternalRunsCreateResponse, error)

	ApiInternalRunsCreateWithResponse(ctx context.Context, body ApiInternalRunsCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*ApiInternalRunsCreateResponse, error)

	// ApiInternalV2RunsCancel request with any body
	ApiInternalV2RunsCancelWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ApiInternalV2RunsCancelResponse, error)

	ApiInternalV2RunsCancelWithResponse(ctx context.Context, body ApiInternalV2RunsCancelJSONRequestBody, reqEditors ...RequestEditorFn) (*ApiInternalV2RunsCancelResponse, error)

	// ApiInternalHighlevelConnectionStatus request with any body
	ApiInternalHighlevelConnectionStatusWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ApiInternalHighlevelConnectionStatusResponse, error)

	ApiInternalHighlevelConnectionStatusWithResponse(ctx context.Context, body ApiInternalHighlevelConnectionStatusJSONRequestBody, reqEditors ...RequestEditorFn) (*ApiInternalHighlevelConnectionStatusResponse, error)

	// ApiInternalV2RunsCreate request with any body
	ApiInternalV2RunsCreateWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ApiInternalV2RunsCreateResponse, error)

	ApiInternalV2RunsCreateWithResponse(ctx context.Context, body ApiInternalV2RunsCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*ApiInternalV2RunsCreateResponse, error)

	// ApiInternalV2RecipientsStatus request with any body
	ApiInternalV2RecipientsStatusWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ApiInternalV2RecipientsStatusResponse, error)

	ApiInternalV2RecipientsStatusWithResponse(ctx context.Context, body ApiInternalV2RecipientsStatusJSONRequestBody, reqEditors ...RequestEditorFn) (*ApiInternalV2RecipientsStatusResponse, error)

	// ApiInternalVersion request
	ApiInternalVersionWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ApiInternalVersionResponse, error)
}

type ApiInternalRunsCreateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON207      *RunsCreated
	JSON400      *Error
}

// Status returns HTTPResponse.Status
func (r ApiInternalRunsCreateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ApiInternalRunsCreateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ApiInternalV2RunsCancelResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON207      *RunsCanceled
	JSON400      *Error
}

// Status returns HTTPResponse.Status
func (r ApiInternalV2RunsCancelResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ApiInternalV2RunsCancelResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ApiInternalHighlevelConnectionStatusResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *HighLevelRecipientStatus
	JSON400      *Error
}

// Status returns HTTPResponse.Status
func (r ApiInternalHighlevelConnectionStatusResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ApiInternalHighlevelConnectionStatusResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ApiInternalV2RunsCreateResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON207      *RunsCreated
}

// Status returns HTTPResponse.Status
func (r ApiInternalV2RunsCreateResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ApiInternalV2RunsCreateResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ApiInternalV2RecipientsStatusResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *[]RecipientStatus
	JSON400      *Error
}

// Status returns HTTPResponse.Status
func (r ApiInternalV2RecipientsStatusResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ApiInternalV2RecipientsStatusResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type ApiInternalVersionResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *Version
}

// Status returns HTTPResponse.Status
func (r ApiInternalVersionResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r ApiInternalVersionResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// ApiInternalRunsCreateWithBodyWithResponse request with arbitrary body returning *ApiInternalRunsCreateResponse
func (c *ClientWithResponses) ApiInternalRunsCreateWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ApiInternalRunsCreateResponse, error) {
	rsp, err := c.ApiInternalRunsCreateWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseApiInternalRunsCreateResponse(rsp)
}

func (c *ClientWithResponses) ApiInternalRunsCreateWithResponse(ctx context.Context, body ApiInternalRunsCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*ApiInternalRunsCreateResponse, error) {
	rsp, err := c.ApiInternalRunsCreate(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseApiInternalRunsCreateResponse(rsp)
}

// ApiInternalV2RunsCancelWithBodyWithResponse request with arbitrary body returning *ApiInternalV2RunsCancelResponse
func (c *ClientWithResponses) ApiInternalV2RunsCancelWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ApiInternalV2RunsCancelResponse, error) {
	rsp, err := c.ApiInternalV2RunsCancelWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseApiInternalV2RunsCancelResponse(rsp)
}

func (c *ClientWithResponses) ApiInternalV2RunsCancelWithResponse(ctx context.Context, body ApiInternalV2RunsCancelJSONRequestBody, reqEditors ...RequestEditorFn) (*ApiInternalV2RunsCancelResponse, error) {
	rsp, err := c.ApiInternalV2RunsCancel(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseApiInternalV2RunsCancelResponse(rsp)
}

// ApiInternalHighlevelConnectionStatusWithBodyWithResponse request with arbitrary body returning *ApiInternalHighlevelConnectionStatusResponse
func (c *ClientWithResponses) ApiInternalHighlevelConnectionStatusWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ApiInternalHighlevelConnectionStatusResponse, error) {
	rsp, err := c.ApiInternalHighlevelConnectionStatusWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseApiInternalHighlevelConnectionStatusResponse(rsp)
}

func (c *ClientWithResponses) ApiInternalHighlevelConnectionStatusWithResponse(ctx context.Context, body ApiInternalHighlevelConnectionStatusJSONRequestBody, reqEditors ...RequestEditorFn) (*ApiInternalHighlevelConnectionStatusResponse, error) {
	rsp, err := c.ApiInternalHighlevelConnectionStatus(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseApiInternalHighlevelConnectionStatusResponse(rsp)
}

// ApiInternalV2RunsCreateWithBodyWithResponse request with arbitrary body returning *ApiInternalV2RunsCreateResponse
func (c *ClientWithResponses) ApiInternalV2RunsCreateWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ApiInternalV2RunsCreateResponse, error) {
	rsp, err := c.ApiInternalV2RunsCreateWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseApiInternalV2RunsCreateResponse(rsp)
}

func (c *ClientWithResponses) ApiInternalV2RunsCreateWithResponse(ctx context.Context, body ApiInternalV2RunsCreateJSONRequestBody, reqEditors ...RequestEditorFn) (*ApiInternalV2RunsCreateResponse, error) {
	rsp, err := c.ApiInternalV2RunsCreate(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseApiInternalV2RunsCreateResponse(rsp)
}

// ApiInternalV2RecipientsStatusWithBodyWithResponse request with arbitrary body returning *ApiInternalV2RecipientsStatusResponse
func (c *ClientWithResponses) ApiInternalV2RecipientsStatusWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*ApiInternalV2RecipientsStatusResponse, error) {
	rsp, err := c.ApiInternalV2RecipientsStatusWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseApiInternalV2RecipientsStatusResponse(rsp)
}

func (c *ClientWithResponses) ApiInternalV2RecipientsStatusWithResponse(ctx context.Context, body ApiInternalV2RecipientsStatusJSONRequestBody, reqEditors ...RequestEditorFn) (*ApiInternalV2RecipientsStatusResponse, error) {
	rsp, err := c.ApiInternalV2RecipientsStatus(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseApiInternalV2RecipientsStatusResponse(rsp)
}

// ApiInternalVersionWithResponse request returning *ApiInternalVersionResponse
func (c *ClientWithResponses) ApiInternalVersionWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*ApiInternalVersionResponse, error) {
	rsp, err := c.ApiInternalVersion(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseApiInternalVersionResponse(rsp)
}

// ParseApiInternalRunsCreateResponse parses an HTTP response from a ApiInternalRunsCreateWithResponse call
func ParseApiInternalRunsCreateResponse(rsp *http.Response) (*ApiInternalRunsCreateResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ApiInternalRunsCreateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 207:
		var dest RunsCreated
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON207 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	}

	return response, nil
}

// ParseApiInternalV2RunsCancelResponse parses an HTTP response from a ApiInternalV2RunsCancelWithResponse call
func ParseApiInternalV2RunsCancelResponse(rsp *http.Response) (*ApiInternalV2RunsCancelResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ApiInternalV2RunsCancelResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 207:
		var dest RunsCanceled
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON207 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	}

	return response, nil
}

// ParseApiInternalHighlevelConnectionStatusResponse parses an HTTP response from a ApiInternalHighlevelConnectionStatusWithResponse call
func ParseApiInternalHighlevelConnectionStatusResponse(rsp *http.Response) (*ApiInternalHighlevelConnectionStatusResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ApiInternalHighlevelConnectionStatusResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest HighLevelRecipientStatus
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	}

	return response, nil
}

// ParseApiInternalV2RunsCreateResponse parses an HTTP response from a ApiInternalV2RunsCreateWithResponse call
func ParseApiInternalV2RunsCreateResponse(rsp *http.Response) (*ApiInternalV2RunsCreateResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ApiInternalV2RunsCreateResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 207:
		var dest RunsCreated
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON207 = &dest

	}

	return response, nil
}

// ParseApiInternalV2RecipientsStatusResponse parses an HTTP response from a ApiInternalV2RecipientsStatusWithResponse call
func ParseApiInternalV2RecipientsStatusResponse(rsp *http.Response) (*ApiInternalV2RecipientsStatusResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ApiInternalV2RecipientsStatusResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest []RecipientStatus
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest Error
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	}

	return response, nil
}

// ParseApiInternalVersionResponse parses an HTTP response from a ApiInternalVersionWithResponse call
func ParseApiInternalVersionResponse(rsp *http.Response) (*ApiInternalVersionResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &ApiInternalVersionResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest Version
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	}

	return response, nil
}
