/*
 * @FilePath: /api/v1/kibana_types.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-10 10:40:59
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-10-13 18:08:57
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

// KibanaSpec defines the desired state of Kibana
type KibanaSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Size      int32                          `json:"size,omitempty"`
	Image     string                         `json:"image,omitempty"`
	EsInfo    EsInfo                         `json:"esInfo,omitempty"`
	Resources k8scorev1.ResourceRequirements `json:"resource,omitempty"`
}

type EsInfo struct {
	//+kubebuilder:default:= http
	Schema string `json:"schema,omitempty"`
	Host   string `json:"host,omitempty"`
	//+kubebuilder:default:= 9200
	Port int  `json:"port,omitempty"`
	Auth Auth `json:"auth,omitempty"`
}

type Auth struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// KibanaStatus defines the observed state of Kibana
type KibanaStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=db,path=kibanas,shortName=kb,singular=kibana
//+kubebuilder:printcolumn:JSONPath=.spec.size,name=Size,type=integer

// Kibana is the Schema for the kibanas API
type Kibana struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   KibanaSpec   `json:"spec,omitempty"`
	Status KibanaStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// KibanaList contains a list of Kibana
type KibanaList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Kibana `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Kibana{}, &KibanaList{})
}
