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
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sort"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

type ServiceAccount struct{
	Name string `json:"name,omitempty"`
	Namespace string `json:"namespace,omitempty"`
}

// AzureExtensionIdentitySpec defines the desired state of AzureExtensionIdentity
type AzureExtensionIdentitySpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	ServiceAccounts []ServiceAccount `json:"serviceAccounts,omitempty"`
	TokenNamespace string `json:"tokenNamespace,omitempty"`
}

// AzureExtensionIdentityStatus defines the observed state of AzureExtensionIdentity
type AzureExtensionIdentityStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true
// AzureExtensionIdentity is the Schema for the azureextensionidentities API
type AzureExtensionIdentity struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureExtensionIdentitySpec   `json:"spec,omitempty"`
	Status AzureExtensionIdentityStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AzureExtensionIdentityList contains a list of AzureExtensionIdentity
type AzureExtensionIdentityList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureExtensionIdentity `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureExtensionIdentity{}, &AzureExtensionIdentityList{})
}

func(this AzureExtensionIdentitySpec) GetServiceAccounts() []v1.Subject {
	var subjects []v1.Subject
	for _, serviceAccount := range this.ServiceAccounts {
		subjects = append(subjects, v1.Subject{
			Kind:      v1.ServiceAccountKind,
			Name:      serviceAccount.Name,
			Namespace: serviceAccount.Namespace,
		})
	}
	return subjects
}

// Union the Service Account with the existing ones
func(this AzureExtensionIdentitySpec) AddServiceAccount(list []ServiceAccount) bool  {
	changed := false
	for _, serviceAccount := range list {
		index := sort.Search(len(this.ServiceAccounts), func(i int) bool {
			return this.ServiceAccounts[i].Namespace == serviceAccount.Namespace &&
				this.ServiceAccounts[i].Name == serviceAccount.Name
		})
		if index < 0 {
			this.ServiceAccounts = append(this.ServiceAccounts, serviceAccount)
			changed = true
		}
	}
	return changed
}