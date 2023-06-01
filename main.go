package main

import (
	"fmt"
	"github.com/souravbiswassanto/crd/pkg/apis/makecrd.com/v1alpha1"
	_ "k8s.io/code-generator"
)

func main() {
	k := v1alpha1.Crd{}
	fmt.Println(k)
}
