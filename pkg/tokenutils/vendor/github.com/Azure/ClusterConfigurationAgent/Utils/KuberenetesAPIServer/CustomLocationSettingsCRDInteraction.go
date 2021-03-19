package KuberenetesAPIServer

import (
	"context"
	"fmt"

	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	"github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/CustomLocationSettings/v1beta1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type CustomLocationSettingsClient struct {
	LocationSettings *v1beta1.CustomLocationSettings
	ApiClient        *KubeAPIServerQueries
	Logger           *LogHelper.Logger
}

func (config *CustomLocationSettingsClient) GetCRDKind() ConfigurationKind {
	return CustomLocationSettings
}

func (config *CustomLocationSettingsClient) IsMarkedForDeletion() bool {
	return false
}

func (extensionClient *CustomLocationSettingsClient) GetStatus() (interface{}, error) {
	objectKey, _ := client.ObjectKeyFromObject(extensionClient.LocationSettings)
	locationSettings := &v1beta1.CustomLocationSettings{}

	apiConfig := extensionClient.ApiClient.Client
	err := apiConfig.Get(context.Background(), objectKey, locationSettings)
	if err != nil {
		return nil, err
	}
	return locationSettings, nil
}

func (extensionClient *CustomLocationSettingsClient) PutStatus() error {
	return fmt.Errorf("no status for customLocationSettings")
}

// No need for SetSynced in this case, just taking to  complete the interface contract
func (extensionClient *CustomLocationSettingsClient) CreateOrUpdate(setDelete bool) error {
	// creates the in-cluster config
	key, _ := client.ObjectKeyFromObject(extensionClient.LocationSettings)
	apiConfig := extensionClient.ApiClient.Client

	retrieve := extensionClient.LocationSettings.DeepCopy()
	locationSettings := extensionClient.LocationSettings

	err := apiConfig.Get(context.Background(), key, retrieve)

	objectExists := err == nil || !errors.IsNotFound(err)
	extensionClient.Logger.Info("Object Found : %v", objectExists)

	if locationSettings.Spec.ExtensionRegistrationTime == 0 ||
		(locationSettings.Spec.ClusterRole == "" &&
			locationSettings.Spec.RPAppId == "" &&
			len(locationSettings.Spec.ResourceTypeMappings) == 0) {

		// No Update Required
		return nil
	}

	updateRequired := retrieve.Spec.ExtensionRegistrationTime < locationSettings.Spec.ExtensionRegistrationTime

	// reset err
	err = nil

	if setDelete {
		errDel := extensionClient.deleteConfig(retrieve)
		if errDel != nil {
			extensionClient.Logger.Error("Unable to Delete the existing config due to error %s", errDel)
			return fmt.Errorf("Unable to Delete the existing config due to error %s", errDel)
		}
		return nil
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
		err = apiConfig.Create(context.Background(), locationSettings)
	} else if objectExists && updateRequired {
		//Needed for update
		locationSettings.ObjectMeta.ResourceVersion = retrieve.ObjectMeta.ResourceVersion
		err = apiConfig.Update(context.Background(), locationSettings)
	}

	return err
}

func (extensionClient *CustomLocationSettingsClient) UpdateStatus() error {
	return fmt.Errorf("no status for customLocationSettings")
}

func (extensionClient *CustomLocationSettingsClient) Delete() error {
	return extensionClient.deleteConfig(extensionClient.LocationSettings)
}

func (extensionClient *CustomLocationSettingsClient) deleteConfig(locationSettings *v1beta1.CustomLocationSettings) error {
	apiConfig := extensionClient.ApiClient.Client

	policy := metav1.DeletePropagationForeground
	deleteOption := &client.DeleteOptions{PropagationPolicy: &policy}

	err := apiConfig.Delete(context.Background(), locationSettings, deleteOption)
	return client.IgnoreNotFound(err)
}
