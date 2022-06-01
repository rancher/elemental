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

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher-sandbox/os2/tests/e2e/helpers/misc"
)

var _ = Describe("E2E - Bootstrapping node", Label("bootstrap"), func() {
	var (
		serverId string
	)

	It("Install node and add it in Rancher", func() {
		By("Checking if VM name is set", func() {
			Expect(vmName).To(Not(BeEmpty()))
		})

		By("Configuring iPXE boot script for network installation", func() {
			numberOfFile, err := misc.ConfigureiPXE()
			Expect(err).To(Not(HaveOccurred()))
			Expect(numberOfFile).To(BeNumerically(">=", 1))
		})

		By("Creating and installing VM", func() {
			hostData, err := tools.GetHostNetConfig(".*name='"+vmName+"'.*", netDefaultFileName)
			Expect(err).To(Not(HaveOccurred()))

			// Install VM
			cmd := exec.Command("../scripts/install-vm", vmName, hostData.Mac)
			out, err := cmd.CombinedOutput()
			GinkgoWriter.Printf("%s\n", out)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Checking that the VM is available in Rancher", func() {
			id, err := misc.GetServerId(clusterNS, vmIndex)
			Expect(err).To(Not(HaveOccurred()))
			Expect(id).To(Not(BeEmpty()))

			// Export the id of the newly installed node
			serverId = id
		})

		By("Adding server role to predefined cluster", func() {
			patchCmd := `{"spec":{"clusterName":"` + clusterName + `","config":{"role":"server"}}}`
			_, err := kubectl.Run("patch", "MachineInventories",
				"--namespace", clusterNS, serverId,
				"--type", "merge", "--patch", patchCmd,
			)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Restarting the VM", func() {
			err := exec.Command("virsh", "start", vmName).Run()
			Expect(err).To(Not(HaveOccurred()))

			// Waiting for node to be added to the cluster (maybe can be wrote purely in Go?)
			err = exec.Command("../scripts/wait-for-node").Run()
			Expect(err).To((HaveOccurred()))
		})

		By("Checking that the VM is added in the cluster", func() {
			internalClusterName, err := kubectl.Run("get", "cluster",
				"--namespace", clusterNS, clusterName,
				"-o", "jsonpath={.status.clusterName}")
			Expect(err).To(Not(HaveOccurred()))
			Expect(internalClusterName).To(Not(BeEmpty()))

			internalClusterToken, err := kubectl.Run("get", "MachineInventories",
				"--namespace", clusterNS, serverId,
				"-o", "jsonpath={.status.clusterRegistrationTokenNamespace}")
			Expect(err).To(Not(HaveOccurred()))
			Expect(internalClusterToken).To(Not(BeEmpty()))

			// Check that the VM is added
			Expect(internalClusterName).To(Equal(internalClusterToken))
		})

		By("Checking VM ssh connection", func() {
			hostData, err := tools.GetHostNetConfig(".*name='"+vmName+"'.*", netDefaultFileName)
			Expect(err).To(Not(HaveOccurred()))

			client := &tools.Client{
				Host:     string(hostData.IP) + ":22",
				Username: userName,
				Password: userPassword,
			}

			// Retry the SSH connection, as it can takes time for the user to be created
			Eventually(func() string {
				out, _ := client.RunSSH("uname -n")
				return out
			}, "5m", "5s").Should(ContainSubstring(vmNameRoot))
		})
	})
})
