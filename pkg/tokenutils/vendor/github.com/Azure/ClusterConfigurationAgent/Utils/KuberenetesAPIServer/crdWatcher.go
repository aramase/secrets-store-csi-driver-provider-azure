package KuberenetesAPIServer

import (
	"fmt"

	"k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions/apiextensions/v1beta1"
	utilRuntime "k8s.io/apimachinery/pkg/util/runtime"

	internalinterfaces "k8s.io/apiextensions-apiserver/pkg/client/informers/externalversions/internalinterfaces"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/cache"
)

func (apiClient *KubeAPIServerQueries) WatchCrdResourcesWithFilterOptions(listOptions *metav1.ListOptions, handlers cache.ResourceEventHandlerFuncs) {
	optionsModifier := func(options *metav1.ListOptions) {
		options.LabelSelector = listOptions.LabelSelector
	}

	watchOptions := internalinterfaces.TweakListOptionsFunc(optionsModifier)
	informer := v1beta1.NewFilteredCustomResourceDefinitionInformer(apiClient.crdClientset, 0, cache.Indexers{}, watchOptions)
	stopper := make(chan struct{})
	defer utilRuntime.HandleCrash()

	informer.AddEventHandler(handlers)

	go informer.Run(stopper)
	if !cache.WaitForCacheSync(stopper, informer.HasSynced) {
		utilRuntime.HandleError(fmt.Errorf("timed out waiting for caches to sync"))
		return
	}

	<-stopper
	close(stopper)
}
