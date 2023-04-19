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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
)

func testClusterAvailability(ns, cluster string) {
	Eventually(func() string {
		out, _ := kubectl.Run("get", "cluster",
			"--namespace", ns, cluster,
			"-o", "jsonpath={.metadata.name}")
		return out
	}, misc.SetTimeout(3*time.Minute), 5*time.Second).Should(Equal(clusterName))
}

var _ = Describe("E2E - Uninstall Elemental Operator", Label("uninstall-operator"), func() {
	// Create kubectl context
	// Default timeout is too small, so New() cannot be used
	k := &kubectl.Kubectl{
		Namespace:    "",
		PollTimeout:  misc.SetTimeout(300 * time.Second),
		PollInterval: 500 * time.Millisecond,
	}

	It("Uninstall Elemental Operator", func() {
		By("Testing cluster resource availability BEFORE operator uninstallation", func() {
			testClusterAvailability(clusterNS, clusterName)
		})

		By("Uninstalling Operator via Helm", func() {
			err := kubectl.RunHelmBinaryWithCustomErr(
				"uninstall", "elemental-operator",
				"--namespace", "cattle-elemental-system",
			)
			Expect(err).To(Not(HaveOccurred()))
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
			}, misc.SetTimeout(3*time.Minute), 5*time.Second).Should(ContainSubstring("NotFound"))
		})

		By("Deleting cluster resource", func() {
			_, err := kubectl.Run("delete", "cluster",
				"--namespace", clusterNS, clusterName)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Testing cluster resource unavailability", func() {
			out, err := kubectl.Run("get", "cluster",
				"--namespace", clusterNS, clusterName,
				"-o", "jsonpath={.metadata.name}")
			Expect(err).To(HaveOccurred())
			Expect(out).To(ContainSubstring("NotFound"))
		})
	})

	It("Re-install Elemental Operator", func() {
		By("Installing Operator via Helm", func() {
			operatorChart := "oci://registry.opensuse.org/isv/rancher/elemental/dev/charts/rancher/elemental-operator-chart"
			err := kubectl.RunHelmBinaryWithCustomErr("upgrade", "--install", "elemental-operator",
				operatorChart,
				"--namespace", "cattle-elemental-system",
				"--create-namespace",
			)
			Expect(err).To(Not(HaveOccurred()))

			err = k.WaitForNamespaceWithPod("cattle-elemental-system", "app=elemental-operator")
			Expect(err).To(Not(HaveOccurred()))
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
