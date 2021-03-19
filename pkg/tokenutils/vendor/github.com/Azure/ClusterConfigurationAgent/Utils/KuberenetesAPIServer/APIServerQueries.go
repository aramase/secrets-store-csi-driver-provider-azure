package KuberenetesAPIServer

import (
	_ "bytes"
	"fmt"
	"io/ioutil"
	_ "net/http"
	_ "strconv"
	"strings"
	"time"

	"github.com/Azure/ClusterConfigurationAgent/Utils/Helper"

	"github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/ClusterIdentityRequest/v1beta1"
	settingsv1beta1 "github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/CustomLocationSettings/v1beta1"
	extensionsv1beta1 "github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/ExtensionConfig/v1beta1"
	extensionIdentityv1beta1 "github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/ExtensionIdentity/v1beta1"

	v1beta12 "github.com/Azure/ClusterConfigurationAgent/Utils/CRDModels/GitConfig/v1beta1"

	"github.com/Azure/ClusterConfigurationAgent/LogHelper"
	"golang.org/x/net/context"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	_ "k8s.io/api/rbac/v1beta1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	_ "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	_ "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	_ "k8s.io/client-go/tools/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"

	apiextensionv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"

	apiextensionClientSet "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes/scheme"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"

	_ "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type KubeAPIServerQueries struct {
	Clientset        kubernetes.Interface
	crdClientset     apiextensionClientSet.Interface
	podLogsClientset *corev1.CoreV1Client
	KubeConfig       *rest.Config
	Client           client.Client
	dynamicClient    dynamic.Interface
	logger           LogHelper.LogWriter
	IsClientMocked   bool
	Namespace        string
}

func (config *KubeAPIServerQueries) InitializeClient(
	namespace string,
	logger LogHelper.LogWriter,
	isClusterScoped bool) error {

	return config.initializeClientInternal(namespace, logger, nil, nil, nil, isClusterScoped, nil)
}

func (config *KubeAPIServerQueries) InitializeClientWithConfig(
	namespace string,
	logger LogHelper.LogWriter,
	isClusterScoped bool,
	configFile string) error {

	return config.initializeClientInternal(namespace, logger, nil, nil, nil, isClusterScoped, &configFile)
}

func (config *KubeAPIServerQueries) InitializeTestClient(
	namespace string,
	logger LogHelper.LogWriter,
	isClusterScoped bool,
	definedClient client.Client,
	definedDynamicClient dynamic.Interface,
	definedCrdClient apiextensionClientSet.Interface) error {

	return config.initializeClientInternal(namespace, logger, definedClient, definedDynamicClient, definedCrdClient, isClusterScoped, nil)
}

func (config *KubeAPIServerQueries) initializeClientInternal(
	namespace string,
	logger LogHelper.LogWriter,
	definedClient client.Client,
	definedDynamicClient dynamic.Interface,
	definedCrdClient apiextensionClientSet.Interface,
	isClusterScoped bool,
	configFile *string) error {

	var err error

	if definedClient != nil || definedDynamicClient != nil || definedCrdClient != nil {
		config.Client = definedClient
		config.IsClientMocked = true
		config.crdClientset = definedCrdClient
		config.logger = logger
		config.dynamicClient = definedDynamicClient
		return nil
	}

	if !isClusterScoped {
		config.Namespace = namespace
	}

	if config.IsClientMocked {
		return nil
	}

	config.logger = logger

	if configFile == nil {
		config.KubeConfig, err = rest.InClusterConfig()
	} else {
		config.KubeConfig, err = clientcmd.BuildConfigFromFlags("", *configFile)
	}
	if err != nil {
		config.logger.Error("Unable to get the API Server Address : {%v}", err)
		return err
	}

	// creates the clientset

	config.Clientset, err = kubernetes.NewForConfig(config.KubeConfig)
	if err != nil {
		config.logger.Error("Unable to create the clientSet : {%v}", err)
		return err
	}

	config.dynamicClient, err = dynamic.NewForConfig(config.KubeConfig)
	if err != nil {
		config.logger.Error("Unable to create the dyanmic Client : {%v}", err)
		return err

	}
	schemeOverride := scheme.Scheme
	err = v1.AddToScheme(schemeOverride)
	err = appsv1.AddToScheme(schemeOverride)
	err = rbacv1.AddToScheme(schemeOverride)
	err = apiextensionv1beta1.AddToScheme(schemeOverride)
	err = apiextensionv1.AddToScheme(schemeOverride)
	err = v1beta1.AddToScheme(schemeOverride)
	err = v1beta12.AddToScheme(schemeOverride)
	err = batchv1.AddToScheme(schemeOverride)
	if Helper.IsExtensionsEnabled() {
		err = settingsv1beta1.AddToScheme(schemeOverride)
		err = extensionsv1beta1.AddToScheme(schemeOverride)
		err = extensionIdentityv1beta1.AddToScheme(schemeOverride)
	}

	if err != nil {
		config.logger.Error("Initialize client set scheme : {%v}", err)
		return err
	}

	config.Client, err = client.New(config.KubeConfig, client.Options{Scheme: schemeOverride})
	if err != nil {
		config.logger.Error("Unable to create the clientSet : {%v}", err)
		return err
	}

	config.crdClientset, err = apiextensionClientSet.NewForConfig(config.KubeConfig)
	if err != nil {
		config.logger.Error("Unable to get the API Server Address : {%v}", err)
		return err
	}

	config.podLogsClientset, err = corev1.NewForConfig(config.KubeConfig)
	if err != nil {
		config.logger.Error("Unable to create the clientSet : {%v}", err)
		return err
	}

	return nil
}

