package v1alpha1

import (
	kruisev1alpha1 "github.com/openkruise/kruise-api/apps/v1alpha1"
	kruiseclientset "github.com/openkruise/kruise-api/client/clientset/versioned"
	kruiseinformer "github.com/openkruise/kruise-api/client/informers/externalversions"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"testing"
	"time"
)

func TestName(t *testing.T) {

	cfg := config.GetConfigOrDie()
	clientset := kruiseclientset.NewForConfigOrDie(cfg)
	factory := kruiseinformer.NewSharedInformerFactory(clientset, 0)
	//factory.Apps().V1alpha1().CloneSets().Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
	//	AddFunc: func(obj interface{}) {
	//		t.Log("add cloneSet")
	//	},
	//	UpdateFunc: nil,
	//	DeleteFunc: nil,
	//})
	_, err := factory.ForResource(kruisev1alpha1.GroupVersion.WithResource("clonesets"))
	if err != nil {
		return
	}
	factory.Start(nil)
	factory.WaitForCacheSync(nil)
	t.Log("informer has been synced")
	list, err := factory.Apps().V1alpha1().CloneSets().Lister().CloneSets("").List(labels.NewSelector())
	if err != nil {
		t.Fatal(err)
	}
	for _, cloneSet := range list {
		t.Log("====>", cloneSet.Name)
	}

	time.Sleep(10 * time.Second)
}
