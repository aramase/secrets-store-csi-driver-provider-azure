package TokenProviders

import (
	"time"

	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Helper"
	"github.com/Azure/ClusterConfigurationAgent/Utils/KuberenetesAPIServer"
)

type ManagedIdentityTokenProvider struct {
	TokenProvider
	audience  string
	resourceId string
	apiClient *KuberenetesAPIServer.KubeAPIServerQueries
	crdClient *KuberenetesAPIServer.ClusterIdentityCRDClient
}

func (provider *ManagedIdentityTokenProvider) Init(apiClient *KuberenetesAPIServer.KubeAPIServerQueries, debugLogging bool, logger LogHelper.LogWriter, audience string, resourceId string) {
	provider.apiClient = apiClient
	provider.DebugLogging = debugLogging
	provider.Logger = logger
	provider.resourceId = resourceId
	provider.audience = audience

	if provider.apiClient == nil {
		provider.Logger.Error("Error: Invalid (null) apiClient.")
		return
	}
}

func (provider *ManagedIdentityTokenProvider) GetToken() (*AccessToken, error) {
	// First check that the apiClient is valid
	if provider.apiClient == nil {
		provider.Logger.Error("Error: Initialize the token provider first, by calling Init.")
		return nil, nil
	}

	var err error
	var token interface{}
	err = nil

	// Create the CrdClient object
	if provider.crdClient == nil {
		provider.crdClient = &KuberenetesAPIServer.ClusterIdentityCRDClient{}
		crdName := Helper.GetIdentityRequestCRDName(provider.audience, provider.resourceId)

		provider.crdClient.Init(crdName, Constants.ManagementNamespace, provider.apiClient, provider.Logger, provider.audience, provider.resourceId)

		// Take care of the case where the pod restrats
		token, err = provider.crdClient.GetTokenFromStatus()
	}

	// If the token has less than 1 hour for expiry, renew the token
	if err != nil || provider.crdClient.ExpirationTime.Sub(time.Now()) < Constants.ExpirationInMinutes*time.Minute {
		// Use the ClusterIdentityCRDInteraction object to renew token (or create if it doesn't exist)
		err = provider.crdClient.CreateOrRenewClusterIdentity()

		// Get the token - with retry and a max time
		token, err = Helper.WithRetry(provider.crdClient.GetTokenFromStatus, false, Constants.GetTokenWaitTimeInMinutes*time.Minute)
	}

	// Use the ClusterIdentityCRDInteraction to get the Token Reference

	if err != nil {
		provider.Logger.Error("Error in getting ClusterIdentityToken reference: error {%v}", err)
		return nil, err
	} else if provider.crdClient.Token == "" {
		provider.crdClient.Token = token.(string)
	}

	accessToken := &AccessToken{
		Token:               provider.crdClient.Token,
		HeaderName:          "Authorization",
		AuthorizationScheme: "Bearer ",
	}

	// If the expiration time is still valid, return the token
	return accessToken, nil
}
