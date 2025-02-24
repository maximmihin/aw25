package testclient

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"
)

type HttpClient struct {
	Client *ClientWithResponses
}

func NewTestClient(t *testing.T, baseUrl string) HttpClient {
	usClient, err := NewClientWithResponses(baseUrl, WithHTTPClient(http.DefaultClient))
	require.NoError(t, err, "fail to create http client for user service")

	t.Logf("created user service http client with baseUrl %s", baseUrl)
	return HttpClient{
		Client: usClient,
	}
}

type AuthParams struct {
	Password string
	Username string
}

func (r HttpClient) Auth(t *testing.T, p AuthParams) *PostApiAuthResponse {
	t.Helper()
	t.Logf("Auth send request with id %s", uuid.New().String())

	res, err := r.Client.PostApiAuthWithResponse(t.Context(), PostApiAuthJSONRequestBody{
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

func (r HttpClient) BuyMerch(t *testing.T, p BuyMerchParams) *GetApiBuyItemResponse {
	t.Helper()
	t.Logf("BuyMerch send request with id %s", uuid.New().String())

	res, err := r.Client.GetApiBuyItemWithResponse(context.TODO(), p.MerchItem, WithBearer(p.Auth))
	require.NoError(t, err)
	return res
}

type SendCoinParams struct {
	Auth   string
	Amount int
	ToUser string
}

func (r HttpClient) SendCoins(t *testing.T, p SendCoinParams) *PostApiSendCoinResponse {
	t.Helper()
	t.Logf("SendCoins send request with id %s", uuid.New().String())

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

func (r HttpClient) Info(t *testing.T, p InfoParams) *GetApiInfoResponse {
	t.Helper()
	t.Logf("Info send request with id %s", uuid.New().String())

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
