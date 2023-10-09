/*
 * @FilePath: /api/v1/elasticsearch_types.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-09 13:00:45
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-10-09 15:08:17
 * @Description:
 * Copyright (c) 2023 by kbsonlong, All Rights Reserved.
 */
/*
Copyright 2023.

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

package v1

import (
	k8scorev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ElasticsearchSpec defines the desired state of Elasticsearch
type ElasticsearchSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// Foo is an example field of Elasticsearch. Edit elasticsearch_types.go to remove/update
	Size      int32                          `json:"size,omitempty"`
	Image     string                         `json:"image,omitempty"`
	Resources k8scorev1.ResourceRequirements `json:"resource,omitempty"`
}

// ElasticsearchStatus defines the observed state of Elasticsearch
type ElasticsearchStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Elasticsearch is the Schema for the elasticsearches API
type Elasticsearch struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ElasticsearchSpec   `json:"spec,omitempty"`
	Status ElasticsearchStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ElasticsearchList contains a list of Elasticsearch
type ElasticsearchList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Elasticsearch `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Elasticsearch{}, &ElasticsearchList{})
}
