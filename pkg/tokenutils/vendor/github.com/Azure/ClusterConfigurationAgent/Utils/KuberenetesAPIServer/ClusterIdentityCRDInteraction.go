package KuberenetesAPIServer

import (
	"context"
	"fmt"
	"time"

	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	"github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/ClusterIdentityRequest/v1beta1"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterIdentityCRDClient struct {
	identityRequest *v1beta1.AzureClusterIdentityRequest
	logger          LogHelper.LogWriter
	ApiClient       *KubeAPIServerQueries
	ExpirationTime  time.Time
	Token           string
	namespace       string
}

// Initializer for the ClusterIdentity CRD
func (identity *ClusterIdentityCRDClient) Init(name string, namespace string, apiClient *KubeAPIServerQueries, logger LogHelper.LogWriter, audience, resourceId string) {
	identity.identityRequest = &v1beta1.AzureClusterIdentityRequest{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.AzureClusterIdentityRequestSpec{
			Audience: audience,
		},
	}
	if len(resourceId) > 0 {
		identity.identityRequest.Spec.ResourceId = resourceId
	}
	identity.logger = logger
	identity.ApiClient = apiClient
	identity.ExpirationTime = time.Now() // Initialize with current time, so this is always expired when created
	identity.namespace = namespace
}

//
func (identity *ClusterIdentityCRDClient) GetTokenFromStatus() (interface{}, error) {
	objectKey, _ := client.ObjectKeyFromObject(identity.identityRequest)
	clusterIdentity := &v1beta1.AzureClusterIdentityRequest{}

	apiClient := identity.ApiClient.Client

	// creates the in-cluster identity
	err := apiClient.Get(context.Background(), objectKey, clusterIdentity)
	if err != nil {
		return nil, err
	}
	if clusterIdentity.Status.TokenReference.DataName == "" || clusterIdentity.Status.TokenReference.SecretName == "" || clusterIdentity.Status.ExpirationTime == "" {
		err := fmt.Errorf("status not populated")
		identity.logger.Error("In clusterIdentityCRDInteraction %s", err)
		return nil, err
	}

	// Overwrite the existing one with the new identity retrieved
	identity.identityRequest = clusterIdentity
	identity.ExpirationTime, err = time.Parse(Constants.TimeFormat, clusterIdentity.Status.ExpirationTime)
	if err != nil {
		identity.logger.Error("Unable to parse the expiration time into the correct format %s", err)
		return nil, err
	}

	identity.Token, err = identity.ApiClient.GetSecrets(clusterIdentity.Status.TokenReference.SecretName, identity.namespace, clusterIdentity.Status.TokenReference.DataName)
	return identity.Token, err
}

// This function checks if the Identity exists and if it doesn't, always creates it
// It does not check the validity of the token.  This is to address the case where the token may have been compromised.
func (identity *ClusterIdentityCRDClient) CreateOrRenewClusterIdentity() error {
	// creates the in-cluster identity
	key, _ := client.ObjectKeyFromObject(identity.identityRequest)
	apiClient := identity.ApiClient.Client

	err := apiClient.Get(context.Background(), key, identity.identityRequest)
	objectNotExists := err != nil && client.IgnoreNotFound(err) == nil
	identity.logger.Info("Object Found : %v", !objectNotExists)

	// reset err
	err = nil

	// If not found and the delete operator is not selected then create the identity
	if objectNotExists {
		err = apiClient.Create(context.Background(), identity.identityRequest)
	} else if !objectNotExists {
		err = apiClient.Update(context.Background(), identity.identityRequest)

		//reset status
		identity.identityRequest.Status = v1beta1.AzureClusterIdentityRequestStatus{}
		err = apiClient.Status().Update(context.Background(), identity.identityRequest)
	}

	return err
}
