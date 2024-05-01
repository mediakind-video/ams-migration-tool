package cmd

import (
	"context"
	"fmt"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	migrate "dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/migration"
	"dev.azure.com/mediakind/mkio/ams-migration-tool.git/pkg/mkiosdk"
)

// command line options
var (
	azSubscription  string
	azResourceGroup string
	azAccountName   string
	mkSubscription  string
	migrationFile   string
	apiEndpoint     string

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
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate AMS Assets",
	Long:  `Migrate Assets and StreamingLocators from Azure MediaServices to MKIO.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	Run: func(cmd *cobra.Command, args []string) {

		ctx := context.Background()

		if debug {
			log.Info("Debug enabled")
			log.SetLevel(log.DebugLevel)
		}

		// If we don't export,import,validate what do we do?
		if !exportResources && !importResources && !validateResources {
			log.Fatal("Please select a valid command: [import|export|validate]")
		}

		// Set a timestamp on our migraiton file
		if migrationFile == "" {
			migrationFile = fmt.Sprintf("migration-%v.json", time.Now().Unix())
		}

		migrationContents := migrate.MigrationFileContents{}

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

		// Log into MKIO for the Import. Do this first so we know if it fails before we do any work.
		var mkAssetsClient *mkiosdk.AssetsClient
		var mkAssetFiltersClient *mkiosdk.AssetFiltersClient
		var mkStreamingLocatorsClient *mkiosdk.StreamingLocatorsClient
		var mkStreamingEndpointsClient *mkiosdk.StreamingEndpointsClient
		var mkContentKeyPoliciesClient *mkiosdk.ContentKeyPoliciesClient

		// We need a login for import and validate. We should try that early so we don't do work if we can't login.
		if importResources || validateResources {
			log.Info("Logging into MK/IO")
			var err error

			mkToken := os.Getenv("MKIO_TOKEN")
			if mkToken == "" {
				log.Fatalf("import Error: could not find MKIO_TOKEN environment variable")
			}

			// Create Clients
			mkAssetsClient, err = mkiosdk.NewAssetsClient(ctx, mkSubscription, mkToken, apiEndpoint, nil)
			if err != nil {
				log.Fatalf("error creating MKIO Assets Client: %v", err)
			}
			mkAssetFiltersClient, err = mkiosdk.NewAssetFiltersClient(ctx, mkSubscription, mkToken, apiEndpoint, nil)
			if err != nil {
				log.Fatalf("error creating MKIO Asset Filters Client: %v", err)
			}
			mkStreamingLocatorsClient, err = mkiosdk.NewStreamingLocatorsClient(ctx, mkSubscription, mkToken, apiEndpoint, nil)
			if err != nil {
				log.Fatalf("error creating MKIO StreamingLocators Client: %v", err)
			}
			mkStreamingEndpointsClient, err = mkiosdk.NewStreamingEndpointsClient(ctx, mkSubscription, mkToken, apiEndpoint, nil)
			if err != nil {
				log.Fatalf("error creating MKIO StreamingEndpoints Client: %v", err)
			}
			mkContentKeyPoliciesClient, err = mkiosdk.NewContentKeyPoliciesClient(ctx, mkSubscription, mkToken, apiEndpoint, nil)
			if err != nil {
				log.Fatalf("error creating MKIO ContentKeyPolicies Client: %v", err)

			}
		}

		// Read from Azure and generate an output file w/ the proper resources.
		if exportResources {
			// A couple of simple checks to avoid bad things
			if assetFilters && !assets {
				log.Fatalf("AssetFilter export requires Asset export")
			}

			log.Info("Starting Export from Azure")

			// Handle Assets
			if assets {
				assetList, err := migrate.ExportAssets(ctx, azureClient)
				if err != nil {
					log.Errorf("error exporting assets: %v", err)
				}

				migrationContents.Assets = assetList

				// Handle Asset Filters -- Can only do this if we have a list of assets
				if assetFilters {
					assetFiltersList, err := migrate.ExportAssetFilters(ctx, azureClient, assetList)
					if err != nil {
						log.Errorf("error exporting asset filters: %v", err)
					}

					migrationContents.AssetFilters = assetFiltersList

				}
			}

			// Handle StreamingLocators.
			if streamingLocators {
				streamingLocatorsList, err := migrate.ExportStreamingLocators(ctx, azureClient)
				if err != nil {
					log.Errorf("error exporting streaming locators: %v", err)
				}

				migrationContents.StreamingLocators = streamingLocatorsList
			}

			// Handle StreamingEndpoints. Switching to handle as part of assets. They are related
			if streamingEndpoints {
				se, err := migrate.ExportStreamingEndpoints(ctx, azureClient)
				if err != nil {
					log.Errorf("error exporting streaming locators: %v", err)
				}

				migrationContents.StreamingEndpoints = se
			}

			// Handle StreamingEndpoints. Switching to handle as part of assets. They are related
			if contentKeyPolicies {
				ckp, err := migrate.ExportContentKeyPolicies(ctx, azureClient)
				if err != nil {
					log.Errorf("error exporting content key policies: %v", err)
				}
				migrationContents.ContentKeyPolicies = ckp
			}

			err := migrationContents.WriteMigrationFile(ctx, migrationFile)
			if err != nil {
				// No point continuing w/o this file... Exit
				log.Fatalf("unable to write migration export file contents: %v", err)
			}

			log.Infof("Done exporting. Exported content written to file: %s", migrationFile)
		}

		// Handle Import into MKIO
		if importResources {
			log.Info("Starting Import to MK/IO")

			// Read migration file & populate migration contents from it
			contents := migrate.MigrationFileContents{}
			err := contents.ReadMigrationFile(ctx, migrationFile)
			if err != nil {
				log.Fatalf("could not read migration file: %v", err)
			}

			// Handling ConentKeyPolicies. This should happen before StreamingLocators
			if contentKeyPolicies {
				err := migrate.ImportContentKeyPolicies(ctx, mkContentKeyPoliciesClient, contents.ContentKeyPolicies, overwrite)
				if err != nil {
					log.Errorf("error importing content key policies: %v", err)
				}
			}

			// Handling Assets
			if assets {
				err := migrate.ImportAssets(ctx, mkAssetsClient, contents.Assets, overwrite)
				if err != nil {
					log.Errorf("error importing assets: %v", err)
				}
			}

			// Handling Asset Filters. These require an asset, so import after assets
			if assetFilters {
				err := migrate.ImportAssetFilters(ctx, mkAssetFiltersClient, contents.AssetFilters, overwrite)
				if err != nil {
					log.Errorf("error importing asset filters: %v", err)
				}
			}

			// Handling StreamingLocators
			if streamingLocators {
				err := migrate.ImportStreamingLocators(ctx, mkStreamingLocatorsClient, contents.StreamingLocators, overwrite)
				if err != nil {
					log.Errorf("error importing streaming locators: %v", err)
				}
			}

			// Handling StreamingEndpoints
			if streamingEndpoints {
				err := migrate.ImportStreamingEndpoints(ctx, mkStreamingEndpointsClient, contents.StreamingEndpoints, overwrite)
				if err != nil {
					log.Errorf("error importing streaming endpoints: %v", err)
				}
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
				err := migrate.ValidateStreamingLocators(ctx, mkStreamingLocatorsClient, mkStreamingEndpointsClient, contents.StreamingLocators)
				if err != nil {
					log.Errorf("error validating streamingLocators: %v", err)
				}
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
	rootCmd.PersistentFlags().StringVar(&mkSubscription, "mediakind-subscription", "", "Mediakind Subscription ID for MKIO")
	rootCmd.PersistentFlags().StringVar(&apiEndpoint, "api-endpoint", "https://api.mk.io", "MKIO API endpoint")

	rootCmd.PersistentFlags().StringVar(&migrationFile, "migration-file", "", "Migration filename")

	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "Enable debug logging")
	rootCmd.PersistentFlags().BoolVar(&exportResources, "export", false, "Toggle export from AMS")
	rootCmd.PersistentFlags().BoolVar(&importResources, "import", false, "Toggle import into MKIO")
	rootCmd.PersistentFlags().BoolVar(&validateResources, "validate", false, "Toggle validate in MKIO")
	rootCmd.PersistentFlags().BoolVar(&overwrite, "overwrite", false, "overwrite resources that already exist")

	rootCmd.PersistentFlags().BoolVar(&assets, "assets", false, "Run Export/Import on Assets")
	rootCmd.PersistentFlags().BoolVar(&assetFilters, "asset-filters", false, "Run Export/Import on Asset Filters")
	rootCmd.PersistentFlags().BoolVar(&contentKeyPolicies, "content-key-policies", false, "Run Export/Import on ContentKeyPolicies")
	rootCmd.PersistentFlags().BoolVar(&streamingLocators, "streaming-locators", false, "run Export/Import on StreamingLocators")
	rootCmd.PersistentFlags().BoolVar(&streamingEndpoints, "streaming-endpoints", false, "run Export/Import on StreamingEndpoints")

	// Configure Logger
	// log.SetFormatter(&log.JSONFormatter{})

}
