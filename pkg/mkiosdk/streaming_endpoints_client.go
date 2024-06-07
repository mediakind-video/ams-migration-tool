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

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
)

// StreamingEndpointsClient contains the methods for the StreamingEndpoints group.
// Don't use this type directly, use NewStreamingEndpointsClient() instead.
type StreamingEndpointsClient struct {
	MkioClient
}

// NewStreamingEndpointsClient creates a new instance of StreamingEndpointsClient with the specified values.
// subscriptionID - The unique identifier for a Microsoft Azure subscription.
// token - used to authorize requests.
// apiEndpoint - used to specify the MKIO API endpoint.
// options - pass nil to accept the default values.
func NewStreamingEndpointsClient(ctx context.Context, subscriptionName string, token string, apiEndpoint string, options *ClientOptions) (*StreamingEndpointsClient, error) {
	if options == nil {
		options = &ClientOptions{
			host: apiEndpoint,
		}
	}
	hc := &http.Client{}
	client := &StreamingEndpointsClient{
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

// CreateOrUpdate - Creates or updates an StreamingEndpoint in the Media Services account
// If the operation fails it returns an error type.
// streamingEndpointName - The StreamingEndpoint name.
// parameters - The request parameters
// options - StreamingEndpointClientCreateOrUpdateOptions contains the optional parameters for the StreamingEndpointsClient.CreateOrUpdate method.
func (client *StreamingEndpointsClient) CreateOrUpdate(ctx context.Context, streamingEndpointName string, parameters armmediaservices.StreamingEndpoint, options *armmediaservices.StreamingEndpointsClientBeginCreateOptions) (armmediaservices.StreamingEndpointsClientCreateResponse, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, streamingEndpointName, parameters, options)
	if err != nil {
		return armmediaservices.StreamingEndpointsClientCreateResponse{}, fmt.Errorf("unable to generate Create/Update request: %v", err)
	}
	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		// We hit some error we and failed retry loop. Return error
		return armmediaservices.StreamingEndpointsClientCreateResponse{}, err

	}
	return client.createOrUpdateHandleResponse(resp)
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *StreamingEndpointsClient) createOrUpdateCreateRequest(ctx context.Context, streamingEndpointName string, parameters armmediaservices.StreamingEndpoint, options *armmediaservices.StreamingEndpointsClientBeginCreateOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingEndpoints/{streamingEndpointName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{streamingEndpointName}", url.PathEscape(streamingEndpointName))
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
func (client *StreamingEndpointsClient) createOrUpdateHandleResponse(resp *http.Response) (armmediaservices.StreamingEndpointsClientCreateResponse, error) {
	result := armmediaservices.StreamingEndpointsClientCreateResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.StreamingEndpointsClientCreateResponse{}, err
	}
	if err := json.Unmarshal(body, &result.StreamingEndpoint); err != nil {
		return armmediaservices.StreamingEndpointsClientCreateResponse{}, err
	}
	return result, nil
}

// Get - Get the details of a Streaming Endpoint in the Media Services account
// If the operation fails it returns an *ResponseError type.
// streamingEndpointName - The StreamingEndpoint name.
// options - StreamingEndpointsClientGetOptions contains the optional parameters for the StreamingEndpointsClient.Get method.
func (client *StreamingEndpointsClient) Get(ctx context.Context, streamingEndpointName string, options *armmediaservices.StreamingEndpointsClientGetOptions) (armmediaservices.StreamingEndpointsClientGetResponse, error) {
	req, err := client.getCreateRequest(ctx, streamingEndpointName, options)
	if err != nil {
		return armmediaservices.StreamingEndpointsClientGetResponse{}, err
	}
	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		// We hit some error we and failed retry loop. Return error
		return armmediaservices.StreamingEndpointsClientGetResponse{}, err
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *StreamingEndpointsClient) getCreateRequest(ctx context.Context, streamingEndpointName string, options *armmediaservices.StreamingEndpointsClientGetOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingEndpoints/{streamingEndpointName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{streamingEndpointName}", url.PathEscape(streamingEndpointName))
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
func (client *StreamingEndpointsClient) getHandleResponse(resp *http.Response) (armmediaservices.StreamingEndpointsClientGetResponse, error) {
	result := armmediaservices.StreamingEndpointsClientGetResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.StreamingEndpointsClientGetResponse{}, err
	}
	if err := json.Unmarshal(body, &result.StreamingEndpoint); err != nil {
		return armmediaservices.StreamingEndpointsClientGetResponse{}, err
	}
	return result, nil
}

// List - List the Streaming Endpoints in the Media Services account
// If the operation fails it returns an *ResponseError type.
// options - StreamingEndpointsClientGetOptions contains the optional parameters for the StreamingEndpointsClient.Get method.
func (client *StreamingEndpointsClient) List(ctx context.Context, options *armmediaservices.StreamingEndpointsClientListOptions) (armmediaservices.StreamingEndpointsClientListResponse, error) {
	req, err := client.listCreateRequest(ctx, options)
	if err != nil {
		return armmediaservices.StreamingEndpointsClientListResponse{}, err
	}
	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		// We hit some error we and failed retry loop. Return error
		return armmediaservices.StreamingEndpointsClientListResponse{}, err
	}
	return client.listHandleResponse(resp)
}

// listCreateRequest creates the List request.
func (client *StreamingEndpointsClient) listCreateRequest(ctx context.Context, options *armmediaservices.StreamingEndpointsClientListOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingEndpoints"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionID cannot be empty")
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

// listHandleResponse handles the List response.
func (client *StreamingEndpointsClient) listHandleResponse(resp *http.Response) (armmediaservices.StreamingEndpointsClientListResponse, error) {
	result := armmediaservices.StreamingEndpointsClientListResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.StreamingEndpointsClientListResponse{}, err
	}
	if err := json.Unmarshal(body, &result.StreamingEndpointListResult); err != nil {
		return armmediaservices.StreamingEndpointsClientListResponse{}, err
	}
	return result, nil
}

// Delete - Deletes a Streaming Endpoint in the Media Services account
// If the operation fails it returns an *ResponseError type.
// streamingEndpointName - The StreamingEndpoint name.
// options - StreamingEndpointsClientDeleteOptions contains the optional parameters for the StreamingEndpointsClient.Delete
// method.
func (client *StreamingEndpointsClient) Delete(ctx context.Context, streamingEndpointName string, options *armmediaservices.StreamingEndpointsClientBeginDeleteOptions) (armmediaservices.AssetsClientDeleteResponse, error) {
	req, err := client.deleteCreateRequest(ctx, streamingEndpointName, options)
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
func (client *StreamingEndpointsClient) deleteCreateRequest(ctx context.Context, streamingEndpointName string, options *armmediaservices.StreamingEndpointsClientBeginDeleteOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingEndpoints/{streamingEndpointName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{streamingEndpointName}", url.PathEscape(streamingEndpointName))
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

// lookupStreamingEndpoints Get streaming endpoints from mk.io. Remove pagination
func (client *StreamingEndpointsClient) LookupStreamingEndpoints(ctx context.Context) ([]*armmediaservices.StreamingEndpoint, error) {

	req, err := client.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	se := []*armmediaservices.StreamingEndpoint{}

	se = append(se, req.StreamingEndpointListResult.Value...)

	return se, nil
}
