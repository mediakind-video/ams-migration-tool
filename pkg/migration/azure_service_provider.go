package migrate

import (
	"context"
	"fmt"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
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
	streamingPoliciesClient  *armmediaservices.StreamingPoliciesClient
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
	// Get a Azure MediaServices StreamingPolicies Client
	streamingPoliciesClient, err := armmediaservices.NewStreamingPoliciesClient(subscription, credential, nil)
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
		streamingPoliciesClient:  streamingPoliciesClient,
		credential:               credential,
	}, nil
}

func generateFilter(before string, after string) string {
	filter := ""
	// Handle after date
	if after != "" {
		filter = fmt.Sprintf("properties/created gt %s", after)
	}
	// Before & after. Add an and in between
	if before != "" && after != "" {
		filter = fmt.Sprintf("%s and", filter)
	}
	// Handle before date
	if before != "" {
		filter = fmt.Sprintf("%s properties/created lt %s", filter, before)
	}
	return filter
}

// lookupAssets  Get assets from Azure MediaServices. Remove pagination
func (a *AzureServiceProvider) lookupAssets(ctx context.Context, before string, after string) ([]*armmediaservices.Asset, error) {
	client := a.assetsClient

	// Generate the filter
	filter := generateFilter(before, after)

	// If we have a filter apply it
	options := &armmediaservices.AssetsClientListOptions{Orderby: to.Ptr("properties/created")}
	if filter != "" {
		options.Filter = to.Ptr(filter)
	}

	pager := client.NewListPager(a.resourceGroup, a.accountName, options)

	assets := []*armmediaservices.Asset{}

	// We get pages back. Loop through pages and create a list of assets
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return assets, fmt.Errorf("failed to advance page: %v", err)
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
			return assetFilters, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			// log.Debugf("Id: %s, Name: %s, Type: %s, Container: %s, StorageAccountName: %s, AssetId: %s\n", *v.ID, *v.Name, *v.Type, *v.Properties.Container, *v.Properties.StorageAccountName, *v.Properties.AssetID)
			assetFilters = append(assetFilters, v)
		}
	}
	return assetFilters, nil
}

// lookupStreamingLocators Get StreamingLocators from Azure MediaServices. Remove pagination
func (a *AzureServiceProvider) lookupStreamingLocators(ctx context.Context, before string, after string) ([]*armmediaservices.StreamingLocator, error) {
	client := a.streamingLocatorsClient
	sl := []*armmediaservices.StreamingLocator{}

	// Generate the filter
	filter := generateFilter(before, after)

	// If we have a filter apply it
	options := &armmediaservices.StreamingLocatorsClientListOptions{Orderby: to.Ptr("properties/created")}
	if filter != "" {
		options.Filter = to.Ptr(filter)
	}

	pager := client.NewListPager(a.resourceGroup, a.accountName, options)

	// Paginated result. We just need a list. loop through and generate that list
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return sl, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			contentKeys, err := client.ListContentKeys(ctx, a.resourceGroup, a.accountName, *v.Name, nil)
			if err != nil {
				log.Error("Failed to get content keys for streaming locator: ", *v.Name)
				continue
			}
			v.Properties.ContentKeys = contentKeys.ContentKeys
			log.Debugf("Id: %s, Name: %s, Type: %s, AssetName: %s, StreamingLocatorID: %s, StreamingPolicyName: %s\n", *v.ID, *v.Name, *v.Type, *v.Properties.AssetName, *v.Properties.StreamingLocatorID, *v.Properties.StreamingPolicyName)
			sl = append(sl, v)
		}
	}
	return sl, nil
}

// lookupStreamingPolicies Get StreamingPolicies from Azure MediaServices. Remove pagination
func (a *AzureServiceProvider) lookupStreamingPolicies(ctx context.Context, before string, after string) ([]*armmediaservices.StreamingPolicy, error) {
	client := a.streamingPoliciesClient
	sp := []*armmediaservices.StreamingPolicy{}

	// Generate the filter
	filter := generateFilter(before, after)

	// If we have a filter apply it
	options := &armmediaservices.StreamingPoliciesClientListOptions{Orderby: to.Ptr("properties/created")}
	if filter != "" {
		options.Filter = to.Ptr(filter)
	}

	pager := client.NewListPager(a.resourceGroup, a.accountName, options)

	// Paginated result. We just need a list. loop through and generate that list
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return sp, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			// Skip the predefined. They should also be in MKIO
			if !strings.HasPrefix(*v.Name, "Predefined_") {
				log.Debugf("Id: %s, Name: %s, Type: %s\n", *v.ID, *v.Name, *v.Type)
				sp = append(sp, v)
			}
		}
	}
	return sp, nil
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
			return se, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			log.Debugf("Id: %s, Name: %s, Type: %s, Location: %s\n", *v.ID, *v.Name, *v.Type, *v.Location)
			// Clean up the exported resource to import properly
			v.Properties.Created = nil
			v.Properties.LastModified = nil
			if !*v.Properties.CdnEnabled {
				v.Properties.CdnProvider = nil
				v.Properties.CdnProfile = nil
			}
			se = append(se, v)
		}
	}
	return se, nil
}

// lookupContentKeyPolicies Get contentKeyPolicy from Azure MediaServices. Remove pagination
func (a *AzureServiceProvider) lookupContentKeyPolicies(ctx context.Context, before string, after string) ([]*armmediaservices.ContentKeyPolicy, error) {
	client := a.contentKeyPoliciesClient

	// Generate the filter
	filter := generateFilter(before, after)

	// If we have a filter apply it
	options := &armmediaservices.ContentKeyPoliciesClientListOptions{Orderby: to.Ptr("properties/created")}
	if filter != "" {
		options.Filter = to.Ptr(filter)
	}

	pager := client.NewListPager(a.resourceGroup, a.accountName, options)

	ckp := []*armmediaservices.ContentKeyPolicy{}

	ckpFailures := []string{}
	// We get pages back. Loop through pages and create a list of assets
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return ckp, fmt.Errorf("failed to advance page: %v", err)
		}
		for _, v := range nextResult.Value {
			// log.Debugf("Id: %s, Name: %s, Type: %s, Container: %s, StorageAccountName: %s, AssetId: %s\n", *v.ID, *v.Name, *v.Type, *v.Properties.Container, *v.Properties.StorageAccountName, *v.Properties.AssetID)
			props, err := client.GetPolicyPropertiesWithSecrets(ctx, a.resourceGroup, a.accountName, *v.Name, nil)
			if err != nil {
				ckpFailures = append(ckpFailures, *v.Name)
				log.Errorf("unable to get content key policy %v: %v", *v.Name, err)
				continue
			}
			v.Properties = &props.ContentKeyPolicyProperties
			ckp = append(ckp, v)
		}
	}

	if len(ckpFailures) > 0 {
		// Return error after we've looped through all the content key policies
		return ckp, fmt.Errorf("unable to get content key policy %v", ckpFailures)
	}
	return ckp, nil
}
