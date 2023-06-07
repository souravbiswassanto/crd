package controller

import (
	"fmt"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/tools/cache"
	"log"
	"time"
)

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