func (config *KubeAPIServerQueries) WatchForServiceAccountToken(AccountName string,
	timeoutSeconds time.Duration) bool {

	to := int64(timeoutSeconds.Seconds())
	watcher, err := config.Clientset.CoreV1().ServiceAccounts(config.Namespace).Watch(context.Background(), metav1.ListOptions{
		FieldSelector:  fmt.Sprintf("metadata.name=%s", AccountName),
		TimeoutSeconds: &to,
	})

	if err != nil {
		return false
	}

	for event := range watcher.ResultChan() {
		p, ok := event.Object.(*v1.ServiceAccount)
		if !ok {
			// Invalid secret found
			watcher.Stop()
			return false
		}
		if len(p.Secrets) >= 1 {
			watcher.Stop()
			return true
		}
	}
	config.logger.Info("No valid secret found")
	watcher.Stop()
	return false
}

//Waits for Secret to be give data with identity 10
func (config *KubeAPIServerQueries) WatchForSecrets(SecretName string, timeoutSeconds time.Duration) watch.Interface {
	to := int64(timeoutSeconds.Seconds())
	watcher, err := config.Clientset.CoreV1().Secrets(config.Namespace).Watch(context.Background(), metav1.ListOptions{
		FieldSelector:  fmt.Sprintf("metadata.name=%s", SecretName),
		TimeoutSeconds: &to,
	})

	if err != nil {
		return nil
	}

	return watcher
}

func (config *KubeAPIServerQueries) WatchForDataInSecret(watcher watch.Interface, key string) bool {
	for event := range watcher.ResultChan() {
		p, ok := event.Object.(*v1.Secret)
		if !ok {
			// Invalid secret found
			return false
		}
		if len(p.Data) >= 1 {
			_, ok = p.Data[key]
			watcher.Stop()
			config.logger.Info("Valid %s create : %v", key, ok)
			return ok
		}
	}
	config.logger.Info("No Valid %s found", key)
	return false
}

