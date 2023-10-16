package mkiosdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
)

// AssetFiltersClient contains the methods for the Asset Filters group.
// Don't use this type directly, use NewAssetFiltersClient() instead.
type AssetFiltersClient struct {
	host             string
	subscriptionName string
	token            string
	hc               *http.Client
}

// NewAssetFiltersClient creates a new instance of AssetFiltersClient with the specified values.
// subscriptionName - The subscription (project) name for the .
// token - used to authorize requests. Usually a credential from azidentity.
// options - pass nil to accept the default values.
func NewAssetFiltersClient(subscriptionName string, token string, options *ClientOptions) (*AssetFiltersClient, error) {
	if options == nil {
		options = &ClientOptions{
			host: "https://dev.io.mediakind.com",
		}
	}
	hc := &http.Client{}
	client := &AssetFiltersClient{
		subscriptionName: subscriptionName,
		host:             options.host,
		token:            token,
		hc:               hc,
	}
	return client, nil
}

// CreateOrUpdate - Creates or updates an Asset in the Media Services account
// If the operation fails it returns an error type.
// assetFilterName - The Asset name.
// parameters - The request parameters
// options - AssetFiltersClientCreateOrUpdateOptions contains the optional parameters for the AssetFiltersClient.CreateOrUpdate method.
func (client *AssetFiltersClient) CreateOrUpdate(ctx context.Context, assetName string, assetFilterName string, parameters *armmediaservices.AssetFilter, options *armmediaservices.AssetFiltersClientCreateOrUpdateOptions) (armmediaservices.AssetFiltersClientCreateOrUpdateResponse, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, assetName, assetFilterName, parameters, options)
	if err != nil {
		return armmediaservices.AssetFiltersClientCreateOrUpdateResponse{}, err
	}
	resp, err := client.hc.Do(req)
	if err != nil {
		return armmediaservices.AssetFiltersClientCreateOrUpdateResponse{}, err
	}
	if !HasStatusCode(resp, http.StatusOK, http.StatusCreated) {
		return armmediaservices.AssetFiltersClientCreateOrUpdateResponse{}, NewResponseError(resp)
	}
	return client.createOrUpdateHandleResponse(resp)
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *AssetFiltersClient) createOrUpdateCreateRequest(ctx context.Context, assetName string, assetFilterName string, parameters *armmediaservices.AssetFilter, options *armmediaservices.AssetFiltersClientCreateOrUpdateOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/assets/{assetName}/filters/{assetFilterName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{assetName}", url.PathEscape(assetName))
	urlPath = strings.ReplaceAll(urlPath, "{assetFilterName}", url.PathEscape(assetFilterName))
	body, err := json.Marshal(parameters)
	if err != nil {
		return nil, err
	}
	path, err := url.JoinPath(client.host, urlPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPut, path, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-mkio-token", client.token)
	return req, nil
}

// createOrUpdateHandleResponse handles the CreateOrUpdate response.
func (client *AssetFiltersClient) createOrUpdateHandleResponse(resp *http.Response) (armmediaservices.AssetFiltersClientCreateOrUpdateResponse, error) {
	result := armmediaservices.AssetFiltersClientCreateOrUpdateResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.AssetFiltersClientCreateOrUpdateResponse{}, err
	}
	if err := json.Unmarshal(body, &result.AssetFilter); err != nil {
		return armmediaservices.AssetFiltersClientCreateOrUpdateResponse{}, err
	}
	return result, nil
}

// Delete - Deletes an Asset in the Media Services account
// If the operation fails it returns an ResponseError type.
// assetFilterName - The Asset name.
// options - AssetFiltersClientDeleteOptions contains the optional parameters for the AssetFiltersClient.Delete method.
func (client *AssetFiltersClient) Delete(ctx context.Context, assetFilterName string, options *armmediaservices.AssetFiltersClientDeleteOptions) (armmediaservices.AssetFiltersClientDeleteResponse, error) {
	req, err := client.deleteCreateRequest(ctx, assetFilterName, options)
	if err != nil {
		return armmediaservices.AssetFiltersClientDeleteResponse{}, err
	}
	resp, err := client.hc.Do(req)
	if err != nil {
		return armmediaservices.AssetFiltersClientDeleteResponse{}, err
	}
	if !HasStatusCode(resp, http.StatusOK, http.StatusNoContent) {
		return armmediaservices.AssetFiltersClientDeleteResponse{}, NewResponseError(resp)
	}
	return armmediaservices.AssetFiltersClientDeleteResponse{}, nil
}

// deleteCreateRequest creates the Delete request.
func (client *AssetFiltersClient) deleteCreateRequest(ctx context.Context, assetFilterName string, options *armmediaservices.AssetFiltersClientDeleteOptions) (*http.Request, error) {

	urlPath := "/api/ams/{subscriptionName}/assetFilters/{assetFilterName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{assetFilterName}", url.PathEscape(assetFilterName))
	path, err := url.JoinPath(client.host, urlPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-mkio-token", client.token)

	return req, nil
}

// Get - Get the details of a Asset in the Media Services account
// If the operation fails it returns an *ResponseError type.
// assetFilterName - The Asset name.
// options - AssetClientGetOptions contains the optional parameters for the AssetClient.Get method.
func (client *AssetFiltersClient) Get(ctx context.Context, assetFilterName string, options *armmediaservices.AssetFiltersClientGetOptions) (armmediaservices.AssetFiltersClientGetResponse, error) {
	req, err := client.getCreateRequest(ctx, assetFilterName, options)
	if err != nil {
		return armmediaservices.AssetFiltersClientGetResponse{}, err
	}
	resp, err := client.hc.Do(req)
	if err != nil {
		return armmediaservices.AssetFiltersClientGetResponse{}, err
	}
	if !HasStatusCode(resp, http.StatusOK) {
		return armmediaservices.AssetFiltersClientGetResponse{}, NewResponseError(resp)
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *AssetFiltersClient) getCreateRequest(ctx context.Context, assetFilterName string, options *armmediaservices.AssetFiltersClientGetOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/assetFilters/{assetFilterName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{assetFilterName}", url.PathEscape(assetFilterName))
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
func (client *AssetFiltersClient) getHandleResponse(resp *http.Response) (armmediaservices.AssetFiltersClientGetResponse, error) {
	result := armmediaservices.AssetFiltersClientGetResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.AssetFiltersClientGetResponse{}, err
	}
	if err := json.Unmarshal(body, &result.AssetFilter); err != nil {
		return armmediaservices.AssetFiltersClientGetResponse{}, err
	}
	return result, nil
}
