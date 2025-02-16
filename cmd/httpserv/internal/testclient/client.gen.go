// Package testclient provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/oapi-codegen/oapi-codegen/v2 version v2.4.2-0.20250102212541-8bbe226927c9 DO NOT EDIT.
package testclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/oapi-codegen/runtime"
)

const (
	BearerAuthScopes = "BearerAuth.Scopes"
)

// AuthRequest defines model for AuthRequest.
type AuthRequest struct {
	// Password Пароль для аутентификации.
	Password string `json:"password"`

	// Username Имя пользователя для аутентификации.
	Username string `json:"username"`
}

// AuthResponse defines model for AuthResponse.
type AuthResponse struct {
	// Token JWT-токен для доступа к защищенным ресурсам.
	Token *string `json:"token,omitempty"`
}

// ErrorResponse defines model for ErrorResponse.
type ErrorResponse struct {
	// Errors Сообщение об ошибке, описывающее проблему.
	Errors *string `json:"errors,omitempty"`
}

// InfoResponse defines model for InfoResponse.
type InfoResponse struct {
	CoinHistory *struct {
		Received *[]struct {
			// Amount Количество полученных монет.
			Amount *int `json:"amount,omitempty"`

			// FromUser Имя пользователя, который отправил монеты.
			FromUser *string `json:"fromUser,omitempty"`
		} `json:"received,omitempty"`
		Sent *[]struct {
			// Amount Количество отправленных монет.
			Amount *int `json:"amount,omitempty"`

			// ToUser Имя пользователя, которому отправлены монеты.
			ToUser *string `json:"toUser,omitempty"`
		} `json:"sent,omitempty"`
	} `json:"coinHistory,omitempty"`

	// Coins Количество доступных монет.
	Coins     *int `json:"coins,omitempty"`
	Inventory *[]struct {
		// Quantity Количество предметов.
		Quantity *int `json:"quantity,omitempty"`

		// Type Тип предмета.
		Type *string `json:"type,omitempty"`
	} `json:"inventory,omitempty"`
}

// SendCoinRequest defines model for SendCoinRequest.
type SendCoinRequest struct {
	// Amount Количество монет, которые необходимо отправить.
	Amount int `json:"amount"`

	// ToUser Имя пользователя, которому нужно отправить монеты.
	ToUser string `json:"toUser"`
}

// PostApiAuthJSONRequestBody defines body for PostApiAuth for application/json ContentType.
type PostApiAuthJSONRequestBody = AuthRequest

// PostApiSendCoinJSONRequestBody defines body for PostApiSendCoin for application/json ContentType.
type PostApiSendCoinJSONRequestBody = SendCoinRequest

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
	// PostApiAuthWithBody request with any body
	PostApiAuthWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostApiAuth(ctx context.Context, body PostApiAuthJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetApiBuyItem request
	GetApiBuyItem(ctx context.Context, item string, reqEditors ...RequestEditorFn) (*http.Response, error)

	// GetApiInfo request
	GetApiInfo(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error)

	// PostApiSendCoinWithBody request with any body
	PostApiSendCoinWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error)

	PostApiSendCoin(ctx context.Context, body PostApiSendCoinJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error)
}

func (c *Client) PostApiAuthWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostApiAuthRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostApiAuth(ctx context.Context, body PostApiAuthJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostApiAuthRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetApiBuyItem(ctx context.Context, item string, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetApiBuyItemRequest(c.Server, item)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) GetApiInfo(ctx context.Context, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewGetApiInfoRequest(c.Server)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostApiSendCoinWithBody(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostApiSendCoinRequestWithBody(c.Server, contentType, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

func (c *Client) PostApiSendCoin(ctx context.Context, body PostApiSendCoinJSONRequestBody, reqEditors ...RequestEditorFn) (*http.Response, error) {
	req, err := NewPostApiSendCoinRequest(c.Server, body)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	if err := c.applyEditors(ctx, req, reqEditors); err != nil {
		return nil, err
	}
	return c.Client.Do(req)
}

// NewPostApiAuthRequest calls the generic PostApiAuth builder with application/json body
func NewPostApiAuthRequest(server string, body PostApiAuthJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostApiAuthRequestWithBody(server, "application/json", bodyReader)
}

// NewPostApiAuthRequestWithBody generates requests for PostApiAuth with any type of body
func NewPostApiAuthRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/auth")
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

// NewGetApiBuyItemRequest generates requests for GetApiBuyItem
func NewGetApiBuyItemRequest(server string, item string) (*http.Request, error) {
	var err error

	var pathParam0 string

	pathParam0, err = runtime.StyleParamWithLocation("simple", false, "item", runtime.ParamLocationPath, item)
	if err != nil {
		return nil, err
	}

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/buy/%s", pathParam0)
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

// NewGetApiInfoRequest generates requests for GetApiInfo
func NewGetApiInfoRequest(server string) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/info")
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

// NewPostApiSendCoinRequest calls the generic PostApiSendCoin builder with application/json body
func NewPostApiSendCoinRequest(server string, body PostApiSendCoinJSONRequestBody) (*http.Request, error) {
	var bodyReader io.Reader
	buf, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	bodyReader = bytes.NewReader(buf)
	return NewPostApiSendCoinRequestWithBody(server, "application/json", bodyReader)
}

// NewPostApiSendCoinRequestWithBody generates requests for PostApiSendCoin with any type of body
func NewPostApiSendCoinRequestWithBody(server string, contentType string, body io.Reader) (*http.Request, error) {
	var err error

	serverURL, err := url.Parse(server)
	if err != nil {
		return nil, err
	}

	operationPath := fmt.Sprintf("/api/sendCoin")
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
	// PostApiAuthWithBodyWithResponse request with any body
	PostApiAuthWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostApiAuthResponse, error)

	PostApiAuthWithResponse(ctx context.Context, body PostApiAuthJSONRequestBody, reqEditors ...RequestEditorFn) (*PostApiAuthResponse, error)

	// GetApiBuyItemWithResponse request
	GetApiBuyItemWithResponse(ctx context.Context, item string, reqEditors ...RequestEditorFn) (*GetApiBuyItemResponse, error)

	// GetApiInfoWithResponse request
	GetApiInfoWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetApiInfoResponse, error)

	// PostApiSendCoinWithBodyWithResponse request with any body
	PostApiSendCoinWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostApiSendCoinResponse, error)

	PostApiSendCoinWithResponse(ctx context.Context, body PostApiSendCoinJSONRequestBody, reqEditors ...RequestEditorFn) (*PostApiSendCoinResponse, error)
}

type PostApiAuthResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *AuthResponse
	JSON400      *ErrorResponse
	JSON401      *ErrorResponse
	JSON500      *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r PostApiAuthResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostApiAuthResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetApiBuyItemResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON400      *ErrorResponse
	JSON401      *ErrorResponse
	JSON500      *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r GetApiBuyItemResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetApiBuyItemResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type GetApiInfoResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON200      *InfoResponse
	JSON400      *ErrorResponse
	JSON401      *ErrorResponse
	JSON500      *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r GetApiInfoResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r GetApiInfoResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

type PostApiSendCoinResponse struct {
	Body         []byte
	HTTPResponse *http.Response
	JSON400      *ErrorResponse
	JSON401      *ErrorResponse
	JSON500      *ErrorResponse
}

// Status returns HTTPResponse.Status
func (r PostApiSendCoinResponse) Status() string {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.Status
	}
	return http.StatusText(0)
}

// StatusCode returns HTTPResponse.StatusCode
func (r PostApiSendCoinResponse) StatusCode() int {
	if r.HTTPResponse != nil {
		return r.HTTPResponse.StatusCode
	}
	return 0
}

// PostApiAuthWithBodyWithResponse request with arbitrary body returning *PostApiAuthResponse
func (c *ClientWithResponses) PostApiAuthWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostApiAuthResponse, error) {
	rsp, err := c.PostApiAuthWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostApiAuthResponse(rsp)
}

