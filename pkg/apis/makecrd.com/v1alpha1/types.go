package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type CrdSpec struct {
	Name      string        `json:"name,omitempty"`
	Replicas  *int32        `json:"replicas"`
	Container ContainerSpec `json:"container,container"`
}
type ContainerSpec struct {
	Image string `json:"image,omitempty"`
	Port  int32  `json:"port,omitempty"`
}

type CredStatus struct {
	AvailableReplicas int32 `json:"availableReplicas"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type Crd struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              CrdSpec    `json:"spec"`
	Status            CredStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CrdList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Crd `json:"items"`
}
