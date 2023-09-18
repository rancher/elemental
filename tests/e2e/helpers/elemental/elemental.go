/*
Copyright © 2022 - 2023 SUSE LLC

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

package elemental

import (
	"fmt"
	"strings"

	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"gopkg.in/yaml.v3"
)

/**
 * Add node selector
 * @remarks A nodeSelector field is added
 * @param key key to add in YAML
 * @param value value to set the key to
 * @returns The YAML structure or an error
 */
func AddSelector(key, value string) ([]byte, error) {
	type selectorYaml struct {
		MatchLabels map[string]string `yaml:"matchLabels,omitempty"`
	}

	type selector struct {
		SelectorYaml selectorYaml `yaml:"nodeSelector,omitempty"`
	}

	v := selectorYaml{map[string]string{key: value}}
	s := selector{v}
	out, err := yaml.Marshal(s)
	if err != nil {
		return nil, err
	}

	// Add indent at the beginning
	out = append([]byte("  "), out...)

	return out, nil
}

/**
 * Get container URI from ManagedOSVersion
 * @param clusterNS Namespace
 * @param os OS version to get URI from
 * @returns URI of container image
 */
func GetImageURI(clusterNS, os string) (string, error) {
	uri, err := kubectl.Run("get", "ManagedOSVersion",
		"--namespace", clusterNS, os,
		"-o", "jsonpath={.spec.metadata.uri}")

	if err != nil {
		return "", err
	}

	return uri, nil
}

/**
 * Get Machine from MachineInventory
 * @remarks This is useful to link machine name from Elemental to the Rancher Manager one
 * @param clusterNS Namespace
 * @param machineInventory Machine name as seen by Elemental
 * @returns Corresponding internal machine name
 */
func GetInternalMachine(clusterNS, machineInventory string) (string, error) {
	machine, err := kubectl.Run("get", "Machine",
		"--namespace", clusterNS,
		"-o", "jsonpath={.items[?(@.status.nodeRef.name==\""+machineInventory+"\")].metadata.name}")
	if err != nil {
		return "", err
	}

	return machine, nil
}

/**
 * Get container image used for Elemental operator
 * @remarks Image used is returned
 * @returns The container image used or an error
 */
func GetOperatorImage() (string, error) {
	operatorImage, err := kubectl.Run("get", "pod",
		"--namespace", "cattle-elemental-system",
		"-l", "app=elemental-operator", "-o", "jsonpath={.items[*].status.containerStatuses[*].image}")
	if err != nil {
		return "", err
	}

	return operatorImage, nil
}

/**
 * Get Elemental operator version
 * @remarks Version is returned
 * @returns the Elemental operator version or an error
 */
func GetOperatorVersion() (string, error) {
	operatorImage, err := GetOperatorImage()
	if err != nil {
		return "", err
	}

	// Extract version
	operatorVersion := strings.Split(operatorImage, ":")

	return operatorVersion[1], nil
}

/**
 * Get MachineInventory name (aka. server id)
 * @param clusterNS Namespace
 * @param index URL of the repository
 * @returns The name/id of the server or an error
 */
func GetServerID(clusterNS string, index int) (string, error) {
	serverID, err := kubectl.Run("get", "MachineInventories",
		"--namespace", clusterNS,
		"-o", "jsonpath={.items["+fmt.Sprint(index-1)+"].metadata.name}")
	if err != nil {
		return "", err
	}

	return serverID, nil
}

/**
 * Set hostname of the node
 * @remarks Define the hostname base on baseName and node index
 * @param baseName Basename to use, "empty" if nothing provided
 * @param index index of the node
 * @returns Full hostname of the node
 */
func SetHostname(baseName string, index int) string {
	if baseName == "" {
		baseName = "emtpy"
	}

	if index < 0 {
		index = 0
	}

	return baseName + "-" + fmt.Sprintf("%03d", index)
}

/**
 * Set a label on MachineInventory
 * @remarks Define a custom label on a MachineInventory
 * @param clusterNS Name of the repository
 * @param node Name of the node
 * @param key Label to set
 * @param value Value to set on Label
 * @returns Nothing or an error
 */
func SetMachineInventoryLabel(clusterNS, node, key, value string) error {
	_, err := kubectl.Run("label", "machineinventory",
		"--namespace", clusterNS, node,
		"--overwrite", key+"="+value)
	if err != nil {
		return err
	}

	return nil
}
