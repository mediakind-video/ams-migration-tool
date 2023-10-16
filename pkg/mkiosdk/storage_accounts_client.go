package mkiosdk

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
)

// StorageAccountsClient contains the methods for the StorageAccounts group.
// Don't use this type directly, use NewStorageAccountsClient() instead.
type StorageAccountsClient struct {
	host             string
	customerId       string
	subscriptionName string
	token            string
	hc               *http.Client
}

// NewStorageAccountsClient creates a new instance of StorageAccountsClient with the specified values.
// customerId - The customer ID for the storage account.
// subscriptionName - The subscription Name for the storage account.
// token - used to authorize requests.
// options - pass nil to accept the default values.
func NewStorageAccountsClient(customerId string, subscriptionName string, token string, options *ClientOptions) (*StorageAccountsClient, error) {
	if options == nil {
		options = &ClientOptions{
			host: "https://dev.io.mediakind.com",
		}
	}
	hc := &http.Client{}
	client := &StorageAccountsClient{
		customerId:       customerId,
		subscriptionName: subscriptionName,
		host:             options.host,
		token:            token,
		hc:               hc,
	}
	return client, nil
}

// Get - Get the details of a Asset in the Media Services account
// If the operation fails it returns an *ResponseError type.
// assetName - The Asset name.
// options - AssetClientGetOptions contains the optional parameters for the AssetClient.Get method.
func (client *StorageAccountsClient) Get(ctx context.Context, options *armmediaservices.ClientGetOptions) (armmediaservices.ClientGetResponse, error) {
	req, err := client.getCreateRequest(ctx, options)
	if err != nil {
		return armmediaservices.ClientGetResponse{}, err
	}
	resp, err := client.hc.Do(req)
	if err != nil {
		return armmediaservices.ClientGetResponse{}, err
	}
	if !HasStatusCode(resp, http.StatusOK) {
		return armmediaservices.ClientGetResponse{}, NewResponseError(resp)
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *StorageAccountsClient) getCreateRequest(ctx context.Context, options *armmediaservices.ClientGetOptions) (*http.Request, error) {
	urlPath := "/api/accounts/{customerId}/subscription/{subscriptionName}/storage/"

	if client.customerId == "" {
		return nil, errors.New("parameter client.customerId cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{customerId}", url.PathEscape(client.customerId))
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	path, err := url.JoinPath(client.host, urlPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-mkio-token", client.token)
	return req, nil
}

// getHandleResponse handles the Get response.
func (client *StorageAccountsClient) getHandleResponse(resp *http.Response) (armmediaservices.ClientGetResponse, error) {
	result := armmediaservices.ClientGetResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.ClientGetResponse{}, err
	}
	if err := json.Unmarshal(body, &result.Properties.StorageAccounts); err != nil {
		return armmediaservices.ClientGetResponse{}, err
	}
	return result, nil
}
