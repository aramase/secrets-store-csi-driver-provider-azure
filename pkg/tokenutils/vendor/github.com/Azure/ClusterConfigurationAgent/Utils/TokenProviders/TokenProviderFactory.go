package TokenProviders

import (
	"crypto/rsa"
	"log"
	"strings"
	"time"

	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Helper"
	"github.com/Azure/ClusterConfigurationAgent/Utils/KuberenetesAPIServer"
)

func TokenProviderFactoryWithOverride(overrideValue OverrideTokenProviderValues,
	logger LogHelper.LogWriter,
	apiClient *KuberenetesAPIServer.KubeAPIServerQueries) TokenProviders {

	return TokenProviderFactory(overrideValue, logger, apiClient, "") // Don't need clusterType when we have an override value
}

func TokenProviderFactory(overrideValue OverrideTokenProviderValues,
	logger LogHelper.LogWriter,
	apiClient *KuberenetesAPIServer.KubeAPIServerQueries,
	clusterType string) TokenProviders {

	return TokenProviderFactoryWithAudience(overrideValue, logger, apiClient, clusterType, Constants.ClusterConfigConnectedClusterFirstPartyAppId)
}

func ExtensionTokenProviderWithClient(audience, extensionName string, logger LogHelper.LogWriter, apiClient *KuberenetesAPIServer.KubeAPIServerQueries) TokenProviders {
	baseProvider := TokenProvider{false, logger}
	tokenProvider := &ManagedIdentityTokenProvider{TokenProvider: baseProvider}
	tokenProvider.Init(apiClient, false, logger, audience, extensionName)
	return tokenProvider
}

func TokenProviderFactoryWithAudience(overrideValue OverrideTokenProviderValues,
	logger LogHelper.LogWriter,
	apiClient *KuberenetesAPIServer.KubeAPIServerQueries,
	clusterType string,
	audience string) TokenProviders {

	DebugLogging := Helper.IsDebugLogging()
	baseProvider := TokenProvider{DebugLogging, logger}

	switch overrideValue {
	case NoOverride:
		// Create the appropriate provider and initialize them
		switch strings.ToLower(clusterType) {
		case Constants.ManagedClusters:
			tokenProvider := &BearerTokenProviderClientId{TokenProvider: baseProvider}
			tokenProvider.Init(DebugLogging, logger)
			return tokenProvider
		case Constants.ConnectedClusters, Constants.Appliances:
			if Helper.IsNullAuthEnabled() {
				tokenProvider := &NullMessageTokenProvider{TokenProvider: baseProvider}
				tokenProvider.Init(DebugLogging, logger)
				return tokenProvider
			}
			if Helper.IsManagedIdentityAuthEnabled() {
				tokenProvider := &ManagedIdentityTokenProvider{TokenProvider: baseProvider}
				tokenProvider.Init(apiClient, DebugLogging, logger, audience, "")
				return tokenProvider
			} else {
				tokenProvider := CreateSignedMessageProvider(apiClient, logger, baseProvider, DebugLogging)
				return tokenProvider
			}
		}
	case ManagedIdentity:
		tokenProvider := &ManagedIdentityTokenProvider{TokenProvider: baseProvider}
		tokenProvider.Init(apiClient, DebugLogging, logger, audience, "")
		return tokenProvider
	case SignedMessage:
		// Get the resource onboarding Secret
		tokenProvider := CreateSignedMessageProvider(apiClient, logger, baseProvider, DebugLogging)
		return tokenProvider
	case BearerTokenCertificate:
		tokenProvider := &BearerTokenProviderCertificate{TokenProvider: baseProvider}
		tokenProvider.Init(DebugLogging, logger)
		return tokenProvider
	case BearerTokenClientId:
		tokenProvider := &BearerTokenProviderClientId{TokenProvider: baseProvider}
		tokenProvider.Init(DebugLogging, logger)
		return tokenProvider
	}

	return nil
}

func CreateSignedMessageProvider(apiClient *KuberenetesAPIServer.KubeAPIServerQueries, logger LogHelper.LogWriter, baseProvider TokenProvider, DebugLogging bool) *SignedMessageTokenProvider {
	secret, err := Helper.WithRetry(apiClient.GetPrivateKeyFromSecret, false, 2*time.Hour)
	if err != nil || secret == nil {
		logger.Error("Unable to retrieve the onboarding secret make sure that the connect agent is running")
		log.Fatal("unable to get the resource onboarding key")
	}
	tokenProvider := &SignedMessageTokenProvider{TokenProvider: baseProvider}
	tokenProvider.Init(DebugLogging, logger)
	tokenProvider.PopulateMetadata(
		Helper.GetEnvironmentVar(Constants.LabelAzureSubscriptionId),
		Helper.GetEnvironmentVar(Constants.LabelAzureResourceGroup),
		Helper.GetEnvironmentVar(Constants.LabelClusterType),
		Helper.GetEnvironmentVar(Constants.LabelAzureResourceName),
		secret.(*rsa.PrivateKey))
	return tokenProvider
}
