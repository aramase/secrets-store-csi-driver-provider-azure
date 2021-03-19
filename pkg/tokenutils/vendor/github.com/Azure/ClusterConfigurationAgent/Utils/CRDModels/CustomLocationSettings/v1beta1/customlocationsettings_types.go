/*
Copyright 2020 The Kubernetes authors.

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

// CustomLocationSettingsSpec defines the desired state of CustomLocationSettings
type CustomLocationSettingsSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of CustomLocationSettings. Edit CustomLocationSettings_types.go to remove/update
	RPAppId                   string                `json:"RPAppId,omitempty"`
	ClusterRole               string                `json:"ClusterRole,omitempty"`
	ResourceTypeMappings      []ResourceTypeMapping `json:"EnabledResourceTypes,omitempty"`
	ExtensionRegistrationTime int64                 `json:"ExtensionRegistrationTime,omitempty"`
}

type ResourceTypeMapping struct {
	APIVersion                string                     `json:"ApiVersion,omitempty"`
	ResourceType              string                     `json:"ResourceType,omitempty"`
	ResourceProviderNamespace string                     `json:"ResourceProviderNamespace,omitempty"`
	ResourceMapping           CustomTypeGroupVersionKind `json:"ResourceMapping,omitempty"`
}

type CustomTypeGroupVersionKind struct {
	Version string `json:"Version,omitempty"`
	Group   string `json:"Group,omitempty"`
	Kind    string `json:"Kind,omitempty"`
	Name    string `json:"Name,omitempty"`
}

// CustomLocationSettingsStatus defines the observed state of CustomLocationSettings
type CustomLocationSettingsStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +kubebuilder:object:root=true

// CustomLocationSettings is the Schema for the customlocationsettings API
type CustomLocationSettings struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomLocationSettingsSpec   `json:"spec,omitempty"`
	Status CustomLocationSettingsStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// CustomLocationSettingsList contains a list of CustomLocationSettings
type CustomLocationSettingsList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CustomLocationSettings `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CustomLocationSettings{}, &CustomLocationSettingsList{})
}