func (config *KubeAPIServerQueries) GetResourceName(CRDName string) (string, error) {
	crdObject, err := config.crdClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(context.Background(), CRDName, metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	name := crdObject.Status.AcceptedNames.Plural
	if name == "" {
		name = crdObject.Spec.Names.Plural
	}
	return name, nil
}

func (config *KubeAPIServerQueries) CRDExists(CRDName string) (bool, error) {
	_, err := config.crdClientset.ApiextensionsV1beta1().CustomResourceDefinitions().Get(context.Background(), CRDName, metav1.GetOptions{})

	if err != nil {
		if errors.IsNotFound(err) == false {
			config.logger.Error("Unable to get CRD %s : {%v}", CRDName, err)
			return false, err
		}
		return false, nil
	}

	return true, nil
}

func (config *KubeAPIServerQueries) GetPodsWithSelector(selector string) (*v1.PodList, error) {
	pods, err := config.Clientset.CoreV1().Pods(config.Namespace).List(context.Background(), metav1.ListOptions{
		LabelSelector: selector,
	})

	if err != nil {
		config.logger.Error("Unable to get the pods from the client : {%v}", err)
		return nil, err
	}

	config.logger.Info("There are %d pods in the cluster\n", len(pods.Items))
	return pods, nil
}

func (config *KubeAPIServerQueries) DeletePodWithSelector(selector string) error {
	filteredPods, err := config.GetPodsWithSelector(selector)
	if err != nil {
		config.logger.Error("Unable to get the pods from the client : {%v}", err)
		return err
	}

	for _, pod := range filteredPods.Items {
		err = config.Client.Delete(context.Background(), &pod)
		if err != nil {
			config.logger.Error("Unable delete the pods from the client : {%v}", err)
		}
	}
	return err
}

func (config *KubeAPIServerQueries) GetLogs(selector string, logOptions *v1.PodLogOptions) (string, error) {
	pods := config.podLogsClientset.Pods(config.Namespace)

	filteredPods, err := config.GetPodsWithSelector(selector)

	if err != nil {
		config.logger.Error("Unable to get the pods from the client : {%v}", err)
		return "", err
	}

	// Assumes only 1 pod for the selector per namespace
	if len(filteredPods.Items) != 1 {
		err := fmt.Errorf("no or more than 1 flux operator per namespace is not supported")
		config.logger.Error("no or more than 1 flux operator per namespace is not supported : "+
			"number of pods : {%v} ; selector : {%v}", len(filteredPods.Items), selector)
		return "", err
	}

	tailCount := int64(50)

	podlogOptions := logOptions
	if podlogOptions == nil {
		podlogOptions = &v1.PodLogOptions{
			TailLines: &tailCount,
		}
	}

	// get the logs for the pod
	logRequest := pods.GetLogs(filteredPods.Items[0].ObjectMeta.Name, podlogOptions)

	// Do a 5 minute time out on the log request
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute*5)

	response, err := logRequest.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer response.Close()

	logs, err := ioutil.ReadAll(response)
	if err != nil {
		return "", err
	}
	cancelFunc()

	return string(logs), nil
	//(pods.Items[0].ObjectMeta.Name, )
}

func (config *KubeAPIServerQueries) CreateJobAfterRenaming(yamlFile, resourceName string) (runtime.Object, error) {
	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(yamlFile), nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Error while decoding YAML object. yamlFile: %s Err was: %s", yamlFile, err)
		config.logger.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	kind, _ := meta.NewAccessor().Kind(obj)

	newObject := &batchv1.Job{}
	if strings.ToLower(kind) == "job" {

		newObject = obj.(*batchv1.Job)
		newObject.Name = resourceName
	} else {
		return nil, fmt.Errorf("Incorrect yaml provided")
	}

	obj, err = config.applyObject(obj, false, false, false, false)
	if err != nil {
		config.logger.Error("Unable to deploy the yaml : %s with Error: {%v}", yamlFile, err)
		return nil, err
	}

	return obj, err
}

func (config *KubeAPIServerQueries) CreateResource(yamlFile string) (runtime.Object, error) {
	return config.applyYamlInternal(yamlFile, false, false, false, false)
}
func (config *KubeAPIServerQueries) DeleteResource(yamlFile string, deleteDependent bool) (runtime.Object, error) {
	return config.applyYamlInternal(yamlFile, true, true, deleteDependent, false)
}
func (config *KubeAPIServerQueries) DeleteCRD(yamlFile string, deleteDependent bool) (runtime.Object, error) {
	return config.applyYamlInternal(yamlFile, true, true, deleteDependent, true)
}
func (config *KubeAPIServerQueries) CreateOrUpdateResource(yamlFile string) (runtime.Object, error) {
	return config.applyYamlInternal(yamlFile, false, true, false, false)
}

