package KuberenetesAPIServer

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	v1beta12 "github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/GitConfig/v1beta1"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type GitConfigurationClient struct {
	Config            *v1beta12.GitConfig
	ApiClient         *KubeAPIServerQueries
	Logger            *LogHelper.Logger
	ConfigCreatedTime time.Time
}

func (config *GitConfigurationClient) IsMarkedForDeletion() bool {
	return config.Config.Spec.DeleteOperator || !config.Config.DeletionTimestamp.IsZero()
}

func (config *GitConfigurationClient) GetStatus() (interface{}, error) {
	return config.GetStatusWithConfig(true)
}
func (config *GitConfigurationClient) GetCRDKind() ConfigurationKind {
	return GitConfiguration
}

func (config *GitConfigurationClient) GetStatusWithConfig(failIfStatusEmpty bool) (interface{}, error) {
	objectKey, _ := client.ObjectKeyFromObject(config.Config)
	gitConfig := &v1beta12.GitConfig{}

	apiConfig := config.ApiClient.Client

	// creates the in-cluster config
	err := apiConfig.Get(context.Background(), objectKey, gitConfig)
	if err != nil {
		return nil, err
	}
	if gitConfig.Status.Status == "" && failIfStatusEmpty {
		return nil, fmt.Errorf("status not populated")
	}

	// Overwrite the existing one with the new config retrieved
	config.Config = gitConfig

	// empty initilize the array
	if gitConfig.Status.MostRecentEventsFromFlux == nil {
		gitConfig.Status.MostRecentEventsFromFlux = []string{}
	}
	return &gitConfig.Status, nil
}

func (config *GitConfigurationClient) PutStatus() error {
	objectKey, _ := client.ObjectKeyFromObject(config.Config)
	gitConfig := &v1beta12.GitConfig{}

	apiConfig := config.ApiClient.Client

	// creates the in-cluster config
	err := apiConfig.Get(context.Background(), objectKey, gitConfig)
	if err != nil {
		return err
	}

	gitConfig.Status = config.Config.Status
	err = apiConfig.Status().Update(context.Background(), gitConfig)

	return err
}

func (config *GitConfigurationClient) CreateOrUpdate(setDelete bool) error {
	// creates the in-cluster config
	key, _ := client.ObjectKeyFromObject(config.Config)
	apiConfig := config.ApiClient.Client
	gitConfig := config.Config
	gitConfig.Spec.DeleteOperator = setDelete

	objectExists := false
	retrieve := config.Config.DeepCopy()
	configList := &v1beta12.GitConfigList{}
	err := apiConfig.List(context.Background(), configList)
	if err != nil {
		config.Logger.Error("Unable to retrieve current config error %s", err)
	} else {
		// Check for recreate scenario
		// Recreate scenario is valid if config created time in DP is not zero and
		// retrieved creation timestamp on cluster is before config timestamp in DP
		for _, retrievedConfig := range configList.Items {
			if strings.EqualFold(retrievedConfig.Name, gitConfig.Name) {
				if !config.ConfigCreatedTime.IsZero() && retrievedConfig.CreationTimestamp.UTC().Before(config.ConfigCreatedTime.UTC()) {
					config.Logger.Info("Recreate scenario deleting the old config")
					err := config.deleteConfig(&retrievedConfig)
					if err != nil {
						config.Logger.Error("Unable to Delete the existing config due to error %s", err)
						return fmt.Errorf("Unable to Delete the existing config due to error %s", err)
					}
				} else {
					// If not recreate scenario, update retrieved config
					if retrievedConfig.Namespace != gitConfig.Namespace {
						config.Logger.Error("Update of the Namespace is not a valid scenario")
						return fmt.Errorf("Update of the Namespace is not a valid scenario")
					}
					retrieve = retrievedConfig.DeepCopy()
					objectExists = true
				}
				break
			}
		}
		// Do validation that we aren't creating a different config with an operator with the same instance name in the same namespace
		for _, retrievedConfig := range configList.Items {
			if strings.EqualFold(gitConfig.Spec.OperatorInstanceName, retrievedConfig.Spec.OperatorInstanceName) &&
				strings.EqualFold(gitConfig.Namespace, retrievedConfig.Namespace) && !strings.EqualFold(gitConfig.Name, retrievedConfig.Name){
				config.Logger.Error("Unable to create the config with this operator instance name because this instance name already exists in the configuration namespace")
				return fmt.Errorf("Unable to create the config with this operator instance name because this instance name already exists in the configuration namespace")
			}
		}
	}

	// Update protected parameters before checking hash to validate if parameters have changed
	err = config.updateProtectedParameters(gitConfig)
	if err != nil {
		config.Logger.Error("Unable to provision protected parameters %s", err)
	}

	config.Logger.Info("Object Found: %v", objectExists)
	// Spec Has changed thus need to add the spec
	updateRequired := retrieve.Spec.Hash() != gitConfig.Spec.Hash()

	config.Logger.Info("Updated Required: %v", updateRequired)

	// reset err
	err = nil

	// If not found and the delete operator is not selected then create the config
	if !objectExists && !gitConfig.Spec.DeleteOperator {
		config.Logger.Info("Attempting to create the new configuration: %v", gitConfig.Name)
		// Create Namespace if not exists // as the CRD needs to be applied on the CRD
		namespace := v1.Namespace{
			ObjectMeta: metav1.ObjectMeta{
				Name: key.Namespace,
			},
		}
		gitConfig.Spec.DeleteOperator = false
		// Try creating the namespace and ignore the error as it might already be present
		_ = apiConfig.Create(context.Background(), &namespace)

		err = apiConfig.Create(context.Background(), gitConfig)
	} else if objectExists && updateRequired {
		config.Logger.Info("Attempting to update the old configuration to a newer configuration: %v", gitConfig.Name)
		// Remove the Status so that it is retried
		retrieve.Status = v1beta12.GitConfigStatus{MostRecentEventsFromFlux: []string{}}
		err = apiConfig.Status().Update(context.Background(), retrieve)

		if client.IgnoreNotFound(err) != nil {
			config.Logger.Error("Failed to update the status of the configuration: %v with an error: %v", gitConfig.Name, err)
			return err
		}
		config.Logger.Info("Successfully updated the status to empty for an update of the config: %v", gitConfig.Name)

		gitConfig.ResourceVersion = retrieve.ResourceVersion
		gitConfig.Status = retrieve.Status
		err = apiConfig.Update(context.Background(), gitConfig)
		if err != nil {
			config.Logger.Error("Failed to update the spec of the configuration: %v with an error: %v", gitConfig.Name, err)
			return err
		}
		config.Logger.Info("Successfully updated the spec of the config: %v", gitConfig.Name)
	}
	return nil
}

