package controller

import (
	"context"
	"fmt"
	controllerv1 "github.com/souravbiswassanto/crd/pkg/apis/makecrd.com/v1alpha1"
	clientset "github.com/souravbiswassanto/crd/pkg/client/clientset/versioned"
	informer "github.com/souravbiswassanto/crd/pkg/client/informers/externalversions/makecrd.com/v1alpha1"
	lister "github.com/souravbiswassanto/crd/pkg/client/listers/makecrd.com/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	appsinformers "k8s.io/client-go/informers/apps/v1"
	appslisters "k8s.io/client-go/listers/apps/v1"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"time"
)

type Controller struct {
	kubeclientset   kubernetes.Interface
	sampleclientset clientset.Interface

	deploymentsLister appslisters.DeploymentLister
	deploymentsSynced cache.InformerSynced

	crdLister lister.CrdLister
	crdSynced cache.InformerSynced

	workQueue workqueue.RateLimitingInterface
}

// Newcontroller returns a new sapmple controller

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

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s", key))
		return nil
	}
	// Get the Crd resource with this namespace/name
	crd, err := c.crdLister.Crds(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("foo '%s' in workqueue no logner exists", key))
			return nil
		}
		return err
	}
	if err := c.DeploymentHandler(crd); err != nil {
		utilruntime.HandleError(fmt.Errorf("error while handling deployment: %s", err.Error()))
	}
	return nil
}

func (c *Controller) DeploymentHandler(crd *controllerv1.Crd) error {
	deploymentName := crd.Spec.Name
	if deploymentName == "" {
		utilruntime.HandleError(fmt.Errorf(" deployment name must be specified"))
		return nil
	}
	namespace := crd.Namespace
	//name := crd.Name
	deployment, err := c.deploymentsLister.Deployments(namespace).Get(deploymentName)
	if errors.IsNotFound(err) {
		deployment, err = c.kubeclientset.AppsV1().Deployments(namespace).Create(context.TODO(), c.newDeployment(crd), metav1.CreateOptions{})
		if err != nil {
			return err
		}
	}
	if err != nil {
		return err
	}
	if crd.Spec.Replicas != nil && *crd.Spec.Replicas != *deployment.Spec.Replicas {
		*deployment.Spec.Replicas = *crd.Spec.Replicas
		deployment, err = c.kubeclientset.AppsV1().Deployments(namespace).Update(context.TODO(), deployment, metav1.UpdateOptions{})
		if err != nil {
			return err
		}
	}
	log.Println("Updating status")
	err = c.updateCrdStatus(crd, deployment)
	if err != nil {
		fmt.Println("Sourav")
		fmt.Println(err.Error())
	}
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) newDeployment(crd *controllerv1.Crd) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      crd.Spec.Name,
			Namespace: crd.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: crd.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": "my-app",
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": "my-app",
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "my-app",
							Image: crd.Spec.Container.Image,
							Ports: []corev1.ContainerPort{
								{
									Name:          "http",
									Protocol:      corev1.ProtocolTCP,
									ContainerPort: crd.Spec.Container.Port,
								},
							},
						},
					},
				},
			},
		},
	}

}

func (c *Controller) updateCrdStatus(crd *controllerv1.Crd, deployment *appsv1.Deployment) error {
	crdCopy := crd.DeepCopy()
	crdCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
	_, err := c.sampleclientset.MakecrdV1alpha1().Crds(crd.Namespace).Update(context.TODO(), crdCopy, metav1.UpdateOptions{})
	return err
}