func (c *ClientWithResponses) PostApiAuthWithResponse(ctx context.Context, body PostApiAuthJSONRequestBody, reqEditors ...RequestEditorFn) (*PostApiAuthResponse, error) {
	rsp, err := c.PostApiAuth(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostApiAuthResponse(rsp)
}

// GetApiBuyItemWithResponse request returning *GetApiBuyItemResponse
func (c *ClientWithResponses) GetApiBuyItemWithResponse(ctx context.Context, item string, reqEditors ...RequestEditorFn) (*GetApiBuyItemResponse, error) {
	rsp, err := c.GetApiBuyItem(ctx, item, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetApiBuyItemResponse(rsp)
}

// GetApiInfoWithResponse request returning *GetApiInfoResponse
func (c *ClientWithResponses) GetApiInfoWithResponse(ctx context.Context, reqEditors ...RequestEditorFn) (*GetApiInfoResponse, error) {
	rsp, err := c.GetApiInfo(ctx, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParseGetApiInfoResponse(rsp)
}

// PostApiSendCoinWithBodyWithResponse request with arbitrary body returning *PostApiSendCoinResponse
func (c *ClientWithResponses) PostApiSendCoinWithBodyWithResponse(ctx context.Context, contentType string, body io.Reader, reqEditors ...RequestEditorFn) (*PostApiSendCoinResponse, error) {
	rsp, err := c.PostApiSendCoinWithBody(ctx, contentType, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostApiSendCoinResponse(rsp)
}

func (c *ClientWithResponses) PostApiSendCoinWithResponse(ctx context.Context, body PostApiSendCoinJSONRequestBody, reqEditors ...RequestEditorFn) (*PostApiSendCoinResponse, error) {
	rsp, err := c.PostApiSendCoin(ctx, body, reqEditors...)
	if err != nil {
		return nil, err
	}
	return ParsePostApiSendCoinResponse(rsp)
}

// ParsePostApiAuthResponse parses an HTTP response from a PostApiAuthWithResponse call
func ParsePostApiAuthResponse(rsp *http.Response) (*PostApiAuthResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostApiAuthResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest AuthResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseGetApiBuyItemResponse parses an HTTP response from a GetApiBuyItemWithResponse call
func ParseGetApiBuyItemResponse(rsp *http.Response) (*GetApiBuyItemResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetApiBuyItemResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParseGetApiInfoResponse parses an HTTP response from a GetApiInfoWithResponse call
func ParseGetApiInfoResponse(rsp *http.Response) (*GetApiInfoResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &GetApiInfoResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 200:
		var dest InfoResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON200 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}

// ParsePostApiSendCoinResponse parses an HTTP response from a PostApiSendCoinWithResponse call
func ParsePostApiSendCoinResponse(rsp *http.Response) (*PostApiSendCoinResponse, error) {
	bodyBytes, err := io.ReadAll(rsp.Body)
	defer func() { _ = rsp.Body.Close() }()
	if err != nil {
		return nil, err
	}

	response := &PostApiSendCoinResponse{
		Body:         bodyBytes,
		HTTPResponse: rsp,
	}

	switch {
	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 400:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON400 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 401:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON401 = &dest

	case strings.Contains(rsp.Header.Get("Content-Type"), "json") && rsp.StatusCode == 500:
		var dest ErrorResponse
		if err := json.Unmarshal(bodyBytes, &dest); err != nil {
			return nil, err
		}
		response.JSON500 = &dest

	}

	return response, nil
}
