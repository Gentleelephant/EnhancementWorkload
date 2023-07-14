package kruise

import (
	"github.com/Gentleelephant/EnhancementWorkload/pkg/api"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/apiserver/query"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/models/resources/v1alpha1"
	kruisev1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseinformer "github.com/openkruise/kruise-api/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/runtime"
	"strings"
)

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

const (
	statusStopped  = "stopped"
	statusRunning  = "running"
	statusUpdating = "updating"
)

type CloneSetObjectGetter struct {
	informer kruiseinformer.SharedInformerFactory
}

func NewCloneSetObjectGetter(informer kruiseinformer.SharedInformerFactory) v1alpha1.Interface {
	return &CloneSetObjectGetter{informer: informer}
}

func (s *CloneSetObjectGetter) Get(namespace, name string) (runtime.Object, error) {
	return s.informer.Apps().V1alpha1().CloneSets().Lister().CloneSets(namespace).Get(name)
}

func (s *CloneSetObjectGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	objs, err := s.informer.Apps().V1alpha1().CloneSets().Lister().CloneSets(namespace).List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}

	return v1alpha1.DefaultList(result, query, s.compare, s.filter), nil
}

func (s *CloneSetObjectGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftObj, ok := left.(*kruisev1alpha1.CloneSet)
	if !ok {
		return false
	}

	rightObj, ok := right.(*kruisev1alpha1.CloneSet)
	if !ok {
		return false
	}

	return v1alpha1.DefaultObjectMetaCompare(leftObj.ObjectMeta, rightObj.ObjectMeta, field)
}

func (s *CloneSetObjectGetter) filter(obj runtime.Object, filter query.Filter) bool {
	scaledObject, ok := obj.(*kruisev1alpha1.CloneSet)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(cloneSetStatus(scaledObject.Status), string(filter.Value)) == 0
	default:
		return v1alpha1.DefaultObjectMetaFilter(scaledObject.ObjectMeta, filter)
	}
}

func cloneSetStatus(status kruisev1alpha1.CloneSetStatus) string {
	if status.Replicas == 0 && status.ReadyReplicas == 0 {
		return statusStopped
	} else if status.ReadyReplicas == status.Replicas {
		return statusRunning
	} else {
		return statusUpdating
	}
}

type SidecarSetObjectGetter struct {
	informer kruiseinformer.SharedInformerFactory
}

func NewSidecarSetGetter(informer kruiseinformer.SharedInformerFactory) v1alpha1.Interface {
	return &SidecarSetObjectGetter{informer: informer}
}

func (s *SidecarSetObjectGetter) Get(namespace, name string) (runtime.Object, error) {
	return s.informer.Apps().V1alpha1().SidecarSets().Lister().Get(name)
}

func (s *SidecarSetObjectGetter) List(namespace string, query *query.Query) (*api.ListResult, error) {
	objs, err := s.informer.Apps().V1alpha1().SidecarSets().Lister().List(query.Selector())
	if err != nil {
		return nil, err
	}

	var result []runtime.Object
	for _, obj := range objs {
		result = append(result, obj)
	}

	return v1alpha1.DefaultList(result, query, s.compare, s.filter), nil
}

func (s *SidecarSetObjectGetter) compare(left runtime.Object, right runtime.Object, field query.Field) bool {
	leftObj, ok := left.(*kruisev1alpha1.SidecarSet)
	if !ok {
		return false
	}

	rightObj, ok := right.(*kruisev1alpha1.SidecarSet)
	if !ok {
		return false
	}

	return v1alpha1.DefaultObjectMetaCompare(leftObj.ObjectMeta, rightObj.ObjectMeta, field)
}

func (s *SidecarSetObjectGetter) filter(obj runtime.Object, filter query.Filter) bool {
	sidecarset, ok := obj.(*kruisev1alpha1.SidecarSet)
	if !ok {
		return false
	}
	switch filter.Field {
	case query.FieldStatus:
		return strings.Compare(sidecarSetStatus(sidecarset.Status), string(filter.Value)) == 0
	default:
		return v1alpha1.DefaultObjectMetaFilter(sidecarset.ObjectMeta, filter)
	}
}

func sidecarSetStatus(status kruisev1alpha1.SidecarSetStatus) string {
	if status.MatchedPods == 0 && status.ReadyPods == 0 {
		return statusStopped
	} else if status.MatchedPods == status.ReadyPods {
		return statusRunning
	} else {
		return statusUpdating
	}
}
