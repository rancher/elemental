/*
Copyright © 2023 SUSE LLC

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
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

func testClusterAvailability(ns, cluster string) {
	Eventually(func() string {
		out, _ := kubectl.Run("get", "cluster",
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
		By("Testing cluster resource availability BEFORE operator uninstallation", func() {
			testClusterAvailability(clusterNS, clusterName)
		})

		By("Uninstalling Operator via Helm", func() {
			for _, chart := range []string{"elemental-operator", "elemental-operator-crds"} {
				if strings.Contains(chart, "-crds") {
					// Check if CRDs chart is available (not always the case in older versions)
					chartList, err := exec.Command("helm",
						"list",
						"--no-headers",
						"--namespace", "cattle-elemental-system",
					).CombinedOutput()
					Expect(err).To(Not(HaveOccurred()))

					if !strings.Contains(string(chartList), chart) {
						continue
					}
				}
				err := kubectl.RunHelmBinaryWithCustomErr(
					"uninstall", chart,
					"--namespace", "cattle-elemental-system",
				)
				Expect(err).To(Not(HaveOccurred()))
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

		By("Deleting cluster resource", func() {
			Eventually(func() error {
				_, err := kubectl.Run("delete", "cluster",
					"--namespace", clusterNS, clusterName)
				return err
			}, tools.SetTimeout(2*time.Minute), 10*time.Second).Should(Not(HaveOccurred()))
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
			for _, chart := range []string{"elemental-operator-crds", "elemental-operator"} {
				// Check if CRDs chart is available (not always the case in older versions)
				// Anyway, if it is needed and missing the next chart installation will fail too
				if strings.Contains(chart, "-crds") {
					noChart := kubectl.RunHelmBinaryWithCustomErr("show", "readme", operatorRepo+"/"+chart+"-chart")
					if noChart != nil {
						continue
					}
				}
				err := kubectl.RunHelmBinaryWithCustomErr("upgrade", "--install", chart,
					operatorRepo+"/"+chart+"-chart",
					"--namespace", "cattle-elemental-system",
					"--create-namespace",
				)
				Expect(err).To(Not(HaveOccurred()))
			}

			// Wait for pod to be started
			err := rancher.CheckPod(k, [][]string{{"cattle-elemental-system", "app=elemental-operator"}})
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
