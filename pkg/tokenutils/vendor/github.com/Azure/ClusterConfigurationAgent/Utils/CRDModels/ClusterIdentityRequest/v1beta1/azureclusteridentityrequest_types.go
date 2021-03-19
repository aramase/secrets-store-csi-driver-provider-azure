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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// AzureClusterIdentityRequestSpec defines the desired state of AzureClusterIdentityRequest
type AzureClusterIdentityRequestSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Audience   string `json:"audience,omitempty"`
	ResourceId string `json:"resourceId,omitempty"`
	ApiVersion string `json:"apiVersion,omitempty"`
}

type TokenReference struct {
	SecretName string `json:"secretName,omitempty"`
	DataName   string `json:"dataName,omitempty"`
}

// AzureClusterIdentityRequestStatus defines the observed state of AzureClusterIdentityRequest
type AzureClusterIdentityRequestStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
	TokenReference TokenReference `json:"tokenReference,omitempty"`
	ExpirationTime string `json:"expirationTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// AzureClusterIdentityRequest is the Schema for the azureclusteridentityrequests API
type AzureClusterIdentityRequest struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AzureClusterIdentityRequestSpec   `json:"spec,omitempty"`
	Status AzureClusterIdentityRequestStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// AzureClusterIdentityRequestList contains a list of AzureClusterIdentityRequest
type AzureClusterIdentityRequestList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AzureClusterIdentityRequest `json:"items"`
}

func init() {
	SchemeBuilder.Register(&AzureClusterIdentityRequest{}, &AzureClusterIdentityRequestList{})
}
