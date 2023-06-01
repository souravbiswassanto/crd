package main

import (
	"flag"
	"fmt"
	client "github.com/souravbiswassanto/crd/pkg/client/clientset/versioned"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	_ "k8s.io/code-generator"
	"path/filepath"
)

func main() {
	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "optional")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "optional")
	}
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("error %s\n", err.Error())
	}
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("error while building kubernetes clientset %s\n", err.Error())
	}
	myClient, err := client.NewForConfig(config)
	if err != nil {
		fmt.Printf("error while building custom klientset %s\n", err.Error())
	}

}
