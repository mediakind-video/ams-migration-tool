package mkiosdk

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Generic types

type MkioClient struct {
	host             string
	subscriptionName string
	token            string
	hc               *http.Client
}

func generateFilter(before string, after string) string {
	filter := ""
	// Handle after date
	if after != "" {
		filter = fmt.Sprintf("properties/created gt %s", after)
	}
	// Before & after. Add an and in between
	if before != "" && after != "" {
		filter = fmt.Sprintf("%s and", filter)
	}
	// Handle before date
	if before != "" && after == "" {
		filter = fmt.Sprintf("properties/created lt %s", before)
	} else if before != "" {
		filter = fmt.Sprintf("%s properties/created lt %s", filter, before)
	}
	return filter
}

type ClientOptions struct {
	host string
}

// Do backoff to handle rate limiting. Rate limit should last about 1 minute. This is over 2 minutes, so we should be good
var backoffSchedule = []time.Duration{
	1 * time.Second,
	5 * time.Second,
	10 * time.Second,
	15 * time.Second,
	25 * time.Second,
	30 * time.Second,
	45 * time.Second,
}

// GetProfile - Get the Media Services account
// If the operation fails it returns an error.
func (client *MkioClient) GetProfile(ctx context.Context) error {
	req, err := client.getProfileRequest(ctx)
	if err != nil {
		return err
	}
	resp, err := client.hc.Do(req)
	if err != nil {
		return err
	}
	if !HasStatusCode(resp, http.StatusOK, http.StatusCreated) {
		return NewResponseError(resp)
	}
	return nil
}

// getProfileRequest creates the GetProfile request.
func (client *MkioClient) getProfileRequest(ctx context.Context) (*http.Request, error) {
	urlPath := "/api/profile"
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

func (client *MkioClient) DoRequestWithBackoff(request *http.Request) (*http.Response, error) {
	var resp *http.Response

	// loop through backoff schedule. Hopefully we don't actually have to loop, but this will trigger if we get rate limited
	for _, backoff := range backoffSchedule {
		var err error

		resp, err = client.hc.Do(request)
		if err != nil {
			// Return an error from the Request
			return resp, err
		}

		// Check the status code, expectations of response are consistent across functions
		if request.Method == http.MethodPut {
			// CreateOrUpdate
			if HasStatusCode(resp, http.StatusOK, http.StatusCreated) {
				return resp, nil
			}
		} else if request.Method == http.MethodPost {
			// List Paths
			if HasStatusCode(resp, http.StatusOK) {
				return resp, nil
			}
		} else if request.Method == http.MethodDelete {
			// Delete
			if HasStatusCode(resp, http.StatusOK, http.StatusNoContent) {
				return resp, nil
			}
		} else if request.Method == http.MethodGet {
			// Get/List
			if HasStatusCode(resp, http.StatusOK) {
				return resp, nil
			}
		}

		// Unhandled status Codes. The only one we care about retrying for is TooManyRequests. Return an error from the status code
		if !HasStatusCode(resp, http.StatusTooManyRequests) {
			return resp, NewResponseError(resp)
		}

		// We have a TooManyRequests status code. Sleep for the backoff duration and try again
		time.Sleep(backoff)
	}

	// We have exhausted the backoff schedule. Return the last response & corresponding Error
	return resp, NewResponseError(resp)
}
