/*
Copyright © 2024 SUSE LLC

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
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/elemental"
)

func deleteFinalizers(ns, object, value string) {
	_, err := kubectl.RunWithoutErr("patch", object,
		"--namespace", ns, value, "--type", "merge",
		"--patch", "{\"metadata\":{\"finalizers\":null}}")
	Expect(err).To(Not(HaveOccurred()))

}

func testClusterAvailability(ns, cluster string) {
	Eventually(func() string {
		out, _ := kubectl.RunWithoutErr("get", "cluster.v1.provisioning.cattle.io",
			"--namespace", ns, cluster,
			"-o", "jsonpath={.metadata.name}")
		return out
	}, tools.SetTimeout(3*time.Minute), 5*time.Second).Should(Equal(clusterName))
}

var _ = Describe("E2E - Uninstall Elemental Operator", Label("uninstall-operator"), func() {
	// Create kubectl context
	// Default timeout is too small, so New() cannot be used
	k := &kubectl.Kubectl{
		Namespace:    "",
		PollTimeout:  tools.SetTimeout(300 * time.Second),
		PollInterval: 500 * time.Millisecond,
	}

	It("Uninstall Elemental Operator", func() {
		// Report to Qase
		testCaseID = 70

		By("Testing cluster resource availability BEFORE operator uninstallation", func() {
			testClusterAvailability(clusterNS, clusterName)
		})

		By("Uninstalling Operator via Helm", func() {
			for _, chart := range []string{"elemental-operator", "elemental-operator-crds"} {
				RunHelmCmdWithRetry(
					"uninstall", chart,
					"--namespace", "cattle-elemental-system",
				)
			}
		})

		By("Testing cluster resource availability AFTER operator uninstallation", func() {
			testClusterAvailability(clusterNS, clusterName)
		})

		By("Checking that Elemental resources are gone", func() {
			Eventually(func() string {
				out, _ := kubectl.Run("get", "MachineInventorySelectorTemplate",
					"--namespace", clusterNS,
					"-o", "jsonpath={.items[*].metadata.name}")
				return out
			}, tools.SetTimeout(3*time.Minute), 5*time.Second).Should(ContainSubstring("NotFound"))
		})

		// NOTE: the operator cannot be reinstall now because there are still CRDs pending to be removed
		By("Checking that Elemental operator CRDs cannot be reinstalled", func() {
			chart := "elemental-operator-crds"
			out, err := kubectl.RunHelmBinaryWithOutput("upgrade", "--install", chart,
				operatorRepo+"/"+chart+"-chart",
				"--namespace", "cattle-elemental-system",
				"--create-namespace",
				"--wait", "--wait-for-jobs",
			)
			Expect(err).To(HaveOccurred(), out)
			Expect(out).To(ContainSubstring("CRDs from previous installations are pending to be removed"))
		})

		// NOTE: we have to run this in background to be able to apply the workaround!
		var wg sync.WaitGroup
		wg.Add(1)
		go func(ns, name string) {
			defer wg.Done()
			defer GinkgoRecover()

			By("Deleting cluster resource", func() {
				Eventually(func() error {
					_, err := kubectl.RunWithoutErr("delete", "cluster.v1.provisioning.cattle.io",
						"--namespace", ns, name)
					return err
				}, tools.SetTimeout(2*time.Minute), 10*time.Second).Should(Not(HaveOccurred()))
			})
		}(clusterNS, clusterName)

		// Removing finalizers from MachineInventory and Machine
		By("Removing finalizers from MachineInventory/Machine", func() {
			// NOTE: wait a bit for the cluster deletion to be started (it's running in background)
			time.Sleep(1 * time.Minute)

			machineList, err := kubectl.RunWithoutErr("get", "MachineInventory",
				"--namespace", clusterNS, "-o", "jsonpath={.items[*].metadata.name}")
			Expect(err).To(Not(HaveOccurred()))

			for _, machine := range strings.Fields(machineList) {
				internalMachine, err := elemental.GetInternalMachine(clusterNS, machine)
				Expect(err).To(Not(HaveOccurred()))

				// Delete blocking Finalizers
				GinkgoWriter.Printf("Deleting Finalizers for MachineInventory '%s'...\n", machine)
				deleteFinalizers(clusterNS, "MachineInventory", machine)

				// Only if Machine is still present
				if internalMachine != "" {
					GinkgoWriter.Printf("Deleting Finalizers for Machine '%s'...\n", internalMachine)
					deleteFinalizers(clusterNS, "Machine", internalMachine)
				}
			}
		})

		// Wait for cluster deletion to be completed
		wg.Wait()

		By("Testing cluster resource unavailability", func() {
			out, err := kubectl.Run("get", "cluster.v1.provisioning.cattle.io",
				"--namespace", clusterNS, clusterName,
				"-o", "jsonpath={.metadata.name}")
			Expect(err).To(HaveOccurred(), out)
			Expect(out).To(ContainSubstring("NotFound"))
		})
	})

	It("Re-install Elemental Operator", func() {
		// Report to Qase
		testCaseID = 62

		By("Installing Operator via Helm", func() {
			for _, chart := range []string{"elemental-operator-crds", "elemental-operator"} {
				RunHelmCmdWithRetry("upgrade", "--install", chart,
					operatorRepo+"/"+chart+"-chart",
					"--namespace", "cattle-elemental-system",
					"--create-namespace",
					"--wait", "--wait-for-jobs",
				)
			}

			// Wait for pod to be started
			Eventually(func() error {
				return rancher.CheckPod(k, [][]string{{"cattle-elemental-system", "app=elemental-operator"}})
			}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(BeNil())
		})

		By("Creating a dumb MachineRegistration", func() {
			err := kubectl.Apply(clusterNS, dumbRegistrationYaml)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Creating cluster", func() {
			// NOTE: we can re-use clusterYaml as it has already been configured correctly
			err := kubectl.Apply(clusterNS, clusterYaml)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Testing cluster resource availability", func() {
			testClusterAvailability(clusterNS, clusterName)
		})
	})
})
