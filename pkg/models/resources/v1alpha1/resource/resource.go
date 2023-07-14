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

package resource

import (
	"errors"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/constants"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/models/resources/v1alpha1/kruise"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/cache"

	"github.com/Gentleelephant/EnhancementWorkload/pkg/api"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/apiserver/query"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/informers"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/models/resources/v1alpha1"
	kruisev1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
)

var ErrResourceNotSupported = errors.New("resource is not supported")

type ResourceGetter struct {
	namespacedResourceGetters map[schema.GroupVersionResource]v1alpha1.Interface
	clusterResourceGetters    map[schema.GroupVersionResource]v1alpha1.Interface
}

func NewResourceGetter(factory informers.InformerFactory, cache cache.Cache) *ResourceGetter {
	namespacedResourceGetters := make(map[schema.GroupVersionResource]v1alpha1.Interface)
	clusterResourceGetters := make(map[schema.GroupVersionResource]v1alpha1.Interface)

	namespacedResourceGetters[kruisev1alpha1.SchemeGroupVersion.WithResource(constants.CloneSetType)] = kruise.NewCloneSetObjectGetter(factory.KruiseInformerFactory())
	clusterResourceGetters[kruisev1alpha1.SchemeGroupVersion.WithResource(constants.SidecarSetType)] = kruise.NewSidecarSetGetter(factory.KruiseInformerFactory())
	return &ResourceGetter{
		namespacedResourceGetters: namespacedResourceGetters,
		clusterResourceGetters:    clusterResourceGetters,
	}
}

// TryResource will retrieve a getter with resource name, it doesn't guarantee find resource with correct group version
// need to refactor this use schema.GroupVersionResource
func (r *ResourceGetter) TryResource(clusterScope bool, resource string) v1alpha1.Interface {
	if clusterScope {
		for k, v := range r.clusterResourceGetters {
			if k.Resource == resource {
				return v
			}
		}
	}
	for k, v := range r.namespacedResourceGetters {
		if k.Resource == resource {
			return v
		}
	}
	return nil
}

func (r *ResourceGetter) Get(resource, namespace, name string) (runtime.Object, error) {
	clusterScope := namespace == ""
	getter := r.TryResource(clusterScope, resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.Get(namespace, name)
}

func (r *ResourceGetter) List(resource, namespace string, query *query.Query) (*api.ListResult, error) {
	clusterScope := namespace == ""
	getter := r.TryResource(clusterScope, resource)
	if getter == nil {
		return nil, ErrResourceNotSupported
	}
	return getter.List(namespace, query)
}
