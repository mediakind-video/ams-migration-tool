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

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
)

// AssetsClient contains the methods for the Assets group.
// Don't use this type directly, use NewAssetsClient() instead.
type AssetsClient struct {
	MkioClient
}

// NewAssetsClient creates a new instance of AssetsClient with the specified values.
// subscriptionName - The subscription (project) name for the .
// token - used to authorize requests. Usually a credential from azidentity.
// apiEndpoint - used to specify the MKIO API endpoint.
// options - pass nil to accept the default values.
func NewAssetsClient(ctx context.Context, subscriptionName string, token string, apiEndpoint string, options *ClientOptions) (*AssetsClient, error) {
	if options == nil {
		options = &ClientOptions{
			host: apiEndpoint,
		}
	}
	hc := &http.Client{}
	client := &AssetsClient{MkioClient{
		subscriptionName: subscriptionName,
		host:             options.host,
		token:            token,
		hc:               hc,
	},
	}

	// Test that our token is valid
	err := client.GetProfile(ctx)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// CreateOrUpdate - Creates or updates an Asset in the Media Services account
// If the operation fails it returns an error type.
// assetName - The Asset name.
// parameters - The request parameters
// options - AssetsClientCreateOrUpdateOptions contains the optional parameters for the AssetsClient.CreateOrUpdate method.
func (client *AssetsClient) CreateOrUpdate(ctx context.Context, assetName string, parameters *armmediaservices.Asset, options *armmediaservices.AssetsClientCreateOrUpdateOptions) (armmediaservices.AssetsClientCreateOrUpdateResponse, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, assetName, parameters, options)
	if err != nil {
		return armmediaservices.AssetsClientCreateOrUpdateResponse{}, err
	}

	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		return armmediaservices.AssetsClientCreateOrUpdateResponse{}, err
	}

	return client.createOrUpdateHandleResponse(resp)
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *AssetsClient) createOrUpdateCreateRequest(ctx context.Context, assetName string, parameters *armmediaservices.Asset, options *armmediaservices.AssetsClientCreateOrUpdateOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/assets/{assetName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{assetName}", url.PathEscape(assetName))
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
func (client *AssetsClient) createOrUpdateHandleResponse(resp *http.Response) (armmediaservices.AssetsClientCreateOrUpdateResponse, error) {
	result := armmediaservices.AssetsClientCreateOrUpdateResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.AssetsClientCreateOrUpdateResponse{}, err
	}
	if err := json.Unmarshal(body, &result.Asset); err != nil {
		return armmediaservices.AssetsClientCreateOrUpdateResponse{}, err
	}
	return result, nil
}

// Delete - Deletes an Asset in the Media Services account
// If the operation fails it returns an ResponseError type.
// assetName - The Asset name.
// options - AssetsClientDeleteOptions contains the optional parameters for the AssetsClient.Delete method.
func (client *AssetsClient) Delete(ctx context.Context, assetName string, options *armmediaservices.AssetsClientDeleteOptions) (armmediaservices.AssetsClientDeleteResponse, error) {
	req, err := client.deleteCreateRequest(ctx, assetName, options)
	if err != nil {
		return armmediaservices.AssetsClientDeleteResponse{}, err
	}

	// Try to do request, handle retries if tooManyRequests
	_, err = client.DoRequestWithBackoff(req)
	if err != nil {
		return armmediaservices.AssetsClientDeleteResponse{}, err
	}

	return armmediaservices.AssetsClientDeleteResponse{}, nil
}

// deleteCreateRequest creates the Delete request.
func (client *AssetsClient) deleteCreateRequest(ctx context.Context, assetName string, options *armmediaservices.AssetsClientDeleteOptions) (*http.Request, error) {

	urlPath := "/api/ams/{subscriptionName}/assets/{assetName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{assetName}", url.PathEscape(assetName))
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
// assetName - The Asset name.
// options - AssetClientGetOptions contains the optional parameters for the AssetClient.Get method.
func (client *AssetsClient) Get(ctx context.Context, assetName string, options *armmediaservices.AssetsClientGetOptions) (armmediaservices.AssetsClientGetResponse, error) {
	req, err := client.getCreateRequest(ctx, assetName, options)
	if err != nil {
		return armmediaservices.AssetsClientGetResponse{}, err
	}

	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		return armmediaservices.AssetsClientGetResponse{}, err
	}

	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *AssetsClient) getCreateRequest(ctx context.Context, assetName string, options *armmediaservices.AssetsClientGetOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/assets/{assetName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{assetName}", url.PathEscape(assetName))
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
func (client *AssetsClient) getHandleResponse(resp *http.Response) (armmediaservices.AssetsClientGetResponse, error) {
	result := armmediaservices.AssetsClientGetResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.AssetsClientGetResponse{}, err
	}
	if err := json.Unmarshal(body, &result.Asset); err != nil {
		return armmediaservices.AssetsClientGetResponse{}, err
	}
	return result, nil
}

// List - List Assets in the mk.io account
// If the operation fails it returns an *ResponseError type.
// options - AssetClientGetOptions contains the optional parameters for the AssetClient.Get method.
func (client *AssetsClient) List(ctx context.Context, options *armmediaservices.AssetsClientListOptions) (armmediaservices.AssetsClientListResponse, error) {
	req, err := client.listCreateRequest(ctx, options)
	if err != nil {
		return armmediaservices.AssetsClientListResponse{}, err
	}

	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		return armmediaservices.AssetsClientListResponse{}, err
	}

	return client.listHandleResponse(resp)
}

// listCreateRequest creates the list request.
func (client *AssetsClient) listCreateRequest(ctx context.Context, options *armmediaservices.AssetsClientListOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/assets"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	path, err := url.JoinPath(client.host, urlPath)
	if err != nil {
		return nil, err
	}

	// Apply filters to query
	filter := ""
	if options.Filter != nil {
		filter = `$filter=` + *options.Filter
	}
	q, err := url.ParseQuery(filter)
	if err == nil {
		path = path + "?" + q.Encode()
	}

	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-mkio-token", client.token)
	return req, nil
}

// listHandleResponse handles the List response.
func (client *AssetsClient) listHandleResponse(resp *http.Response) (armmediaservices.AssetsClientListResponse, error) {
	result := armmediaservices.AssetsClientListResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.AssetsClientListResponse{}, err
	}
	if err := json.Unmarshal(body, &result.AssetCollection); err != nil {
		return armmediaservices.AssetsClientListResponse{}, err
	}
	return result, nil
}

// lookupAssets  Get assets from mk.io
func (client *AssetsClient) LookupAssets(ctx context.Context, before string, after string) ([]*armmediaservices.Asset, error) {
	// Generate the filter
	filter := generateFilter(before, after)

	// If we have a filter apply it
	options := &armmediaservices.AssetsClientListOptions{Orderby: to.Ptr("properties/created")}
	if filter != "" {
		options.Filter = to.Ptr(filter)
	}
	req, err := client.List(ctx, options)
	if err != nil {
		return nil, err
	}

	assets := []*armmediaservices.Asset{}

	assets = append(assets, req.AssetCollection.Value...)

	return assets, nil
}
