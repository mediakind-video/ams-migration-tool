package migrate

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/mkiosdk"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
	log "github.com/sirupsen/logrus"
)

// ExportAzAssetFilters creates a file containing all AssetFilters from an AzureMediaService Subscription
func ExportAzAssetFilters(ctx context.Context, azSp *AzureServiceProvider, assets []*armmediaservices.Asset, workers int) (map[string][]*armmediaservices.AssetFilter, error) {
	log.Info("Exporting AssetFilters")

	allAssetFilters := map[string][]*armmediaservices.AssetFilter{}
	skipped := []string{}

	// Waitgroup to wait for all goroutines to finish
	wg := new(sync.WaitGroup)

	// Create channels to communicate between workers
	filterChan := make(chan map[string][]*armmediaservices.AssetFilter, len(assets))
	skippedChan := make(chan string, len(assets))
	jobs := make(chan string, len(assets))

	// Setup worker pool. This will start X workers to handle jobs
	for w := 1; w <= workers; w++ {
		log.Infof("Starting AssetFilter worker %d", w)
		go azSp.lookupAssetFiltersWorker(ctx, wg, jobs, filterChan, skippedChan)
	}

	// Loop through assets and add them to the jobs channel
	for _, a := range assets {
		// Add to waitgroup to wait for all jobs to finish
		wg.Add(1)
		// Start a job for the worker to handle
		jobs <- *a.Name
	}
	log.Info("Waiting for AssetFilter workers to finish")
	wg.Wait()
	log.Info("Done Processing Asset Filters")

	close(jobs)
	close(filterChan)
	for f := range filterChan {
		for k, v := range f {
			allAssetFilters[k] = v
		}
	}
	close(skippedChan)
	for result := range skippedChan {
		if result != "" {
			skipped = append(skipped, result)
		}
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
func ImportAssetFilters(ctx context.Context, client *mkiosdk.AssetFiltersClient, assetFilters map[string][]*armmediaservices.AssetFilter, overwrite bool) (int, int, []string, error) {

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
		return successCount, skipped, failedAssetFilters, fmt.Errorf("failed to import %d Asset Filters: %v", len(failedAssetFilters), failedAssetFilters)
	}

	return successCount, skipped, failedAssetFilters, nil
}

// ValidateAssetFilters
func ValidateAssetFilters(ctx context.Context) error {
	log.Info("Validating MKIO AssetFilters")

	return nil
}
