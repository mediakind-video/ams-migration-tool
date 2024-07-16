package cmd

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	migrate "dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/migration"
	"dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/mkiosdk"
)

// command line options
var (
	azSubscription       string
	azResourceGroup      string
	azAccountName        string
	mkImportSubscription string
	mkExportSubscription string
	migrationFile        string
	apiEndpoint          string
	createdBefore        string
	createdAfter         string

	workers int

	debug             bool
	importResources   bool
	exportResources   bool
	validateResources bool
	overwrite         bool

	assets             bool
	assetFilters       bool
	contentKeyPolicies bool
	streamingLocators  bool
	streamingEndpoints bool
	streamingPolicies  bool

	fairplayAmsCompatibility bool
)

const ASSETS = "assets"
const ASSETFILTERS = "assetFilters"
const STREAMINGPOLICIES = "streamingPolicies"
const STREAMINGLOCATORS = "streamingLocators"
const STREAMINGENDPOINTS = "streamingEndpoints"
const CONTENTKEYPOLICIES = "contentKeyPolicies"
const EXPORT = "export"
const IMPORT = "import"

type results struct {
	resource  string
	operation string
	duration  time.Duration
	failures  []string
	skipped   int
	migrated  int
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate AMS Assets",
	Long:  `Migrate Assets and StreamingLocators from Azure MediaServices to mk.io.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		ctx := context.Background()

		if debug {
			log.Info("Debug enabled")
			log.SetLevel(log.DebugLevel)
		}

		var timings []results
		// If we don't export,import,validate what do we do?
		if !exportResources && !importResources && !validateResources {
			log.Fatal("Please select a valid command: [import|export|validate]")
		}

		// Set a timestamp on our migraiton file
		if migrationFile == "" {
			migrationFile = fmt.Sprintf("migration-%v.json", time.Now().Unix())
		}

		migrationContents := migrate.MigrationFileContents{}

		// Log into MKIO for the Import. Do this first so we know if it fails before we do any work.
		var mkImportAssetsClient *mkiosdk.AssetsClient
		var mkImportAssetFiltersClient *mkiosdk.AssetFiltersClient
		var mkImportStreamingLocatorsClient *mkiosdk.StreamingLocatorsClient
		var mkImportStreamingPoliciesClient *mkiosdk.StreamingPoliciesClient
		var mkImportStreamingEndpointsClient *mkiosdk.StreamingEndpointsClient
		var mkImportContentKeyPoliciesClient *mkiosdk.ContentKeyPoliciesClient

		// We need a login for import and validate. We should try that early so we don't do work if we can't login.
		if importResources || validateResources {
			log.Info("Logging into mk.io")
			var err error

			mkToken := os.Getenv("MKIO_TOKEN")
			if mkToken == "" {
				log.Fatalf("import Error: could not find MKIO_TOKEN environment variable")
			}

			// Create Clients
			mkImportAssetsClient, err = mkiosdk.NewAssetsClient(ctx, mkImportSubscription, mkToken, apiEndpoint, nil)
			if err != nil {
				log.Fatalf("error creating mk.io Assets Client: %v", err)
			}
			mkImportAssetFiltersClient, err = mkiosdk.NewAssetFiltersClient(ctx, mkImportSubscription, mkToken, apiEndpoint, nil)
			if err != nil {
				log.Fatalf("error creating mk.io Asset Filters Client: %v", err)
			}
			mkImportStreamingPoliciesClient, err = mkiosdk.NewStreamingPoliciesClient(ctx, mkImportSubscription, mkToken, apiEndpoint, nil)
			if err != nil {
				log.Fatalf("error creating mk.io StreamingPolicies Client: %v", err)
			}
			mkImportStreamingLocatorsClient, err = mkiosdk.NewStreamingLocatorsClient(ctx, mkImportSubscription, mkToken, apiEndpoint, nil)
			if err != nil {
				log.Fatalf("error creating mk.io StreamingLocators Client: %v", err)
			}
			mkImportStreamingEndpointsClient, err = mkiosdk.NewStreamingEndpointsClient(ctx, mkImportSubscription, mkToken, apiEndpoint, nil)
			if err != nil {
				log.Fatalf("error creating mk.io StreamingEndpoints Client: %v", err)
			}
			mkImportContentKeyPoliciesClient, err = mkiosdk.NewContentKeyPoliciesClient(ctx, mkImportSubscription, mkToken, apiEndpoint, nil)
			if err != nil {
				log.Fatalf("error creating mk.io ContentKeyPolicies Client: %v", err)
			}
		}

		// Read from Azure and generate an output file w/ the proper resources.
		if exportResources {
			// A couple of simple checks to avoid bad things
			if assetFilters && !assets {
				log.Fatalf("AssetFilter export requires Asset export")
			}

			if (azSubscription != "" || azResourceGroup != "" || azAccountName != "") && mkExportSubscription != "" {
				log.Fatal("export Error: cannot export from both Azure and mk.io subscription")
			}

			// Handle Azure Export
			if azSubscription != "" && azResourceGroup != "" && azAccountName != "" {
				// Login to Azure for the export
				var azureClient *migrate.AzureServiceProvider
				if exportResources {
					log.Info("Logging into Azure")
					var err error
					azureClient, err = migrate.NewAzureServiceProvider(azSubscription, azResourceGroup, azAccountName)
					if err != nil {
						log.Fatalf("unable to log into Azure: %v", err)
					}

				}

				// A couple of simple checks to avoid bad things
				if assetFilters && !assets {
					log.Fatalf("AssetFilter export requires Asset export")
				}

				log.Info("Starting Export from Azure")

				// Handle Assets
				if assets {
					start := time.Now()
					assetList, err := migrate.ExportAzAssets(ctx, azureClient, createdBefore, createdAfter)
					if err != nil {
						log.Errorf("error exporting assets: %v", err)
					}

					migrationContents.Assets = assetList
					timings = append(timings, results{resource: ASSETS, operation: EXPORT, duration: time.Since(start), migrated: len(assetList)})

					// Handle Asset Filters -- Can only do this if we have a list of assets
					if assetFilters {
						start = time.Now()
						assetFiltersList, err := migrate.ExportAzAssetFilters(ctx, azureClient, assetList, workers)
						if err != nil {
							log.Errorf("error exporting asset filters: %v", err)
						}
						count := 0
						// How many did we export?
						for _, v := range assetFiltersList {
							count = count + len(v)
						}
						migrationContents.AssetFilters = assetFiltersList

						timings = append(timings, results{resource: ASSETFILTERS, operation: EXPORT, duration: time.Since(start), migrated: count})
					}
				}

				// Handle Streaming Policies. These are used by StreamingLocators, so do it first
				if streamingPolicies {
					start := time.Now()

					sp, err := migrate.ExportAzStreamingPolicies(ctx, azureClient, createdBefore, createdAfter)
					if err != nil {
						log.Errorf("error exporting streaming policies: %v", err)
					}

					timings = append(timings, results{resource: STREAMINGPOLICIES, operation: EXPORT, duration: time.Since(start), migrated: len(sp)})

					migrationContents.StreamingPolicies = sp
				}

				// Handle StreamingLocators.
				if streamingLocators {
					start := time.Now()
					streamingLocatorsList, err := migrate.ExportAzStreamingLocators(ctx, azureClient, createdBefore, createdAfter)
					if err != nil {
						log.Errorf("error exporting streaming locators: %v", err)
					}

					timings = append(timings, results{resource: STREAMINGLOCATORS, operation: EXPORT, duration: time.Since(start), migrated: len(streamingLocatorsList)})

					migrationContents.StreamingLocators = streamingLocatorsList
				}

				// Handle StreamingEndpoints. Switching to handle as part of assets. They are related
				if streamingEndpoints {
					start := time.Now()
					se, err := migrate.ExportAzStreamingEndpoints(ctx, azureClient)
					if err != nil {
						log.Errorf("error exporting streaming locators: %v", err)
					}

					timings = append(timings, results{resource: STREAMINGENDPOINTS, operation: EXPORT, duration: time.Since(start), migrated: len(se)})

					migrationContents.StreamingEndpoints = se
				}
				// Handle StreamingEndpoints. Switching to handle as part of assets. They are related
				if contentKeyPolicies {
					start := time.Now()
					ckp, err := migrate.ExportAzContentKeyPolicies(ctx, azureClient, createdBefore, createdAfter)
					if err != nil {
						log.Errorf("error exporting content key policies: %v", err)
					}

					timings = append(timings, results{resource: CONTENTKEYPOLICIES, operation: EXPORT, duration: time.Since(start), migrated: len(ckp)})
					migrationContents.ContentKeyPolicies = ckp
				}
			} else if mkExportSubscription != "" {
				// Log into MKIO for the Export.
				mkToken := os.Getenv("MKIO_TOKEN")
				if mkToken == "" {
					log.Fatalf("import Error: could not find MKIO_TOKEN environment variable")
				}

				mkExportAssetsClient, err := mkiosdk.NewAssetsClient(ctx, mkExportSubscription, mkToken, apiEndpoint, nil)
				if err != nil {
					log.Fatalf("error creating mk.io Assets Client: %v", err)
				}
				mkExportAssetFiltersClient, err := mkiosdk.NewAssetFiltersClient(ctx, mkExportSubscription, mkToken, apiEndpoint, nil)
				if err != nil {
					log.Fatalf("error creating mk.io Asset Filters Client: %v", err)
				}
				mkExportStreamingLocatorsClient, err := mkiosdk.NewStreamingLocatorsClient(ctx, mkImportSubscription, mkToken, apiEndpoint, nil)
				if err != nil {
					log.Fatalf("error creating mk.io StreamingLocators Client: %v", err)
				}
				mkExportStreamingPoliciesClient, err := mkiosdk.NewStreamingPoliciesClient(ctx, mkImportSubscription, mkToken, apiEndpoint, nil)
				if err != nil {
					log.Fatalf("error creating mk.io StreamingPolicies Client: %v", err)
				}
				mkExportStreamingEndpointsClient, err := mkiosdk.NewStreamingEndpointsClient(ctx, mkImportSubscription, mkToken, apiEndpoint, nil)
				if err != nil {
					log.Fatalf("error creating mk.io StreamingEndpoints Client: %v", err)
				}
				mkExportContentKeyPoliciesClient, err := mkiosdk.NewContentKeyPoliciesClient(ctx, mkImportSubscription, mkToken, apiEndpoint, nil)
				if err != nil {
					log.Fatalf("error creating mk.io ContentKeyPolicies Client: %v", err)
				}

				log.Info("Starting Export from mk.io")

				// Handle Assets
				if assets {
					start := time.Now()
					assetList, err := migrate.ExportMkAssets(ctx, mkExportAssetsClient, createdBefore, createdAfter)
					if err != nil {
						log.Errorf("error exporting assets: %v", err)
					}
					timings = append(timings, results{resource: ASSETS, operation: EXPORT, duration: time.Since(start), migrated: len(assetList)})
					migrationContents.Assets = assetList

					// Handle Asset Filters -- Can only do this if we have a list of assets
					if assetFilters {
						start = time.Now()
						assetFiltersList, err := migrate.ExportMkAssetFilters(ctx, mkExportAssetFiltersClient, assetList)
						if err != nil {
							log.Errorf("error exporting asset filters: %v", err)
						}
						count := 0
						// How many did we export?
						for _, v := range assetFiltersList {
							count = count + len(v)
						}
						timings = append(timings, results{resource: ASSETFILTERS, operation: EXPORT, duration: time.Since(start), migrated: count})
						migrationContents.AssetFilters = assetFiltersList

					}
				}

				// Handle Streaming Policies. These are used by StreamingLocators, so do it first
				if streamingPolicies {
					start := time.Now()
					sp, err := migrate.ExportMkStreamingPolicies(ctx, mkExportStreamingPoliciesClient, createdBefore, createdAfter)
					if err != nil {
						log.Errorf("error exporting streaming policies: %v", err)
					}
					timings = append(timings, results{resource: STREAMINGPOLICIES, operation: EXPORT, duration: time.Since(start), migrated: len(sp)})

					migrationContents.StreamingPolicies = sp
				}

				// Handle StreamingLocators.
				if streamingLocators {
					start := time.Now()
					streamingLocatorsList, err := migrate.ExportMkStreamingLocators(ctx, mkExportStreamingLocatorsClient, createdBefore, createdAfter)
					if err != nil {
						log.Errorf("error exporting streaming locators: %v", err)
					}
					timings = append(timings, results{resource: STREAMINGLOCATORS, operation: EXPORT, duration: time.Since(start), migrated: len(streamingLocatorsList)})
					migrationContents.StreamingLocators = streamingLocatorsList
				}

				// // Handle StreamingEndpoints. Switching to handle as part of assets. They are related
				if streamingEndpoints {
					start := time.Now()
					se, err := migrate.ExportMkStreamingEndpoints(ctx, mkExportStreamingEndpointsClient)
					if err != nil {
						log.Errorf("error exporting streaming locators: %v", err)
					}
					timings = append(timings, results{resource: STREAMINGENDPOINTS, operation: EXPORT, duration: time.Since(start), migrated: len(se)})

					migrationContents.StreamingEndpoints = se
				}
				// // Handle StreamingEndpoints. Switching to handle as part of assets. They are related
				if contentKeyPolicies {
					start := time.Now()
					ckp, err := migrate.ExportMkContentKeyPolicies(ctx, mkExportContentKeyPoliciesClient, createdBefore, createdAfter)
					if err != nil {
						log.Errorf("error exporting content key policies: %v", err)
					}
					timings = append(timings, results{resource: CONTENTKEYPOLICIES, operation: EXPORT, duration: time.Since(start), migrated: len(ckp)})
					migrationContents.ContentKeyPolicies = ckp
				}
				// } else {
				// 	log.Fatal("export Error: cannot export without Azure or mk.io subscription information")
				// }
			}

			err := migrationContents.WriteMigrationFile(ctx, migrationFile)
			if err != nil {
				// No point continuing w/o this file... Exit
				log.Fatalf("unable to write migration export file contents: %v", err)
			}
			log.Infof("Done exporting. Exported content written to file: %s", migrationFile)
		}

		// Handle Import into mk.io
		if importResources {
			log.Info("Starting Import to mk.io")

			// Read migration file & populate migration contents from it
			contents := migrate.MigrationFileContents{}
			err := contents.ReadMigrationFile(ctx, migrationFile)
			if err != nil {
				log.Fatalf("could not read migration file: %v", err)
			}

			// Handling ConentKeyPolicies. This should happen before StreamingLocators
			if contentKeyPolicies {
				start := time.Now()
				success, skipped, failureList, err := migrate.ImportContentKeyPolicies(ctx, mkImportContentKeyPoliciesClient, contents.ContentKeyPolicies, overwrite, fairplayAmsCompatibility)
				if err != nil {
					log.Errorf("error importing content key policies: %v", err)
				}
				timings = append(timings, results{resource: CONTENTKEYPOLICIES, operation: IMPORT, duration: time.Since(start), skipped: skipped, failures: failureList, migrated: success})
			}

			// Handling Assets
			if assets {
				start := time.Now()
				success, skipped, failureList, err := migrate.ImportAssets(ctx, mkImportAssetsClient, contents.Assets, overwrite)
				if err != nil {
					log.Errorf("error importing assets: %v", err)
				}
				timings = append(timings, results{resource: ASSETS, operation: IMPORT, duration: time.Since(start), skipped: skipped, failures: failureList, migrated: success})
			}

			// Handling Asset Filters. These require an asset, so import after assets
			if assetFilters {
				start := time.Now()
				success, skipped, failureList, err := migrate.ImportAssetFilters(ctx, mkImportAssetFiltersClient, contents.AssetFilters, overwrite)
				if err != nil {
					log.Errorf("error importing asset filters: %v", err)
				}
				timings = append(timings, results{resource: ASSETFILTERS, operation: IMPORT, duration: time.Since(start), skipped: skipped, failures: failureList, migrated: success})
			}

			// Handling StreamingPolicies
			if streamingPolicies {
				start := time.Now()
				success, skipped, failureList, err := migrate.ImportStreamingPolicies(ctx, mkImportStreamingPoliciesClient, contents.StreamingPolicies, overwrite)
				if err != nil {
					log.Errorf("error importing streaming policies: %v", err)
				}
				timings = append(timings, results{resource: STREAMINGPOLICIES, operation: IMPORT, duration: time.Since(start), skipped: skipped, failures: failureList, migrated: success})
			}
			// Handling StreamingLocators
			if streamingLocators {
				start := time.Now()
				success, skipped, failureList, err := migrate.ImportStreamingLocators(ctx, mkImportStreamingLocatorsClient, contents.StreamingLocators, overwrite)
				if err != nil {
					log.Errorf("error importing streaming locators: %v", err)
				}
				timings = append(timings, results{resource: STREAMINGLOCATORS, operation: IMPORT, duration: time.Since(start), skipped: skipped, failures: failureList, migrated: success})
			}

			// Handling StreamingEndpoints
			if streamingEndpoints {
				start := time.Now()
				success, skipped, failureList, err := migrate.ImportStreamingEndpoints(ctx, mkImportStreamingEndpointsClient, contents.StreamingEndpoints, overwrite)
				if err != nil {
					log.Errorf("error importing streaming endpoints: %v", err)
				}
				timings = append(timings, results{resource: STREAMINGENDPOINTS, operation: IMPORT, duration: time.Since(start), skipped: skipped, failures: failureList, migrated: success})
			}
		}

		// Handle Validation of imported Streaming Locators/Endpoints
		if validateResources {
			mkToken := os.Getenv("MKIO_TOKEN")
			if mkToken == "" {
				log.Fatalf("validation Error: could not find MKIO_TOKEN environment variable")
			}

			// Read migration file & populate migration contents from it
			contents := migrate.MigrationFileContents{}
			err := contents.ReadMigrationFile(ctx, migrationFile)
			if err != nil {
				log.Fatalf("could not read migration file: %v", err)
			}

			// Handling StreamingLocators
			if streamingLocators {
				err := migrate.ValidateStreamingLocators(ctx, mkImportStreamingLocatorsClient, mkImportStreamingEndpointsClient, contents.StreamingLocators)
				if err != nil {
					log.Errorf("error validating streamingLocators: %v", err)
				}
			}
		}

		// Write out results
		fmt.Println("Results:")
		w := tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		_, _ = fmt.Fprintf(w, "Operation\tResource\tMigrated\tSkipped\tFailed\tDuration\n")
		for _, v := range timings {
			// Some output to give stats at the end
			if v.operation == EXPORT {
				_, _ = fmt.Fprintf(w, "%v\t%v\t%d\t-\t-\t%v\n", v.operation, v.resource, v.migrated, v.duration)
			} else if v.operation == IMPORT {
				_, _ = fmt.Fprintf(w, "%v\t%v\t%d\t%d\t%d\t%v\n", v.operation, v.resource, v.migrated, v.skipped, len(v.failures), v.duration)
			}
		}
		w.Flush()

		fmt.Println("\nFailures:")
		for _, v := range timings {
			if len(v.failures) > 0 {
				fmt.Printf("\tFailed to %v %v: %v\n", v.operation, v.resource, v.failures)
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&azSubscription, "azure-subscription", "", "Azure Subscription ID for existing AMS")
	rootCmd.PersistentFlags().StringVar(&azResourceGroup, "azure-resource-group", "", "Resource Group for existing AMS")
	rootCmd.PersistentFlags().StringVar(&azAccountName, "azure-account-name", "", "Account Name for existing AMS")
	rootCmd.PersistentFlags().StringVar(&mkImportSubscription, "mediakind-import-subscription", "", "Mediakind Subscription ID for import in mk.io")
	rootCmd.PersistentFlags().StringVar(&mkExportSubscription, "mediakind-export-subscription", "", "Mediakind Subscription ID for export in mk.io")
	rootCmd.PersistentFlags().StringVar(&apiEndpoint, "api-endpoint", "https://api.mk.io", "mk.io API endpoint")
	rootCmd.PersistentFlags().StringVar(&createdBefore, "created-before", "", "filter export for resources created before date")
	rootCmd.PersistentFlags().StringVar(&createdAfter, "created-after", "", "filter export for resources created after date")
	rootCmd.PersistentFlags().IntVar(&workers, "workers", 1, "number of workers to run in parallel")

	rootCmd.PersistentFlags().StringVar(&migrationFile, "migration-file", "", "Migration filename")

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&exportResources, "export", false, "Toggle export from AMS")
	rootCmd.PersistentFlags().BoolVar(&importResources, "import", false, "Toggle import into mk.io")
	rootCmd.PersistentFlags().BoolVar(&validateResources, "validate", false, "Toggle validate in mk.io")
	rootCmd.PersistentFlags().BoolVar(&overwrite, "overwrite", false, "overwrite resources that already exist")

	rootCmd.PersistentFlags().BoolVar(&assets, "assets", false, "Run Export/Import on Assets")
	rootCmd.PersistentFlags().BoolVar(&assetFilters, "asset-filters", false, "Run Export/Import on Asset Filters")
	rootCmd.PersistentFlags().BoolVar(&contentKeyPolicies, "content-key-policies", false, "Run Export/Import on ContentKeyPolicies")
	rootCmd.PersistentFlags().BoolVar(&streamingLocators, "streaming-locators", false, "run Export/Import on StreamingLocators")
	rootCmd.PersistentFlags().BoolVar(&streamingEndpoints, "streaming-endpoints", false, "run Export/Import on StreamingEndpoints")
	rootCmd.PersistentFlags().BoolVar(&streamingPolicies, "streaming-policies", false, "run Export/Import on StreamingPolicies")

	rootCmd.PersistentFlags().BoolVar(&fairplayAmsCompatibility, "fairplay-ams-compatibility", false, "set fairPlayAmsCompatibility=true for all fairplay content key policies")

	// Configure Logger
	// log.SetFormatter(&log.JSONFormatter{})

}
