package KuberenetesAPIServer

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"

	"github.com/Azure/ClusterConfigurationAgent/Utils/Constants"
	"github.com/Azure/ClusterConfigurationAgent/Utils/Helper"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (apiClient *KubeAPIServerQueries) GetPrivateKeyFromSecret() (interface{}, error) {
	secretName := Helper.GetEnvironmentVar(Constants.SecretName)
	secretNamespace := Helper.GetEnvironmentVar(Constants.SecretNamespace)
	secret, err := apiClient.GetSecrets(secretName, secretNamespace, "privateKey")
	if err != nil {
		return nil, err
	}

	privateKey, err := ParseRsaPrivateKeyFromPemStr([]byte(secret))
	if err != nil {
		log.Printf("Parsing the Private Key failed : error {%v} ; secret's {name, namespace} %s, %s", err, secretName, secretNamespace)
		return nil, err
	}
	return privateKey, err
}

func (apiClient *KubeAPIServerQueries) GetSecretData(name string, namespace string) (map[string][]byte, error) {
	clientset := apiClient.Client
	objectKey := &client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}

	//TODO: Returns err if secret doesn't exist or other error
	secret := &v1.Secret{}
	err := clientset.Get(context.Background(), *objectKey, secret)
	if err != nil {
		log.Printf("Error in the Getting Config {%v}", err)
		return nil, err
	}
	return secret.Data, nil
}

func (apiClient *KubeAPIServerQueries) GetSecretsAsBytes(name string, namespace string, keyName string) ([]byte, error) {
	mapData, err := apiClient.GetSecretData(name, namespace)
	if err != nil {
		log.Printf("Error in getting secret data from config {%v}", err)
		return []byte{}, err
	}
	return mapData[keyName], err
}

func (apiClient *KubeAPIServerQueries) GetSecrets(name string, namespace string, keyName string) (string, error) {
	mapData, err := apiClient.GetSecretData(name, namespace)
	if err != nil {
		log.Printf("Error in getting secret data from config {%v}", err)
		return "", err
	}
	data := string(mapData[keyName])
	return data, nil
}

func (apiClient *KubeAPIServerQueries) PutSecrets(name string, data map[string][]byte) (string, error) {
	return apiClient.PutSecretsToNamespace(name, apiClient.Namespace, data)
}

func (apiClient *KubeAPIServerQueries) PutSecretsToNamespace(name string, namespace string, data map[string][]byte) (string, error) {
	clientset := apiClient.Client
	objectKey := &client.ObjectKey{
		Namespace: namespace,
		Name:      name,
	}

	secret := &v1.Secret{}
	err := clientset.Get(context.Background(), *objectKey, secret)
	objectExists := err == nil || !errors.IsNotFound(err)
	log.Printf("Object Found : %v ObjectKey %v", objectExists, objectKey)

	// reset err
	err = nil
	if objectExists {
		secret.Data = data
		err = clientset.Update(context.Background(), secret)
	} else {
		secret = &v1.Secret{
			ObjectMeta: metaV1.ObjectMeta{
				Name:      name,
				Namespace: namespace,
			},
			Data: data,
		}
		err = clientset.Create(context.Background(), secret)
	}

	return secret.ResourceVersion, err
}

func (apiClient *KubeAPIServerQueries) DeleteSecret(name string) error {
	return apiClient.DeleteSecretFromNamespace(name, apiClient.Namespace)
}

func (apiClient *KubeAPIServerQueries) DeleteSecretFromNamespace(name string, namespace string) error {
	clientset := apiClient.Client
	secret := &v1.Secret{
		ObjectMeta: metaV1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metaV1.TypeMeta{
			Kind: "Secret",
		},
		Type: v1.SecretTypeOpaque,
	}
	err := clientset.Delete(context.Background(), secret)
	secretNotExist := err != nil && client.IgnoreNotFound(err) == nil
	if !secretNotExist && err != nil {
		return err
	}
	return nil
}

func ParseRsaPrivateKeyFromPemStr(privPEM []byte) (*rsa.PrivateKey, error) {
	block, _ := pem.Decode(privPEM)
	if block == nil {
		return nil, fmt.Errorf("failed to parse PEM block containing the key")
	}

	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		return nil, err
	}

	return priv, nil
}
