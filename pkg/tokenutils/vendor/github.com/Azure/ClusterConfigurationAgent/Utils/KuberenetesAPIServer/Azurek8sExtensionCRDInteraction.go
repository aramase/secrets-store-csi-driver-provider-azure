package KuberenetesAPIServer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"

	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	settingsv1beta1 "github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/CustomLocationSettings/v1beta1"
	v1beta12 "github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/ExtensionConfig/v1beta1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExtensionConfigClient struct {
	Extension              *v1beta12.ExtensionConfig
	locationSettingsClient *CustomLocationSettingsClient
	ApiClient              *KubeAPIServerQueries
	Logger                 *LogHelper.Logger
	ConfigCreatedTime      time.Time
}

func (config *ExtensionConfigClient) IsMarkedForDeletion() bool {
	return !config.Extension.DeletionTimestamp.IsZero()
}

func (config *ExtensionConfigClient) GetCRDKind() ConfigurationKind {
	return ExtensionConfiguration
}

func (config *ExtensionConfigClient) InitLocationSettingsClient() {

	// Create the client for settings
	if config.Extension.CustomLocationSettings == nil {
		config.Extension.CustomLocationSettings = &settingsv1beta1.CustomLocationSettings{
			ObjectMeta: metav1.ObjectMeta{
				Name:      config.Extension.Name,
				Namespace: Constants.ManagementNamespace,
			},
		}
	}

	config.locationSettingsClient = &CustomLocationSettingsClient{
		LocationSettings: config.Extension.CustomLocationSettings,
		ApiClient:        config.ApiClient,
		Logger:           config.Logger,
	}
}

func (extensionClient *ExtensionConfigClient) GetStatus() (interface{}, error) {
	return extensionClient.GetStatusWithConfig(true)
}

func (extensionClient *ExtensionConfigClient) GetStatusWithConfig(failIfStatusEmpty bool) (interface{}, error) {
	objectKey, _ := client.ObjectKeyFromObject(extensionClient.Extension)
	extension := &v1beta12.ExtensionConfig{}

	apiConfig := extensionClient.ApiClient.Client

	// creates the in-cluster config
	err := apiConfig.Get(context.Background(), objectKey, extension)
	if err != nil {
		return nil, err
	}
	if extension.Status.Status == "" && failIfStatusEmpty {
		return nil, fmt.Errorf("status not populated")
	}

	// Get Settings attached with the extension
	locationSettings, err := extensionClient.locationSettingsClient.GetStatus()
	if err == nil && locationSettings != nil {
		extension.CustomLocationSettings = locationSettings.(*settingsv1beta1.CustomLocationSettings)
		extensionClient.locationSettingsClient.LocationSettings = extension.CustomLocationSettings
	}

	// Overwrite the existing one with the new config retrieved
	extensionClient.Extension = extension

	return extension, nil
}

func (extensionClient *ExtensionConfigClient) PutStatus() error {
	objectKey, _ := client.ObjectKeyFromObject(extensionClient.Extension)
	extension := &v1beta12.ExtensionConfig{}

	apiConfig := extensionClient.ApiClient.Client

	// creates the in-cluster config
	err := apiConfig.Get(context.Background(), objectKey, extension)
	if err != nil {
		return err
	}

	extension.Status = extensionClient.Extension.Status
	err = apiConfig.Status().Update(context.Background(), extension)

	return err
}

