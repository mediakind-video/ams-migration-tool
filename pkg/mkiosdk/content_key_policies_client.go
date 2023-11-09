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

// ContentKeyPoliciesClient contains the methods for the ContentKeyPolicies group.
// Don't use this type directly, use NewContentKeyPoliciesClient() instead.
type ContentKeyPoliciesClient struct {
	host             string
	subscriptionName string
	token            string
	hc               *http.Client
}

// NewContentKeyPoliciesClient creates a new instance of ContentKeyPoliciesClient with the specified values.
// subscriptionName - The subscription (project) name for the .
// token - used to authorize requests. Usually a credential from azidentity.
// apiEndpoint - used to specify the MKIO API endpoint.
// options - pass nil to accept the default values.
func NewContentKeyPoliciesClient(subscriptionName string, token string, apiEndpoint string, options *ClientOptions) (*ContentKeyPoliciesClient, error) {
	if options == nil {
		options = &ClientOptions{
			host: apiEndpoint,
		}
	}
	hc := &http.Client{}
	client := &ContentKeyPoliciesClient{
		subscriptionName: subscriptionName,
		host:             options.host,
		token:            token,
		hc:               hc,
	}
	return client, nil
}

// CreateOrUpdate - Creates or updates an ContentKeyPolicy in the Media Services account
// If the operation fails it returns an error type.
// contentKeyPolicyName - The contentKeyPolicy name.
// parameters - The request parameters
// options - ContentKeyPoliciesClientCreateOrUpdateOptions contains the optional parameters for the ContentKeyPoliciesClient.CreateOrUpdate method.
func (client *ContentKeyPoliciesClient) CreateOrUpdate(ctx context.Context, contentKeyPolicyName string, parameters *armmediaservices.ContentKeyPolicy, options *armmediaservices.ContentKeyPoliciesClientCreateOrUpdateOptions) (armmediaservices.ContentKeyPoliciesClientCreateOrUpdateResponse, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, contentKeyPolicyName, parameters, options)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientCreateOrUpdateResponse{}, err
	}
	resp, err := client.hc.Do(req)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientCreateOrUpdateResponse{}, err
	}
	if !HasStatusCode(resp, http.StatusOK, http.StatusCreated) {
		return armmediaservices.ContentKeyPoliciesClientCreateOrUpdateResponse{}, NewResponseError(resp)
	}
	return client.createOrUpdateHandleResponse(resp)
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *ContentKeyPoliciesClient) createOrUpdateCreateRequest(ctx context.Context, contentKeyPolicyName string, parameters *armmediaservices.ContentKeyPolicy, options *armmediaservices.ContentKeyPoliciesClientCreateOrUpdateOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/contentKeyPolicies/{contentKeyPolicyName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{contentKeyPolicyName}", url.PathEscape(contentKeyPolicyName))
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
func (client *ContentKeyPoliciesClient) createOrUpdateHandleResponse(resp *http.Response) (armmediaservices.ContentKeyPoliciesClientCreateOrUpdateResponse, error) {
	result := armmediaservices.ContentKeyPoliciesClientCreateOrUpdateResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientCreateOrUpdateResponse{}, err
	}
	if err := json.Unmarshal(body, &result.ContentKeyPolicy); err != nil {
		return armmediaservices.ContentKeyPoliciesClientCreateOrUpdateResponse{}, err
	}
	return result, nil
}

// Delete - Deletes an ContentKeyPolicy in the Media Services account
// If the operation fails it returns an ResponseError type.
// contentKeyPolicyName - The ContentKeyPolicy name.
// options - ContentKeyPoliciesClientDeleteOptions contains the optional parameters for the ContentKeyPoliciesClient.Delete method.
func (client *ContentKeyPoliciesClient) Delete(ctx context.Context, contentKeyPolicyName string, options *armmediaservices.ContentKeyPoliciesClientDeleteOptions) (armmediaservices.ContentKeyPoliciesClientDeleteResponse, error) {
	req, err := client.deleteCreateRequest(ctx, contentKeyPolicyName, options)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientDeleteResponse{}, err
	}
	resp, err := client.hc.Do(req)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientDeleteResponse{}, err
	}
	if !HasStatusCode(resp, http.StatusOK, http.StatusNoContent) {
		return armmediaservices.ContentKeyPoliciesClientDeleteResponse{}, NewResponseError(resp)
	}
	return armmediaservices.ContentKeyPoliciesClientDeleteResponse{}, nil
}

// deleteCreateRequest creates the Delete request.
func (client *ContentKeyPoliciesClient) deleteCreateRequest(ctx context.Context, contentKeyPolicyName string, options *armmediaservices.ContentKeyPoliciesClientDeleteOptions) (*http.Request, error) {

	urlPath := "/api/ams/{subscriptionName}/contentKeyPolicies/{contentKeyPolicyName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{contentKeyPolicyName}", url.PathEscape(contentKeyPolicyName))
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

// Get - Get the details of a ContentKeyPolicy in the Media Services account
// If the operation fails it returns an *ResponseError type.
// contentKeyPolicyName - The contentKeyPolicy name.
// options - ContentKeyPoliciesClientGetOptions contains the optional parameters for the ContentKeyPolicisClient.Get method.
func (client *ContentKeyPoliciesClient) Get(ctx context.Context, contentKeyPolicyName string, options *armmediaservices.ContentKeyPoliciesClientGetOptions) (armmediaservices.ContentKeyPoliciesClientGetResponse, error) {
	req, err := client.getCreateRequest(ctx, contentKeyPolicyName, options)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientGetResponse{}, err
	}
	resp, err := client.hc.Do(req)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientGetResponse{}, err
	}
	if !HasStatusCode(resp, http.StatusOK) {
		return armmediaservices.ContentKeyPoliciesClientGetResponse{}, NewResponseError(resp)
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *ContentKeyPoliciesClient) getCreateRequest(ctx context.Context, contentKeyPolicyName string, options *armmediaservices.ContentKeyPoliciesClientGetOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/contentKeyPolicies/{contentKeyPolicyName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{contentKeyPolicyName}", url.PathEscape(contentKeyPolicyName))
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
func (client *ContentKeyPoliciesClient) getHandleResponse(resp *http.Response) (armmediaservices.ContentKeyPoliciesClientGetResponse, error) {
	result := armmediaservices.ContentKeyPoliciesClientGetResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientGetResponse{}, err
	}
	if err := json.Unmarshal(body, &result.ContentKeyPolicy); err != nil {
		return armmediaservices.ContentKeyPoliciesClientGetResponse{}, err
	}
	return result, nil
}
