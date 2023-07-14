package apiserver

import (
	"context"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/informers"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/kapis/v1alpha1"
	restfulspec "github.com/emicklei/go-restful-openapi/v2"
	"github.com/emicklei/go-restful/v3"
	kruisev1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseclientset "github.com/openkruise/kruise-api/client/clientset/versioned"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type APIServer struct {
	ServerCount int

	Server *http.Server

	InformerFactory informers.InformerFactory

	KruiseClient kruiseclientset.Interface

	// k8s client
	K8sclient kubernetes.Interface

	Client client.Client
	// webservice container, where all webservice defines
	Container *restful.Container
}

func (s *APIServer) installKruiseAPI() {
	runtime.Must(v1alpha1.AddToContainer(s.Container, s.InformerFactory, s.KruiseClient, s.K8sclient))
}

func (s *APIServer) PrepareRun(stopCh <-chan struct{}) error {
	s.Container = restful.NewContainer()
	s.Container.Router(restful.CurlyRouter{})

	// Add container filter to enable CORS
	cors := restful.CrossOriginResourceSharing{
		ExposeHeaders:  []string{"X-My-Header"},
		AllowedHeaders: []string{"Content-Type", "Accept"},
		AllowedMethods: []string{http.MethodGet, http.MethodPost, http.MethodDelete, http.MethodPut},
		CookiesAllowed: false,
		Container:      s.Container}
	s.Container.Filter(cors.Filter)
	// Add container filter to respond to OPTIONS
	s.Container.Filter(s.Container.OPTIONSFilter)

	s.installKruiseAPI()

	s.Server.Handler = s.Container

	//add openapi
	config := restfulspec.Config{
		WebServices:                   s.Container.RegisteredWebServices(), // you control what services are visible
		APIPath:                       "/apidocs.json",
		PostBuildSwaggerObjectHandler: v1alpha1.SwaggerObject}
	s.Container.Add(restfulspec.NewOpenAPIService(config))
	//OpenAPI

	for _, ws := range s.Container.RegisteredWebServices() {
		routes := ws.Routes()
		for _, route := range routes {
			klog.Infof("Method:%s Path: %s", route.Method, route.Path)
		}
	}

	return nil
}

func (s *APIServer) Run(ctx context.Context) error {
	ctx, cancle := context.WithCancel(ctx)
	defer cancle()

	err := s.waitForResourceSync(ctx)
	if err != nil {
		return err
	}

	go func() {
		<-ctx.Done()
		_ = s.Server.Shutdown(context.Background())
	}()

	s.Server.ListenAndServe()

	return nil
}

func (s *APIServer) waitForResourceSync(ctx context.Context) error {
	klog.V(0).Info("Start cache objects")

	stopCh := ctx.Done()

	informerFactory := s.InformerFactory

	cloneSetInformer, err := informerFactory.KruiseInformerFactory().ForResource(kruisev1alpha1.GroupVersion.WithResource("clonesets"))
	if err != nil {
		return err
	}
	cloneSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

		},
		UpdateFunc: func(oldObj, newObj interface{}) {

		},
		DeleteFunc: func(obj interface{}) {

		},
	})
	sidecarSetInformer, err := informerFactory.KruiseInformerFactory().ForResource(kruisev1alpha1.GroupVersion.WithResource("sidecarsets"))
	if err != nil {
		return err
	}
	sidecarSetInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {

		},
		UpdateFunc: func(oldObj, newObj interface{}) {

		},
		DeleteFunc: func(obj interface{}) {

		},
	})

	s.InformerFactory.Start(stopCh)
	s.InformerFactory.WaitForCacheSync(stopCh)

	klog.V(0).Info("Finished caching objects")
	return nil
}
