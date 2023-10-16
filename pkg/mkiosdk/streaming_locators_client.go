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

// StreamingLocatorsClient contains the methods for the StreamingLocators group.
// Don't use this type directly, use NewStreamingLocatorsClient() instead.
type StreamingLocatorsClient struct {
	host             string
	subscriptionName string
	token            string
	hc               *http.Client
}

// NewStreamingLocatorsClient creates a new instance of StreamingLocatorsClient with the specified values.
// subscriptionName - The subscription (project) name for the .
// token - used to authorize requests. Usually a credential from azidentity.
// options - pass nil to accept the default values.
func NewStreamingLocatorsClient(subscriptionName string, token string, options *ClientOptions) (*StreamingLocatorsClient, error) {
	if options == nil {
		options = &ClientOptions{
			host: "https://dev.io.mediakind.com",
		}
	}
	hc := &http.Client{}
	client := &StreamingLocatorsClient{
		subscriptionName: subscriptionName,
		host:             options.host,
		token:            token,
		hc:               hc,
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
	resp, err := client.hc.Do(req)
	if err != nil {
		return armmediaservices.StreamingLocatorsClientCreateResponse{}, err
	}
	if !HasStatusCode(resp, http.StatusOK, http.StatusCreated) {
		return armmediaservices.StreamingLocatorsClientCreateResponse{}, NewResponseError(resp)
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
	resp, err := client.hc.Do(req)
	if err != nil {
		return armmediaservices.AssetsClientDeleteResponse{}, err
	}
	if !HasStatusCode(resp, http.StatusOK, http.StatusNoContent) {
		return armmediaservices.AssetsClientDeleteResponse{}, NewResponseError(resp)
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
	resp, err := client.hc.Do(req)
	if err != nil {
		return armmediaservices.StreamingLocatorsClientGetResponse{}, err
	}
	if !HasStatusCode(resp, http.StatusOK) {
		return armmediaservices.StreamingLocatorsClientGetResponse{}, NewResponseError(resp)
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
	resp, err := client.hc.Do(req)
	if err != nil {
		return armmediaservices.StreamingLocatorsClientListPathsResponse{}, err
	}
	if !HasStatusCode(resp, http.StatusOK) {
		return armmediaservices.StreamingLocatorsClientListPathsResponse{}, NewResponseError(resp)
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
