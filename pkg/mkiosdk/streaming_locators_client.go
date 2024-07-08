package mkiosdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
)

// StreamingLocatorsClient contains the methods for the StreamingLocators group.
// Don't use this type directly, use NewStreamingLocatorsClient() instead.
type StreamingLocatorsClient struct {
	MkioClient
}

// NewStreamingLocatorsClient creates a new instance of StreamingLocatorsClient with the specified values.
// subscriptionName - The subscription (project) name for the .
// token - used to authorize requests. Usually a credential from azidentity.
// apiEndpoint - used to specify the MKIO API endpoint.
// options - pass nil to accept the default values.
func NewStreamingLocatorsClient(ctx context.Context, subscriptionName string, token string, apiEndpoint string, options *ClientOptions) (*StreamingLocatorsClient, error) {
	if options == nil {
		options = &ClientOptions{
			host: apiEndpoint,
		}
	}
	hc := &http.Client{}
	client := &StreamingLocatorsClient{
		MkioClient{
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

// CreateOrUpdate - Creates or updates an StreamingLocators in the Media Services account
// If the operation fails it returns an error type.
// streamingLocatorName - The StreamingLocator name.
// parameters - The request parameters
// options - StreamingLocatorsClientCreateOrUpdateOptions contains the optional parameters for the StreamingLocatorsClient.CreateOrUpdate method.
func (client *StreamingLocatorsClient) CreateOrUpdate(ctx context.Context, streamingLocatorName string, parameters armmediaservices.StreamingLocator, options *armmediaservices.StreamingLocatorsClientCreateOptions) (armmediaservices.StreamingLocatorsClientCreateResponse, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, streamingLocatorName, parameters, options)
	if err != nil {
		return armmediaservices.StreamingLocatorsClientCreateResponse{}, fmt.Errorf("unable to generate Create/Update request: %v", err)
	}
	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		// We hit some error we and failed retry loop. Return error
		return armmediaservices.StreamingLocatorsClientCreateResponse{}, err
	}
	return client.createOrUpdateHandleResponse(resp)
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *StreamingLocatorsClient) createOrUpdateCreateRequest(ctx context.Context, streamingLocatorName string, parameters armmediaservices.StreamingLocator, options *armmediaservices.StreamingLocatorsClientCreateOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingLocators/{streamingLocatorName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{streamingLocatorName}", url.PathEscape(streamingLocatorName))
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
func (client *StreamingLocatorsClient) createOrUpdateHandleResponse(resp *http.Response) (armmediaservices.StreamingLocatorsClientCreateResponse, error) {
	result := armmediaservices.StreamingLocatorsClientCreateResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.StreamingLocatorsClientCreateResponse{}, err
	}
	if err := json.Unmarshal(body, &result.StreamingLocator); err != nil {
		return armmediaservices.StreamingLocatorsClientCreateResponse{}, err
	}
	return result, nil
}

// Delete - Deletes a Streaming Locator in the Media Services account
// If the operation fails it returns an *ResponseError type.
// streamingLocatorName - The StreamingLocator name.
// options - StreamingLocatorsClientDeleteOptions contains the optional parameters for the StreamingLocatorsClient.Delete
// method.
func (client *StreamingLocatorsClient) Delete(ctx context.Context, streamingLocatorName string, options *armmediaservices.StreamingLocatorsClientDeleteOptions) (armmediaservices.AssetsClientDeleteResponse, error) {
	req, err := client.deleteCreateRequest(ctx, streamingLocatorName, options)
	if err != nil {
		return armmediaservices.AssetsClientDeleteResponse{}, err
	}
	// Try to do request, handle retries if tooManyRequests
	_, err = client.DoRequestWithBackoff(req)
	if err != nil {
		// We hit some error we and failed retry loop. Return error
		return armmediaservices.AssetsClientDeleteResponse{}, err
	}
	return armmediaservices.AssetsClientDeleteResponse{}, nil
}

// deleteCreateRequest creates the Delete request.
func (client *StreamingLocatorsClient) deleteCreateRequest(ctx context.Context, streamingLocatorName string, options *armmediaservices.StreamingLocatorsClientDeleteOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingLocators/{streamingLocatorName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{streamingLocatorName}", url.PathEscape(streamingLocatorName))
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

// Get - Get the details of a Streaming Locator in the Media Services account
// If the operation fails it returns an *ResponseError type.
// streamingLocatorName - The StreamingLocator name.
// options - StreamingLocatorsClientGetOptions contains the optional parameters for the StreamingLocatorsClient.Get method.
func (client *StreamingLocatorsClient) Get(ctx context.Context, streamingLocatorName string, options *armmediaservices.StreamingLocatorsClientGetOptions) (armmediaservices.StreamingLocatorsClientGetResponse, error) {
	req, err := client.getCreateRequest(ctx, streamingLocatorName, options)
	if err != nil {
		return armmediaservices.StreamingLocatorsClientGetResponse{}, err
	}
	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		// We hit some error we and failed retry loop. Return error
		return armmediaservices.StreamingLocatorsClientGetResponse{}, err
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *StreamingLocatorsClient) getCreateRequest(ctx context.Context, streamingLocatorName string, options *armmediaservices.StreamingLocatorsClientGetOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingLocators/{streamingLocatorName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{streamingLocatorName}", url.PathEscape(streamingLocatorName))
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
func (client *StreamingLocatorsClient) getHandleResponse(resp *http.Response) (armmediaservices.StreamingLocatorsClientGetResponse, error) {
	result := armmediaservices.StreamingLocatorsClientGetResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.StreamingLocatorsClientGetResponse{}, err
	}
	if err := json.Unmarshal(body, &result.StreamingLocator); err != nil {
		return armmediaservices.StreamingLocatorsClientGetResponse{}, err
	}
	return result, nil
}

// ListPaths - List Paths supported by this Streaming Locator
// If the operation fails it returns an *ResponseError type.
// streamingLocatorName - The Streaming Locator name.
// options - StreamingLocatorsClientListPathsOptions contains the optional parameters for the StreamingLocatorsClient.ListPaths
// method.
func (client *StreamingLocatorsClient) ListPaths(ctx context.Context, streamingLocatorName string, options *armmediaservices.StreamingLocatorsClientListPathsOptions) (armmediaservices.StreamingLocatorsClientListPathsResponse, error) {
	req, err := client.listPathsCreateRequest(ctx, streamingLocatorName, options)
	if err != nil {
		return armmediaservices.StreamingLocatorsClientListPathsResponse{}, err
	}
	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		// We hit some error we and failed retry loop. Return error
		return armmediaservices.StreamingLocatorsClientListPathsResponse{}, err
	}

	return client.listPathsHandleResponse(resp)
}

// listPathsCreateRequest creates the ListPaths request.
func (client *StreamingLocatorsClient) listPathsCreateRequest(ctx context.Context, streamingLocatorName string, options *armmediaservices.StreamingLocatorsClientListPathsOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingLocators/{streamingLocatorName}/listPaths"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{streamingLocatorName}", url.PathEscape(streamingLocatorName))
	path, err := url.JoinPath(client.host, urlPath)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-mkio-token", client.token)

	return req, nil
}

// listPathsHandleResponse handles the ListPaths response.
func (client *StreamingLocatorsClient) listPathsHandleResponse(resp *http.Response) (armmediaservices.StreamingLocatorsClientListPathsResponse, error) {
	result := armmediaservices.StreamingLocatorsClientListPathsResponse{}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.StreamingLocatorsClientListPathsResponse{}, err
	}
	if err := json.Unmarshal(body, &result.ListPathsResponse); err != nil {
		return armmediaservices.StreamingLocatorsClientListPathsResponse{}, err
	}
	return result, nil
}

// List - List Streaming Locators
// If the operation fails it returns an *ResponseError type.
// options - StreamingLocatorsClientListOptions contains the optional parameters for the StreamingLocatorsClient.List
// method.
func (client *StreamingLocatorsClient) List(ctx context.Context, options *armmediaservices.StreamingLocatorsClientListOptions) (armmediaservices.StreamingLocatorsClientListResponse, error) {
	req, err := client.listCreateRequest(ctx, options)
	if err != nil {
		return armmediaservices.StreamingLocatorsClientListResponse{}, err
	}
	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		// We hit some error we and failed retry loop. Return error
		return armmediaservices.StreamingLocatorsClientListResponse{}, err
	}

	return client.listHandleResponse(resp)
}

// listCreateRequest creates the ListPaths request.
func (client *StreamingLocatorsClient) listCreateRequest(ctx context.Context, options *armmediaservices.StreamingLocatorsClientListOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingLocators"
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
func (client *StreamingLocatorsClient) listHandleResponse(resp *http.Response) (armmediaservices.StreamingLocatorsClientListResponse, error) {
	result := armmediaservices.StreamingLocatorsClientListResponse{}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.StreamingLocatorsClientListResponse{}, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return armmediaservices.StreamingLocatorsClientListResponse{}, err
	}
	return result, nil
}

// lookupStreamingLocators Get streaming locators from mk.io. Remove pagination
func (client *StreamingLocatorsClient) LookupStreamingLocators(ctx context.Context, before string, after string) ([]*armmediaservices.StreamingLocator, error) {

	// Generate the filter
	filter := generateFilter(before, after)

	// If we have a filter apply it
	options := &armmediaservices.StreamingLocatorsClientListOptions{Orderby: to.Ptr("properties/created")}
	if filter != "" {
		options.Filter = to.Ptr(filter)
	}

	req, err := client.List(ctx, options)
	if err != nil {
		return nil, err
	}

	sl := []*armmediaservices.StreamingLocator{}

	// Unlike in Azure, we don't need additional calls to get content keys
	sl = append(sl, req.StreamingLocatorCollection.Value...)

	return sl, nil
}
