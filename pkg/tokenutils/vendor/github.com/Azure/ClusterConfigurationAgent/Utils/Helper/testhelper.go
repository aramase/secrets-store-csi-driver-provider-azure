package Helper

import (
	"fmt"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

func SetMockEnvironments(extensionEnable bool) map[string]map[string][]byte {
	_ = os.Setenv("AZURE_SUBSCRIPTION_ID", "testSub")
	_ = os.Setenv("AZURE_RESOURCE_GROUP", "testRG")
	_ = os.Setenv("AZURE_RESOURCE_NAME", "testCluster")
	_ = os.Setenv("AZURE_REGION", "devEnvironment")
	_ = os.Setenv("AZURE_TENANT_ID", "72f988bf-86f1-41af-91ab-2d7cd011db47")
	_ = os.Setenv("HELM_AUTO_UPDATE_CHECK_FREQUENCY_IN_MINUTES", "1")
	_ = os.Setenv("ARC_AGENT_RELEASE_TRAIN", "dev")
	_ = os.Setenv("ARC_AGENT_HELM_CHART_NAME", "k8s_agent")
	_ = os.Setenv("CLUSTER_TYPE", "ConnectedClusters") // ManagedClusters (AKS) or ConnectedClusters
	_ = os.Setenv(Constants.SecretName, Constants.WellKnownKubernetesSecret)
	_ = os.Setenv(Constants.SecretNamespace, Constants.WellKnownAzureArcManagementNamespace)
	_ = os.Setenv("AZURE_ARC_AUTOUPDATE", "false")
	_ = os.Setenv("EXTENSION_OPERATOR_ENABLED", strconv.FormatBool(extensionEnable))
	_ = os.Setenv("FLUX_CLIENT_DEFAULT_LOCATION", GetFluxCtlPath())
	_ = os.Setenv("LOCAL_TESTING_ROOT_PATH", filepath.Join(GetTestPath(), "data"))
	_ = os.Setenv("HELM_LOCAL_TESTING_CACHE_PATH", filepath.Join(GetTestPath(), "helmcache"))
	_ = os.Setenv("KUBERNETES_DISTRO", "override-distro")
	_ = os.Setenv("KUBERNETES_INFRA", "override-infra")

	proxySecrets := map[string]map[string][]byte{}
	proxySecrets[Constants.ProxyConfigSecretPrefix] = map[string][]byte{
		"HTTPS_PROXY": []byte("override-httpsproxy"),
		"HTTP_PROXY": []byte("override-httpproxy"),
		"NO_PROXY":   []byte("override-no-proxy"),
	}
	proxySecrets[Constants.ProxyCertSecretPrefix] = map[string][]byte{
		Constants.ProxyCertFileName: []byte("override-cert"),
	}
	return proxySecrets
}

func GetTestPath() string {
	folder, _ := os.Getwd()
	folder = strings.TrimSuffix(folder, string(os.PathSeparator))
	index := strings.Index(folder, "ClusterConfigurationAgent")
	return filepath.Join(folder[:index], "ClusterConfigurationAgent", "TestFiles")
}

func GetExecutablePath(path string) string {
	if runtime.GOOS == "windows" {
		return path + ".exe"
	}
	return path
}

func GetFluxCtlPath() string {
	if runtime.GOOS == "windows" {
		return fmt.Sprintf("azurearcfork8sdev.azurecr.io/arc-preview/fluxctl:%s-win", Constants.FluxCTLCurrentVersion)
	} else {
		return fmt.Sprintf("azurearcfork8sdev.azurecr.io/arc-preview/fluxctl:%s", Constants.FluxCTLCurrentVersion)
	}
}
