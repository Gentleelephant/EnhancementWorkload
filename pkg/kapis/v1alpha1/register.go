package v1alpha1

import (
	"github.com/Gentleelephant/EnhancementWorkload/pkg/api"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/apiserver/query"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/apiserver/runtime"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/constants"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/informers"
	serrors "github.com/Gentleelephant/EnhancementWorkload/pkg/server/errors"
	openapi "github.com/emicklei/go-restful-openapi"
	"github.com/emicklei/go-restful/v3"
	"github.com/go-openapi/spec"
	"github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseclientset "github.com/openkruise/kruise-api/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes"
	"net/http"
)

var GroupVersion = schema.GroupVersion{Group: "", Version: "v1alpha1"}

func SwaggerObject(swo *spec.Swagger) {
	swo.Info = &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "UserService",
			Description: "Resource for managing Users",
			Contact: &spec.ContactInfo{
				ContactInfoProps: spec.ContactInfoProps{
					Name:  "john",
					Email: "john@doe.rp",
					URL:   "http://johndoe.org",
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: nil,
				},
			},
			License: &spec.License{
				LicenseProps: spec.LicenseProps{
					Name: "MIT",
					URL:  "http://mit.org",
				},
				VendorExtensible: spec.VendorExtensible{
					Extensions: nil,
				},
			},
			Version: "1.0.0",
		},
	}
	swo.Tags = []spec.Tag{spec.Tag{TagProps: spec.TagProps{
		Name:        "users",
		Description: "Managing users"}}}
}

func AddToContainer(container *restful.Container, informers informers.InformerFactory, clientset kruiseclientset.Interface, k8sclient kubernetes.Interface) error {

	ws := runtime.NewWebService(GroupVersion)
	h := NewKruiseHandler(informers, clientset, k8sclient)

	// list all cloneset/sidecarset in all namespaces
	ws.Route(ws.GET("/{resources}").
		To(h.ListResource).
		Doc("List the clonesets object or sidecarsets object in all namespace").
		Metadata(openapi.KeyOpenAPITags, []string{constants.Common}).
		Param(ws.PathParameter("resources", "known values include clonesets, sidecarsets").Required(true)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. ascending=false").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Writes(api.ListResult{Items: []interface{}{}}).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}))

	//ws.Route(ws.GET("/pod").
	//	To(h.ListPod).
	//	Doc("Get the pod object in the specified namespace").
	//	Metadata(openapi.KeyOpenAPITags, []string{constants.PodType}).
	//	Param(ws.QueryParameter("resource", "known values include cloneset, sidecarset").Required(true)).
	//	Param(ws.QueryParameter("namespace", "name of the namespace").Required(false)).
	//	Param(ws.QueryParameter("name", "name of the pod").Required(true)).
	//	Writes(api.ListResult{Items: []interface{}{}}).
	//	Produces(restful.MIME_JSON).
	//	ReturnsError(http.StatusInternalServerError, api.StatusError, api.ErrorMessage{}).
	//	Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}))

	registerCloneSetApi(ws, h)
	registerSidecarSetApi(ws, h)

	container.Add(ws)
	return nil
}

func registerCloneSetApi(ws *restful.WebService, h *Handler) {

	// list clonesets in a namespaces
	ws.Route(ws.GET("/namespaces/{namespace}/{resources}").
		To(h.ListResource).
		Doc("List the clonesets object in all namespace").
		Metadata(openapi.KeyOpenAPITags, []string{constants.CloneSetType}).
		Param(ws.PathParameter("resources", "known values include sidecarsets").Required(true)).
		Param(ws.QueryParameter(query.ParameterPage, "page").Required(false).DataFormat("page=%d").DefaultValue("page=1")).
		Param(ws.QueryParameter(query.ParameterLimit, "limit").Required(false)).
		Param(ws.QueryParameter(query.ParameterAscending, "sort parameters, e.g. ascending=false").Required(false).DefaultValue("ascending=false")).
		Param(ws.QueryParameter(query.ParameterOrderBy, "sort parameters, e.g. orderBy=createTime")).
		Writes(api.ListResult{Items: []interface{}{}}).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, api.StatusOK, api.ListResult{Items: []interface{}{}}))

	// get clonesets
	ws.Route(ws.GET("/namespaces/{namespace}/{resources}/{name}").
		To(h.GetResource).
		Doc("Get the clonesets object in the specified namespace").
		Metadata(openapi.KeyOpenAPITags, []string{constants.CloneSetType}).
		Param(ws.PathParameter("namespace", "name of the namespace").Required(true)).
		Param(ws.PathParameter("resources", "known values include cloneset, sidecarset").Required(true)).
		Param(ws.PathParameter("name", "name of the cloneset").Required(true)).
		Writes(v1alpha1.CloneSet{}).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.CloneSet{}))

	// create clonesets
	ws.Route(ws.POST("/namespaces/{namespace}/{resources}").
		To(h.CreateResource).
		Doc("create a cloneset").
		Metadata(openapi.KeyOpenAPITags, []string{constants.CloneSetType}).
		Param(ws.PathParameter("namespace", "namespace of the Resource").Required(true)).
		Param(ws.PathParameter("resources", "known values include cloneset").Required(true)).
		Writes(v1alpha1.CloneSet{}).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.CloneSet{}))

	// update clonesets
	ws.Route(ws.PUT("/namespaces/{namespace}/{resources}/{name}").
		To(h.UpdateResource).
		Doc("create a cloneset").
		Metadata(openapi.KeyOpenAPITags, []string{constants.CloneSetType}).
		Param(ws.PathParameter("namespace", "namespace of the Resource").Required(true)).
		Param(ws.PathParameter("resources", "known values include cloneset, sidecarset").Required(true)).
		Param(ws.PathParameter("name", "name of the scaledobject").Required(true)).
		Writes(v1alpha1.CloneSet{}).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.CloneSet{}))

	// delete clonesets
	ws.Route(ws.DELETE("/namespaces/{namespace}/{resources}/{name}").
		To(h.DeleteResource).
		Doc("delete the specified cloneset").
		Metadata(openapi.KeyOpenAPITags, []string{constants.CloneSetType}).
		Param(ws.PathParameter("namespaces", "namespace of sidecarset").Required(true)).
		Param(ws.PathParameter("resources", "known values include cloneset").Required(true)).
		Param(ws.PathParameter(query.ParameterName, "the name of the resource").Required(true)).
		Writes(serrors.None).
		Returns(http.StatusOK, api.StatusOK, serrors.None))
}

