package options

import (
	"fmt"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/apiserver"
	"github.com/Gentleelephant/EnhancementWorkload/pkg/informers"
	kruiseclientset "github.com/openkruise/kruise-api/client/clientset/versioned"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

type ServerRunOptions struct {
}

func NewServerRunOptions() *ServerRunOptions {
	return &ServerRunOptions{}
}

func (s *ServerRunOptions) NewApiServer(stopCh <-chan struct{}) (*apiserver.APIServer, error) {
	apiServer := &apiserver.APIServer{}

	server := &http.Server{
		Addr: fmt.Sprintf(":%d", 8080),
	}

	cfg := config.GetConfigOrDie()

	kruiseClientset := kruiseclientset.NewForConfigOrDie(cfg)
	dynamicClient := dynamic.NewForConfigOrDie(cfg)
	kubernetesClient := kubernetes.NewForConfigOrDie(cfg)
	informerFactory := informers.NewInformerFactories(kubernetesClient, kruiseClientset, dynamicClient)

	apiServer.Server = server
	apiServer.List = make(chan interface{}, 12)
	apiServer.InformerFactory = informerFactory
	apiServer.KruiseClient = kruiseClientset
	apiServer.K8sclient = kubernetesClient

	return apiServer, nil

}
