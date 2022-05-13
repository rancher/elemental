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
	"bytes"
	"os/exec"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher-sandbox/os2/tests/e2e/helpers/misc"
)

var _ = Describe("E2E - Bootstrapping node", Label("bootstrap"), func() {
	It("Install RancherOS node", func() {
		By("Creating and installing VM", func() {
			/*
				cmd := exec.Command("virt-install",
					"--name", vmName,
					"--os-type", "Linux",
					"--os-variant", "opensuse-unknown",
					"--virt-type", "kvm",
					"--machine", "q35",
					"--boot", "bios.useserial=on",
					"--ram", "2048",
					"--vcpus", "2",
					"--cpu", "host",
					"--disk", "path=hdd.img,bus=virtio,size=35",
					"--check", "disk_size=off",
					"--graphics", "none",
					"--serial", "pty",
					"--console", "pty,target_type=virtio",
					"--rng", "random",
					"--tpm", "emulator,model=tpm-crb,version=2.0",
					"--noreboot",
					"--pxe",
					"--network", "network=default,bridge=virbr0,model=virtio,mac=52:54:00:00:00:01",
				)
			*/
			// TODO: Create a native Go function for this
			netDefaultFileName := "../assets/net-default.xml"
			mac, err := exec.Command("sed", "-n", "/name='"+vmName+"'/s/.*mac='\\(.*\\)'.*/\\1/p", netDefaultFileName).CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			mac = bytes.Trim(mac, "\n")

			// Install VM
			cmd := exec.Command("../scripts/install-vm", vmName, mac)
			out, err := cmd.CombinedOutput()
			GinkgoWriter.Printf("%s\n", out)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Checking that the VM is available in Rancher", func() {
			misc.GetServerId(clusterNS)
		})
	})

	It("Add server "+vmName+" in "+clusterName, func() {
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
			err := exec.Command("virsh", "start", vmName).Run()
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
			// TODO: Create a native Go function for this
			netDefaultFileName := "../assets/net-default.xml"
			ip, err := exec.Command("sed", "-n", "/name='"+vmName+"'/s/.*ip='\\(.*\\)'.*/\\1/p", netDefaultFileName).CombinedOutput()
			ip = bytes.Trim(ip, "\n")
			GinkgoWriter.Printf("IP=%s\n", ip)
			Expect(err).NotTo(HaveOccurred())

			client := &tools.Client{
				Host:     string(ip) + ":22",
				Username: userName,
				Password: userPassword,
			}

			// Retry the SSH connection, as it can takes time for the user to be created
			Eventually(func() string {
				out, _ := client.RunSSH("uname -n")
				return out
			}, "5m", "5s").Should(ContainSubstring(vmName))
		})
	})
})
