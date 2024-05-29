package mkiosdk

import (
	"encoding/json"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/mediaservices/armmediaservices"
)

// armmediaservices structs implements the JSON Marshaler interface, and as
// such needs to be overridden to include the FairPlayAmsCompatibility field.

type FPContentKeyPolicy struct {
	armmediaservices.ContentKeyPolicy
	FPProperties *FPContentKeyPolicyProperties
}

func (c *FPContentKeyPolicy) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populate(objectMap, "id", c.ID)
	populate(objectMap, "name", c.Name)
	populate(objectMap, "properties", c.FPProperties)
	populate(objectMap, "systemData", c.SystemData)
	populate(objectMap, "type", c.Type)
	return json.Marshal(objectMap)
}

type FPContentKeyPolicyProperties struct {
	armmediaservices.ContentKeyPolicyProperties
	FairPlayAmsCompatibility *bool
}

func (c *FPContentKeyPolicyProperties) MarshalJSON() ([]byte, error) {
	objectMap := make(map[string]interface{})
	populateTimeRFC3339(objectMap, "created", c.Created)
	populate(objectMap, "description", c.Description)
	populateTimeRFC3339(objectMap, "lastModified", c.LastModified)
	populate(objectMap, "options", c.Options)
	populate(objectMap, "policyId", c.PolicyID)
	populate(objectMap, "fairPlayAmsCompatibility", c.FairPlayAmsCompatibility)
	return json.Marshal(objectMap)
}
