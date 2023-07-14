package v1alpha1

import (
	"fmt"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/api"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/apiserver/query"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/constants"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/informers"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/models/v1alpha1"
	serrors "github.com/Gentleelephant/EnhancementWorkload/pkg/server/errors"
	"github.com/emicklei/go-restful/v3"
	v1alpha12 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseclientset "github.com/openkruise/kruise-api/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

type Handler struct {
	operator v1alpha1.Operator
	list     chan interface{}
}

func NewKruiseHandler(informers informers.InformerFactory, clietset kruiseclientset.Interface, k8sclient kubernetes.Interface) *Handler {
	return &Handler{
		operator: v1alpha1.NewOperator(informers, clietset, k8sclient),
	}
}

func (h *Handler) ListPod(request *restful.Request, response *restful.Response) {
	namespace := request.QueryParameter("namespace")
	name := request.QueryParameter("name")
	resource := request.QueryParameter("resource")

	pods, err := h.operator.ListPods(namespace, resource, name)
	handleResponse(request, response, pods, err)
}

func (h *Handler) ListResource(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	resources := request.PathParameter("resources")
	q := query.ParseQueryParameter(request)

	user := request.HeaderParameter(constants.UserAgent)

	if resources == constants.SidecarSetType {
		labelSelector := q.LabelSelector
		if labelSelector == "" {
			labelSelector = fmt.Sprintf("%s=%s", constants.UserAgent, user)
		} else {
			labelSelector = fmt.Sprintf("%s,%s=%s", labelSelector, constants.UserAgent, user)
		}
		q.LabelSelector = labelSelector
	}

	objs, err := h.operator.List(namespace, resources, q)
	handleResponse(request, response, objs, err)
}

func (h *Handler) GetResource(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	resources := request.PathParameter("resources")
	name := request.PathParameter("name")

	if !h.operator.IsKnownResource(resources) {
		api.HandleBadRequest(response, request, serrors.New("unknown resource type %s", resources))
		return
	}
	obj, err := h.operator.Get(namespace, resources, name)

	deepCopy := obj.DeepCopyObject()
	sidecarset, ok := deepCopy.(*v1alpha12.SidecarSet)
	if ok {
		labels := sidecarset.GetLabels()
		if labels != nil {
			user := request.HeaderParameter(constants.UserAgent)
			if user == "" {
				obj = nil
			}
			labelUser, exist := labels[constants.UserAgent]
			if !exist {
				obj = nil
			} else if labelUser != user {
				klog.Errorf("user [%s] can not get sidecarset %s", user, name)
				obj = nil
				err = errors.NewForbidden(v1alpha12.Resource("sidecarset"), name, serrors.New("user [%s] can not get sidecarset %s", user, name))
			}
		} else {
			obj = nil
		}
	}
	handleResponse(request, response, obj, err)
}

func (h *Handler) CreateResource(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	resources := request.PathParameter("resources")

	if !h.operator.IsKnownResource(resources) {
		api.HandleBadRequest(response, request, serrors.New("unknown resource type %s", resources))
		return
	}

	obj := h.operator.GetObject(resources)
	if err := request.ReadEntity(obj); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if resources == constants.SidecarSetType {
		sidecarSet, ok := obj.(*v1alpha12.SidecarSet)
		if ok {
			labels := sidecarSet.GetLabels()
			if labels == nil {
				labels = make(map[string]string)
			}
			user := request.HeaderParameter(constants.UserAgent)
			labels[constants.UserAgent] = user
			sidecarSet.SetLabels(labels)
		}
		obj = sidecarSet
	}

	if err := h.operator.VerifyResouces(namespace, obj); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	created, err := h.operator.Create(namespace, resources, obj)
	handleResponse(request, response, created, err)
}

func (h *Handler) UpdateResource(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	resources := request.PathParameter("resources")
	name := request.PathParameter("name")

	if !h.operator.IsKnownResource(resources) {
		api.HandleBadRequest(response, request, serrors.New("unknown resource type %s", resources))
		return
	}

	obj := h.operator.GetObject(resources)
	if err := request.ReadEntity(obj); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	if err := h.operator.VerifyResouces(namespace, obj); err != nil {
		api.HandleBadRequest(response, request, err)
		return
	}

	updated, err := h.operator.Update(namespace, resources, name, obj)
	handleResponse(request, response, updated, err)
}

func (h *Handler) DeleteResource(request *restful.Request, response *restful.Response) {
	namespace := request.PathParameter("namespace")
	resources := request.PathParameter("resources")
	name := request.PathParameter("name")

	if !h.operator.IsKnownResource(resources) {
		api.HandleBadRequest(response, request, serrors.New("unknown resource type %s", resources))
		return
	}

	handleResponse(request, response, serrors.None, h.operator.Delete(namespace, resources, name))
}

func handleResponse(req *restful.Request, resp *restful.Response, obj interface{}, err error) {
	if err != nil {
		klog.Error(err)
		if errors.IsNotFound(err) {
			api.HandleNotFound(resp, req, err)
			return
		} else if errors.IsConflict(err) {
			api.HandleConflict(resp, req, err)
			return
		}
		api.HandleBadRequest(resp, req, err)
		return
	}

	_ = resp.WriteEntity(obj)
}
