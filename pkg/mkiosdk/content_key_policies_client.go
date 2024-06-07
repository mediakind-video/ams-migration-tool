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

// ContentKeyPoliciesClient contains the methods for the ContentKeyPolicies group.
// Don't use this type directly, use NewContentKeyPoliciesClient() instead.
type ContentKeyPoliciesClient struct {
	MkioClient
}

// NewContentKeyPoliciesClient creates a new instance of ContentKeyPoliciesClient with the specified values.
// subscriptionName - The subscription (project) name for the .
// token - used to authorize requests. Usually a credential from azidentity.
// apiEndpoint - used to specify the MKIO API endpoint.
// options - pass nil to accept the default values.
func NewContentKeyPoliciesClient(ctx context.Context, subscriptionName string, token string, apiEndpoint string, options *ClientOptions) (*ContentKeyPoliciesClient, error) {
	if options == nil {
		options = &ClientOptions{
			host: apiEndpoint,
		}
	}
	hc := &http.Client{}
	client := &ContentKeyPoliciesClient{
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

// CreateOrUpdate - Creates or updates an ContentKeyPolicy in the Media Services account
// If the operation fails it returns an error type.
// contentKeyPolicyName - The contentKeyPolicy name.
// parameters - The request parameters
// options - ContentKeyPoliciesClientCreateOrUpdateOptions contains the optional parameters for the ContentKeyPoliciesClient.CreateOrUpdate method.
func (client *ContentKeyPoliciesClient) CreateOrUpdate(ctx context.Context, contentKeyPolicyName string, parameters *FPContentKeyPolicy, options *armmediaservices.ContentKeyPoliciesClientCreateOrUpdateOptions) (armmediaservices.ContentKeyPoliciesClientCreateOrUpdateResponse, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, contentKeyPolicyName, parameters, options)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientCreateOrUpdateResponse{}, err
	}

	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientCreateOrUpdateResponse{}, err
	}

	return client.createOrUpdateHandleResponse(resp)
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *ContentKeyPoliciesClient) createOrUpdateCreateRequest(ctx context.Context, contentKeyPolicyName string, parameters *FPContentKeyPolicy, options *armmediaservices.ContentKeyPoliciesClientCreateOrUpdateOptions) (*http.Request, error) {
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
	// Try to do request, handle retries if tooManyRequests
	_, err = client.DoRequestWithBackoff(req)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientDeleteResponse{}, err
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

	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientGetResponse{}, err
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

// GetPolicyPropertiesWithSWecrets - Get the details of a ContentKeyPolicy in the mk.io account
// If the operation fails it returns an *ResponseError type.
// contentKeyPolicyName - The contentKeyPolicy name.
// options - ContentKeyPoliciesClientGetOptions contains the optional parameters for the ContentKeyPolicisClient.Get method.
func (client *ContentKeyPoliciesClient) GetPolicyPropertiesWithSecrets(ctx context.Context, contentKeyPolicyName string, options *armmediaservices.ContentKeyPoliciesClientGetOptions) (armmediaservices.ContentKeyPoliciesClientGetResponse, error) {
	req, err := client.getWithSecretsCreateRequest(ctx, contentKeyPolicyName, options)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientGetResponse{}, err
	}

	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientGetResponse{}, err
	}

	return client.getWithSecretsHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *ContentKeyPoliciesClient) getWithSecretsCreateRequest(ctx context.Context, contentKeyPolicyName string, options *armmediaservices.ContentKeyPoliciesClientGetOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/contentKeyPolicies/{contentKeyPolicyName}/getPolicyPropertiesWithSecrets"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{contentKeyPolicyName}", url.PathEscape(contentKeyPolicyName))
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

// getHandleResponse handles the Get response.
func (client *ContentKeyPoliciesClient) getWithSecretsHandleResponse(resp *http.Response) (armmediaservices.ContentKeyPoliciesClientGetResponse, error) {
	result := armmediaservices.ContentKeyPoliciesClientGetResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientGetResponse{}, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return armmediaservices.ContentKeyPoliciesClientGetResponse{}, err
	}
	return result, nil
}

// List - List ContentKeyPolicy in the mk.io account
// If the operation fails it returns an *ResponseError type.
// options - ContentKeyPoliciesClientListOptions contains the optional parameters for the ContentKeyPolicisClient.Get method.
func (client *ContentKeyPoliciesClient) List(ctx context.Context, options *armmediaservices.ContentKeyPoliciesClientListOptions) (armmediaservices.ContentKeyPoliciesClientListResponse, error) {
	req, err := client.listCreateRequest(ctx, options)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientListResponse{}, err
	}

	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientListResponse{}, err
	}

	return client.listHandleResponse(resp)
}

// listCreateRequest creates the Get request.
func (client *ContentKeyPoliciesClient) listCreateRequest(ctx context.Context, options *armmediaservices.ContentKeyPoliciesClientListOptions) (*http.Request, error) {
	urlPath := "/api/ams/{subscriptionName}/contentKeyPolicies"
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

// listHandleResponse handles the list response.
func (client *ContentKeyPoliciesClient) listHandleResponse(resp *http.Response) (armmediaservices.ContentKeyPoliciesClientListResponse, error) {
	result := armmediaservices.ContentKeyPoliciesClientListResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.ContentKeyPoliciesClientListResponse{}, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return armmediaservices.ContentKeyPoliciesClientListResponse{}, err
	}
	return result, nil
}

// lookupContentKeyPolicies  Get content key policies from mk.io
func (client *ContentKeyPoliciesClient) LookupContentKeyPolicies(ctx context.Context) ([]*armmediaservices.ContentKeyPolicy, error) {

	req, err := client.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	ckp := []*armmediaservices.ContentKeyPolicy{}
	skipped := []string{}

	for _, v := range req.ContentKeyPolicyCollection.Value {
		// get content key policy with secrets
		props, err := client.GetPolicyPropertiesWithSecrets(ctx, *v.Name, nil)
		if err != nil {
			skipped = append(skipped, *v.Name)
			continue
		}
		ckp = append(ckp, &props.ContentKeyPolicy)
	}

	if len(skipped) > 0 {
		// Return error after we've looped through all the content key policies
		return ckp, fmt.Errorf("unable to get content key policy %v", skipped)
	}

	return ckp, nil
}
