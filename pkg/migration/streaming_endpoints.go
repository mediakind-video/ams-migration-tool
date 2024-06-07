package migrate

import (
	"context"
	"fmt"
	"strings"

	"dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/mkiosdk"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
	log "github.com/sirupsen/logrus"
)

// ExportAzStreamingEndponts creates a file containing all StreamingEndpoints from an AzureMediaService Subscription
func ExportAzStreamingEndpoints(ctx context.Context, azSp *AzureServiceProvider) ([]*armmediaservices.StreamingEndpoint, error) {
	log.Info("Exporting Streaming Endpoints")

	// Lookup StreamingEndpoins
	se, err := azSp.lookupStreamingEndpoints(ctx)
	if err != nil {
		return se, fmt.Errorf("encountered error while exporting StreamingEndpoints From Azure: %v", err)
	}

	return se, nil
}

// ExportMkStreamingEndponts creates a file containing all StreamingEndpoints from a mk.io Subscription
func ExportMkStreamingEndpoints(ctx context.Context, client *mkiosdk.StreamingEndpointsClient) ([]*armmediaservices.StreamingEndpoint, error) {
	log.Info("Exporting Streaming Endpoints")

	// Lookup StreamingEndpoins
	se, err := client.LookupStreamingEndpoints(ctx)
	if err != nil {
		return se, fmt.Errorf("encountered error while exporting StreamingEndpoints From mk.io: %v", err)
	}

	return se, nil
}

// ImportStreamingEndpoints reads a file containing StreamingEndpoints in JSON format. Insert each asset into MKIO
func ImportStreamingEndpoints(ctx context.Context, client *mkiosdk.StreamingEndpointsClient, streamingEndpoints []*armmediaservices.StreamingEndpoint, overwrite bool) error {
	log.Info("Importing Streaming Endpoints")

	// Some values to output at the end
	successCount := 0
	skipped := 0
	failedSE := []string{}

	// Create each streamingEndpoint
	for _, se := range streamingEndpoints {
		found := true
		// Check if StreamingEndpoint already exists. We can't update them, so need to delete and recreate
		_, err := client.Get(ctx, *se.Name, nil)
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
			_, err := client.Delete(ctx, *se.Name, nil)
			if err != nil {
				log.Errorf("unable to delete old StreamingEndpoint %v for overwrite: %v", *se.Name, err)
			}
		}

		// We don't have an existing resource... We can create one
		log.Debugf("Creating StreamingEndpoint in MKIO: %v", *se.Name)

		// Location mismatch between Azure and MKIO
		if *se.Location == "East US" {
			log.Debugf("Location mismatch for %v. Setting to eastus", *se.Name)
			eastus := "eastus"
			se.Location = &eastus
		} else if *se.Location == "West US 2" {
			log.Debugf("Location mismatch for %v. Setting to westus2", *se.Name)
			westus := "westus2"
			se.Location = &westus
		} else if *se.Location == "West Europe" {
			log.Debugf("Location mismatch for %v. Setting to westeurope", *se.Name)
			westeurope := "westeurope"
			se.Location = &westeurope
		}

		// Not supported CDN Provider. Set to Akamai, with user input
		if se.Properties.CdnProvider != nil && *se.Properties.CdnProvider != "Akamai" {
			log.Info("CDN Provider mismatch. User input required")
			var setProvider string
			fmt.Printf("CDN Provider mismatch for %v. Change to Akamai [y/N]\n", *se.Name)
			fmt.Scan(&setProvider)
			if setProvider == "y" || setProvider == "Y" {
				log.Infof("Setting CDN Provider to StandardAkamai for %v", *se.Name)
				akamai := "StandardAkamai"
				se.Properties.CdnProvider = &akamai
			}
		}
		_, err = client.CreateOrUpdate(ctx, *se.Name, *se, nil)
		if err != nil {
			failedSE = append(failedSE, *se.Name)

			log.Errorf("unable to import streamingEndpoint %v: %v", *se.Name, err)
		} else {
			successCount++
		}
	}

	log.Infof("Skipped %d existing streamingEndpoints", skipped)
	log.Infof("Imported %d streamingEndpoints", successCount)

	if len(failedSE) > 0 {
		return fmt.Errorf("failed to import %d StreamingEndpoints: %v", len(failedSE), failedSE)
	}
	return nil
}
