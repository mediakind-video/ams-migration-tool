package migrate

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	log "github.com/sirupsen/logrus"
)

type AzureServiceProvider struct {
	// config               AzureServiceProviderConfig
	credential               *azidentity.DefaultAzureCredential
	subscriptionId           string
	resourceGroup            string
	accountName              string
	assetsClient             *armmediaservices.AssetsClient
	assetFiltersClient       *armmediaservices.AssetFiltersClient
	streamingLocatorsClient  *armmediaservices.StreamingLocatorsClient
	streamingEndpointsClient *armmediaservices.StreamingEndpointsClient
	contentKeyPoliciesClient *armmediaservices.ContentKeyPoliciesClient

	// storageClientFactory *armstorage.ClientFactory
	accountsClient *armstorage.AccountsClient
}

func NewAzureServiceProvider(subscription string, resourceGroup string, accountName string) (*AzureServiceProvider, error) {

	log.Info("Logging into Azure")
	credential, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %v", err)
	}
	// Get a Azure MediaServices Assets Client
	assetsClient, err := armmediaservices.NewAssetsClient(subscription, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Media Service Client: %v", err)
	}
	// Get a Azure MediaServices Asset filters Client
	assetFiltersClient, err := armmediaservices.NewAssetFiltersClient(subscription, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Media Service Client: %v", err)
	}
	// Get a Azure MediaServices StreamingLocator Client
	streamingLocatorsClient, err := armmediaservices.NewStreamingLocatorsClient(subscription, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Media Service Client: %v", err)
	}
	// Get a Azure MediaServices StreamingEndpoints Client
	streamingEndpointsClient, err := armmediaservices.NewStreamingEndpointsClient(subscription, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Media Service Client: %v", err)
	}
	// Get a Azure MediaServices StreamingEndpoints Client
	contentKeyPoliciesClient, err := armmediaservices.NewContentKeyPoliciesClient(subscription, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Media Service Client: %v", err)
	}
	accountsClient, err := armstorage.NewAccountsClient(subscription, credential, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create accounts client: %v", err)
	}

	return &AzureServiceProvider{
		subscriptionId:           subscription,
		accountName:              accountName,
		resourceGroup:            resourceGroup,
		assetsClient:             assetsClient,
		assetFiltersClient:       assetFiltersClient,
		streamingLocatorsClient:  streamingLocatorsClient,
		streamingEndpointsClient: streamingEndpointsClient,
		accountsClient:           accountsClient,
		contentKeyPoliciesClient: contentKeyPoliciesClient,
		credential:               credential,
	}, nil
}

// lookupAssets  Get assets from Azure MediaServices. Remove pagination
func (a *AzureServiceProvider) lookupAssets(ctx context.Context) ([]*armmediaservices.Asset, error) {
	client := a.assetsClient

	pager := client.NewListPager(a.resourceGroup, a.accountName, nil)

	assets := []*armmediaservices.Asset{}

	// We get pages back. Loop through pages and create a list of assets
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			log.Debugf("Id: %s, Name: %s, Type: %s, Container: %s, StorageAccountName: %s, AssetId: %s\n", *v.ID, *v.Name, *v.Type, *v.Properties.Container, *v.Properties.StorageAccountName, *v.Properties.AssetID)
			assets = append(assets, v)
		}
	}
	return assets, nil
}

// lookupAssetFilters  Get asset filters from Azure MediaServices. Remove pagination
func (a *AzureServiceProvider) lookupAssetFilters(ctx context.Context, assetName string) ([]*armmediaservices.AssetFilter, error) {
	client := a.assetFiltersClient

	pager := client.NewListPager(a.resourceGroup, a.accountName, assetName, nil)

	assetFilters := []*armmediaservices.AssetFilter{}

	// We get pages back. Loop through pages and create a list of asset filters
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			// log.Debugf("Id: %s, Name: %s, Type: %s, Container: %s, StorageAccountName: %s, AssetId: %s\n", *v.ID, *v.Name, *v.Type, *v.Properties.Container, *v.Properties.StorageAccountName, *v.Properties.AssetID)
			assetFilters = append(assetFilters, v)
		}
	}
	return assetFilters, nil
}

// lookupStreamingLocators Get StreamingLocators from Azure MediaServices. Remove pagination
func (a *AzureServiceProvider) lookupStreamingLocators(ctx context.Context) ([]*armmediaservices.StreamingLocator, error) {
	client := a.streamingLocatorsClient
	sl := []*armmediaservices.StreamingLocator{}

	pager := client.NewListPager(a.resourceGroup, a.accountName, nil)

	// Paginated result. We just need a list. loop through and generate that list
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			log.Debugf("Id: %s, Name: %s, Type: %s, AssetName: %s, StreamingLocatorID: %s, StreamingPolicyName: %s\n", *v.ID, *v.Name, *v.Type, *v.Properties.AssetName, *v.Properties.StreamingLocatorID, *v.Properties.StreamingPolicyName)
			sl = append(sl, v)
		}
	}
	return sl, nil
}

// lookupStreamingEndpoints Get StreamingEndpoints from Azure MediaServices. Remove pagination
func (a *AzureServiceProvider) lookupStreamingEndpoints(ctx context.Context) ([]*armmediaservices.StreamingEndpoint, error) {
	client := a.streamingEndpointsClient
	se := []*armmediaservices.StreamingEndpoint{}

	pager := client.NewListPager(a.resourceGroup, a.accountName, nil)

	// Paginated result. We just need a list. loop through and generate that list
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			log.Debugf("Id: %s, Name: %s, Type: %s, Location: %s\n", *v.ID, *v.Name, *v.Type, *v.Location)
			se = append(se, v)
		}
	}
	return se, nil
}

// lookupContentKeyPolicies Get contentKeyPolicy from Azure MediaServices. Remove pagination
func (a *AzureServiceProvider) lookupContentKeyPolicies(ctx context.Context) ([]*armmediaservices.ContentKeyPolicy, error) {
	client := a.contentKeyPoliciesClient

	pager := client.NewListPager(a.resourceGroup, a.accountName, nil)

	ckp := []*armmediaservices.ContentKeyPolicy{}

	// We get pages back. Loop through pages and create a list of assets
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			// log.Debugf("Id: %s, Name: %s, Type: %s, Container: %s, StorageAccountName: %s, AssetId: %s\n", *v.ID, *v.Name, *v.Type, *v.Properties.Container, *v.Properties.StorageAccountName, *v.Properties.AssetID)
			ckp = append(ckp, v)
		}
	}
	return ckp, nil
}
