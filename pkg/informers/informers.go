/*
Copyright 2023 The KubeSphere Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package informers

import (
	kruiseclientset "github.com/openkruise/kruise-api/client/clientset/versioned"
	kruiseinformer "github.com/openkruise/kruise-api/client/informers/externalversions"
	"k8s.io/client-go/kubernetes"
	"reflect"
	"time"

	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	k8sinformers "k8s.io/client-go/informers"
)

// default re-sync period for all informer factories
const defaultResync = 600 * time.Second

// InformerFactory is a group all shared informer factories which kubesphere needed
// callers should check if the return value is nil
type InformerFactory interface {
	KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory
	//ApiExtensionSharedInformerFactory() apiextensionsinformers.SharedInformerFactory
	DynamicSharedInformerFactory() dynamicinformer.DynamicSharedInformerFactory
	KruiseInformerFactory() kruiseinformer.SharedInformerFactory
	// Start shared informer factory one by one if they are not nil
	Start(stopCh <-chan struct{})
	WaitForCacheSync(ch <-chan struct{})
}

type GenericInformerFactory interface {
	Start(stopCh <-chan struct{})
	WaitForCacheSync(stopCh <-chan struct{}) map[reflect.Type]bool
}

type informerFactories struct {
	kubernetesShareInformerFactory k8sinformers.SharedInformerFactory
	//apiextensionsInformerFactory   apiextensionsinformers.SharedInformerFactory
	dynamicInformerFactory dynamicinformer.DynamicSharedInformerFactory
	kruiseInformerFactory  kruiseinformer.SharedInformerFactory
}

func (f *informerFactories) KruiseInformerFactory() kruiseinformer.SharedInformerFactory {
	return f.kruiseInformerFactory
}

func NewInformerFactories(
	client kubernetes.Interface,
	kruiseClient kruiseclientset.Interface,
	dynamicClient dynamic.Interface) InformerFactory {
	factory := &informerFactories{}

	if client != nil {
		factory.kubernetesShareInformerFactory = k8sinformers.NewSharedInformerFactory(client, defaultResync)
	}
	//
	//if apiextensionsClient != nil {
	//	factory.apiextensionsInformerFactory = apiextensionsinformers.NewSharedInformerFactory(apiextensionsClient, defaultResync)
	//}

	if dynamicClient != nil {
		factory.dynamicInformerFactory = dynamicinformer.NewDynamicSharedInformerFactory(dynamicClient, defaultResync)
	}

	if kruiseClient != nil {
		factory.kruiseInformerFactory = kruiseinformer.NewSharedInformerFactory(kruiseClient, defaultResync)
	}

	return factory
}

func (f *informerFactories) DynamicSharedInformerFactory() dynamicinformer.DynamicSharedInformerFactory {
	return f.dynamicInformerFactory
}

func (f *informerFactories) KubernetesSharedInformerFactory() k8sinformers.SharedInformerFactory {
	return f.kubernetesShareInformerFactory
}

//
//func (f *informerFactories) ApiExtensionSharedInformerFactory() apiextensionsinformers.SharedInformerFactory {
//	return f.apiextensionsInformerFactory
//}

func (f *informerFactories) Start(stopCh <-chan struct{}) {
	if f.kubernetesShareInformerFactory != nil {
		f.kubernetesShareInformerFactory.Start(stopCh)
	}
	//
	//if f.apiextensionsInformerFactory != nil {
	//	f.apiextensionsInformerFactory.Start(stopCh)
	//}

	if f.dynamicInformerFactory != nil {
		f.dynamicInformerFactory.Start(stopCh)
	}

	if f.kruiseInformerFactory != nil {
		f.kruiseInformerFactory.Start(stopCh)
	}
}

func (f *informerFactories) WaitForCacheSync(stopCh <-chan struct{}) {
	if f.kubernetesShareInformerFactory != nil {
		f.kubernetesShareInformerFactory.WaitForCacheSync(stopCh)
	}
	//
	//if f.apiextensionsInformerFactory != nil {
	//	f.apiextensionsInformerFactory.WaitForCacheSync(stopCh)
	//}

	if f.dynamicInformerFactory != nil {
		f.dynamicInformerFactory.WaitForCacheSync(stopCh)
	}

	if f.kruiseInformerFactory != nil {
		f.kruiseInformerFactory.WaitForCacheSync(stopCh)
	}
}
