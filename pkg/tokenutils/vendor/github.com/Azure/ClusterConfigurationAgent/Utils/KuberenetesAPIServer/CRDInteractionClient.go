package KuberenetesAPIServer

import (
	"context"
	"fmt"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Helper"
	"log"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"time"

	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	settingsv1beta1 "github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/CustomLocationSettings/v1beta1"
	"github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/ExtensionConfig/v1beta1"
	v1beta12 "github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/GitConfig/v1beta1"

	"k8s.io/apimachinery/pkg/runtime"
)

type ConfigurationKind int

const (
	GitConfiguration = iota
	ExtensionConfiguration
	CustomLocationSettings
)

func (cs ConfigurationKind) String() string {
	return [...]string{
		Constants.GitConfigKind,
		Constants.ExtensionConfigKind}[cs]
}


type LocalConfiguration struct {
	LocalCrdObject runtime.Object
	ConfigurationKind ConfigurationKind
	CorrelationId	string
	Name 	string
}

type CRDInteractionClient interface {
	PutStatus() error
	GetStatus() (interface{}, error)
	CreateOrUpdate(bool) error
	UpdateStatus() error
	Delete() error
	GetCRDKind() ConfigurationKind
	IsMarkedForDeletion() bool
}

func CRDInteractionClientFactory(
	localConfig  LocalConfiguration,
	logger *LogHelper.Logger,
	apiClient *KubeAPIServerQueries,
	configCreatedTime time.Time) (CRDInteractionClient, error) {
	switch localConfig.ConfigurationKind {
	case GitConfiguration:
		client := &GitConfigurationClient{
			Config:            localConfig.LocalCrdObject.(*v1beta12.GitConfig),
			ApiClient:         apiClient,
			ConfigCreatedTime: configCreatedTime,
			Logger:            logger,
		}
		return client, nil
	case ExtensionConfiguration:
		client := &ExtensionConfigClient{
			Extension:         localConfig.LocalCrdObject.(*v1beta1.ExtensionConfig),
			ApiClient:         apiClient,
			ConfigCreatedTime: configCreatedTime,
			Logger:            logger,
		}
		client.InitLocationSettingsClient()
		return client, nil
	case CustomLocationSettings:
		client := &CustomLocationSettingsClient{
			LocationSettings:  localConfig.LocalCrdObject.(*settingsv1beta1.CustomLocationSettings),
			ApiClient:         apiClient,
			Logger:            logger,
		}
		return client, nil
	default:
		return nil, fmt.Errorf("Unsupported Configuration Kind")
	}
}


func (config *KubeAPIServerQueries) GetAllConfigurations(onlyChanged bool, filter []ConfigurationKind) ([]LocalConfiguration, error) {
	var list []LocalConfiguration

	if contains(filter, GitConfiguration) {
		gitconfigList := &v1beta12.GitConfigList{}
		// Get all the namespaces and all the GitConfigs to se the flag and ad to
		err := config.Client.List(context.Background(), gitconfigList, &client.ListOptions{
			Namespace:     config.Namespace,
		})
		if err != nil || len(gitconfigList.Items) == 0 {
			log.Printf("Could not get the list of configs : %v", err)
		}
		for _, config := range gitconfigList.Items {

			if !onlyChanged || config.Status.IsSyncedWithAzure == false {
				list = append(list, LocalConfiguration{
					LocalCrdObject:    config.DeepCopy(),
					ConfigurationKind: GitConfiguration,
					CorrelationId:     config.Spec.CorrelationId,
					Name:              config.Name,
				})
			}
		}
	}

	if !Helper.IsExtensionsEnabled() || !contains(filter, ExtensionConfiguration)	{
		return list, nil
	}


	extensionList := &v1beta1.ExtensionConfigList{}
	// Get all the namespaces and all the GitConfigs to se the flag and ad to
	err := config.Client.List(context.Background(), extensionList, &client.ListOptions{
		Namespace:     config.Namespace,
	})
	if err != nil || len(extensionList.Items) == 0 {
		log.Printf("Could not get the list of configs : %v", err)
	}

	customLocationSettings := &settingsv1beta1.CustomLocationSettingsList{}
	// Get all the namespaces and all the GitConfigs to se the flag and ad to
	if contains(filter, CustomLocationSettings) {
		err = config.Client.List(context.Background(), customLocationSettings, &client.ListOptions{
		Namespace:     Constants.ManagementNamespace,
	} )
	}
	customLocationSettingsMap := make(map[string]*settingsv1beta1.CustomLocationSettings)
	for _, customLocationSetting := range customLocationSettings.Items {
		customLocationSettingsMap[customLocationSetting.Name] = customLocationSetting.DeepCopy()
	}

	for _, config := range extensionList.Items {
		isSynced := config.Status.DataPlaneSyncStatus.IsSyncedWithAzure
		lastSync, err := time.Parse(Constants.TimeFormat, config.Status.DataPlaneSyncStatus.LastSyncTime)
		if err == nil {
			// if not synced in last 10 minutes we need to sync again
			isSynced = config.Status.DataPlaneSyncStatus.IsSyncedWithAzure && time.Now().Sub(lastSync) < 10 * time.Minute
		}

		if !onlyChanged || !isSynced {
			 localConfig := LocalConfiguration{
				LocalCrdObject:    config.DeepCopy(),
				ConfigurationKind: ExtensionConfiguration,
				CorrelationId: config.Spec.CorrelationId,
				Name: config.Name,
			}
			if setting , ok := customLocationSettingsMap[config.Name] ; ok {
				localConfig.LocalCrdObject.(*v1beta1.ExtensionConfig).CustomLocationSettings = setting
			}
			list = append(list,localConfig)
		}

	}
	return list, nil
}


func contains(arr []ConfigurationKind, item ConfigurationKind) bool {
	for _, a := range arr {
		if a == item {
			return true
		}
	}
	return false
}
