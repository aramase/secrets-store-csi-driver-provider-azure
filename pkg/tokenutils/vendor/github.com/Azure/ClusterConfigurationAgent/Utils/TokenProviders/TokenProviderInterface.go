package TokenProviders

import (
	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	"time"
)

type AccessToken struct {
	Token      string
	Expiration time.Time
	HeaderName  string
	AuthorizationScheme string
}

type TokenProviders interface {
	GetToken() (*AccessToken, error)
}

type TokenProvider struct {
	DebugLogging bool
	Logger       LogHelper.LogWriter
}