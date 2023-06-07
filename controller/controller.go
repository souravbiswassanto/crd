package controller

import (
	"context"
	"fmt"
	makecrd_com "github.com/souravbiswassanto/crd/pkg/apis/makecrd.com"
	controllerv1 "github.com/souravbiswassanto/crd/pkg/apis/makecrd.com/v1alpha1"
	clientset "github.com/souravbiswassanto/crd/pkg/client/clientset/versioned"
	lister "github.com/souravbiswassanto/crd/pkg/client/listers/makecrd.com/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	appslisters "k8s.io/client-go/listers/apps/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/util/workqueue"
	"log"
	"strings"
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

func (c *Controller) syncHandler(key string) error {
	namespace, name, err := cache.SplitMetaNamespaceKey(key)
	if err != nil {
		utilruntime.HandleError(fmt.Errorf("invalid resource key: %s\n", key))
		return nil
	}
	// Get the Crd resource with this namespace/name
	crd, err := c.crdLister.Crds(namespace).Get(name)
	if err != nil {
		if errors.IsNotFound(err) {
			utilruntime.HandleError(fmt.Errorf("foo '%s' in workqueue no logner exists\n", key))
			return nil
		}
		return err
	}
	if err := c.deploymentHandler(crd); err != nil {
		utilruntime.HandleError(fmt.Errorf("error while handling deployment: %s", err.Error()))
	}

	if err := c.serviceHandler(crd); err != nil {
		utilruntime.HandleError(fmt.Errorf("error while handling service\n"))
	}

	return nil
}

func (c *Controller) deploymentHandler(crd *controllerv1.Crd) error {
	deploymentName := crd.Spec.Name
	if deploymentName == "" {
		deploymentName = strings.Join(buildSlice(crd.Name, makecrd_com.Deployment), "-")
	}
	fmt.Println(deploymentName, " This is deployment name")
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

func (c *Controller) serviceHandler(crd *controllerv1.Crd) error {
	serviceName := crd.Spec.Name
	if serviceName == "" {
		serviceName = strings.Join(buildSlice(crd.Name, makecrd_com.Service), makecrd_com.Dash)
	}
	service, err := c.kubeclientset.CoreV1().Services(crd.Namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
	if errors.IsNotFound(err) {
		service, err = c.kubeclientset.CoreV1().Services(crd.Namespace).Create(context.TODO(), c.newService(crd), metav1.CreateOptions{})

		if err != nil {
			log.Println(err)
			return err
		}
		log.Printf("Service %s created\n", service.Name)
	}
	if err != nil {
		log.Println(err)
		return err
	}
	_, err = c.kubeclientset.CoreV1().Services(crd.Namespace).Update(context.TODO(), c.newService(crd), metav1.UpdateOptions{})
	if err != nil {
		return err
	}
	return nil
}

func (c *Controller) newDeployment(crd *controllerv1.Crd) *appsv1.Deployment {
	deploymentName := crd.Spec.Name
	if deploymentName == "" {
		deploymentName = strings.Join(buildSlice(crd.Name, makecrd_com.Deployment), "-")
	}
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      deploymentName,
			Namespace: crd.Namespace,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(crd, controllerv1.SchemeGroupVersion.WithKind(makecrd_com.ResourceDefinition)),
			},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: crd.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					makecrd_com.AppLabel:  makecrd_com.AppValue,
					makecrd_com.NameLabel: crd.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						makecrd_com.AppLabel:  makecrd_com.AppValue,
						makecrd_com.NameLabel: crd.Name,
					},
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  makecrd_com.AppValue,
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

func (c *Controller) newService(crd *controllerv1.Crd) *corev1.Service {
	serviceName := crd.Spec.Name
	if serviceName == "" {
		serviceName = strings.Join(buildSlice(crd.Name, makecrd_com.Service), makecrd_com.Dash)
	}
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name: serviceName,
			OwnerReferences: []metav1.OwnerReference{
				*metav1.NewControllerRef(crd, controllerv1.SchemeGroupVersion.WithKind(makecrd_com.ResourceDefinition)),
			},
		},
		Spec: corev1.ServiceSpec{
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				makecrd_com.AppLabel:  makecrd_com.AppValue,
				makecrd_com.NameLabel: crd.Name,
			},
			Ports: []corev1.ServicePort{
				{
					Port:       makecrd_com.ServicePort,
					TargetPort: intstr.FromInt(int(crd.Spec.Container.Port)),
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
