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
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher-sandbox/os2/tests/e2e/helpers/misc"
)

var _ = Describe("E2E - Upgrading node", Label("upgrade"), func() {
	var localVmName = vmName + "-02"

	It("Install RancherOS node", func() {
		By("Creating and installing VM", func() {
			netDefaultFileName := "../assets/net-default.xml"
			hostData, err := tools.GetHostNetConfig(".*name='"+localVmName+"'.*", netDefaultFileName)
			Expect(err).NotTo(HaveOccurred())

			// Install VM
			cmd := exec.Command("../scripts/install-vm", localVmName, hostData.Mac)
			out, err := cmd.CombinedOutput()
			GinkgoWriter.Printf("%s\n", out)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Checking that the VM is available in Rancher", func() {
			misc.GetServerId(clusterNS)
		})
	})

	It("Add server "+localVmName+" in "+clusterName, func() {
		By("Adding server role to predefined cluster", func() {
			serverId := misc.GetServerId(clusterNS)
			patchCmd := `{"spec":{"clusterName":"` + clusterName + `","config":{"role":"server"}}}`
			_, err := kubectl.Run("patch", "MachineInventories",
				"--namespace", clusterNS, serverId,
				"--type", "merge", "--patch", patchCmd,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Restarting the VM", func() {
			err := exec.Command("virsh", "start", localVmName).Run()
			Expect(err).NotTo(HaveOccurred())

			// Waiting for node to be added to the cluster (maybe can be wrote purely in Go?)
			err = exec.Command("../scripts/wait-for-node").Run()
			Expect(err).NotTo(HaveOccurred())
		})

		By("Checking that the VM is added in the cluster", func() {
			serverId, err := kubectl.Run("get", "MachineInventories",
				"--namespace", clusterNS,
				"-o", "jsonpath={.items[0].metadata.name}")
			Expect(err).NotTo(HaveOccurred())
			Expect(serverId).ToNot(Equal(""))

			internalClusterName, err := kubectl.Run("get", "cluster",
				"--namespace", clusterNS, clusterName,
				"-o", "jsonpath={.status.clusterName}")
			Expect(err).NotTo(HaveOccurred())
			Expect(internalClusterName).ToNot(Equal(""))

			internalClusterToken, err := kubectl.Run("get", "MachineInventories",
				"--namespace", clusterNS, serverId,
				"-o", "jsonpath={.status.clusterRegistrationTokenNamespace}")
			Expect(err).NotTo(HaveOccurred())
			Expect(internalClusterToken).ToNot(Equal(""))

			// Check that the VM is added
			Expect(internalClusterName).To(Equal(internalClusterToken))
		})

		By("Checking VM ssh connection", func() {
			netDefaultFileName := "../assets/net-default.xml"
			hostData, err := tools.GetHostNetConfig(".*name='"+localVmName+"'.*", netDefaultFileName)
			Expect(err).NotTo(HaveOccurred())

			client := &tools.Client{
				Host:     string(hostData.IP) + ":22",
				Username: userName,
				Password: userPassword,
			}

			// Retry the SSH connection, as it can takes time for the user to be created
			Eventually(func() string {
				out, _ := client.RunSSH("uname -n")
				return out
			}, "5m", "5s").Should(ContainSubstring(localVmName))
		})
	})

	It("Upgrade RancherOS", func() {
		/*
			By("Adding UpgradeChannel in Rancher", func() {
				err := kubectl.Apply(clusterNS, "../../rancheros-*.upgradechannel-*.yaml")
				Expect(err).NotTo(HaveOccurred())
			})

			By("Triggering Upgrade in Rancher", func() {
				err := kubectl.Apply(clusterNS, "../assets/upgrade-with-managedOSVersion.yaml")
				Expect(err).NotTo(HaveOccurred())
			})
		*/

		By("Triggering Upgrade in Rancher", func() {
			upgradeWithOsImageYaml := "../assets/upgrade-with-osImage.yaml"

			err := tools.Sed("%OS_IMAGE%", osImage, upgradeWithOsImageYaml)
			Expect(err).NotTo(HaveOccurred())
			err = tools.Sed("%CLUSTER_NAME%", clusterName, upgradeWithOsImageYaml)
			Expect(err).NotTo(HaveOccurred())
			err = kubectl.Apply(clusterNS, upgradeWithOsImageYaml)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Checking VM upgrade", func() {
			netDefaultFileName := "../assets/net-default.xml"
			hostData, err := tools.GetHostNetConfig(".*name='"+localVmName+"'.*", netDefaultFileName)
			Expect(err).NotTo(HaveOccurred())

			client := &tools.Client{
				Host:     string(hostData.IP) + ":22",
				Username: userName,
				Password: userPassword,
			}

			version := strings.Split(osImage, ":")[1]
			Eventually(func() string {
				// Use grep here in case of comment in the file!
				out, _ := client.RunSSH("eval $(grep -v ^# /usr/lib/os-release) && echo ${VERSION_ID}")
				out = strings.Trim(out, "\n")
				return out
			}, "10m", "30s").Should(Equal(version))
		})
	})
})
