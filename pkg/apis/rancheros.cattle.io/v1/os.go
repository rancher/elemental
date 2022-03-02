/*
Copyright © 2022 SUSE LLC

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
	fleet "github.com/rancher/fleet/pkg/apis/fleet.cattle.io/v1alpha1"
	upgradev1 "github.com/rancher/system-upgrade-controller/pkg/apis/upgrade.cattle.io/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type ManagedOSImage struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ManagedOSImageSpec   `json:"spec"`
	Status ManagedOSImageStatus `json:"status"`
}

type ManagedOSImageSpec struct {
	OSImage      string                `json:"osImage,omitempty"`
	CloudConfig  *fleet.GenericMap     `json:"cloudConfig,omitempty"`
	NodeSelector *metav1.LabelSelector `json:"nodeSelector,omitempty"`
	Concurrency  *int64                `json:"concurrency,omitempty"`

	Prepare *upgradev1.ContainerSpec `json:"prepare,omitempty"`
	Cordon  *bool                    `json:"cordon,omitempty"`
	Drain   *upgradev1.DrainSpec     `json:"drain,omitempty"`

	ClusterRolloutStrategy *fleet.RolloutStrategy `json:"clusterRolloutStrategy,omitempty"`
	Targets                []fleet.BundleTarget   `json:"clusterTargets,omitempty"`
}

type ManagedOSImageStatus struct {
	fleet.BundleStatus
}
