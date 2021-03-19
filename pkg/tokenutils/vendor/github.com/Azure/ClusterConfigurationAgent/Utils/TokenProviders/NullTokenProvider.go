package TokenProviders

import (
	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
)

type NullMessageTokenProvider struct {
	TokenProvider
	SubscriptionID, GroupName, ClusterType, ResourceName string
}

func (provider *NullMessageTokenProvider) Init(debugLogging bool, logger LogHelper.LogWriter) {
	provider.Logger = logger
	provider.DebugLogging = debugLogging
}

func (provider *NullMessageTokenProvider) GetToken() (*AccessToken, error) {
	accessToken := &AccessToken{
		Token:               "",
		HeaderName:          "Authorization",
		AuthorizationScheme: "",
	}
	return accessToken, nil
}
