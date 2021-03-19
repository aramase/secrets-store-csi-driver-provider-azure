package Constants

import (
	"io/ioutil"
	"strings"
	"time"
)

var (
	ManagementNamespace  string
	FluxctlWithBase64Git = [2]string{"0.1.9", "0.1.13"}
)

// This is run on initialization of the module
func init() {
	// Initialize management namespace to the current pod's namespace
	ManagementNamespace = WellKnownAzureArcManagementNamespace
	namespace, err := ioutil.ReadFile(NamespaceMountPath)
	if err == nil {
		namespaceAsStr := strings.TrimSpace(string(namespace))
		if namespaceAsStr != "" {
			ManagementNamespace = namespaceAsStr
		}
	}
}

const (
	ListConfigurationsEndpoint  = "/subscriptions/%s/resourceGroups/%s/provider/%s/clusters/%s/configurations/getPendingConfigs?api-version=2019-11-01-Preview"
	RegionBasedEndpointTemplate = "https://%s.dp.kubernetesconfiguration.azure.%s"
	PutConfigurationEndpoint    = "/subscriptions/%s/resourceGroups/%s/provider/%s/clusters/%s/configurations/%s/postStatus?api-version=2019-11-01-Preview"

	//TODO: Understand how sovereign clouds will work
	HISFallbackGlobalEndpoint = "https://gbl.his.hybridcompute.azure-automation.net/discovery?location=%s&api-version=1.0-preview"
	HISOverrideVariable       = "HIS_ENDPOINT"
	HISGlobalEndpointTemplate = "https://gbl.his.arc.azure.%s/discovery?location=%s&api-version=1.0-preview"

	// Environment variables' names
	LabelAksCredentialLocation     = "AKS_CREDENTIAL_LOCATION"
	LabelArcAgentHelmChartName     = "ARC_AGENT_HELM_CHART_NAME"
	LabelArcAgentHelmVersion       = "HELM_CHART_CURRENT_VERSION"
	LabelArcAgentReleaseTrain      = "ARC_AGENT_RELEASE_TRAIN"
	LabelTenantId                  = "AZURE_TENANT_ID"
	LabelAzureSubscriptionId       = "AZURE_SUBSCRIPTION_ID"
	LabelAzureResourceGroup        = "AZURE_RESOURCE_GROUP"
	LabelAzureResourceName         = "AZURE_RESOURCE_NAME"
	LabelAzureRegion               = "AZURE_REGION"
	LabelAzureArcAgentVersion      = "AZURE_ARC_AGENT_VERSION"
	LabelClusterType               = "CLUSTER_TYPE"
	LabelAzureEnvironment          = "AZURE_ENVIRONMENT"
	LabelDebugLogging              = "DEBUG_LOGGING"
	LabelFluxClientDefaultLocation = "FLUX_CLIENT_DEFAULT_LOCATION"
	LabelClusterConfigApiResource  = "CLUSTERCONFIG_API_RESOURCE"
	LabelWebSocketPattern          = "FLUX_WEB_SOCKET_PATTERN"
	LabelFluxEventPattern          = "FLUX_EVENT_HTTP_PATTERN"
	LabelFluxUpstreamEnabled       = "FLUX_UPSTREAM_SERVICE_ENABLED"
	ExtensionOperatorEnabled       = "EXTENSION_OPERATOR_ENABLED"
	GitOpsEnabled                  = "GITOPS_ENABLED"
	ClusterConnectAgentEnabled     = "CLUSTER_CONNECT_AGENT_ENABLED"

	HISApiVersion                     = "1.1-preview"
	RsaSignedHash                     = "rsapss %s"
	ApiVersionQueryParam              = "api-version"
	IdQueryParam                      = "id"
	HisApplicationId                  = "eec53b1f-b9a4-4479-acf5-6b247c6a49f2"
	HISCertificateClientIdKeyName     = "azure-identity-client-id"
	ParentResourceIdQueryParam        = "parentResourceId"
	ConfigurationIdentityResourceType = "ConnectedClusterConfiguration"
	HISRegionalEndpoint               = "/type/%s/identity"
	ArmID                             = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.Kubernetes/%s/%s"
	// Creating dedicated ARMID for appliance to not disrupt existing code - TODO is to merge it
	ArmIDForAppliance                = "/subscriptions/%s/resourceGroups/%s/providers/Microsoft.ResourceConnector/%s/%s"
	ExtensionArmID                   = "%s/providers/Microsoft.KubernetesConfiguration/extensions/%s"
	PostLogsEndpoint                 = "/subscriptions/%s/resourceGroups/%s/provider/%s/clusters/%s/agentDiagnostics/postLogs?api-version=2019-11-01-Preview"
	GetUpdateConfigEndpoint          = "/subscriptions/%s/resourceGroups/%s/provider/%s/clusters/%s/getUpdateConfig?api-version=2019-11-01-Preview"
	GetExtensionUpdateConfigEndpoint = "/subscriptions/%s/resourceGroups/%s/provider/%s/clusters/%s/getExtensionUpdateConfigs?api-version=2019-11-01-Preview"
	SyncResourcesEndpoint            = "/subscriptions/%s/resourceGroups/%s/provider/%s/clusters/%s/syncResources?api-version=2019-11-01-Preview"

	// These are same of the connect agent , need to modify them if there is a change
	WellKnownAzureArcManagementNamespace = "azure-arc"
	NamespaceMountPath                   = "/run/secrets/kubernetes.io/serviceaccount/namespace"
	WellKnownKubernetesSecret            = "azure-arc-connect-privatekey"
	HISCertificateKeyName                = "azure-identity-certificate"
	HISCertificateExpirationTimeKeyName  = "azure-identity-certificate-expiration-time"
	HISCertificateRenewAfterKeyName      = "cert-renew-after"
	HISExtensionCertificateKeyName       = "extension-identity-certificate-%s"
	ClusterIdentityToken                 = "cluster-identity-token"
	IdentityRequestCRDName               = "identity-request-%x" // %x representing the sha256 hash of audience
	ExpirationInMinutes                  = 60
	GetTokenWaitTimeInMinutes            = 5

	// Proxy configuration values
	WellKnownAzureProxyConfigMap    = "azure-proxy-config"
	WellKnownAzureProxyConfigSecret = "proxy-config"
	WellKnownAzureProxyCertSecret   = "proxy-cert"
	ProxyCertSecretVolumeName       = "proxy-certstore"
	ProxyCertFileName               = "proxy-cert.crt"
	SSLCertVolumeName               = "ssl-certs"
	ProxyCertMountPath              = "/usr/local/share/ca-certificates/proxy-cert.crt"
	SSLCertMountPath                = "/etc/ssl/certs/"

	DefaultFluxCtlLocation = "https://github.com/fluxcd/flux/releases/download/1.14.2/fluxctl_linux_amd64"
	TimeFormat             = "2006-01-02T15:04:05.000Z"
	DefaultMessageLevel    = 3

	// This corresponds roughly 1 hour of retry before giving up
	MaxRetriesForNonCompliantState          = 60
	FluxHelmOperatorChartRepo               = "mcr.microsoft.com/oss/fluxcd/helm-operator-chart"
	FluxHelmOperatorImageRepo               = "mcr.microsoft.com/oss/fluxcd/helm-operator"
	FluxHelmMinSupportedOperatorVersion     = "0.3.0"
	FluxHelmDefaultSupportedOperatorVersion = "1.2.0"
	FluxHelmOperatorChartName               = "helm-operator"

	FluxInitContainerImageRepo = "mcr.microsoft.com/azurearck8s/arc-preview/flux-init-container:0.0.1"

	ConnectedClusters                            = "connectedclusters"
	ManagedClusters                              = "managedclusters"
	Appliances                                   = "appliances"
	ApplianceRPNamespace                         = "Microsoft.ResourceConnector"
	ApplianceConnectAgentToDataPlaneAppId        = "d22ea4d1-2678-4a7b-aa5e-f340c2a7d993"
	ClusterConfigAksFirstPartyAppId              = "03db181c-e9d3-4868-9097-f0b728327182"
	ClusterConfigConnectedClusterFirstPartyAppId = "c699bf69-fb1d-4eaf-999b-99e6b2ae4d85"
	AksUserAssignedIdentityClientId              = "AKS_USER_ASSIGNED_IDENTITY_CLIENT_ID"

	DefaultArcAgentHelmChartName  = "arc-k8s-agents"
	DefaultArcAgentReleaseTrain   = "Stable"
	VaultSuffix                   = "-token"
	MANAGED_IDENTITY_AUTH         = "MANAGED_IDENTITY_AUTH"
	NO_AUTH                       = "NO_AUTH_HEADER_DATA_PLANE"
	DefaultarcAgentReleaseTrain   = "stable"
	AnnoationOCIArtifactImageName = "org.opencontainers.image.title"

	AzureClusterIdentityRequestsLeaderElectionId = "azidentityreqleaderid"
	GitConfigLeaderElectionId                    = "gitconfigleaderid"
	ExtensionControllerLeaderElectionId          = "extensioncontrollerleaderid"
	GitConfigFinalizerString                     = "GitConfig.Controller.Finalizer"
	ExtensionConfigFinalizerString               = "Extension.Controller.Finalizer"

	//if Dp operation failed for 15 minutes then agent readiness will marked as failure
	MaxWaitBeforeMarkReadinessFailureInMinute = 15

	MaxContentLengthForDataPlaneCallInBytes = 1024 * 1024
	AnnotationProviderNamespace             = "resourceSync.arc.azure.com/ProviderNamespace"
	AnnotationResourceType                  = "resourceSync.arc.azure.com/ResourceType"
	AnnotationResourceId                    = "resourceSync.arc.azure.com/AzureResourceId"
	AnnotationLastSynced                    = "resourceSync.arc.azure.com/LastSynced"

	AnnotationLastSyncedHelmPropertiesHash = "gitconfig.arc.azure.com/HelmOperatorSpecHash"

	//Error constants
	ExitCodeWhenNoKeyIsGenerated = 11

	GitConfigKind                   = "Git"
	ExtensionConfigKind             = "Extension"
	ActiveDirectoryEndpointTemplate = "https://login.microsoftonline.%s/"

	TokenSecretDataKey  = "token"
	CaCertSecretDataKey = "ca.crt"

	ProtectedParametersSecretPrefix = "protected-parameters"

	SSHPrivateKeySecretPrefix  = "ssh-private-key"
	SSHPrivateKeyParameterName = "sshPrivateKey"

	GitAuthSecretPrefix      = "git-auth"
	GitAuthHelmSecretPrefix  = "git-auth-helm"
	GitAuthUserParameterName = "httpsUser"
	GitAuthKeyParameterName  = "httpsKey"

	ProxyConfigSecretPrefix = "proxy-config"
	ProxyCertSecretPrefix   = "proxy-cert"

	HelmCacheStorePaths = "/opt/helm/cache"

	SecretKindName = "secrets"
	GetVerbs       = "get"
	WatchVerbs     = "watch"
	ListVerbs      = "list"
	RoleKind       = "Role"

	FluxCTLCurrentVersion = "0.1.16"
	ArcZoneAnnotationKey  = "CustomLocationId"
	ExtensionResource     = "extensionconfigs"

	MinFluxctlWithPrivateAuthAndProxySupport = "0.1.15"
	MinFluxctlWithMemcachedDeleteSupport     = "0.1.16"

	CertExpirationTimePeriod = time.Hour * 24 * 90
	CertRenewalTimePeriod    = time.Hour * 24 * 46
	UpdateJobName            = "updatearcagentjob"
)

