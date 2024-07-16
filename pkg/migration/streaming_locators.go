package migrate

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/mkiosdk"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
	log "github.com/sirupsen/logrus"
)

// ExportAzStreamingLocators creates a file containing all StreamingLocators from an AzureMediaService Subscription
func ExportAzStreamingLocators(ctx context.Context, azSp *AzureServiceProvider, before string, after string) ([]*armmediaservices.StreamingLocator, error) {
	log.Info("Exporting Streaming Locators")

	// Lookup StreamingLocators
	sl, err := azSp.lookupStreamingLocators(ctx, before, after)
	if err != nil {
		return sl, fmt.Errorf("encountered error while exporting StreamingLocators From Azure: %v", err)
	}

	return sl, nil
}

// ExportMkStreamingLocators creates a file containing all StreamingLocators from a mk.io Subscription
func ExportMkStreamingLocators(ctx context.Context, client *mkiosdk.StreamingLocatorsClient, before string, after string) ([]*armmediaservices.StreamingLocator, error) {
	log.Info("Exporting Streaming Locators")

	// Lookup StreamingLocators
	sl, err := client.LookupStreamingLocators(ctx, before, after)
	if err != nil {
		return sl, fmt.Errorf("encountered error while exporting StreamingLocators From mk.io: %v", err)
	}

	return sl, nil
}

// ImportStreamingLocators reads a file containing StreamingLocators in JSON format. Insert each asset into MKIO
func ImportStreamingLocators(ctx context.Context, client *mkiosdk.StreamingLocatorsClient, streamingLocators []*armmediaservices.StreamingLocator, overwrite bool) (int, int, []string, error) {

	log.Info("Importing Streaming Locators")

	// Some values to output at the end
	successCount := 0
	skipped := 0
	failedSL := []string{}

	// Create each streamingLocator
	for _, sl := range streamingLocators {
		found := true
		// Check if StreamingLocator already exists. We can't update them, so need to delete and recreate
		_, err := client.Get(ctx, *sl.Name, nil)
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
			_, err := client.Delete(ctx, *sl.Name, nil)
			if err != nil {
				log.Errorf("unable to delete old StreamingLocator %v for overwrite: %v", *sl.Name, err)
			}
		}

		// We don't have an existing resource... We can create one
		log.Debugf("Creating StreamingLocator in MKIO: %v", *sl.Name)

		if strings.HasPrefix(*sl.Properties.StreamingPolicyName, "Predefined_") {
			log.Infof("removing customer ContentKeys from StreamingLocator with Predefined Streaming Policy: %v", *sl.Name)
			sl.Properties.ContentKeys = nil
		}

		_, err = client.CreateOrUpdate(ctx, *sl.Name, *sl, nil)
		if err != nil {
			failedSL = append(failedSL, *sl.Name)

			log.Errorf("unable to import streamingLocator %v: %v", *sl.Name, err)
		} else {
			successCount++
		}
	}

	log.Infof("Skipped %d existing streamingLocators", skipped)
	log.Infof("Imported %d streamingLocators", successCount)

	if len(failedSL) > 0 {
		return successCount, skipped, failedSL, fmt.Errorf("failed to import %d StreamingLocators: %v", len(failedSL), failedSL)
	}
	return successCount, skipped, failedSL, nil
}

// ValidateStreamingLocators validates that streaming locators exist in MKIO and produce output.
func ValidateStreamingLocators(ctx context.Context, slClient *mkiosdk.StreamingLocatorsClient, seClient *mkiosdk.StreamingEndpointsClient, streamingLocators []*armmediaservices.StreamingLocator) error {
	log.Info("Validating MKIO StreamingLocators")

	streamingEndpoints, err := seClient.List(ctx, nil)
	if err != nil {
		return fmt.Errorf("unable to list streaming Endpoints: %v", err)

	}

	log.Info(streamingEndpoints)
	var streamingEndpoint armmediaservices.StreamingEndpoint
	// Find a streaming Endpoint we can use to test
	for _, se := range streamingEndpoints.Value {
		if *se.Properties.ResourceState == "Running" {
			streamingEndpoint = *se
		}
	}

	if *streamingEndpoint.Properties.HostName == "" {
		return fmt.Errorf("unable to find HostName of Running StreamingEndpoint for testing")
	}
	log.Infof("Found streamingEndpoint for testing: %v", *streamingEndpoint.Name)

	failedSL := []string{}
	successCount := 0
	missingSL := []string{}

	httpClient := &http.Client{}

	for _, sl := range streamingLocators {
		found := true
		// Check if streaming locator exists in MKIO
		_, err := slClient.Get(ctx, *sl.Name, nil)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				found = false
				// Probably important that we expect it but can't find it
				missingSL = append(missingSL, *sl.Name)
			}
		}

		if found {
			log.Debugf("Found StreamingLocator in MKIO: %s", *sl.Name)
			resp, err := slClient.ListPaths(ctx, *sl.Name, nil)
			if err != nil {
				log.Errorf("unable to list paths for streamingLocator %v: %v", *sl.Name, err)
				failedSL = append(failedSL, *sl.Name)
			}
			for _, sp := range resp.StreamingPaths {
				for _, path := range sp.Paths {
					url := fmt.Sprintf("https://%v%v", *streamingEndpoint.Properties.HostName, *path)

					log.Debugf("Found StreamingLocator Path: %v", *path)
					r, err := httpClient.Get(url)
					if err != nil {
						log.Errorf("encountered error running GET %v: %v ", url, err)
						failedSL = append(failedSL, url)
						continue
					}
					if r.StatusCode != 200 {
						log.Errorf("bad status %v: %v ", url, r.StatusCode)
						failedSL = append(failedSL, url)
					} else {
						successCount++
					}
				}
			}
		}
	}

	log.Infof("Validated %d streamingLocators", successCount)
	if len(missingSL) > 0 {
		log.Errorf("failed to get %d StreamingLocators: %v", len(missingSL), missingSL)
	}

	if len(failedSL) > 0 {
		log.Errorf("failed to validate %d StreamingLocators: %v", len(failedSL), failedSL)
	}

	if len(failedSL) > 0 || len(missingSL) > 0 {
		return fmt.Errorf("validation failed")
	}

	return nil
}
