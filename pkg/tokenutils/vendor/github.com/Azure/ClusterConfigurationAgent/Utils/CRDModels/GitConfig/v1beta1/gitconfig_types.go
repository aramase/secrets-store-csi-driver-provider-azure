package v1beta1

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"strings"

	"github.com/Azure/ClusterConfigurationAgent/Utils/Helper"

	"github.com/prometheus/common/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ConfigState int

const (
	InitialState ConfigState = iota
	StartedDownload
	DownloadSucceeded
	Installing
	InstallSucceeded
	FailedInstall
	Uninstalling
	UninstallSucceeded
	FailedUninstall
	RetryingToGetPublicKey

	// Failed to get public key but operator is installed
	FailedToGetPublicKey
)

const (
	OperatorScopeCluster    = "cluster"
	OperatorScopeNamespaced = "namespace"
)

func (cs ConfigState) String() string {
	// Ensure none of these strings are substrings of other strings as we perform compliance check with strings.Contains
	return [...]string{
		"Haven't started installing the operator",
		"Started downloading client",
		"Downloading client succeeded ",
		"Installing the operator yaml",
		"Successfully installed the operator",
		"Failed the install of the operator",
		"Uninstalling the operator",
		"Successfully uninstalled the operator",
		"Failed the uninstall of the operator",
		"Installed the operator but trying to get public key",
		"Installed the operator but couldn't get public key",
	}[cs]
}

// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// GitConfigSpec defines the desired state of GitConfig
type GitConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	OperatorParams          string                     `json:"operatorParams"`
	OperationClientLocation string                     `json:"operatorClientLocation"`
	OperatorType            string                     `json:"operatorType"`
	GitUrl                  string                     `json:"giturl"`
	SSHKnownHostsContents   string                     `json:"sshKnownHostsContents"`
	OperatorInstanceName    string                     `json:"operatorInstanceName"`
	OperatorScope           string                     `json:"operatorScope"`
	DeleteOperator          bool                       `json:"deleteOperator"`
	EnableHelmOperator      bool                       `json:"enableHelmOperator"`
	CorrelationId           string                     `json:"correlationId"`
	HelmOperatorProperties  HelmOperatorPropertiesSpec `json:"helmOperatorProperties"`
	ProtectedParameters     ProtectedParametersSpec    `json:"protectedParameters"`
	// Important: Run "make" to regenerate code after modifying this file
}

type HelmOperatorPropertiesSpec struct {
	RepoUrl      string `json:"repoUrl"`
	ChartName    string `json:"chartName"`
	ChartVersion string `json:"chartVersion"`
	ChartValues  string `json:"chartValues"`
}

type ProtectedParametersSpec struct {
	ReferenceName string                 `json:"referenceName"`
	Version       string                 `json:"version"`
	RawValues     map[string]interface{} `json:"-"`
}

// GitConfigStatus defines the observed state of GitConfig
type GitConfigStatus struct {
	ConfigAppliedTime        string   `json:"configAppliedTime"`
	Status                   string   `json:"status"`
	Message                  string   `json:"message"`
	PublicKey                string   `json:"publicKey"`
	LastGitCommitSynced      string   `json:"lastGitCommitInformation"`
	MostRecentEventsFromFlux []string `json:"mostRecentEventsFromFlux"`
	ErrorsInTheLastSynced    string   `json:"errorsInTheLastSynced"`
	LastPolledStatusTime     string   `json:"lastPolledStatusTime"`
	ProxyConfigHash          string   `json:"proxyConfigHash"`
	IsSyncedWithAzure        bool     `json:"isSyncedWithAzure"`
	RetryCountPublicKey      int      `json:"retryCountPublicKey"`
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// GitConfig is the Schema for the gitconfigs API
type GitConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   GitConfigSpec   `json:"spec,omitempty"`
	Status GitConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// GitConfigList contains a list of GitConfig
type GitConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []GitConfig `json:"items"`
}

type ProxyConfigProperties struct {
	ProxyConfigReferenceName string
	ProxyConfigHash          string
	ProxyCertReferenceName   string
}

type ProtectedParameterProperties struct {
	ReferenceName			   string
	Version					   string
	SSHPrivateKeyReferenceName string
	GitAuthReferenceName       string
	GitAuthHelmReferenceName   string
}

type OperatorProperties struct {
	OperatorParams                     string
	GitUrl                             string
	SSHKnownHostsContents              string
	OperatorInstanceName               string
	NamespacedScope                    bool
	EnabledHelmOperator                bool
	ExtensionOperatorPropertiesChanged bool
	HelmOperatorProperties             *HelmOperatorPropertiesSpec
	CrdInstanceName                    string
	ProxyConfigProperties              *ProxyConfigProperties
	ProtectedParametersProperties      *ProtectedParameterProperties
}

func (config *GitConfig) GetOperatorProperties() *OperatorProperties {
	spec := config.Spec
	properties := &OperatorProperties{}
	properties.OperatorParams = spec.OperatorParams
	properties.NamespacedScope = strings.ToLower(spec.OperatorScope) == OperatorScopeNamespaced
	properties.GitUrl = spec.GitUrl
	properties.SSHKnownHostsContents = spec.SSHKnownHostsContents
	properties.OperatorInstanceName = spec.OperatorInstanceName
	properties.ProtectedParametersProperties = &ProtectedParameterProperties{ReferenceName: spec.ProtectedParameters.ReferenceName, Version: spec.ProtectedParameters.Version}
	properties.ProxyConfigProperties = &ProxyConfigProperties{ProxyConfigHash: Helper.GetEnvironmentVar("PROXY_CONFIG_HASH")}
	properties.EnabledHelmOperator = spec.EnableHelmOperator
	if spec.EnableHelmOperator {
		properties.HelmOperatorProperties = &spec.HelmOperatorProperties
	}
	properties.CrdInstanceName = config.Name
	return properties
}

func init() {
	SchemeBuilder.Register(&GitConfig{}, &GitConfigList{})
}

func (spec *GitConfigSpec) Hash() string {
	jsonBytes, _ := json.Marshal(spec)
	log.Infof("Hashed Object for Git Spec : %s", (string)(jsonBytes))
	return hash(jsonBytes)
}

func hash(content []byte) string {
	b := md5.Sum(content)
	return base64.StdEncoding.EncodeToString(b[:])
}
