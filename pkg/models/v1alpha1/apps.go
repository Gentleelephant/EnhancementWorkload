package v1alpha1

import (
	"context"
	"fmt"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/api"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/apiserver/query"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/constants"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/informers"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/models/resources/v1alpha1/resource"
	"github.com/duke-git/lancet/v2/slice"
	"github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseclientset "github.com/openkruise/kruise-api/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
)

type Operator interface {
	List(namespace, resource string, query *query.Query) (*api.ListResult, error)
	Get(namespace, resource, name string) (runtime.Object, error)
	Create(namespace, resource string, obj runtime.Object) (runtime.Object, error)
	Update(namespace, resource, name string, obj runtime.Object) (runtime.Object, error)
	Delete(namespace, resource, name string) error
	ListPods(namespace, resource, name string) (*api.ListResult, error)

	VerifyResouces(namespace string, obj runtime.Object) error
	IsKnownResource(resource string) bool
	GetObject(resource string) runtime.Object
}

type operator struct {
	kubernetesclientset kubernetes.Interface
	kruiseclientset     kruiseclientset.Interface
	resourceGetter      *resource.ResourceGetter
}

func (c *operator) ListPods(namespace, resource, name string) (*api.ListResult, error) {

	if !slice.Contain([]string{constants.SidecarSetType, constants.CloneSetType}, resource) {
		return nil, errors.NewBadRequest("resource type is not supported")
	}

	workload, err := c.resourceGetter.Get(resource, namespace, name)
	if err != nil {
		return nil, err
	}

	switch resource {
	case constants.CloneSetType:
		cloneSet := workload.(*v1alpha1.CloneSet)
		selector := cloneSet.Spec.Selector
		matchLabels, err := v1.LabelSelectorAsMap(selector)
		if err != nil {
			return nil, err
		}
		return c.listPods(namespace, matchLabels)
	case constants.SidecarSetType:
		sidecarSet := workload.(*v1alpha1.SidecarSet)
		selector := sidecarSet.Spec.Selector
		matchLabels, err := v1.LabelSelectorAsMap(selector)
		if err != nil {
			return nil, err
		}
		return c.listPods(namespace, matchLabels)
	default:
		return nil, errors.NewInternalError(nil)
	}
}

func (c *operator) listPods(namespace string, matchLabels map[string]string) (*api.ListResult, error) {
	podList, err := c.kubernetesclientset.CoreV1().Pods(namespace).List(context.Background(), v1.ListOptions{LabelSelector: labels.Set(matchLabels).String()})
	if err != nil {
		return nil, err
	}

	if len(podList.Items) == 0 {
		return &api.ListResult{
			Items:      []interface{}{},
			TotalItems: 0,
		}, nil
	}

	var objs []interface{}
	for _, item := range podList.Items {
		objs = append(objs, &item)
	}
	return &api.ListResult{
		Items:      objs,
		TotalItems: len(podList.Items),
	}, err
}

func (c *operator) List(namespace, resource string, query *query.Query) (*api.ListResult, error) {
	return c.resourceGetter.List(resource, namespace, query)
}

func (c *operator) Get(namespace, resource, name string) (runtime.Object, error) {
	obj, err := c.resourceGetter.Get(resource, namespace, name)
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (c *operator) Create(namespace, resource string, obj runtime.Object) (runtime.Object, error) {
	switch resource {
	case constants.CloneSetType:
		cloneset, ok := obj.(*v1alpha1.CloneSet)
		if !ok {
			return nil, fmt.Errorf("object is not a CloneSet")
		}
		return c.kruiseclientset.AppsV1alpha1().CloneSets(namespace).Create(context.Background(), cloneset, v1.CreateOptions{})
	case constants.SidecarSetType:
		sidecarset, ok := obj.(*v1alpha1.SidecarSet)
		if !ok {
			return nil, fmt.Errorf("object is not a SidecarSet")
		}
		return c.kruiseclientset.AppsV1alpha1().SidecarSets().Create(context.Background(), sidecarset, v1.CreateOptions{})
	default:
		return nil, errors.NewInternalError(nil)
	}
}

func (c *operator) Update(namespace, resource, name string, obj runtime.Object) (runtime.Object, error) {
	old, err := c.resourceGetter.Get(resource, namespace, name)
	if err != nil {
		return nil, err
	}

	switch resource {
	case constants.CloneSetTag:
		oldScaledObject := old.(*v1alpha1.CloneSet)
		newCloneset, ok := obj.(*v1alpha1.CloneSet)
		if !ok {
			return nil, fmt.Errorf("object is not a CloneSet")
		}
		newCloneset.SetResourceVersion(oldScaledObject.ResourceVersion)
		return c.kruiseclientset.AppsV1alpha1().CloneSets(namespace).Update(context.Background(), newCloneset, v1.UpdateOptions{})
	case constants.SidecarSetType:
		oldScaledJob := old.(*v1alpha1.SidecarSet)
		NewScaledJob := obj.(*v1alpha1.SidecarSet)
		NewScaledJob.SetResourceVersion(oldScaledJob.ResourceVersion)
		return c.kruiseclientset.AppsV1alpha1().SidecarSets().Update(context.Background(), NewScaledJob, v1.UpdateOptions{})
	default:
		return nil, errors.NewInternalError(nil)
	}
}

func (c *operator) Delete(namespace, resource, name string) error {

	switch resource {
	case constants.CloneSetType:
		return c.kruiseclientset.AppsV1alpha1().CloneSets(namespace).Delete(context.Background(), name, v1.DeleteOptions{})
	case constants.SidecarSetType:
		return c.kruiseclientset.AppsV1alpha1().SidecarSets().Delete(context.Background(), name, v1.DeleteOptions{})
	default:
		return errors.NewInternalError(nil)
	}

}

func (c *operator) VerifyResouces(namespace string, obj runtime.Object) error {
	return nil
}

func (c *operator) IsKnownResource(resource string) bool {
	if c.GetObject(resource) == nil {
		return false
	}
	return true
}

func (c *operator) GetObject(resource string) runtime.Object {
	switch resource {
	case constants.CloneSetType:
		return &v1alpha1.CloneSet{}
	case constants.SidecarSetType:
		return &v1alpha1.SidecarSet{}
	default:
		return nil
	}
}

func NewOperator(informers informers.InformerFactory, clientset kruiseclientset.Interface, k8sclient kubernetes.Interface) Operator {
	return &operator{
		kruiseclientset:     clientset,
		kubernetesclientset: k8sclient,
		resourceGetter:      resource.NewResourceGetter(informers, nil),
	}
}
