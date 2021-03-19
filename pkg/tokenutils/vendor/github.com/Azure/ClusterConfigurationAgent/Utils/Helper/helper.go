package Helper

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/errors"

	v1beta12 "github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/CustomLocationSettings/v1beta1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"

	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type retryableFunction func() (interface{}, error)

var (
	agentHeartbeat = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "agent_heartbeat",
		Help: "Heartbeat metric for an arc agent",
	})
)

func RegisterHeartbeatMetric() {
	err := metrics.Registry.Register(agentHeartbeat)

	if err != nil {
		fmt.Println("Error registering heartbeat metric. Error: %v", err)
	} else {
		fmt.Println("Registered heartbeat metric successfully")
	}

}

// Retry for time duration and then let the docker container crash
func WithRetry(retryLogic retryableFunction, fatalError bool, duration time.Duration) (interface{}, error) {
	var result interface{}
	channel := make(chan interface{}, 1)
	stop := make(chan interface{})
	var lastErr error
	go func() {
		for {
			functionResult, err := retryLogic()
			if err == nil {
				channel <- functionResult
				break
			}
			lastErr = err
			select {
			case <-stop:
				//timed out
				return
			default:
				time.Sleep(30 * time.Second)
			}
		}
	}()

	select {
	case channelResult := <-channel:
		result = channelResult
	case <-time.After(duration):
		errorMsg := fmt.Sprintf("Error : Retry for given duration didn't get any results with err {%v}", lastErr)
		if fatalError {
			log.Fatalf(errorMsg)
		} else {
			close(stop)
			log.Printf(errorMsg)
			return nil, fmt.Errorf(errorMsg)
		}
	}
	close(stop)
	return result, nil
}

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GetEnvironmentVar(envKey string) (envValue string) {
	envValue = os.Getenv(envKey)
	if len(envValue) <= 0 {
		return ""
	}

	return envValue
}

// https://book.kubebuilder.io/reference/using-finalizers.html
func RemoveStringFromSlice(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}

func GetEnvironmentVarWithDefault(envKey string, defaultVaule string) (envValue string) {
	envValue = os.Getenv(envKey)
	if len(envValue) <= 0 {
		return defaultVaule
	}

	return envValue
}

func IsManagedIdentityAuthEnabled() bool {
	return GetEnvironmentBoolWithDefault(Constants.MANAGED_IDENTITY_AUTH, false)
}

func IsNullAuthEnabled() bool {
	return GetEnvironmentBoolWithDefault(Constants.NO_AUTH, false)
}

func IsExtensionsEnabled() bool {
	return GetEnvironmentBoolWithDefault(Constants.ExtensionOperatorEnabled, false)
}

func IsGitOpsEnabled() bool {
	return GetEnvironmentBoolWithDefault(Constants.GitOpsEnabled, true)
}

func IsDebugLogging() bool {
	return GetEnvironmentBoolWithDefault(Constants.LabelDebugLogging, false)
}

func IsFluxUpstreamEnabled() bool {
	return GetEnvironmentBoolWithDefault(Constants.LabelFluxUpstreamEnabled, false)
}

func IsClusterConnectAgentEnabled() bool {
	return GetEnvironmentBoolWithDefault(Constants.ClusterConnectAgentEnabled, false)
}

func GetEnvironmentBoolWithDefault(envKey string, defaultVaule bool) bool {
	envBool, err := strconv.ParseBool(GetEnvironmentVar(envKey))
	if err != nil {
		envBool = defaultVaule
	}
	return envBool
}

func GetClusterLocation() string {
	return GetEnvironmentVar(Constants.LabelAzureRegion)
}

func HISGlobalEndpoint() string {
	var suffix string
	location := strings.ToLower(GetEnvironmentVar(Constants.LabelAzureRegion))

	// Return the override endpoint if variable is set
	overrideEndpoint := os.Getenv(Constants.HISOverrideVariable)
	if overrideEndpoint != "" {
		return overrideEndpoint
	}

	if strings.HasPrefix(location, "usgov") || strings.HasPrefix(location, "usdod") {
		suffix = "us"
	} else {
		suffix = "com"
	}
	return fmt.Sprintf(Constants.HISGlobalEndpointTemplate, suffix, location)
}

func ActiveDirectoryEndpoint() string {
	var suffix string
	location := strings.ToLower(GetEnvironmentVar(Constants.LabelAzureRegion))

	if strings.HasPrefix(location, "usgov") || strings.HasPrefix(location, "usdod") {
		suffix = "us"
	} else {
		suffix = "com"
	}
	return fmt.Sprintf(Constants.ActiveDirectoryEndpointTemplate, suffix)
}

