package migrate

import (
	"context"
	"fmt"
	"sync"

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

// ImportAssetsWorker reads a file containing Assets in JSON format. Insert each asset into MKIO
func ImportAssetsWorker(ctx context.Context, client *mkiosdk.AssetsClient, overwrite bool, wg *sync.WaitGroup, jobs chan *armmediaservices.Asset, successChan chan string, skippedChan chan string, failedChan chan string) {

	for asset := range jobs {
		log.Debugf("Importing Asset in MKIO: %v", *asset.Name)

		found := true
		// Check if asset already exists. Skip update unless overwrite is set
		_, err := client.Get(ctx, *asset.Name, nil)
		if err != nil {
			found = false
		}
		if found && !overwrite {
			// Found something and we're not overwriting. We should skip it
			log.Debugf("Asset already exists in MKIO, skipping: %v", *asset.Name)
			skippedChan <- *asset.Name
		} else {

			log.Debugf("Creating Asset in MKIO: %v", *asset.Name)

			_, err = client.CreateOrUpdate(ctx, *asset.Name, asset, nil)
			if err != nil {
				failedChan <- *asset.Name
				log.Errorf("unable to import asset %v: %v", *asset.Name, err)
			} else {
				successChan <- *asset.Name
			}
		}
		wg.Done()
	}
}

// ImportAssets reads a file containing Assets in JSON format. Insert each asset into MKIO
func ImportAssets(ctx context.Context, client *mkiosdk.AssetsClient, assets []*armmediaservices.Asset, overwrite bool, workers int) (int, int, []string, error) {
	log.Info("Importing Assets")

	// Waitgroup to wait for all goroutines to finish
	wg := new(sync.WaitGroup)

	// Create channels to communicate between workers
	successChan := make(chan string, len(assets))
	skippedChan := make(chan string, len(assets))
	failedChan := make(chan string, len(assets))
	jobs := make(chan *armmediaservices.Asset, len(assets))

	// Setup worker pool. This will start X workers to handle jobs
	for w := 1; w <= workers; w++ {
		log.Infof("Starting Asset worker %d", w)
		go ImportAssetsWorker(ctx, client, overwrite, wg, jobs, successChan, skippedChan, failedChan)
	}

	failedAssets := []string{}
	skipped := 0
	successCount := 0
	// Create each asset
	for _, asset := range assets {
		wg.Add(1)
		jobs <- asset
	}

	log.Info("Waiting for Assets workers to finish")
	wg.Wait()
	log.Info("Done importing Assets")

	close(jobs)
	close(successChan)
	close(skippedChan)
	close(failedChan)
	for f := range successChan {
		if f != "" {
			successCount++
		}
	}
	for result := range skippedChan {
		if result != "" {
			skipped++
		}
	}
	for result := range failedChan {
		if result != "" {
			failedAssets = append(failedAssets, result)
		}
	}

	log.Infof("Skipped %d existing assets", skipped)
	log.Infof("Imported %d assets", successCount)

	if len(failedAssets) > 0 {
		return successCount, skipped, failedAssets, fmt.Errorf("failed to import %d assets: %v", len(failedAssets), failedAssets)
	}

	return successCount, skipped, failedAssets, nil
}

// ValidateAssets
func ValidateAssets(ctx context.Context) error {
	log.Info("Validating MKIO Assets")

	return nil
}
