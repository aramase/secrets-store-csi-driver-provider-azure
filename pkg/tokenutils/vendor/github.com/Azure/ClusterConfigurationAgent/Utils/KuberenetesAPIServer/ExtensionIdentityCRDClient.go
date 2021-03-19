package KuberenetesAPIServer

import (
	"context"

	"github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/ExtensionIdentity/v1beta1"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ExtensionIdentityCRDClient struct {
	ApiClient *KubeAPIServerQueries
	crd       *v1beta1.AzureExtensionIdentity
}

func (this *ExtensionIdentityCRDClient) Init(apiClient *KubeAPIServerQueries, extensionName string) error {
	this.ApiClient = apiClient
	this.crd = &v1beta1.AzureExtensionIdentity{
		ObjectMeta: metav1.ObjectMeta{
			Name:      extensionName,
			Namespace: Constants.ManagementNamespace,
		},
		Spec: v1beta1.AzureExtensionIdentitySpec{
			ServiceAccounts: nil,
			TokenNamespace:  Constants.ManagementNamespace,
		},
	}
	objectKey, _ := client.ObjectKeyFromObject(this.crd)
	err := this.ApiClient.Client.Get(context.Background(), objectKey, this.crd)
	return err
}

func (this *ExtensionIdentityCRDClient) AddAnnotationsAndLabel(annotation map[string]string , labels map[string]string) {
	currAnnotations := this.crd.GetAnnotations()
	currAnnotations = mergeMaps(annotation, currAnnotations)
	this.crd.SetAnnotations(currAnnotations)

	currentLabels := this.crd.GetLabels()
	currentLabels = mergeMaps(labels, currentLabels)
	this.crd.SetLabels(currentLabels)
}

func mergeMaps(mergingMap map[string]string, initialMap map[string]string) map[string]string {
	if initialMap == nil {
		initialMap = map[string]string{}
	}
	for k, v := range mergingMap {
		if _, isPresent := initialMap[k]; !isPresent {
			initialMap[k] = v
		}
	}
	return initialMap
}

func (this *ExtensionIdentityCRDClient) GetTokenNamespace() string {
	return this.crd.Spec.TokenNamespace
}

func (this *ExtensionIdentityCRDClient) Create(listServiceAccount []v1beta1.ServiceAccount) error {
	this.crd.Spec.ServiceAccounts = listServiceAccount
	return this.ApiClient.Client.Create(context.Background(), this.crd)
}

func (this *ExtensionIdentityCRDClient) Delete() error {
	return this.ApiClient.Client.Delete(context.Background(), this.crd)
}

func (this *ExtensionIdentityCRDClient) AddAccess(listServiceAccount []v1beta1.ServiceAccount) error {
	// Update with the newer one
	if this.crd.Spec.AddServiceAccount(listServiceAccount) {
		return this.ApiClient.Client.Update(context.Background(), this.crd)
	}
	return nil
}

func (this *ExtensionIdentityCRDClient) GiveAccessToSecret(secretName string, secretNamespace string) error {

	objectKey := metav1.ObjectMeta{
		Name:      secretName,
		Namespace: secretNamespace,
	}
	role := rbacv1.Role{
		ObjectMeta: objectKey,
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{Constants.GetVerbs, Constants.WatchVerbs, Constants.ListVerbs},
				APIGroups:     []string{""},
				Resources:     []string{Constants.SecretKindName},
				ResourceNames: []string{secretName},
			},
		},
	}
	err := this.ApiClient.CreateOrUpdateObject(&role)
	if err != nil {
		return err
	}
	rolebinding := rbacv1.RoleBinding{
		ObjectMeta: objectKey,
		RoleRef: rbacv1.RoleRef{
			APIGroup: rbacv1.GroupName,
			Kind:     Constants.RoleKind,
			Name:     role.Name,
		},
		Subjects: this.crd.Spec.GetServiceAccounts(),
	}
	err = this.ApiClient.CreateOrUpdateObject(&rolebinding)
	if err != nil {
		return err
	}
	return nil
}