var (
	ValidKeyTypes = [...]string{"ssh-rsa", "ssh-ed25519", "ecdsa-sha2", "ssh-dss", "rsa-sha2"}
)

const (
	HelmDefaultValueResourceId          = "Azure.Cluster.ResourceId"
	HelmDefaultValueRegion              = "Azure.Cluster.Region"
	HelmDefaultValueDistro              = "Azure.Cluster.Distribution"
	HelmDefaultValueInfra               = "Azure.Cluster.Infrastructure"
	HelmDefaultValueProxyEnabled        = "Azure.proxySettings.isProxyEnabled"
	HelmDefaultValueProxyHttpsUri       = "Azure.proxySettings.httpsProxy"
	HelmDefaultValueProxyHttpUri        = "Azure.proxySettings.httpProxy"
	HelmDefaultValueProxyNoProxy        = "Azure.proxySettings.noProxy"
	HelmDefaultValueProxyCert           = "Azure.proxySettings.proxyCert"
	HelmDefaultValueIdentityEnabled     = "Azure.Identity.isEnabled"
	HelmDefaultValueIdentityType        = "Azure.Identity.Type"
	HelmDefaultValueExtensionName       = "Azure.Extension.Name"
	HelmDefaultValueExtensionResourceId = "Azure.Extension.ResourceId"
)
