package TokenProviders

type OverrideTokenProviderValues int

const (
	SignedMessage = iota
	ManagedIdentity
	BearerTokenClientId
	BearerTokenCertificate
	NoOverride
)