func (config *GitConfigurationClient) UpdateStatus() error {
	config.Logger.Info("Attempting to update the status of the config: %v", config.Config.Name)
	// creates the in-cluster config
	apiConfig := config.ApiClient.Client
	key, _ := client.ObjectKeyFromObject(config.Config)
	retrieve := &v1beta12.GitConfig{}

	// Gets the config from the cluster if it exists
	// If it exists, update config object status to be synced with Azure
	err := apiConfig.Get(context.Background(), key, retrieve)
	objectExists := err == nil || !errors.IsNotFound(err)
	config.Logger.Info("Git config object Found : %v", objectExists)

	if client.IgnoreNotFound(err) != nil {
		config.Logger.Error("Unable to retrieve current config error %s", err)
		return err
	} else if objectExists {
		retrieve.Status.IsSyncedWithAzure = true
		err = apiConfig.Status().Update(context.Background(), retrieve)
	}
	if err != nil {
		config.Logger.Info("Successfully updated the status of the config: %v", config.Config.Name)
	}
	return err
}

func (config *GitConfigurationClient) Delete() error {
	return config.deleteConfig(config.Config)
}

func (config *GitConfigurationClient) deleteConfig(configToDelete *v1beta12.GitConfig) error {
	config.Logger.Info("Attempting to delete the configuration {%v}", configToDelete.Name)
	apiConfig := config.ApiClient.Client
	policy := metav1.DeletePropagationForeground
	deleteOption := &client.DeleteOptions{PropagationPolicy: &policy}

	err := apiConfig.Delete(context.Background(), configToDelete, deleteOption)
	if client.IgnoreNotFound(err) != nil {
		config.Logger.Error("Failed to delete the configuration {%v} with err: %v", configToDelete.Name, err)
		return err
	}
	err = config.ApiClient.DeleteSecretFromNamespace(configToDelete.Spec.ProtectedParameters.ReferenceName, Constants.ManagementNamespace)
	if client.IgnoreNotFound(err) != nil {
		config.Logger.Error("Failed to delete the protected parameters from the namespace with err: %v", err)
	}
	config.Logger.Info("Successfully deleted the configuration {%v}", configToDelete.Name)
	return nil
}

func (config *GitConfigurationClient) updateProtectedParameters(gitConfig *v1beta12.GitConfig) error {
	config.Logger.Info("Attempting to update protected parameters for config: %v", config.Config.Name)
	protectedParams := gitConfig.Spec.ProtectedParameters.RawValues
	apiClient := config.ApiClient

	protectedParameterRef := strings.ToLower(fmt.Sprintf(`%s-%s`, Constants.ProtectedParametersSecretPrefix, gitConfig.Name))
	// secrets must be numbers, lower case letters, ".", or "-"
	gitConfig.Spec.ProtectedParameters.ReferenceName = protectedParameterRef

	var parsedParams = make(map[string][]byte)

	for key, val := range protectedParams {
		data, err := base64.StdEncoding.DecodeString(val.(string))
		if err != nil {
			return fmt.Errorf("error: unable to parse protected parameter as base64 encoded string")
		}
		parsedParams[key] = []byte(data)
	}

	version, err := apiClient.PutSecretsToNamespace(gitConfig.Spec.ProtectedParameters.ReferenceName, Constants.ManagementNamespace, parsedParams)
	gitConfig.Spec.ProtectedParameters.Version = version
	config.Logger.Info("Successfully put the new version of the protected parameters into the management namespace config: %v with version: %v", config.Config.Name, version)
	return err
}
