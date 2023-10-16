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

		// Leting users select both will overwrite the file... Might be unexpected
		if exportResources && migrationFile != "" {
			log.Fatal("Select --export or --migration-file. Selecting both will overwrite your current file.")
		}

		// Set a timestamp on our migraiton file
		if migrationFile == "" {
			migrationFile = fmt.Sprintf("migration-%v.json", time.Now().Unix())
		}

		migrationContents := migrate.MigrationFileContents{}

		// Export from AMS
		if exportResources {
			// A couple of simple checks to avoid bad things
			if assetFilters && !assets {
				log.Fatalf("AssetFilter export requires Asset export")
			}

			log.Info("Starting Export from Azure")
			// Read from Azure and generate an output file w/ the proper resources.
			client, err := migrate.NewAzureServiceProvider(azSubscription, azResourceGroup, azAccountName)
			if err != nil {
				log.Fatalf("unable to log into Azure: %v", err)
			}

			// Handle Assets
			if assets {
				assetList, err := migrate.ExportAssets(ctx, client)
				if err != nil {
					log.Errorf("error exporting assets: %v", err)
				}

				migrationContents.Assets = assetList

				// Handle Asset Filters -- Can only do this if we have a list of assets
				if assetFilters {
					assetFiltersList, err := migrate.ExportAssetFilters(ctx, client, assetList)
					if err != nil {
						log.Errorf("error exporting asset filters: %v", err)
					}

					migrationContents.AssetFilters = assetFiltersList

				}
			}

			// Handle StreamingLocators.
			if streamingLocators {
				streamingLocatorsList, err := migrate.ExportStreamingLocators(ctx, client)
				if err != nil {
					log.Errorf("error exporting streaming locators: %v", err)
				}

				migrationContents.StreamingLocators = streamingLocatorsList
			}

			// Handle StreamingEndpoints. Switching to handle as part of assets. They are related
			if streamingEndpoints {
				se, err := migrate.ExportStreamingEndpoints(ctx, client)
				if err != nil {
					log.Errorf("error exporting streaming locators: %v", err)
				}

				migrationContents.StreamingEndpoints = se
			}

			// Handle StreamingEndpoints. Switching to handle as part of assets. They are related
			if contentKeyPolicies {
				ckp, err := migrate.ExportContentKeyPolicies(ctx, client)
				if err != nil {
					log.Errorf("error exporting content key policies: %v", err)
				}
				migrationContents.ContentKeyPolicies = ckp
			}

			err = migrationContents.WriteMigrationFile(ctx, migrationFile)
			if err != nil {
				// No point continuing w/o this file... Exit
				log.Fatalf("unable to write migration export file contents: %v", err)
			}

			log.Infof("Done exporting. Exported content written to file: %s", migrationFile)
		}

		// Handle Import into MKIO
		if importResources {
			log.Info("Starting Import to MK/IO")

			mkToken := os.Getenv("MKIO_TOKEN")
			if mkToken == "" {
				log.Fatalf("import Error: could not find MKIO_TOKEN environment variable")
			}

			// Create Clients
			assetsClient, err := mkiosdk.NewAssetsClient(mkSubscription, mkToken, nil)
			if err != nil {
				log.Fatalf("error creating MKIO Assets Client: %v", err)
			}
			assetFiltersClient, err := mkiosdk.NewAssetFiltersClient(mkSubscription, mkToken, nil)
			if err != nil {
				log.Fatalf("error creating MKIO Asset Filters Client: %v", err)
			}
			streamingLocatorsClient, err := mkiosdk.NewStreamingLocatorsClient(mkSubscription, mkToken, nil)
			if err != nil {
				log.Fatalf("error creating MKIO StreamingLocators Client: %v", err)
			}
			streamingEndpointsClient, err := mkiosdk.NewStreamingEndpointsClient(mkSubscription, mkToken, nil)
			if err != nil {
				log.Fatalf("error creating MKIO StreamingEndpoints Client: %v", err)
			}
			contentKeyPoliciesClient, err := mkiosdk.NewContentKeyPoliciesClient(mkSubscription, mkToken, nil)
			if err != nil {
				log.Fatalf("error creating MKIO ContentKeyPolicies Client: %v", err)

			}

			// Read migration file & populate migration contents from it
			contents := migrate.MigrationFileContents{}
			err = contents.ReadMigrationFile(ctx, migrationFile)
			if err != nil {
				log.Fatalf("could not read migration file: %v", err)
			}

			// Handling Asset Filters
			if assetFilters {
				err := migrate.ImportAssetFilters(ctx, assetFiltersClient, contents.AssetFilters, overwrite)
				if err != nil {
					log.Errorf("error importing asset filters: %v", err)
				}
			}

			// Handling Assets
			if assets {
				err := migrate.ImportAssets(ctx, assetsClient, contents.Assets, overwrite)
				if err != nil {
					log.Errorf("error importing assets: %v", err)
				}
			}

			// Handling ConentKeyPolicies. This should happen before StreamingLocators
			if contentKeyPolicies {
				err := migrate.ImportContentKeyPolicies(ctx, contentKeyPoliciesClient, contents.ContentKeyPolicies, overwrite)
				if err != nil {
					log.Errorf("error importing content key policies: %v", err)
				}
			}

			// Handling StreamingLocators
			if streamingLocators {
				err := migrate.ImportStreamingLocators(ctx, streamingLocatorsClient, contents.StreamingLocators, overwrite)
				if err != nil {
					log.Errorf("error importing streaming locators: %v", err)
				}
			}
			// Handling StreamingEndpoints
			if streamingEndpoints {
				err := migrate.ImportStreamingEndpoints(ctx, streamingEndpointsClient, contents.StreamingEndpoints, overwrite)
				if err != nil {
					log.Errorf("error importing streaming endpoints: %v", err)
				}
			}
		}

		// Handle Import into MKIO
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

			// Create clients
			streamingLocatorsClient, err := mkiosdk.NewStreamingLocatorsClient(mkSubscription, mkToken, nil)
			if err != nil {
				log.Fatalf("error creating MKIO StreamingLocators Client: %v", err)
			}
			streamingEndpointsClient, err := mkiosdk.NewStreamingEndpointsClient(mkSubscription, mkToken, nil)
			if err != nil {
				log.Fatalf("error creating MKIO StreamingEndpoints Client: %v", err)
			}

			// Handling StreamingLocators
			if streamingLocators {
				err := migrate.ValidateStreamingLocators(ctx, streamingLocatorsClient, streamingEndpointsClient, contents.StreamingLocators)
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
