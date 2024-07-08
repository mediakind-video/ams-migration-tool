package migrate

import (
	"context"
	"fmt"

	"dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/mkiosdk"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
	log "github.com/sirupsen/logrus"
)

// ExportAzAssets creates a file containing all Assets from an AzureMediaService Subscription
func ExportAzAssets(ctx context.Context, azSp *AzureServiceProvider, before string, after string) ([]*armmediaservices.Asset, error) {
	log.Info("Exporting Assets")

	// Lookup Assets
	assets, err := azSp.lookupAssets(ctx, before, after)
	if err != nil {
		return assets, fmt.Errorf("encountered error while exporting assets from Azure: %v", err)
	}

	return assets, nil
}

// ExportMkAssets creates a file containing all Assets from a mk.io Subscription
func ExportMkAssets(ctx context.Context, client *mkiosdk.AssetsClient, before string, after string) ([]*armmediaservices.Asset, error) {
	log.Info("Exporting Assets")

	// Lookup Assets
	assets, err := client.LookupAssets(ctx, before, after)
	if err != nil {
		return assets, fmt.Errorf("encountered error while exporting assets from mk.io : %v", err)
	}

	return assets, nil
}

// ImportAssets reads a file containing Assets in JSON format. Insert each asset into MKIO
func ImportAssets(ctx context.Context, client *mkiosdk.AssetsClient, assets []*armmediaservices.Asset, overwrite bool) error {
	log.Info("Importing Assets")

	failedAssets := []string{}
	skipped := 0
	successCount := 0
	// Create each asset
	for _, asset := range assets {

		found := true
		// Check if asset already exists. Skip update unless overwrite is set
		_, err := client.Get(ctx, *asset.Name, nil)
		if err != nil {
			found = false
		}
		if found && !overwrite {
			// Found something and we're not overwriting. We should skip it
			log.Debugf("Asset already exists in MKIO, skipping: %v", *asset.Name)
			skipped++
			continue
		}

		log.Debugf("Creating Asset in MKIO: %v", *asset.Name)

		_, err = client.CreateOrUpdate(ctx, *asset.Name, asset, nil)
		if err != nil {
			failedAssets = append(failedAssets, *asset.Name)
			log.Errorf("unable to import asset %v: %v", *asset.Name, err)
		} else {
			successCount++
		}
	}

	log.Infof("Skipped %d existing assets", skipped)
	log.Infof("Imported %d assets", successCount)

	if len(failedAssets) > 0 {
		return fmt.Errorf("failed to import %d assets: %v", len(failedAssets), failedAssets)
	}

	return nil
}

// ValidateAssets
func ValidateAssets(ctx context.Context) error {
	log.Info("Validating MKIO Assets")

	return nil
}
