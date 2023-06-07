package controller

import (
	"fmt"
	clientset "github.com/souravbiswassanto/crd/pkg/client/clientset/versioned"
	informer "github.com/souravbiswassanto/crd/pkg/client/informers/externalversions/makecrd.com/v1alpha1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"time"
)

func NewController(
	kubeclientset kubernetes.Interface,
	sampleclientset clientset.Interface,
	deploymentInformer appsinformers.DeploymentInformer,
	crdInformer informer.CrdInformer) *Controller {
	ctrl := &Controller{
		kubeclientset:     kubeclientset,
		sampleclientset:   sampleclientset,
		deploymentsLister: deploymentInformer.Lister(),
		deploymentsSynced: deploymentInformer.Informer().HasSynced,
		crdLister:         crdInformer.Lister(),
		crdSynced:         crdInformer.Informer().HasSynced,
		workQueue:         workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "Crds"),
	}
	log.Println("Setting up eventhandler")
	crdInformer.Informer().AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: ctrl.enqueueCrds,
		UpdateFunc: func(oldObj, newObj interface{}) {
			fmt.Println("helelelelele kjkljd ljlk")
			ctrl.enqueueCrds(newObj)
		},
		DeleteFunc: func(obj interface{}) {
			fmt.Println("Delete function was called")
			ctrl.enqueueCrds(obj)
		},
	})

	return ctrl
}

func (c *Controller) enqueueCrds(obj interface{}) {
	log.Println("enqueueing custom resource")
	fmt.Printf("Hellolllllllllllllllllllllllll ajaira\n")
	//fmt.Println(obj)
	// ekta object theke key generate kore dei workqueue te add korar jonno
	key, err := cache.MetaNamespaceKeyFunc(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return
	}
	log.Println(key, " key is added in enqueuecrds")
	c.workQueue.AddRateLimited(key)
	log.Println(key, " key is added in second enqueuecrds")
}

func (c *Controller) Run(workers int, stopCh <-chan struct{}) error {
	defer utilruntime.HandleCrash()
	defer c.workQueue.ShuttingDown()
	log.Println("Run is started")
	// wait for cache sync eita sob informer der cache sync howa obdhi wait kore
	// mane cache sync korteche informer gula.
	if ok := cache.WaitForCacheSync(stopCh, c.deploymentsSynced, c.crdSynced); !ok {
		return fmt.Errorf("Failed to wait for cache to sync")
	}

	log.Println("Starting Workers")
	for i := 0; i < workers; i++ {
		go wait.Until(c.runWorker, time.Second, stopCh)
	}
	<-stopCh
	log.Println("shutting down workers")
	return nil
}

func (c *Controller) runWorker() {
	for c.ProcessNextItem() {
		log.Println("One item is processed")
	}
}

func (c *Controller) ProcessNextItem() bool {
	obj, shutdown := c.workQueue.Get()
	if shutdown {
		return false
	}
	fmt.Printf("Hello Hunny bunny Printing obj key in ProcessNext Item")
	fmt.Println(obj)
	err := func(obj interface{}) error {
		defer c.workQueue.Done(obj)
		var key string
		var ok bool
		if key, ok = obj.(string); !ok {
			c.workQueue.Forget(obj)
			utilruntime.HandleError(fmt.Errorf("expected string in workqueue but got %#v\n", obj))
			return nil
		}
		if err := c.syncHandler(key); err != nil {
			c.workQueue.AddRateLimited(key)
			return fmt.Errorf("error syncing '%s': %s, requeuing", err.Error())
		}
		c.workQueue.Forget(obj)
		log.Printf("Successfully synched '%s'\n", key)
		return nil
	}(obj)
	if err != nil {
		utilruntime.HandleError(err)
		return true
	}
	return true
}

func buildSlice(s ...string) []string {
	var t []string
	for _, v := range s {
		t = append(t, v)
	}
	return t
}
