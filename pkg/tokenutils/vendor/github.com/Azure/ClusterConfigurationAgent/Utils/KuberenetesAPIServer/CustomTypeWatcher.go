package KuberenetesAPIServer

import (
	"time"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/tools/cache"
)

//StopWatch : stops watch for the custom type
func (CustomType *CustomType) StopWatch() {
	close(CustomType.stopCh)
}

func (CustomType *CustomType) Watch(actionOnChange func(interface{}, *schema.GroupVersionResource, EventType)) {
	CustomType.WatchOnNamespace(CustomType.ApiServerClient.Namespace, actionOnChange)
}

// Watch get unstructured Type
// To be used async as should not return and keep watching by default
func (CustomType *CustomType) WatchOnNamespace(namespace string, actionOnChange func(interface{}, *schema.GroupVersionResource, EventType)) {
	CustomType.actionOnChange = actionOnChange

	dynClient := CustomType.ApiServerClient.dynamicClient

	// Re sync after 1 hour to cover the DataPlane connectivity down
	factory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(dynClient, time.Hour, namespace, nil)

	// Finally, create our informer for deployments!
	informer := factory.ForResource(CustomType.GvrObject)

	if CustomType.stopCh == nil {
		CustomType.stopCh = make(chan struct{})
	}

	go CustomType.startWatching(CustomType.stopCh, informer.Informer())
}

func (CustomType *CustomType) reconciler(obj interface{}, action EventType) {
	CustomType.actionOnChange(obj, &CustomType.GvrObject, action)
}

func (CustomType *CustomType) startWatching(stopCh <-chan struct{}, s cache.SharedIndexInformer) {
	handlers := cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			CustomType.reconciler(obj, Create)
		},
		UpdateFunc: func(oldObj, obj interface{}) {
			CustomType.reconciler(obj, Update)
		},
		DeleteFunc: func(obj interface{}) {
			CustomType.reconciler(obj, Delete)
		},
	}

	s.AddEventHandler(handlers)
	CustomType.ApiServerClient.logger.Info("Starting watching %s", CustomType.GvrObject.String())
	s.Run(stopCh)
	CustomType.ApiServerClient.logger.Error("Stopped watching %s", CustomType.GvrObject.String())
}

func (eventType EventType) string() string {
	return [...]string{
		"Update",
		"Delete",
		"Create"}[eventType]
}

func GetType(eventType EventType) string {
	return eventType.string()
}
