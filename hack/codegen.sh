#!/bin/bash

set -x

vendor/k8s.io/code-generator/generate-groups.sh all \
  github.com/souravbiswassanto/crd/pkg/client \
  github.com/souravbiswassanto/crd/pkg/apis \
  makecrd.com:v1alpha1 \
  --go-header-file /home/user/go/src/github.com/souravbiswassanto/crd/hack/boilerplate.go.txt
#

controller-gen rbac:roleName=my-crd-controller crd paths=github.com/souravbiswassanto/crd/pkg/apis/makecrd.com/v1alpha1 \
crd:crdVersions=v1 output:crd:dir=/home/user/go/src/github.com/souravbiswassanto/crd/manifests output:stdout