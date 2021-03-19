package TokenProviders

import (
	"context"
	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Helper"
	"github.com/Azure/go-autorest/autorest/adal"
)

type BearerTokenProviderClientId struct {
	TokenProvider
	aksTenantId                     string
	aksClientId                     string
	aksClientSecret                 string
	audience                        string
	aksUserMsi                      bool
	aksUserAssignedIdentityClientId string
}

// Initialize
func (provider *BearerTokenProviderClientId) Init(debugLogging bool, logger LogHelper.LogWriter) {
	provider.DebugLogging = debugLogging
	provider.Logger = logger

	acsFilepath := Helper.GetEnvironmentVar(Constants.LabelAksCredentialLocation)
	credential := &Helper.AksCredential{}
	err := credential.InitializeAcsCredentials(acsFilepath)
	if err != nil {
		logger.Error("Error in creating ACS Credential: error {%v}", err)
		return
	}

	provider.aksClientId = credential.ClientID
	provider.aksClientSecret = credential.ClientSecret
	provider.aksTenantId = credential.TenantID
	provider.audience = Constants.ClusterConfigAksFirstPartyAppId
	provider.aksUserMsi = credential.UseManagedIdentityExtension
	provider.aksUserAssignedIdentityClientId = Helper.GetEnvironmentVar(Constants.AksUserAssignedIdentityClientId) // can be removed from dp
}

func (provider *BearerTokenProviderClientId) GetToken() (*AccessToken, error) {
	var spToken adal.OAuthTokenProvider
	var err error

	if provider.aksUserMsi {
		spToken, err = provider.getMsiToken()
		if spToken == nil || err != nil {
			provider.Logger.Error("Error in getting the token from adal for MSI cluster: error {%v}", err)
			return nil, err
		}
	} else {
		spToken, err = provider.getServicePrincipalToken()
		if spToken == nil || err != nil {
			provider.Logger.Error("Error in getting the token from adal for SPN cluster: error {%v}", err)
			return nil, err
		}
	}
	if refresher, ok := spToken.(adal.RefresherWithContext); ok {
		err = refresher.EnsureFreshWithContext(context.Background())
	} else if refresher, ok := spToken.(adal.Refresher); ok {
		err = refresher.EnsureFresh()
	}
	if err != nil {
		provider.Logger.Error("Error in getting the token from adal: error {%v}", err)
		return nil, err
	}

	token := spToken.OAuthToken()
	if token == "" {
		provider.Logger.Error("Token is nil: {%v}", spToken)
		provider.Logger.Error("aksUserAssignedIdentityClientId is: {%s}", provider.aksUserAssignedIdentityClientId)
		provider.Logger.Error("audience is: {%s}", provider.audience)
	}
	accessToken := &AccessToken{
		Token:               token,
		AuthorizationScheme: "Bearer ",
		HeaderName:          "Authorization",
	}

	return accessToken, err
}

func (provider *BearerTokenProviderClientId) getMsiToken() (token adal.OAuthTokenProvider, err error) {
	// Try to retrieve the token with MSI
	msiEndpoint, err := adal.GetMSIVMEndpoint()
	if err != nil {
		provider.Logger.Error("Failed to get the Managed Service Identity endpoint: %v", err)
		return nil, err
	}

	// User assigned
	adalToken, err := adal.NewServicePrincipalTokenFromMSIWithUserAssignedID(msiEndpoint, provider.audience, provider.aksUserAssignedIdentityClientId)
	if err != nil {
		provider.Logger.Error("Failed to create the Managed Service Identity token: %v", err)
		return nil, err
	}
	return adalToken, nil
}

func (provider *BearerTokenProviderClientId) getServicePrincipalToken() (token adal.OAuthTokenProvider, err error) {
	ActiveDirectoryEndpoint := Helper.ActiveDirectoryEndpoint()
	oauthConfig, err := adal.NewOAuthConfig(ActiveDirectoryEndpoint, provider.aksTenantId)
	if err != nil || oauthConfig == nil {
		provider.Logger.Error("Error in creating OAuth config: error {%v}", err)
		return nil, err
	}
	servicePrincipalToken, err := adal.NewServicePrincipalToken(*oauthConfig, provider.aksClientId, provider.aksClientSecret, provider.audience)

	return servicePrincipalToken, nil
}
