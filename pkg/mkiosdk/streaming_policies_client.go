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

// StreamingPoliciesClient contains the methods for the StreamingPolicies group.
// Don't use this type directly, use NewStreamingPoliciesClient() instead.
type StreamingPoliciesClient struct {
	MkioClient
}

// NewStreamingPoliciesClient creates a new instance of StreamingPoliciesClient with the specified values.
// subscriptionName - The subscription (project) name for the .
// token - used to authorize requests. Usually a credential from azidentity.
// apiEndpoint - used to specify the MKIO API endpoint.
// options - pass nil to accept the default values.
func NewStreamingPoliciesClient(ctx context.Context, subscriptionName string, token string, apiEndpoint string, options *ClientOptions) (*StreamingPoliciesClient, error) {
	if options == nil {
		options = &ClientOptions{
			host: apiEndpoint,
		}
	}
	hc := &http.Client{}
	client := &StreamingPoliciesClient{
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

// CreateOrUpdate - Creates or updates an StreamingPolicies in the Media Services account
// If the operation fails it returns an error type.
// streamingPolicyName - The StreamingPolicy name.
// parameters - The request parameters
// options - StreamingPoliciesClientCreateOrUpdateOptions contains the optional parameters for the StreamingPoliciesClient.CreateOrUpdate method.
func (client *StreamingPoliciesClient) CreateOrUpdate(ctx context.Context, streamingPolicyName string, parameters armmediaservices.StreamingPolicy, options *armmediaservices.StreamingPoliciesClientCreateOptions) (armmediaservices.StreamingPoliciesClientCreateResponse, error) {
	req, err := client.createOrUpdateCreateRequest(ctx, streamingPolicyName, parameters, options)
	if err != nil {
		return armmediaservices.StreamingPoliciesClientCreateResponse{}, fmt.Errorf("unable to generate Create/Update request: %v", err)
	}
	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		// We hit some error we and failed retry loop. Return error
		return armmediaservices.StreamingPoliciesClientCreateResponse{}, err
	}
	return client.createOrUpdateHandleResponse(resp)
}

// createOrUpdateCreateRequest creates the CreateOrUpdate request.
func (client *StreamingPoliciesClient) createOrUpdateCreateRequest(ctx context.Context, streamingPolicyName string, parameters armmediaservices.StreamingPolicy, options *armmediaservices.StreamingPoliciesClientCreateOptions) (*Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingPolicies/{streamingPolicyName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{streamingPolicyName}", url.PathEscape(streamingPolicyName))
	body, err := json.Marshal(parameters)
	if err != nil {
		return nil, err
	}
	path, err := url.JoinPath(client.host, urlPath)
	if err != nil {
		return nil, err
	}

	b := bytes.NewReader(body)
	var rcBody io.ReadCloser
	if body != nil {
		rcBody = io.NopCloser(io.ReadSeeker(b))
	}

	req, err := http.NewRequest(http.MethodPut, path, rcBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("x-mkio-token", client.token)
	return &Request{b, req}, nil
}

// createOrUpdateHandleResponse handles the CreateOrUpdate response.
func (client *StreamingPoliciesClient) createOrUpdateHandleResponse(resp *http.Response) (armmediaservices.StreamingPoliciesClientCreateResponse, error) {
	result := armmediaservices.StreamingPoliciesClientCreateResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.StreamingPoliciesClientCreateResponse{}, err
	}
	if err := json.Unmarshal(body, &result.StreamingPolicy); err != nil {
		return armmediaservices.StreamingPoliciesClientCreateResponse{}, err
	}
	return result, nil
}

// Delete - Deletes a Streaming Policy in the Media Services account
// If the operation fails it returns an *ResponseError type.
// streamingPolicyName - The StreamingPolicy name.
// options - StreamingPoliciesClientDeleteOptions contains the optional parameters for the StreamingPoliciesClient.Delete
// method.
func (client *StreamingPoliciesClient) Delete(ctx context.Context, streamingPolicyName string, options *armmediaservices.StreamingPoliciesClientDeleteOptions) (armmediaservices.AssetsClientDeleteResponse, error) {
	req, err := client.deleteCreateRequest(ctx, streamingPolicyName, options)
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
func (client *StreamingPoliciesClient) deleteCreateRequest(ctx context.Context, streamingPolicyName string, options *armmediaservices.StreamingPoliciesClientDeleteOptions) (*Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingPolicies/{streamingPolicyName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{streamingPolicyName}", url.PathEscape(streamingPolicyName))
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

	return &Request{nil, req}, nil
}

// Get - Get the details of a Streaming Policy in the Media Services account
// If the operation fails it returns an *ResponseError type.
// streamingPolicyName - The StreamingPolicy name.
// options - StreamingPoliciesClientGetOptions contains the optional parameters for the StreamingPoliciesClient.Get method.
func (client *StreamingPoliciesClient) Get(ctx context.Context, streamingPolicyName string, options *armmediaservices.StreamingPoliciesClientGetOptions) (armmediaservices.StreamingPoliciesClientGetResponse, error) {
	req, err := client.getCreateRequest(ctx, streamingPolicyName, options)
	if err != nil {
		return armmediaservices.StreamingPoliciesClientGetResponse{}, err
	}
	// Try to do request, handle retries if tooManyRequests
	resp, err := client.DoRequestWithBackoff(req)
	if err != nil {
		// We hit some error we and failed retry loop. Return error
		return armmediaservices.StreamingPoliciesClientGetResponse{}, err
	}
	return client.getHandleResponse(resp)
}

// getCreateRequest creates the Get request.
func (client *StreamingPoliciesClient) getCreateRequest(ctx context.Context, streamingPolicyName string, options *armmediaservices.StreamingPoliciesClientGetOptions) (*Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingPolicies/{streamingPolicyName}"
	if client.subscriptionName == "" {
		return nil, errors.New("parameter client.subscriptionName cannot be empty")
	}
	urlPath = strings.ReplaceAll(urlPath, "{subscriptionName}", url.PathEscape(client.subscriptionName))
	urlPath = strings.ReplaceAll(urlPath, "{streamingPolicyName}", url.PathEscape(streamingPolicyName))
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
	return &Request{nil, req}, nil
}

// getHandleResponse handles the Get response.
func (client *StreamingPoliciesClient) getHandleResponse(resp *http.Response) (armmediaservices.StreamingPoliciesClientGetResponse, error) {
	result := armmediaservices.StreamingPoliciesClientGetResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.StreamingPoliciesClientGetResponse{}, err
	}
	if err := json.Unmarshal(body, &result.StreamingPolicy); err != nil {
		return armmediaservices.StreamingPoliciesClientGetResponse{}, err
	}
	return result, nil
}

// List - List Streaming Policy in the mk.io account
// If the operation fails it returns an *ResponseError type.
// options - StreamingPoliciesClientListOptions contains the optional parameters for the StreamingPoliciesClient.List method.
func (client *StreamingPoliciesClient) List(ctx context.Context, options *armmediaservices.StreamingPoliciesClientListOptions) (armmediaservices.StreamingPoliciesClientListResponse, error) {
	skipToken := ""
	results := armmediaservices.StreamingPoliciesClientListResponse{}

	for {
		req, err := client.listCreateRequest(ctx, options, skipToken)
		if err != nil {
			return results, err
		}
		// Try to do request, handle retries if tooManyRequests
		resp, err := client.DoRequestWithBackoff(req)
		if err != nil {
			// We hit some error we and failed retry loop. Return error
			return results, err
		}
		listResp, err := client.listHandleResponse(resp)
		if err != nil {
			return results, err
		}
		results.StreamingPolicyCollection.Value = append(results.StreamingPolicyCollection.Value, listResp.StreamingPolicyCollection.Value...)
		if listResp.StreamingPolicyCollection.ODataNextLink == nil {
			// No more pages. Break the loop
			break
		} else {
			// Mor Pages, Update SkipToken
			skipToken = strings.Split(*listResp.StreamingPolicyCollection.ODataNextLink, "skiptoken=")[1]
		}
	}
	return results, nil
}

// listCreateRequest creates the List request.
func (client *StreamingPoliciesClient) listCreateRequest(ctx context.Context, options *armmediaservices.StreamingPoliciesClientListOptions, skipToken string) (*Request, error) {
	urlPath := "/api/ams/{subscriptionName}/streamingPolicies"
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
	if skipToken != "" {
		filter = `$skiptoken=` + skipToken + "&"
	}
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
	return &Request{nil, req}, nil
}

// listHandleResponse handles the List response.
func (client *StreamingPoliciesClient) listHandleResponse(resp *http.Response) (armmediaservices.StreamingPoliciesClientListResponse, error) {
	result := armmediaservices.StreamingPoliciesClientListResponse{}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return armmediaservices.StreamingPoliciesClientListResponse{}, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return armmediaservices.StreamingPoliciesClientListResponse{}, err
	}
	return result, nil
}

// lookupStreamingPolicies Get streaming policies from mk.io
func (client *StreamingPoliciesClient) LookupStreamingPolicies(ctx context.Context, before string, after string) ([]*armmediaservices.StreamingPolicy, error) {
	// Generate the filter
	filter := generateFilter(before, after)

	// If we have a filter apply it
	options := &armmediaservices.StreamingPoliciesClientListOptions{Orderby: to.Ptr("properties/created")}
	if filter != "" {
		options.Filter = to.Ptr(filter)
	}
	req, err := client.List(ctx, options)
	if err != nil {
		return nil, err
	}

	sp := []*armmediaservices.StreamingPolicy{}

	// Only export custom streaming policies
	for _, v := range req.StreamingPolicyCollection.Value {
		if !strings.HasPrefix(*v.Name, "Predefined_") {
			sp = append(sp, v)
		}
	}

	return sp, nil
}
