package Helper

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

//AksCredential represents the credential file for ACS.
type AksCredential struct {
	Cloud                       string `json:"cloud"`
	TenantID                    string `json:"tenantId"`
	SubscriptionID              string `json:"subscriptionId"`
	ClientID                    string `json:"aadClientId"`
	ClientSecret                string `json:"aadClientSecret"`
	ResourceGroup               string `json:"resourceGroup"`
	Region                      string `json:"location"`
	UseManagedIdentityExtension bool   `json:"useManagedIdentityExtension"`
}

// InitializeAcsCredentials returns an AksCredential struct from file path
func (credentials *AksCredential) InitializeAcsCredentials(filePath string) error {
	if filePath == "" {
		log.Printf("Reading ACS credential file failed: Empty filePath")
		return nil
	}

	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		log.Printf("Reading ACS credential file %q failed: %v", filePath, err)
		return err
	}

	// Unmarshal the authentication file.
	if err := json.Unmarshal(data, credentials); err != nil {
		return err
	}

	return nil
}
