package migrate

import (
	"context"
	"fmt"
	"strings"

	"dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/mkiosdk"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
	log "github.com/sirupsen/logrus"
)

// ExportAzAssetFilters creates a file containing all AssetFilters from an AzureMediaService Subscription
func ExportAzAssetFilters(ctx context.Context, azSp *AzureServiceProvider, assets []*armmediaservices.Asset) (map[string][]*armmediaservices.AssetFilter, error) {
	log.Info("Exporting AssetFilters")

	allAssetFilters := map[string][]*armmediaservices.AssetFilter{}
	skipped := []string{}
	for _, a := range assets {

		log.Debugf("exporting filters for asset %v", *a.Name)
		// Lookup AssetFilters
		assetFilters, err := azSp.lookupAssetFilters(ctx, *a.Name)
		if err != nil {
			skipped = append(skipped, *a.Name)
		}
		allAssetFilters[*a.Name] = assetFilters
	}

	if len(skipped) > 0 {
		return allAssetFilters, fmt.Errorf("failed to export %d Asset Filters: %v", len(skipped), skipped)
	}

	return allAssetFilters, nil
}

// ExportMkAssetFilters creates a file containing all AssetFilters from an mk.io Subscription
func ExportMkAssetFilters(ctx context.Context, client *mkiosdk.AssetFiltersClient, assets []*armmediaservices.Asset) (map[string][]*armmediaservices.AssetFilter, error) {
	log.Info("Exporting AssetFilters")

	allAssetFilters := map[string][]*armmediaservices.AssetFilter{}
	skipped := []string{}
	for _, a := range assets {

		log.Debugf("exporting filters for asset %v", *a.Name)
		// Lookup AssetFilters
		assetFilters, err := client.LookupAssetFilters(ctx, *a.Name)
		if err != nil {
			skipped = append(skipped, *a.Name)
		}
		allAssetFilters[*a.Name] = assetFilters
	}

	if len(skipped) > 0 {
		return allAssetFilters, fmt.Errorf("failed to export %d Asset Filters: %v", len(skipped), skipped)
	}

	return allAssetFilters, nil
}

// ImportAssetFilters reads a file containing AssetFilters in JSON format. Insert each asset into MKIO
func ImportAssetFilters(ctx context.Context, client *mkiosdk.AssetFiltersClient, assetFilters map[string][]*armmediaservices.AssetFilter, overwrite bool) error {

	log.Info("Importing AssetFilters")

	failedAssetFilters := []string{}
	skipped := 0
	successCount := 0
	// Create each asset
	// Go through each key/value in map
	for assetName, assetFilterList := range assetFilters {

		// v is a []assetFilter, loop through it to handle each one
		for _, af := range assetFilterList {

			found := true
			// Check if asset already exists. Skip update unless overwrite is set
			_, err := client.Get(ctx, *af.Name, nil)
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

			log.Debugf("Creating AssetFilter in MKIO: %v", *af.Name)

			_, err = client.CreateOrUpdate(ctx, assetName, *af.Name, af, nil)
			if err != nil {
				failedAssetFilters = append(failedAssetFilters, *af.Name)
				log.Errorf("unable to import asset filter %v: %v", *af.Name, err)
			} else {
				successCount++
			}
		}
	}
	log.Infof("Skipped %d existing Asset Filters", skipped)
	log.Infof("Imported %d Asset Filters", successCount)

	if len(failedAssetFilters) > 0 {
		return fmt.Errorf("failed to import %d Asset Filters: %v", len(failedAssetFilters), failedAssetFilters)
	}

	return nil
}

// ValidateAssetFilters
func ValidateAssetFilters(ctx context.Context) error {
	log.Info("Validating MKIO AssetFilters")

	return nil
}
