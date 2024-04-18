/*
Copyright © 2022 - 2024 SUSE LLC

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
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/elemental"
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
	"github.com/rancher/elemental/tests/e2e/helpers/network"
)

var _ = Describe("E2E - Bootstrap node for UI", Label("ui"), func() {
	var (
		bootstrappedNodes int
		wg                sync.WaitGroup
	)

	It("Configure libvirt and bootstrap a node", func() {
		By("Downloading MachineRegistration", func() {
			tokenURL, err := kubectl.RunWithoutErr("get", "MachineRegistration",
				"--namespace", clusterNS,
				"machine-registration", "-o", "jsonpath={.status.registrationURL}")
			Expect(err).To(Not(HaveOccurred()))

			// Get the YAML config file
			Eventually(func() error {
				return tools.GetFileFromURL(tokenURL, installConfigYaml, false)
			}, tools.SetTimeout(2*time.Minute), 10*time.Second).ShouldNot(HaveOccurred())
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

		if !isoBoot && !rawBoot {
			By("Configuring iPXE boot script for network installation", func() {
				numberOfFile, err := network.ConfigureiPXE(httpSrv)
				Expect(err).To(Not(HaveOccurred()))
				Expect(numberOfFile).To(BeNumerically(">=", 1))
			})
		}

		// Loop on node provisionning
		// NOTE: if numberOfVMs == vmIndex then only one node will be provisionned
		bootstrappedNodes = 0
		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := elemental.SetHostname(vmNameRoot, index)
			Expect(hostName).To(Not(BeEmpty()))

			// Add node in network configuration
			err := rancher.AddNode(netDefaultFileName, hostName, index)
			Expect(err).To(Not(HaveOccurred()))

			// Get generated MAC address
			client, macAdrs := GetNodeInfo(hostName)
			Expect(client).To(Not(BeNil()))
			Expect(macAdrs).To(Not(BeEmpty()))

			wg.Add(1)
			go func(s, h, m string, i int, cl *tools.Client) {
				defer wg.Done()
				defer GinkgoRecover()

				By("Installing node "+h, func() {
					// Execute node deployment in parallel
					err := exec.Command(s, h, m).Run()
					Expect(err).To(Not(HaveOccurred()))

					if rawBoot {
						// Report to Qase that we boot from raw image
						testCaseID = 75

						// The VM will boot first on the recovery partition to create the normal partition
						// No need to check the recovery process
						// Only make sure the VM is up and running on the normal partition
						GinkgoWriter.Printf("Checking ssh on VM %s\n", h)
						CheckSSH(cl)
						GinkgoWriter.Printf("Checking ssh OK on VM %s\n", h)

						// Wait for the end of the elemental-register process
						Eventually(func() error {
							_, err := cl.RunSSH("(journalctl --no-pager -u elemental-register.service) | grep -Eiq 'Finished Elemental Register'")
							return err
						}, tools.SetTimeout(4*time.Minute), 10*time.Second).Should(Not(HaveOccurred()))

						// Wait a bit more to be sure the VM is ready and halt it
						time.Sleep(1 * time.Minute)
						GinkgoWriter.Printf("Stopping VM %s\n", h)
						err := exec.Command("sudo", "virsh", "destroy", h).Run()
						Expect(err).To(Not(HaveOccurred()))

						// Make sure VM status is equal to shut-off
						Eventually(func() string {
							out, _ := exec.Command("sudo", "virsh", "domstate", h).Output()
							return strings.Trim(string(out), "\n\n")
						}, tools.SetTimeout(5*time.Minute), 5*time.Second).Should(Equal("shut off"))
					} else {
						// Report to Qase that we boot from ISO
						testCaseID = 9
					}

				})
			}(installVMScript, hostName, macAdrs, index, client)

			// Wait a bit before starting more nodes to reduce CPU and I/O load
			bootstrappedNodes = misc.WaitNodesBoot(index, vmIndex, bootstrappedNodes, numberOfNodesMax)
		}

		// Wait for all parallel jobs
		wg.Wait()
	})

	It("Add the nodes in Rancher Manager", func() {
		// Wait a bit to make sure the VMs is really halted
		// TODO: Find a better way to check this
		time.Sleep(5 * time.Minute)

		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := elemental.SetHostname(vmNameRoot, index)
			Expect(hostName).To(Not(BeEmpty()))

			// Get node information
			client, _ := GetNodeInfo(hostName)
			Expect(client).To(Not(BeNil()))

			// Execute in parallel
			wg.Add(1)
			go func(c, h string, i int, t bool, cl *tools.Client) {
				defer wg.Done()
				defer GinkgoRecover()

				// Restart the node(s)
				By("Restarting "+h+" to add it in the cluster", func() {
					// Wait a little bit to avoid starting all VMs at the same time
					misc.RandomSleep(sequential, i)

					err := exec.Command("sudo", "virsh", "start", h).Run()
					GinkgoWriter.Printf("Starting VM %s\n", h)
					Expect(err).To(Not(HaveOccurred()))
				})

				By("Checking "+h+" SSH connection", func() {
					CheckSSH(cl)
				})

				By("Checking that TPM is correctly configured on "+h, func() {
					testValue := "-c"
					if t == true {
						testValue = "! -e"
					}
					_ = RunSSHWithRetry(cl, "[[ "+testValue+" /dev/tpm0 ]]")
				})

				By("Checking OS version on "+h, func() {
					out := RunSSHWithRetry(cl, "cat /etc/os-release")
					GinkgoWriter.Printf("OS Version on %s:\n%s\n", h, out)
				})
			}(clusterNS, hostName, index, emulateTPM, client)

			// Wait a bit before starting more nodes to reduce CPU and I/O load
			bootstrappedNodes = misc.WaitNodesBoot(index, vmIndex, bootstrappedNodes, numberOfNodesMax)
		}

		// Wait for all parallel jobs
		wg.Wait()
	})
})
