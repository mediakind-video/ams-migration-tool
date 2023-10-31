# AMS Migration Tool

## Introduction

This project allows for the bulk migration from Azure Media Services to MKIO, allowing an easy way to migrate assets.

The tool needs access to both Azure and MKIO to export and import resources, respectively.

## Migration Process

1. Export resources from Azure Media Services as JSON
2. Import resources into MKIO
3. Validate resources in MKIO

## Current State

The project can run in three modes, which can be combined in one execution. [Export,Import,Validate]

Export - Pulls data from Azure MediaServices, creating a JSON file as output
Import - Reads a json file and inserts data into MKIO
Validate - Validates imported assets

These modes are currently command-line flags, rather than separate cobra commands. This gives us the option to run multiple modes at once, allowing users to run a single command to fully migrate items.

### Supported Resources

This migration tool currently works for the following resources:

- Assets
- Asset Filters
- Streaming Endpoints
- Streaming Locators
- Content Key Policies

## Running the migration

## Demo

A detailed demo can be found [here](docs/demo/demo.md)

### Prerequisites

The migration tool doesn't currently handle Get/Create of Storage Accounts. It expects the same Storage Account, with the same name, to be in place in both Azure Media Services and MKIO.

#### Setting up Storage Account in MKIO

1. Navigate to the desired Azure Media Service page in your browser
2. Select `Storage accounts` in the `Settings` section
3. Follow the link to the storage account
   - Note the name of the Media Service account for use in MKIO StorageAccount Creation\
   - There may be more than one here. You will need to complete this process for each. It is expected that this is a limited number. If this turns out to not be the case we should add support to do this automatically
4. Select `Shared access signature` under the `Security + networking` section
   1. Check `Service`, `Object`, and `Container` in `Allowed resource types`
   2. Update the expiry date to be after the desired lifetime of the resources
   3. Click `Generate SAS and connection string`
   4. Copy the `SAS token` to insert into MKIO
   5. Copy the address of the Blob to insert into MKIO
5. Create the Storage Account in MKIO using the information gathered above

#### Azure Integration

This is needed for Export.

Log into Azure from your terminal. Your set Azure account must have access to the Subscription/ResourceGroup you intend to migrate.

#### MKIO integration

This is needed for Import and Validation.

The following instructions contain links to the Dev instance of MKIO. Use similar steps for Prod.

1. Log into [MKIO UI](https://dev-ui.io.mediakind.com/)
2. Get your token (At the moment this only works in an incognito window). [MKIO Token](https://dev.io.mediakind.com/auth/token/)

### Running

You can run the migration from vscode using .vscode/launch.json in this project.

!! Make sure to insert your own Azure Subscription and ResourceGroup and your MKIO token.

To run from the command line use the command:

```bash
go run main.go --export --import ...
```

You can run export and import in the same command, which will automatically import all the exported data into MKIO. You can also run the only export to generate a JSON file, which can then be modified as desired before running the import. This could be useful if storageAccount names differ, or only specific asset migrations are desired.

## Build

### Go Build Command

Run the following command to build the go binary for linux
`GOOS=linux GOARCH=amd64 go build -o mkio-ams-migration`

## Additional Documentation

[MKIO Swagger](https://dev.io.mediakind.com/doc/ui/)

[Azure Media Services SDK](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices#pkg-types)
