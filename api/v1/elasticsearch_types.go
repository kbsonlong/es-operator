/*
 * @FilePath: /api/v1/elasticsearch_types.go
 * @Author: kbsonlong kbsonlong@gmail.com
 * @Date: 2023-10-09 13:00:45
 * @LastEditors: kbsonlong kbsonlong@gmail.com
 * @LastEditTime: 2023-10-11 14:00:41
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

type ElasticsearchHealth string

// Possible traffic light states Elasticsearch health can have.
const (
	ElasticsearchRedHealth     ElasticsearchHealth = "red"
	ElasticsearchYellowHealth  ElasticsearchHealth = "yellow"
	ElasticsearchGreenHealth   ElasticsearchHealth = "green"
	ElasticsearchUnknownHealth ElasticsearchHealth = "unknown"
)

// ElasticsearchOrchestrationPhase is the phase Elasticsearch is in from the controller point of view.
type ElasticsearchOrchestrationPhase string

const (
	// ElasticsearchReadyPhase is operating at the desired spec.
	ElasticsearchReadyPhase ElasticsearchOrchestrationPhase = "Ready"
	// ElasticsearchApplyingChangesPhase controller is working towards a desired state, cluster can be unavailable.
	ElasticsearchApplyingChangesPhase ElasticsearchOrchestrationPhase = "ApplyingChanges"
	// ElasticsearchMigratingDataPhase Elasticsearch is currently migrating data to another node.
	ElasticsearchMigratingDataPhase ElasticsearchOrchestrationPhase = "MigratingData"
	// ElasticsearchNodeShutdownStalledPhase Elasticsearch cannot make progress with a node shutdown during downscale or rolling upgrade.
	ElasticsearchNodeShutdownStalledPhase ElasticsearchOrchestrationPhase = "Stalled"
	// ElasticsearchResourceInvalid is marking a resource as invalid, should never happen if admission control is installed correctly.
	ElasticsearchResourceInvalid ElasticsearchOrchestrationPhase = "Invalid"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ElasticsearchSpec defines the desired state of Elasticsearch
type ElasticsearchSpec struct {
	Size      int32                          `json:"size,omitempty"`
	Image     string                         `json:"image,omitempty"`
	Resources k8scorev1.ResourceRequirements `json:"resource,omitempty"`
}

// ElasticsearchStatus defines the observed state of Elasticsearch
type ElasticsearchStatus struct {

	// AvailableNodes is the number of available all instances.
	AvailableNodes int32 `json:"availableNodes,omitempty"`
	// AvailableDataNodes is the number of available role data instances.
	AvailableDataNodes int32 `json:"availableDataNodes,omitempty"`
	// Version of the stack resource currently running. During version upgrades, multiple versions may run
	// in parallel: this value specifies the lowest version currently running.
	Version string `json:"version,omitempty"`
	//+kubebuilder:default:= unknown
	Health ElasticsearchHealth             `json:"health,omitempty"`
	Phase  ElasticsearchOrchestrationPhase `json:"phase,omitempty"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status
//+kubebuilder:resource:categories=db,path=elasticsearches,shortName=es,singular=elasticsearch
//+kubebuilder:printcolumn:JSONPath=.spec.size,name=Size,type=integer
//+kubebuilder:printcolumn:JSONPath=.status.health,name=Health,type=string
//+kubebuilder:printcolumn:JSONPath=.status.phase,name=Phase,type=string
//+kubebuilder:printcolumn:JSONPath=.status.availableNodes,name=AvailableNodes,type=integer
//+kubebuilder:printcolumn:JSONPath=.status.availableDataNodes,name=AvailableDataNodes,type=integer

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
