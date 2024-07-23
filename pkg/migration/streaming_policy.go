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

// ImportStreamingPolicyWorker - Do the work to import a StreamingPolicy into MKIO
func ImportStreamingPolicyWorker(ctx context.Context, client *mkiosdk.StreamingPoliciesClient, overwrite bool, wg *sync.WaitGroup, jobs chan *armmediaservices.StreamingPolicy, successChan chan string, skippedChan chan string, failedChan chan string) {

	// Create each streamingPolicy
	for sp := range jobs {
		log.Debugf("Importing StreamingPolicy in MKIO: %v", *sp.Name)

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
			skippedChan <- *sp.Name
			wg.Done()
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
			failedChan <- *sp.Name
			log.Errorf("unable to import streamingPolicy %v: %v", *sp.Name, err)
		} else {
			successChan <- *sp.Name
		}
		wg.Done()
	}
}

// ImportStreamingPolicies reads a file containing StreamingPolicies in JSON format. Insert each streaming policy into MKIO
func ImportStreamingPolicies(ctx context.Context, client *mkiosdk.StreamingPoliciesClient, streamingPolicies []*armmediaservices.StreamingPolicy, overwrite bool, workers int) (int, int, []string, error) {
	log.Info("Importing Streaming Policy")

	// Waitgroup to wait for all goroutines to finish
	wg := new(sync.WaitGroup)

	// Create channels to communicate between workers
	successChan := make(chan string, len(streamingPolicies))
	skippedChan := make(chan string, len(streamingPolicies))
	failedChan := make(chan string, len(streamingPolicies))
	jobs := make(chan *armmediaservices.StreamingPolicy, len(streamingPolicies))

	// Setup worker pool. This will start X workers to handle jobs
	for w := 1; w <= workers; w++ {
		log.Infof("Starting StreamingPolicy worker %d", w)
		go ImportStreamingPolicyWorker(ctx, client, overwrite, wg, jobs, successChan, skippedChan, failedChan)
	}

	// Create each streamingPolicy
	for _, sp := range streamingPolicies {
		wg.Add(1)
		jobs <- sp
	}

	// Some values to output at the end
	successCount := 0
	skipped := 0
	failedSP := []string{}

	log.Info("Waiting for Streaming Policy workers to finish")
	wg.Wait()
	log.Info("Done importing Streaming Policies")

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
			failedSP = append(failedSP, result)
		}
	}

	log.Infof("Skipped %d existing streamingPolicy", skipped)
	log.Infof("Imported %d streamingPolicies", successCount)

	if len(failedSP) > 0 {
		return successCount, skipped, failedSP, fmt.Errorf("failed to import %d StreamingPolicies: %v", len(failedSP), failedSP)
	}
	return successCount, skipped, failedSP, nil
}
