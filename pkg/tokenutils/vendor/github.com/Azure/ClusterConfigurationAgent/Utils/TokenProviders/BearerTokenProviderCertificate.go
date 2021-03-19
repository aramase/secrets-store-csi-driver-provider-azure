package TokenProviders

import (
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Helper"
	"github.com/Azure/go-autorest/autorest/adal"
	"log"
)

type BearerTokenProviderCertificate struct {
	TokenProvider
	tenantId string
	clientId string
	resourceId string // audience
	certificate *x509.Certificate
	rsaPrivateKey *rsa.PrivateKey
}

func (provider *BearerTokenProviderCertificate) Init(debugLogging bool, logger LogHelper.LogWriter) {
	provider.DebugLogging = debugLogging
	provider.Logger = logger

	provider.tenantId = Helper.GetEnvironmentVar(Constants.LabelTenantId)
	if provider.tenantId == "" {
		logger.Error("TenantId is empty")
		log.Fatal("TenantId is empty")
	}
}

func (provider *BearerTokenProviderCertificate) PopulateMetadata(clientId string, resourceId string, certificate *x509.Certificate, key *rsa.PrivateKey) {
	provider.clientId, provider.resourceId, provider.certificate, provider.rsaPrivateKey = clientId, resourceId, certificate, key
}

func (provider *BearerTokenProviderCertificate) GetToken() (*AccessToken, error) {
	if provider.tenantId == "" ||
		provider.resourceId == "" ||
		provider.clientId == "" ||
		provider.certificate == nil ||
		provider.rsaPrivateKey == nil {
		err := "error: Initialize token provider first, by calling Init"
		provider.Logger.Error(err)
		return nil, fmt.Errorf(err)
	}

	// Acquiring token using go-rest
	ActiveDirectoryEndpoint := Helper.ActiveDirectoryEndpoint()
	oauthConfig, err := adal.NewOAuthConfig(ActiveDirectoryEndpoint, provider.tenantId)

	spt, err := adal.NewServicePrincipalTokenFromCertificate(*oauthConfig, provider.clientId, provider.certificate, provider.rsaPrivateKey, provider.resourceId)
	if err != nil {
		provider.Logger.Error("Creating new service principal token from certificate call to adal failed %s", err)
		return nil, err
	}

	// call Refresh() per https://github.com/Azure/go-autorest/tree/master/autorest/adal
	freshErr := spt.Refresh()
	if freshErr != nil {
		return nil, freshErr
	}

	t := spt.Token()

	if len(t.AccessToken) == 0 {
		provider.Logger.Error("Failed to retrieve the token from adal")
		err = fmt.Errorf("failed to retrieve token")
		return nil, err
	}

	accessToken:= &AccessToken{
		Token:      t.AccessToken,
		Expiration: t.Expires(),
		HeaderName:  "Authorization",
		AuthorizationScheme: "Bearer ",
	}

	return accessToken, nil
}