func (config *KubeAPIServerQueries) applyYamlInternal(yamlFile string,
	deleteResource bool,
	forceOverWrite bool,
	deleteDependent bool,
	deleteCRD bool) (runtime.Object, error) {
	var err error

	obj, _, err := scheme.Codecs.UniversalDeserializer().Decode([]byte(yamlFile), nil, nil)
	//obj, _, err := decoder([]byte(yamlFile), nil, nil)
	if err != nil {
		msg := fmt.Sprintf("Error while decoding YAML object. yamlFile: %s Err was: %s", yamlFile, err)
		config.logger.Error(msg)
		return nil, fmt.Errorf(msg)
	}

	obj, err = config.applyObject(obj, deleteResource, forceOverWrite, deleteDependent, deleteCRD)
	if err != nil {
		config.logger.Error("Unable to deploy the yaml : %s with Error: {%v}", yamlFile, err)
		return nil, err

	}
	return obj, err
}

func (config *KubeAPIServerQueries) applyObject(obj runtime.Object,
	deleteResource bool,
	forceOverWrite bool,
	deleteDependent bool,
	deleteCRD bool) (runtime.Object, error) {

	kind, _ := meta.NewAccessor().Kind(obj)
	objectKey, _ := client.ObjectKeyFromObject(obj)
	retrievedObject := obj.DeepCopyObject()

	err := config.Client.Get(context.Background(), objectKey, retrievedObject)
	objectNotExists := err != nil && client.IgnoreNotFound(err) == nil
	config.logger.Info("Object Found : %v ObjectKey %v", !objectNotExists, objectKey)

	if !objectNotExists && (!forceOverWrite && !deleteResource) {
		return retrievedObject, &AlreadyExistError{}
	}

	// reset err
	err = nil
	if objectNotExists && !deleteResource {
		err = config.Client.Create(context.Background(), obj)
	} else if !deleteResource {

		// for service updates are more complex and needs to be patched to get the ClusterIP etc.
		if strings.ToLower(kind) == "service" {

			newObject := obj.(*v1.Service)
			retrievedObject := retrievedObject.(*v1.Service)
			newObject.Spec.ClusterIP = retrievedObject.Spec.ClusterIP
			newObject.ResourceVersion = retrievedObject.ResourceVersion
			err = config.Client.Patch(context.Background(), obj, client.MergeFrom(retrievedObject))
		} else if strings.ToLower(kind) == "customresourcedefinition" {

			// Handle CRD updates
			newObject := obj.(*apiextensionv1beta1.CustomResourceDefinition)
			retrievedObject := retrievedObject.(*apiextensionv1beta1.CustomResourceDefinition)
			newObject.ResourceVersion = retrievedObject.ResourceVersion
			err = config.Client.Patch(context.Background(), obj, client.MergeFrom(retrievedObject))
		} else {
			err = config.Client.Update(context.Background(), obj)
		}
	} else if !objectNotExists && deleteResource {

		// No need to delete the CRD as other operators may be using it.
		// Todo [Anubhav] : Mark CRD with helms hook for cleanup on helm uninstall
		if strings.ToLower(kind) != "customresourcedefinition" || deleteCRD {
			if deleteDependent {
				policy := metav1.DeletePropagationBackground
				deleteOption := &client.DeleteOptions{PropagationPolicy: &policy}
				err = config.Client.Delete(context.Background(), obj, deleteOption)
			} else {
				//Not using any delete option, because default is depend on the resource type. For thr flux we dont want to delete the dependencies
				err = config.Client.Delete(context.Background(), obj)
			}

		}
	}

	return obj, err
}

func (config *KubeAPIServerQueries) CreateOrUpdateObject(obj runtime.Object) error {
	err := config.Client.Create(context.Background(), obj)
	if errors.IsAlreadyExists(err) {
		err = config.Client.Update(context.Background(), obj)
	}
	return err
}

func (config *KubeAPIServerQueries) DeepCopy() *KubeAPIServerQueries {
	return &KubeAPIServerQueries{
		Clientset:        config.Clientset,
		podLogsClientset: config.podLogsClientset,
		crdClientset:     config.crdClientset,
		KubeConfig:       config.KubeConfig,
		Client:           config.Client,
		dynamicClient:    config.dynamicClient,
		logger:           config.logger,
		IsClientMocked:   config.IsClientMocked,
		Namespace:        "",
	}
}
