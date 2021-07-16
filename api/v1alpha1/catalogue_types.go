/*
Copyright 2021 pilillo.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// CatalogueSpec defines the desired state of Catalogue
type CatalogueSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Size defines the number of Memcached instances
	Size int32 `json:"size,omitempty"`

	// Foo is an example field of Catalogue. Edit catalogue_types.go to remove/update
	//Foo string `json:"foo,omitempty"`
}

// CatalogueStatus defines the observed state of Catalogue
type CatalogueStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Nodes store the name of the pods which are running Memcached instances
	Nodes []string `json:"nodes,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Catalogue is the Schema for the catalogues API
type Catalogue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CatalogueSpec   `json:"spec,omitempty"`
	Status CatalogueStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// CatalogueList contains a list of Catalogue
type CatalogueList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Catalogue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Catalogue{}, &CatalogueList{})
}
