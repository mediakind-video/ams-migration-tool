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

var fpConfiguration = "#Microsoft.Media.ContentKeyPolicyFairPlayConfiguration"

// ExportAzContentKeyPolicies creates a file containing all ContentKeyPolicies from an AzureMediaService Subscription
func ExportAzContentKeyPolicies(ctx context.Context, azSp *AzureServiceProvider, before string, after string) ([]*armmediaservices.ContentKeyPolicy, error) {
	log.Info("Exporting ContentKeyPolicies")

	// Lookup ContentKeyPolicies
	contentKeyPolicies, err := azSp.lookupContentKeyPolicies(ctx, before, after)
	if err != nil {
		return contentKeyPolicies, fmt.Errorf("encountered error while exporting ConentKeyPolicies from Azure: %v", err)
	}

	return contentKeyPolicies, nil
}

// ExportMkContentKeyPolicies creates a file containing all ContentKeyPolicies from an mk.io Subscription
func ExportMkContentKeyPolicies(ctx context.Context, client *mkiosdk.ContentKeyPoliciesClient, before string, after string) ([]*armmediaservices.ContentKeyPolicy, error) {
	log.Info("Exporting ContentKeyPolicies")

	// Lookup ContentKeyPolicies
	contentKeyPolicies, err := client.LookupContentKeyPolicies(ctx, before, after)
	if err != nil {
		return contentKeyPolicies, fmt.Errorf("encountered error while exporting ConentKeyPolicies from mk.io: %v", err)
	}

	return contentKeyPolicies, nil
}

// ImportContentKeyPoliciesWorker reads a file containing ContentKeyPolicies in JSON format. Insert each ContentKeyPolicy into MKIO
func ImportContentKeyPoliciesWorker(ctx context.Context, client *mkiosdk.ContentKeyPoliciesClient, overwrite bool, wg *sync.WaitGroup, jobs chan *mkiosdk.FPContentKeyPolicy, successChan chan string, skippedChan chan string, failedChan chan string) {

	for contentKeyPolicy := range jobs {
		// Create each ContentKeyPolicy

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
			skippedChan <- *contentKeyPolicy.Name
			wg.Done()
			continue
		}

		if found && overwrite {
			// it exists, but we're overwriting, so we should delete it
			_, err := client.Delete(ctx, *contentKeyPolicy.Name, nil)
			if err != nil {
				failedChan <- *contentKeyPolicy.Name
				log.Errorf("unable to delete old ContentKeyPolicy %v for overwrite: %v", *contentKeyPolicy.Name, err)
			}
		}

		log.Debugf("Creating ContentKeyPolicy in MKIO: %v", *contentKeyPolicy.Name)

		// TODO do something with this response
		_, err = client.CreateOrUpdate(ctx, *contentKeyPolicy.Name, contentKeyPolicy, nil)
		if err != nil {
			failedChan <- *contentKeyPolicy.Name
			log.Errorf("unable to import ContentKeyPolicy %v: %v", *contentKeyPolicy.Name, err)
		} else {
			successChan <- *contentKeyPolicy.Name
		}
		wg.Done()
	}
}

// ImportContentKeyPolicies reads a file containing ContentKeyPolicies in JSON format. Insert each ContentKeyPolicy into MKIO
func ImportContentKeyPolicies(ctx context.Context, client *mkiosdk.ContentKeyPoliciesClient, contentKeyPolicies []*armmediaservices.ContentKeyPolicy, overwrite bool, fairplayAmsCompatibility bool, workers int) (int, int, []string, error) {
	log.Info("Importing ContentKeyPolicies")

	// Waitgroup to wait for all goroutines to finish
	wg := new(sync.WaitGroup)

	// Create channels to communicate between workers
	successChan := make(chan string, len(contentKeyPolicies))
	skippedChan := make(chan string, len(contentKeyPolicies))
	failedChan := make(chan string, len(contentKeyPolicies))
	jobs := make(chan *mkiosdk.FPContentKeyPolicy, len(contentKeyPolicies))

	// Setup worker pool. This will start X workers to handle jobs
	for w := 1; w <= workers; w++ {
		log.Infof("Starting ContentKeyPolicy worker %d", w)
		go ImportContentKeyPoliciesWorker(ctx, client, overwrite, wg, jobs, successChan, skippedChan, failedChan)
	}

	failedContentKeyPolicies := []string{}
	skipped := 0
	successCount := 0
	// Workaround to add FairPlayAmsCompatibility element to ContentKeyPolicy
	fpContentKeyPolicies := make([]*mkiosdk.FPContentKeyPolicy, 0)
	for _, contentKeyPolicy := range contentKeyPolicies {
		val := false
		if fairplayAmsCompatibility {
			for _, option := range contentKeyPolicy.Properties.Options {
				if *option.Configuration.GetContentKeyPolicyConfiguration().ODataType == fpConfiguration {
					val = true
					break
				}
			}
		}
		fpContentKeyPolicies = append(fpContentKeyPolicies,
			&mkiosdk.FPContentKeyPolicy{
				ContentKeyPolicy: *contentKeyPolicy,
				FPProperties: &mkiosdk.FPContentKeyPolicyProperties{
					ContentKeyPolicyProperties: *contentKeyPolicy.Properties,
					FairPlayAmsCompatibility:   &val,
				},
			})
	}

	// create each ContentKeyPolicy
	for _, contentKeyPolicy := range fpContentKeyPolicies {
		wg.Add(1)
		jobs <- contentKeyPolicy
	}

	log.Infof("Skipped %d existing ContentKeyPolicies", skipped)
	log.Infof("Imported %d ContentKeyPolicies", successCount)

	if len(failedContentKeyPolicies) > 0 {
		return successCount, skipped, failedContentKeyPolicies, fmt.Errorf("failed to import %d ContentKeyPolicies: %v", len(failedContentKeyPolicies), failedContentKeyPolicies)
	}

	return successCount, skipped, failedContentKeyPolicies, nil
}

// ValidateContentKeyPolicies TODO
func ValidateContentKeyPolicies(ctx context.Context) error {
	log.Info("Validating MKIO ContentKeyPolicies")

	return nil
}
