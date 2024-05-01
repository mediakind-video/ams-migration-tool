package migrate

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
)

// MigrationFileContents contains the contents for the StreamingEndpoint File
type MigrationFileContents struct {
	AssetFilters       map[string][]*armmediaservices.AssetFilter
	Assets             []*armmediaservices.Asset
	ContentKeyPolicies []*armmediaservices.ContentKeyPolicy

	StreamingEndpoints []*armmediaservices.StreamingEndpoint
	StreamingLocators  []*armmediaservices.StreamingLocator
	StreamingPolicies  []*armmediaservices.StreamingPolicy
}

func (contents MigrationFileContents) WriteMigrationFile(ctx context.Context, fileName string) error {
	migrationBytes, err := json.Marshal(contents)
	if err != nil {
		return fmt.Errorf("unable to marshal assets: %v", err)
	}

	// Create the file
	f, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("unable to create file %v: %v", fileName, err)
	}

	// Close the file when we're done w/ it
	defer f.Close()

	_, err = f.Write(migrationBytes)
	if err != nil {
		return fmt.Errorf("unable to write to file %v: %v", fileName, err)
	}
	f.Sync()

	return nil
}

func (contents *MigrationFileContents) ReadMigrationFile(ctx context.Context, fileName string) error {
	// Read in our migration file
	file, err := os.Open(fileName)
	if err != nil {
		return fmt.Errorf("unable to open migration file: %v: %v", fileName, err)
	}
	defer file.Close()

	// Validate a file exists
	stat, err := file.Stat()
	if err != nil {
		return fmt.Errorf("unable to get migration file stats: %v", err)
	}

	// Read the file into a byte slice
	bs := make([]byte, stat.Size())
	_, err = bufio.NewReader(file).Read(bs)
	if err != nil && err != io.EOF {
		return fmt.Errorf("unable to read migration file: %v", err)
	}

	err = json.Unmarshal(bs, contents)
	if err != nil {
		return fmt.Errorf("unable to unmarshal migration file contents: %v", err)
	}

	return nil
}
