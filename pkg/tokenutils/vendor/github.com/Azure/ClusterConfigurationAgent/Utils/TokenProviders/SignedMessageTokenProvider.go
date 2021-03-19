package TokenProviders

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
)

type AuthorizationSchemes int

const (
	AuthHeaderAsSignedMessage = iota
	AuthHeaderAsSharedKey
)

type SignedMessageTokenProvider struct {
	TokenProvider
	privateKey                                           *rsa.PrivateKey
	AuthorizationScheme                                  AuthorizationSchemes
	SubscriptionID, GroupName, ClusterType, ResourceName string
}

func (provider *SignedMessageTokenProvider) Init(debugLogging bool, logger LogHelper.LogWriter) {
	provider.Logger = logger
	provider.AuthorizationScheme = AuthHeaderAsSharedKey
	provider.DebugLogging = debugLogging
}
func (provider *SignedMessageTokenProvider) PopulateMetadata(SubscriptionID, GroupName, ClusterType, ResourceName string, privateKey *rsa.PrivateKey) {
	provider.privateKey = privateKey
	provider.SubscriptionID, provider.GroupName, provider.ClusterType, provider.ResourceName = SubscriptionID, GroupName, ClusterType, ResourceName
}

func (provider *SignedMessageTokenProvider) GetToken() (*AccessToken, error) {
	headerName := "signedMessage"

	messageTobeSigned, signedMessage, err := provider.getSignedMessage()
	if err != nil {
		provider.Logger.Error("Error: in the Signing the Message message {%s} : error {%v}", messageTobeSigned, err)
		return nil, err
	}

	authorizationScheme := ""
	if provider.AuthorizationScheme == AuthHeaderAsSharedKey {
		authorizationScheme = "SharedKey "
		headerName = "Authorization"
	}

	accessToken := &AccessToken{
		Token:               fmt.Sprintf("%s:%s", messageTobeSigned, signedMessage),
		HeaderName:          headerName,
		AuthorizationScheme: authorizationScheme,
	}

	if provider.DebugLogging {
		provider.Logger.LogSecureString("the access token generated is %v", accessToken)
	}

	return accessToken, err
}

func (provider *SignedMessageTokenProvider) getSignedMessage() (unEncodedMessage string, signedMessage string, err error) {
	unEncodedMessage = fmt.Sprintf("%s/%s/%s/%s", provider.SubscriptionID,
		provider.GroupName, provider.ClusterType, provider.ResourceName)

	privateKey := provider.privateKey

	// sha256 the message
	hashed := sha256.Sum256([]byte(unEncodedMessage))
	var signature []byte

	// For HIS SharedKey we need the PSS signing
	if provider.AuthorizationScheme == AuthHeaderAsSignedMessage {
		opts := rsa.PSSOptions{SaltLength: rsa.PSSSaltLengthEqualsHash}
		signature, err = rsa.SignPSS(rand.Reader, privateKey, crypto.SHA256, hashed[:], &opts)
	} else {
		signature, err = rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	}
	if err != nil {
		return unEncodedMessage, "", err
	}

	encodedSignature := base64.StdEncoding.EncodeToString([]byte(signature))

	if provider.AuthorizationScheme == AuthHeaderAsSignedMessage {
		encodedSignature = fmt.Sprintf(Constants.RsaSignedHash, encodedSignature)
	}


	if provider.DebugLogging {
		provider.Logger.LogSecureString("signature : %s", encodedSignature)
	}

	// In case of PssSigning Need to get the
	return unEncodedMessage, encodedSignature, nil
}
