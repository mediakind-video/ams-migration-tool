package migrate

import (
	"context"
	"fmt"
	"strings"

	"dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/mkiosdk"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
	log "github.com/sirupsen/logrus"
)

// ExportStreamingPolicies creates a file containing all StreamingPolicies from an AzureMediaService Subscription
func ExportStreamingPolicies(ctx context.Context, azSp *AzureServiceProvider) ([]*armmediaservices.StreamingPolicy, error) {
	log.Info("Exporting Streaming Policies")

	// Lookup StreamingPolicies
	sl, err := azSp.lookupStreamingPolicies(ctx)
	if err != nil {
		return sl, fmt.Errorf("encountered error while exporting StreamingPolicies From Azure: %v", err)
	}

	return sl, nil
}

// ImportStreamingPolicies reads a file containing StreamingPolicies in JSON format. Insert each asset into MKIO
func ImportStreamingPolicies(ctx context.Context, client *mkiosdk.StreamingPoliciesClient, streamingPolicies []*armmediaservices.StreamingPolicy, overwrite bool) error {
	log.Info("Importing Streaming Policy")

	// Some values to output at the end
	successCount := 0
	skipped := 0
	failedSP := []string{}

	// Create each streamingPolicy
	for _, sp := range streamingPolicies {
		found := true
		// Check if StreamingPolicy already exists. We can't update them, so need to delete and recreate
		_, err := client.Get(ctx, *sp.Name, nil)
		if err != nil {
			// We are looking for a not found error. If we get this we can add w/o incident
			if strings.Contains(err.Error(), "not found") {
				found = false
			}
		}

		if found && !overwrite {
			// Found something and we're not overwriting. We should skip it
			skipped++
			continue
		}

		if found && overwrite {
			// it exists, but we're overwriting, so we should delete it
			_, err := client.Delete(ctx, *sp.Name, nil)
			if err != nil {
				log.Errorf("unable to delete old StreamingPolicy %v for overwrite: %v", *sp.Name, err)
			}
		}

		// We don't have an existing resource... We can create one
		log.Debugf("Creating StreamingPolicy in MKIO: %v", *sp.Name)

		_, err = client.CreateOrUpdate(ctx, *sp.Name, *sp, nil)
		if err != nil {
			failedSP = append(failedSP, *sp.Name)

			log.Errorf("unable to import streamingPolicy %v: %v", *sp.Name, err)
		} else {
			successCount++
		}
	}

	log.Infof("Skipped %d existing streamingPolicy", skipped)
	log.Infof("Imported %d streamingPolicies", successCount)

	if len(failedSP) > 0 {
		return fmt.Errorf("failed to import %d StreamingPolicies: %v", len(failedSP), failedSP)
	}
	return nil
}

// ValidateStreamingPolicies validates that streaming locators exist in MKIO and produce output.
// func ValidateStreamingPolicies(ctx context.Context, slClient *mkiosdk.StreamingPolicyClient, seClient *mkiosdk.StreamingEndpointsClient, streamingPolicies []*armmediaservices.StreamingPolicy) error {
// 	log.Info("Validating MKIO StreamingLocators")

// 	streamingEndpoints, err := seClient.List(ctx, nil)
// 	if err != nil {
// 		return fmt.Errorf("unable to list streaming Endpoints: %v", err)

// 	}

// 	log.Info(streamingEndpoints)
// 	var streamingEndpoint armmediaservices.StreamingEndpoint
// 	// Find a streaming Endpoint we can use to test
// 	for _, se := range streamingEndpoints.Value {
// 		if *se.Properties.ResourceState == "Running" {
// 			streamingEndpoint = *se
// 		}
// 	}

// 	if *streamingEndpoint.Properties.HostName == "" {
// 		return fmt.Errorf("unable to find HostName of Running StreamingEndpoint for testing")
// 	}
// 	log.Infof("Found streamingEndpoint for testing: %v", *streamingEndpoint.Name)

// 	failedSP := []string{}
// 	successCount := 0
// 	missingSP := []string{}

// 	httpClient := &http.Client{}

// 	for _, sl := range streamingLocators {
// 		found := true
// 		// Check if streaming locator exists in MKIO
// 		_, err := slClient.Get(ctx, *sl.Name, nil)
// 		if err != nil {
// 			if strings.Contains(err.Error(), "not found") {
// 				found = false
// 				// Probably important that we expect it but can't find it
// 				missingSP = append(missingSP, *sl.Name)
// 			}
// 		}

// 		if found {
// 			log.Debugf("Found StreamingLocator in MKIO: %s", *sl.Name)
// 			resp, err := slClient.ListPaths(ctx, *sl.Name, nil)
// 			if err != nil {
// 				log.Errorf("unable to list paths for streamingLocator %v: %v", *sl.Name, err)
// 				failedSP = append(failedSP, *sl.Name)
// 			}
// 			for _, sp := range resp.StreamingPaths {
// 				for _, path := range sp.Paths {
// 					url := fmt.Sprintf("https://%v%v", *streamingEndpoint.Properties.HostName, *path)

// 					log.Debugf("Found StreamingLocator Path: %v", *path)
// 					r, err := httpClient.Get(url)
// 					if err != nil {
// 						log.Errorf("encountered error running GET %v: %v ", url, err)
// 						failedSP = append(failedSP, url)
// 						continue
// 					}
// 					if r.StatusCode != 200 {
// 						log.Errorf("bad status %v: %v ", url, r.StatusCode)
// 						failedSP = append(failedSP, url)
// 					} else {
// 						successCount++
// 					}
// 				}
// 			}
// 		}
// 	}

// 	log.Infof("Validated %d streamingLocators", successCount)
// 	if len(missingSP) > 0 {
// 		log.Errorf("failed to get %d StreamingLocators: %v", len(missingSP), missingSP)
// 	}

// 	if len(failedSP) > 0 {
// 		log.Errorf("failed to validate %d StreamingPolicies: %v", len(failedSP), failedSP)
// 	}

// 	if len(failedSP) > 0 || len(missingSP) > 0 {
// 		return fmt.Errorf("validation failed")
// 	}

// 	return nil
// }
