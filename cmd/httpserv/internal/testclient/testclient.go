package testclient

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

//go:generate oapi-codegen --config=oapi_codegen_client.yaml ../../../../api/v1.yaml

type ServiceClient struct {
	Client *ClientWithResponses
}

func NewTestClient(t *testing.T, host, port string) ServiceClient {
	baseUrl := fmt.Sprintf("http://%s:%s", host, port)

	usClient, err := NewClientWithResponses(baseUrl, WithHTTPClient(http.DefaultClient))
	require.NoError(t, err, "fail to create http client for user service")

	t.Logf("created user service http client with baseUrl %s", baseUrl)
	return ServiceClient{
		Client: usClient,
	}
}

type AuthParams struct {
	Password string
	Username string
}

func (r ServiceClient) Auth(t *testing.T, p AuthParams) *PostApiAuthResponse {
	t.Helper()
	reqId := uuid.NewSHA1(uuid.NameSpaceURL, []byte(t.Name()+"Auth"))
	t.Logf("Auth send request with id %s", reqId.String())

	res, err := r.Client.PostApiAuthWithResponse(context.TODO(), PostApiAuthJSONRequestBody{
		Password: p.Password,
		Username: p.Username,
	})
	require.NoError(t, err)
	return res
}

type BuyMerchParams struct {
	Auth      string
	MerchItem string
}

func (r ServiceClient) BuyMerch(t *testing.T, p BuyMerchParams) *GetApiBuyItemResponse {
	t.Helper()
	reqId := uuid.NewSHA1(uuid.NameSpaceURL, []byte(t.Name()+"BuyMerch"))
	t.Logf("BuyMerch send request with id %s", reqId.String())

	res, err := r.Client.GetApiBuyItemWithResponse(context.TODO(), p.MerchItem, WithBearer(p.Auth))
	require.NoError(t, err)
	return res
}

type SendCoinParams struct {
	Auth   string
	Amount int
	ToUser string
}

func (r ServiceClient) SendCoins(t *testing.T, p SendCoinParams) *PostApiSendCoinResponse {
	t.Helper()
	reqId := uuid.NewSHA1(uuid.NameSpaceURL, []byte(t.Name()+"SendCoins"))
	t.Logf("SendCoins send request with id %s", reqId.String())

	res, err := r.Client.PostApiSendCoinWithResponse(context.TODO(), PostApiSendCoinJSONRequestBody{
		Amount: p.Amount,
		ToUser: p.ToUser,
	}, WithBearer(p.Auth))

	require.NoError(t, err)
	return res
}

type InfoParams struct {
	JwtToken string
}

func (r ServiceClient) Info(t *testing.T, p InfoParams) *GetApiInfoResponse {
	t.Helper()
	reqId := uuid.NewSHA1(uuid.NameSpaceURL, []byte(t.Name()+"Info"))
	t.Logf("Info send request with id %s", reqId.String())

	res, err := r.Client.GetApiInfoWithResponse(context.TODO(), WithBearer(p.JwtToken))
	require.NoError(t, err)
	return res
}

func WithBearer(accessToken string) RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		req.Header.Set("Authorization", "Bearer "+accessToken)
		return nil
	}
}
