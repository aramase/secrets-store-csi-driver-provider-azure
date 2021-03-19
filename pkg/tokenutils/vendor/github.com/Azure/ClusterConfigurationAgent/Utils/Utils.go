package Utils

import (
	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	"github.com/Azure/ClusterConfigurationAgent/Utils/KuberenetesAPIServer"
	"github.com/Azure/ClusterConfigurationAgent/Utils/TokenProviders"
)

func ExtensionTokenProvider(audience, extensionName, tokenNamespace string, logger LogHelper.LogWriter) TokenProviders.TokenProviders {
	apiClient := &KuberenetesAPIServer.KubeAPIServerQueries{}
	err := apiClient.InitializeClient(tokenNamespace, logger, false)
	if err != nil {
		return nil
	}
	return TokenProviders.ExtensionTokenProviderWithClient(audience, extensionName, logger, apiClient)
}
