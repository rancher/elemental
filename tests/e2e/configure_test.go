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

package e2e_test

import (
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
)

var _ = Describe("E2E - Configure test", Label("configure"), func() {
	It("Configure Rancher and libvirt", func() {
		By("Creating a new cluster", func() {
			err := tools.Sed("%CLUSTER_NAME%", clusterName, clusterYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = tools.Sed("%K8S_VERSION%", k8sVersion, clusterYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = kubectl.Apply(clusterNS, clusterYaml)
			Expect(err).To(Not(HaveOccurred()))

			createdCluster, err := kubectl.Run("get", "cluster",
				"--namespace", clusterNS,
				clusterName, "-o", "jsonpath={.metadata.name}")
			Expect(err).To(Not(HaveOccurred()))

			// Check that's the created cluster is the good one
			Expect(createdCluster).To(Equal(clusterName))
		})

		By("Creating cluster selector", func() {
			err := tools.Sed("%CLUSTER_NAME%", clusterName, selectorYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = kubectl.Apply(clusterNS, selectorYaml)
			Expect(err).To(Not(HaveOccurred()))

			// Check that the selector is correctly created
			Eventually(func() string {
				out, _ := kubectl.Run("get", "MachineInventorySelector",
					"--namespace", clusterNS,
					"-o", "jsonpath={.items[*].metadata.name}")
				return out
			}, misc.SetTimeout(3*time.Minute), 5*time.Second).Should(ContainSubstring("selector-" + clusterName))
		})

		By("Adding MachineRegistration", func() {
			err := tools.Sed("%VM_NAME%", vmNameRoot, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = tools.Sed("%USER%", userName, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = tools.Sed("%PASSWORD%", userPassword, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = tools.Sed("%CLUSTER_NAME%", clusterName, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = kubectl.Apply(clusterNS, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Starting HTTP server for network installation", func() {
			// TODO: improve it to run in background!
			// err := tools.HTTPShare("../..", 8000)
			// Expect(err).To(Not(HaveOccurred()))

			// Use Python for now...
			err := exec.Command("../scripts/start-httpd").Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Starting default network", func() {
			// Don't check return code, as the default network could be already removed
			cmds := []string{"net-destroy", "net-undefine"}
			for _, c := range cmds {
				_ = exec.Command("sudo", "virsh", c, "default").Run()
			}

			err := exec.Command("sudo", "virsh", "net-create", netDefaultFileName).Run()
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})