func (extensionClient *ExtensionConfigClient) CreateOrUpdate(setDelete bool) error {
	// creates the in-cluster config
	key, _ := client.ObjectKeyFromObject(extensionClient.Extension)
	apiConfig := extensionClient.ApiClient.Client
	extension := extensionClient.Extension

	objectExists := false
	retrieve := extensionClient.Extension.DeepCopy()
	configList := &v1beta12.ExtensionConfigList{}
	err := apiConfig.List(context.Background(), configList)
	if err != nil {
		extensionClient.Logger.Error("Unable to retrieve current config error %s", err)
	} else {
		for _, retrievedConfig := range configList.Items {
			// Check for recreate scenario
			// Recreate scenario is valid if config created time in DP is not zero and
			// retrieved creation timestamp on cluster is before config timestamp in DP
			if strings.EqualFold(retrievedConfig.Name, extension.Name) {
				if !extensionClient.ConfigCreatedTime.IsZero() && retrievedConfig.CreationTimestamp.UTC().Before(extensionClient.ConfigCreatedTime.UTC()) {
					extensionClient.Logger.Info("Recreate scenario deleting the old config")
					err := extensionClient.deleteConfig(&retrievedConfig)
					if err != nil {
						extensionClient.Logger.Error("Unable to Delete the existing extension config due to error %s", err)
						return fmt.Errorf("Unable to Delete the existing extension config due to error %s", err)
					}
				} else {
					// If not recreate scenario, update retrieved config
					if retrievedConfig.Namespace != extension.Namespace {
						extensionClient.Logger.Error("Update of the Namespace is not a valid scenario")
						return fmt.Errorf("Update of the Namespace is not a valid scenario")
					}
					retrieve = retrievedConfig.DeepCopy()
					objectExists = true
				}
				break
			}
		}
	}

	// Spec Has changed thus need to add the spec
	updateRequired := retrieve.Spec.Hash() != extension.Spec.Hash()

	// reset err
	err = nil

	if setDelete {
		return extensionClient.setExtensionDelete(retrieve)
	}

	// If the update operations are successful then update the settings as well to reference them
	err = extensionClient.locationSettingsClient.CreateOrUpdate(setDelete)
	if err != nil {
		return err
	}

	// If not found and the delete operator is not selected then create the config
	if !objectExists && !setDelete {
		// Create Namespace if not exists // as the CRD needs to be applied on the CRD
		namespace := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: key.Namespace,
			},
		}
		_ = apiConfig.Create(context.Background(), &namespace)
		err = apiConfig.Create(context.Background(), extension)
	} else if objectExists && updateRequired {
		//Needed for update
		extension.ObjectMeta.ResourceVersion = retrieve.ObjectMeta.ResourceVersion
		// Preserve the old status for the CRD
		extension.Status = retrieve.Status
		err = apiConfig.Update(context.Background(), extension)
	}

	return err
}

func (extensionClient *ExtensionConfigClient) UpdateStatus() error {
	apiConfig := extensionClient.ApiClient.Client
	extension := extensionClient.Extension
	key, _ := client.ObjectKeyFromObject(extension)
	retrieve := &v1beta12.ExtensionConfig{}

	err := apiConfig.Get(context.Background(), key, retrieve)
	objectExists := err == nil || !errors.IsNotFound(err)
	extensionClient.Logger.Info("Extension config object Found : %v", objectExists)

	if client.IgnoreNotFound(err) != nil {
		extensionClient.Logger.Error("Unable to retrieve current config error %s", err)
	} else if objectExists {
		extension = retrieve.DeepCopy()

		extension.Status.DataPlaneSyncStatus = v1beta12.SyncStatus{}
		extension.Status.DataPlaneSyncStatus.IsSyncedWithAzure = true
		extension.Status.DataPlaneSyncStatus.LastSyncTime = time.Now().Format(Constants.TimeFormat)

		err = apiConfig.Status().Update(context.Background(), extension)
	}
	return err
}

func (extensionClient *ExtensionConfigClient) Delete() error {
	return extensionClient.deleteConfig(extensionClient.Extension)
}

func (extensionClient *ExtensionConfigClient) deleteConfig(configToDelete *v1beta12.ExtensionConfig) error {
	extensionClient.Logger.Info("Attempting to delete the configuration: %v", configToDelete.Name)
	apiConfig := extensionClient.ApiClient.Client
	policy := metav1.DeletePropagationForeground
	deleteOption := &client.DeleteOptions{PropagationPolicy: &policy}
	key := client.ObjectKey{
		Namespace: configToDelete.Namespace,
		Name:      configToDelete.Name,
	}

	err := apiConfig.Delete(context.Background(), configToDelete, deleteOption)
	if client.IgnoreNotFound(err) != nil {
		extensionClient.Logger.Error("Failed to delete the configuration: %v with err: %v", configToDelete.Name, err)
		return err
	}

	// Wait for deletion to complete to complete
	endtime := time.Now().Add(2*time.Minute)
	for time.Now().Before(endtime) && !errors.IsNotFound(apiConfig.Get(context.Background(), key, configToDelete)) {
		time.Sleep(5*time.Second)
	}

	return nil
}
func (extensionClient *ExtensionConfigClient) setExtensionDelete(retrieve *v1beta12.ExtensionConfig) error {
	apiConfig := extensionClient.ApiClient.Client
	errDel := apiConfig.Delete(context.Background(), retrieve)
	if errDel != nil {
		extensionClient.Logger.Error("Unable to Delete the existing config due to error %s", errDel)
		return fmt.Errorf("Unable to Delete the existing config due to error %s", errDel)
	}

	errDel = extensionClient.locationSettingsClient.Delete()
	if errDel != nil {
		extensionClient.Logger.Error("Unable to Delete the custom location settings crd due to error %s", errDel)
		return fmt.Errorf("Unable to Delete the custom location settings crd due to error %s", errDel)
	}
	// Delete SuccessFull
	return nil
}
