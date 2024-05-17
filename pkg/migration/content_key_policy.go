package migrate

import (
	"context"
	"fmt"
	"strings"

	"dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/mkiosdk"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
	log "github.com/sirupsen/logrus"
)

// ExportAzContentKeyPolicies creates a file containing all ContentKeyPolicies from an AzureMediaService Subscription
func ExportAzContentKeyPolicies(ctx context.Context, azSp *AzureServiceProvider) ([]*armmediaservices.ContentKeyPolicy, error) {
	log.Info("Exporting ContentKeyPolicies")

	// Lookup ContentKeyPolicies
	contentKeyPolicies, err := azSp.lookupContentKeyPolicies(ctx)
	if err != nil {
		return contentKeyPolicies, fmt.Errorf("encountered error while exporting ConentKeyPolicies from Azure: %v", err)
	}

	return contentKeyPolicies, nil
}

// ExportMkContentKeyPolicies creates a file containing all ContentKeyPolicies from an mk.io Subscription
func ExportMkContentKeyPolicies(ctx context.Context, client *mkiosdk.ContentKeyPoliciesClient) ([]*armmediaservices.ContentKeyPolicy, error) {
	log.Info("Exporting ContentKeyPolicies")

	// Lookup ContentKeyPolicies
	contentKeyPolicies, err := client.LookupContentKeyPolicies(ctx)
	if err != nil {
		return contentKeyPolicies, fmt.Errorf("encountered error while exporting ConentKeyPolicies from mk.io: %v", err)
	}

	return contentKeyPolicies, nil
}

// ImportContentKeyPolicies reads a file containing ContentKeyPolicies in JSON format. Insert each ContentKeyPolicy into MKIO
func ImportContentKeyPolicies(ctx context.Context, client *mkiosdk.ContentKeyPoliciesClient, contentKeyPolicies []*armmediaservices.ContentKeyPolicy, overwrite bool) error {
	log.Info("Importing ContentKeyPolicies")

	failedContentKeyPolicies := []string{}
	skipped := 0
	successCount := 0
	// Create each ContentKeyPolicy
	for _, contentKeyPolicy := range contentKeyPolicies {

		found := true
		// Check if ContentKeyPolicy already exists. Skip update unless overwrite is set
		_, err := client.Get(ctx, *contentKeyPolicy.Name, nil)
		if err != nil {
			if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "Not Found") {
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
			_, err := client.Delete(ctx, *contentKeyPolicy.Name, nil)
			if err != nil {
				log.Errorf("unable to delete old ContentKeyPolicy %v for overwrite: %v", *contentKeyPolicy.Name, err)
			}
		}

		log.Debugf("Creating ContentKeyPolicy in MKIO: %v", *contentKeyPolicy.Name)

		// TODO do something with this response
		_, err = client.CreateOrUpdate(ctx, *contentKeyPolicy.Name, contentKeyPolicy, nil)
		if err != nil {
			failedContentKeyPolicies = append(failedContentKeyPolicies, *contentKeyPolicy.Name)
			log.Errorf("unable to import ContentKeyPolicy %v: %v", *contentKeyPolicy.Name, err)
		} else {
			successCount++
		}
	}

	log.Infof("Skipped %d existing ContentKeyPolicies", skipped)
	log.Infof("Imported %d ContentKeyPolicies", successCount)

	if len(failedContentKeyPolicies) > 0 {
		return fmt.Errorf("failed to import %d ContentKeyPolicies: %v", len(failedContentKeyPolicies), failedContentKeyPolicies)
	}

	return nil
}

// ValidateContentKeyPolicies TODO
func ValidateContentKeyPolicies(ctx context.Context) error {
	log.Info("Validating MKIO ContentKeyPolicies")

	return nil
}
