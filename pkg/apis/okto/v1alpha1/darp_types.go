package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// DarpSpec defines the desired state of Darp
// +k8s:openapi-gen=true
type DarpSpec struct {
	Size              int    `json:"size"`
	RootCaSecret      string `json:"rootCaSecret"`
	ServerCertsSecret string `json:"serverCertsSecret"`
	ServerConfigMap   string `json:"serverConfigMap"`
	CertsMountPath    string `json:"certsMountPath"`
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// DarpStatus defines the observed state of Darp
// +k8s:openapi-gen=true
type DarpStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Darp is the Schema for the darps API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type Darp struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   DarpSpec   `json:"spec,omitempty"`
	Status DarpStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// DarpList contains a list of Darp
type DarpList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Darp `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Darp{}, &DarpList{})
}
