/*

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1beta1

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	settingsv1beta1 "github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/CustomLocationSettings/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ExtensionState int

const (
	Installed ExtensionState = iota
	FailedInstall
	Pending
	FailedDelete
)

const (
	OperatorScopeCluster    = "cluster"
	OperatorScopeNamespaced = "namespace"
)

func (cs ExtensionState) String() string {
	return [...]string{
		"Successfully installed the extension",
		"Failed to install the extension",
		"Installing the extension",
		"Failed to delete the extension"}[cs]
}

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ExtensionConfigSpec defines the desired state of ExtensionConfig
type ExtensionConfigSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Parameter               map[string]string `json:"parameter,omitempty"`
	ExtensionType           string            `json:"extensionType,omitempty"`
	RepoUrl                 string            `json:"repoUrl,omitempty"`
	CorrelationId           string            `json:"correlationId,omitempty"`
	Version                 string            `json:"version,omitempty"`
	ReleaseTrain            string            `json:"releaseTrain,omitempty"`
	AutoUpgradeMinorVersion bool              `json:"autoUpgradeMinorVersion,omitempty"`
	LastModifiedTime        metav1.Time       `json:"lastModifiedTime,omitempty"`
}

type SyncStatus struct {
	IsSyncedWithAzure bool   `json:"isSyncedWithAzure"`
	LastSyncTime      string `json:"lastSyncTime"`
}

// ExtensionConfigStatus defines the observed state of ExtensionConfig
type ExtensionConfigStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	ConfigAppliedTime        string     `json:"configAppliedTime"`
	Status                   string     `json:"status"`
	Message                  string     `json:"message"`
	OperatorPropertiesHashed string     `json:"operatorPropertiesHashed"`
	DataPlaneSyncStatus      SyncStatus `json:"syncStatus"`
}

// +kubebuilder:object:root=true

// ExtensionConfig is the Schema for the extensionconfigs API
type ExtensionConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ExtensionConfigSpec   `json:"spec,omitempty"`
	Status ExtensionConfigStatus `json:"status,omitempty"`

	CustomLocationSettings *settingsv1beta1.CustomLocationSettings `json:"-"`
}

// +kubebuilder:object:root=true

// ExtensionConfigList contains a list of ExtensionConfig
type ExtensionConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ExtensionConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ExtensionConfig{}, &ExtensionConfigList{})
}

func (spec *ExtensionConfigSpec) Hash() string {
	jsonBytes, _ := json.Marshal(spec)
	//log.Infof("Hashed Object for ExtensionConfig Spec : %s", (string)(jsonBytes))
	return hash(jsonBytes)
}

func hash(content []byte) string {
	b := md5.Sum(content)
	return base64.StdEncoding.EncodeToString(b[:])
}

func GetStatus(state ExtensionState) string {
	return state.String()
}


func (this *ExtensionConfig) IsInstalledSuccessfully() bool {
	return this.Status.Status == GetStatus(Installed)
}
