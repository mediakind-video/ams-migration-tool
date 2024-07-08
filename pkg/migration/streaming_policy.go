package migrate

import (
	"context"
	"fmt"
	"strings"

	"dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/mkiosdk"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
	log "github.com/sirupsen/logrus"
)

// ExportAzStreamingPolicies creates a file containing all StreamingPolicies from an AzureMediaService Subscription
func ExportAzStreamingPolicies(ctx context.Context, azSp *AzureServiceProvider, before string, after string) ([]*armmediaservices.StreamingPolicy, error) {
	log.Info("Exporting Streaming Policies")

	// Lookup StreamingPolicies
	sl, err := azSp.lookupStreamingPolicies(ctx, before, after)
	if err != nil {
		return sl, fmt.Errorf("encountered error while exporting StreamingPolicies From Azure: %v", err)
	}

	return sl, nil
}

// ExportMkStreamingPolicies creates a file containing all StreamingPolicies from a mk.io Subscription
func ExportMkStreamingPolicies(ctx context.Context, client *mkiosdk.StreamingPoliciesClient, before string, after string) ([]*armmediaservices.StreamingPolicy, error) {
	log.Info("Exporting Streaming Policies")

	// Lookup StreamingPolicies
	sl, err := client.LookupStreamingPolicies(ctx, before, after)
	if err != nil {
		return sl, fmt.Errorf("encountered error while exporting StreamingPolicies From mk.io: %v", err)
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
