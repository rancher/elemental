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

package e2e_test

import (
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

var _ = Describe("E2E - Test the reset feature", Label("reset"), func() {
	It("Reset one node in the cluster", func() {
		// Get the machine inventory name list
		machineInventory, err := kubectl.Run("get", "machineinventory", "-A", "-o", "jsonpath='{.items[*].metadata.name}'")
		Expect(err).To(Not(HaveOccurred()))
		firstMachineInventory := strings.Split(machineInventory, " ")[1]

		By("Configuring reset at machine inventory level", func() {
			// Patch the first machine inventory to enable reset
			_, err = kubectl.Run("patch", "machineinventory", firstMachineInventory, "--namespace", clusterNS, "--type", "merge", "--patch-file", resetMachineInv)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Deleting and removing the node from the cluster", func() {
			out, err := exec.Command("bash", "-c", "kubectl get machines -A | awk '/"+firstMachineInventory+"/ {print $2}'").CombinedOutput()
			Expect(err).To(Not(HaveOccurred()))
			machineToRemove := strings.Trim(string(out), "\n")
			_, err = kubectl.Run("delete", "machines", machineToRemove, "-n", "fleet-default")
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Checking that machine inventory is deleted", func() {
			Eventually(func() string {
				out, _ := kubectl.Run("get", "MachineInventory",
					"--namespace", clusterNS,
					"-o", "jsonpath={.items[*].metadata.name}")
				return out
			}, tools.SetTimeout(5*time.Minute), 5*time.Second).ShouldNot(ContainSubstring(firstMachineInventory))
		})

		By("Checking that machine inventory is back after the reset", func() {
			Eventually(func() string {
				out, _ := kubectl.Run("get", "MachineInventory",
					"--namespace", clusterNS,
					"-o", "jsonpath={.items[*].metadata.name}")
				return out
			}, tools.SetTimeout(8*time.Minute), 5*time.Second).Should(ContainSubstring(firstMachineInventory))
		})

		By("Checking cluster state", func() {
			CheckClusterState(clusterNS, clusterName)
		})
	})
})