func registerSidecarSetApi(ws *restful.WebService, h *Handler) {

	// get sidecarsets
	ws.Route(ws.GET("/{resources}/{name}").
		To(h.GetResource).
		Doc("Get the sidecarset object").
		Metadata(openapi.KeyOpenAPITags, []string{constants.SidecarSetType}).
		Param(ws.PathParameter("resources", "known values include sidecarset").Required(true)).
		Param(ws.PathParameter("name", "name of sidecarset").Required(true)).
		Writes(v1alpha1.SidecarSet{}).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.SidecarSet{}))

	// create sidecarsets
	ws.Route(ws.POST("/{resources}").
		To(h.CreateResource).
		Doc("create a sidecarsets").
		Metadata(openapi.KeyOpenAPITags, []string{constants.SidecarSetType}).
		Param(ws.PathParameter("resources", "known values include sidecarsets").Required(true)).
		Writes(v1alpha1.SidecarSet{}).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.SidecarSet{}))

	// update sidecarsets
	ws.Route(ws.PUT("/{resources}/{name}").
		To(h.UpdateResource).
		Doc("create a sidecarset").
		Metadata(openapi.KeyOpenAPITags, []string{constants.SidecarSetType}).
		Param(ws.PathParameter("resources", "known values include sidecarset").Required(true)).
		Param(ws.PathParameter(query.ParameterName, "name of the sidecarset").Required(true)).
		Writes(v1alpha1.SidecarSet{}).
		Produces(restful.MIME_JSON).
		Returns(http.StatusOK, api.StatusOK, v1alpha1.SidecarSet{}))

	// delete sidecarsets
	ws.Route(ws.DELETE("/{resources}/{name}").
		To(h.DeleteResource).
		Doc("delete the specified sidecarsets").
		Metadata(openapi.KeyOpenAPITags, []string{constants.SidecarSetType}).
		Param(ws.PathParameter("resources", "known values include sidecarset").Required(true)).
		Param(ws.PathParameter(query.ParameterName, "the name of the resource").Required(true)).
		Writes(serrors.None).
		Returns(http.StatusOK, api.StatusOK, serrors.None))
}