func DPEndpoint() string {
	var suffix string
	location := strings.ToLower(GetEnvironmentVar(Constants.LabelAzureRegion))

	if strings.HasPrefix(location, "usgov") || strings.HasPrefix(location, "usdod") {
		suffix = "us"
	} else {
		suffix = "com"
	}
	return fmt.Sprintf(Constants.RegionBasedEndpointTemplate, location, suffix)
}

func PopulateFromEnvVars() (string, string) {
	SubscriptionID := GetEnvironmentVar(Constants.LabelAzureSubscriptionId)
	GroupName := GetEnvironmentVar(Constants.LabelAzureResourceGroup)
	ResourceName := GetEnvironmentVar(Constants.LabelAzureResourceName)
	Location := GetEnvironmentVar(Constants.LabelAzureRegion)
	ClusterType := GetEnvironmentVar(Constants.LabelClusterType) // Managed Cluster (AKS) / Connected Cluster / Appliance
	armID := fmt.Sprintf(Constants.ArmID, SubscriptionID, GroupName, ClusterType, ResourceName)
	if strings.ToLower(ClusterType) == Constants.Appliances {
		armID = fmt.Sprintf(Constants.ArmIDForAppliance, SubscriptionID, GroupName, ClusterType, ResourceName)
	}
	return armID, Location
}

func GetTokenInBytes(accessToken string) []byte {
	return []byte(accessToken)
}

func GetDefaultUpdateCheckFrequency() time.Duration {
	// Override if UpdateCheckFrequencyInMinutes is changed in chart
	defaultUpdateCheckFrequencyInMinutesString := GetEnvironmentVar("HELM_AUTO_UPDATE_CHECK_FREQUENCY_IN_MINUTES")
	defaultUpdateCheckFrequencyInMinutesInt, err := strconv.ParseInt(defaultUpdateCheckFrequencyInMinutesString, 10, 32)
	if err == nil && defaultUpdateCheckFrequencyInMinutesInt > 0 {
		return time.Duration(defaultUpdateCheckFrequencyInMinutesInt) * time.Minute
	}
	if err != nil {
		log.Printf("Unabled to parse:HELM_AUTO_UPDATE_CHECK_FREQUENCY_IN_MINUTES value. default is one hour. Error:{%s}", err)
	}
	return time.Hour
}

func GetCustomLocationSettings(configName string, apiClient client.Client) (*v1beta12.CustomLocationSettings, bool, error) {
	// Get the CRD and the ClusterRoles to be added
	customSettings := &v1beta12.CustomLocationSettings{}
	key := client.ObjectKey{
		Namespace: Constants.ManagementNamespace,
		Name:      configName,
	}
	err := apiClient.Get(context.Background(), key, customSettings)
	objectExists := err == nil || !errors.IsNotFound(err)
	return customSettings, objectExists, err
}

// Set Difference: A - B
func Difference(a, b []string) (diff []string) {
	m := make(map[string]bool)

	for _, item := range b {
		m[item] = true
	}

	for _, item := range a {
		if _, ok := m[item]; !ok {
			diff = append(diff, item)
		}
	}
	return
}

func StartMetricsServer() {
	log.Println("Start metrics server :success")
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}

// Appends the hash of the audience as a suffix to the CRD Name
func GetIdentityRequestCRDName(audience string, resourceId string) string {
	sha256SumBytes := sha256.Sum256([]byte(audience))
	if len(resourceId) > 0 {
		// include resourceId if given
		hashString := []byte(fmt.Sprintf("%s-%s", audience, resourceId))
		sha256SumBytes = sha256.Sum256(hashString)
	}
	return strings.ToLower(fmt.Sprintf(Constants.IdentityRequestCRDName, sha256SumBytes))
}

// Given cluster type, get the first party appID used as ManagedIdentityTokenProvider audience
func GetIdentityTokenAudienceFromClusterType(clusterType string) string {
	if strings.ToLower(clusterType) == Constants.Appliances {
		return Constants.ApplianceConnectAgentToDataPlaneAppId
	}
	return Constants.ClusterConfigConnectedClusterFirstPartyAppId
}

func SetHeartbeatMetric(agentRunningStatus bool) {
	if agentRunningStatus == true {
		agentHeartbeat.Set(1)
	} else {
		agentHeartbeat.Set(2)
	}
}
