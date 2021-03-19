package KuberenetesAPIServer

import (
	"context"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type CustomType struct {
	GvrObject       schema.GroupVersionResource
	ApiServerClient *KubeAPIServerQueries
	stopCh          chan struct{}
	actionOnChange  func(interface{}, *schema.GroupVersionResource, EventType)
}

type EventType int

const (
	Update EventType = iota
	Delete
	Create
	//Patch
)

//Rehydrate or delete the previous objects based on namespace filter
func (CustomType *CustomType) List(namespace string, eventType EventType) error {
	resourceClient := CustomType.ApiServerClient.dynamicClient.Resource(CustomType.GvrObject).Namespace(namespace)
	objects, err := resourceClient.List(context.Background(), v1.ListOptions{})
	if err != nil {
		return err
	}
	for _, obj := range objects.Items {
		CustomType.reconciler(obj, eventType)
	}
	return nil
}

func (CustomType *CustomType) Create(namespace string, unstructured *unstructured.Unstructured) error {
	resourceClient := CustomType.ApiServerClient.dynamicClient.Resource(CustomType.GvrObject).Namespace(namespace)
	_, err := resourceClient.Create(context.Background(), unstructured, v1.CreateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (CustomType *CustomType) Update(namespace string, unstructured *unstructured.Unstructured) error {
	resourceClient := CustomType.ApiServerClient.dynamicClient.Resource(CustomType.GvrObject).Namespace(namespace)
	_, err := resourceClient.Update(context.Background(), unstructured, v1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (CustomType *CustomType) Delete(namespace string, name string) error {
	resourceClient := CustomType.ApiServerClient.dynamicClient.Resource(CustomType.GvrObject).Namespace(namespace)
	err := resourceClient.Delete(context.Background(), name, v1.DeleteOptions{})
	if err != nil {
		return err
	}
	return nil
}
