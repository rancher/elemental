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

func checkClusterAgent(client *tools.Client) {
	// cluster-agent is the pod that communicates to Rancher, wait for it before continuing
	Eventually(func() string {
		out, _ := client.RunSSH("kubectl get pod -n cattle-system -l app=cattle-cluster-agent")
		return out
	}, tools.SetTimeout(5*time.Duration(usedNodes)*time.Minute), 10*time.Second).Should(ContainSubstring("Running"))
}

var _ = Describe("E2E - Bootstrapping node", Label("bootstrap"), func() {
	var (
		bootstrappedNodes int
		wg                sync.WaitGroup
	)

	It("Provision the node", func() {
		// Report to Qase
		testCaseID = 9

		if !isoBoot && !rawBoot {
			By("Downloading MachineRegistration file", func() {
				// Download the new YAML installation config file
				machineRegName := "machine-registration-" + poolType + "-" + clusterName
				tokenURL, err := kubectl.RunWithoutErr("get", "MachineRegistration",
					"--namespace", clusterNS, machineRegName,
					"-o", "jsonpath={.status.registrationURL}")
				Expect(err).To(Not(HaveOccurred()))

				Eventually(func() error {
					return tools.GetFileFromURL(tokenURL, installConfigYaml, false)
				}, tools.SetTimeout(2*time.Minute), 10*time.Second).ShouldNot(HaveOccurred())
			})

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
			_, macAdrs := GetNodeInfo(hostName)
			Expect(macAdrs).To(Not(BeEmpty()))

			wg.Add(1)
			go func(s, h, m string, i int) {
				defer wg.Done()
				defer GinkgoRecover()

				By("Installing node "+h, func() {
					// Wait a little bit to avoid starting all VMs at the same time
					misc.RandomSleep(sequential, i)

					// Execute node deployment in parallel
					err := exec.Command(s, h, m).Run()
					Expect(err).To(Not(HaveOccurred()))
				})
			}(installVMScript, hostName, macAdrs, index)

			// Wait a bit before starting more nodes to reduce CPU and I/O load
			bootstrappedNodes = misc.WaitNodesBoot(index, vmIndex, bootstrappedNodes, numberOfNodesMax)
		}

		// Wait for all parallel jobs
		wg.Wait()

		// Loop on nodes to check that SeedImage cloud-config is correctly applied
		// Only for master pool
		if poolType == "master" && isoBoot {
			for index := vmIndex; index <= numberOfVMs; index++ {
				hostName := elemental.SetHostname(vmNameRoot, index)
				Expect(hostName).To(Not(BeEmpty()))

				client, _ := GetNodeInfo(hostName)
				Expect(client).To(Not(BeNil()))

				wg.Add(1)
				go func(h string, cl *tools.Client) {
					defer wg.Done()
					defer GinkgoRecover()

					By("Checking SeedImage cloud-config on "+h, func() {
						// Wait for SSH to be available
						// NOTE: this also checks that the root password was correctly set by cloud-config
						CheckSSH(cl)

						// Check that the cloud-config is correctly applied by checking the presence of a file
						_ = RunSSHWithRetry(cl, "ls /etc/elemental-test")

						// Check that the installation is completed before halting the VM
						Eventually(func() error {
							// A little bit dirty but this is temporary to keep compatibility with older Stable versions
							_, err := cl.RunSSH("(journalctl --no-pager -u elemental-register.service ; journalctl --no-pager -u elemental-register-install.service) | grep -Eiq 'elemental install.* completed'")
							return err
						}, tools.SetTimeout(8*time.Minute), 10*time.Second).Should(Not(HaveOccurred()))

						// Halt the VM
						_ = RunSSHWithRetry(cl, "setsid -f init 0")
					})
				}(hostName, client)
			}
			wg.Wait()
		}
	})

	It("Add the nodes in Rancher Manager", func() {
		// Report to Qase
		testCaseID = 67

		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := elemental.SetHostname(vmNameRoot, index)
			Expect(hostName).To(Not(BeEmpty()))

			// Execute node deployment in parallel
			wg.Add(1)
			go func(c, h string, i int) {
				defer wg.Done()
				defer GinkgoRecover()

				By("Checking that node "+h+" is available in Rancher", func() {
					Eventually(func() string {
						id, _ := elemental.GetServerID(c, i)
						return id
					}, tools.SetTimeout(1*time.Minute), 5*time.Second).Should(Not(BeEmpty()))
				})
			}(clusterNS, hostName, index)
		}

		// Wait for all parallel jobs
		wg.Wait()

		if vmIndex > 1 {
			By("Checking cluster state", func() {
				WaitCluster(clusterNS, clusterName)
			})
		}

		By("Incrementing number of nodes in "+poolType+" pool", func() {
			// Increase 'quantity' field
			poolName := "pool-" + poolType + "-" + clusterName
			value, err := rancher.SetNodeQuantity(clusterNS, clusterName, poolName, usedNodes)
			Expect(err).To(Not(HaveOccurred()))
			Expect(value).To(BeNumerically(">=", 1))

			// Check that the selector has been correctly created
			Eventually(func() string {
				out, _ := kubectl.RunWithoutErr("get", "MachineInventorySelector",
					"--namespace", clusterNS,
					"-o", "jsonpath={.items[*].metadata.name}")
				return out
			}, tools.SetTimeout(3*time.Minute), 5*time.Second).Should(ContainSubstring("selector-" + poolType + "-" + clusterName))
		})

		By("Waiting for known cluster state before adding the node(s)", func() {
			msg := `(configuring .* node\(s\)|waiting for viable init node)`
			Eventually(func() string {
				clusterMsg, _ := elemental.GetClusterState(clusterNS, clusterName,
					"{.status.conditions[?(@.type==\"Updated\")].message}")

				// Sometimes we can have a different status/condition
				if clusterMsg == "" {
					out, _ := elemental.GetClusterState(clusterNS, clusterName,
						"{.status.conditions[?(@.type==\"Provisioned\")].message}")

					return out
				}

				return clusterMsg
			}, tools.SetTimeout(5*time.Duration(usedNodes)*time.Minute), 10*time.Second).Should(MatchRegexp(msg))
		})

		bootstrappedNodes = 0
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

		if poolType != "worker" {
			for index := vmIndex; index <= numberOfVMs; index++ {
				// Set node hostname
				hostName := elemental.SetHostname(vmNameRoot, index)
				Expect(hostName).To(Not(BeEmpty()))

				// Get node information
				client, _ := GetNodeInfo(hostName)
				Expect(client).To(Not(BeNil()))

				// Execute in parallel
				wg.Add(1)
				go func(h string, cl *tools.Client) {
					defer wg.Done()
					defer GinkgoRecover()

					if strings.Contains(k8sDownstreamVersion, "rke2") {
						By("Configuring kubectl command on node "+h, func() {
							dir := "/var/lib/rancher/rke2/bin"
							kubeCfg := "export KUBECONFIG=/etc/rancher/rke2/rke2.yaml"

							// Wait a little to be sure that RKE2 installation has started
							// Otherwise the directory is not available!
							_ = RunSSHWithRetry(cl, "[[ -d "+dir+" ]]")

							// Configure kubectl
							_ = RunSSHWithRetry(cl, "I="+dir+"/kubectl; if [[ -x ${I} ]]; then ln -s ${I} bin/; echo "+kubeCfg+" >> .bashrc; fi")
						})
					}

					By("Checking kubectl command on "+h, func() {
						// Check if kubectl works
						Eventually(func() string {
							out, _ := cl.RunSSH("kubectl version 2>/dev/null | grep 'Server Version:'")
							return out
						}, tools.SetTimeout(5*time.Minute), 5*time.Second).Should(ContainSubstring(k8sDownstreamVersion))
					})

					By("Checking cluster agent on "+h, func() {
						checkClusterAgent(cl)
					})
				}(hostName, client)
			}

			// Wait for all parallel jobs
			wg.Wait()
		}

		By("Checking cluster state", func() {
			WaitCluster(clusterNS, clusterName)
		})

		if poolType != "worker" {
			for index := vmIndex; index <= numberOfVMs; index++ {
				// Set node hostname
				hostName := elemental.SetHostname(vmNameRoot, index)
				Expect(hostName).To(Not(BeEmpty()))

				// Get node information
				client, _ := GetNodeInfo(hostName)
				Expect(client).To(Not(BeNil()))

				// Execute in parallel
				wg.Add(1)
				go func(h string, cl *tools.Client) {
					defer wg.Done()
					defer GinkgoRecover()

					By("Checking cluster version on "+h, func() {
						Eventually(func() error {
							k8sVer, err := cl.RunSSH("kubectl version 2>/dev/null")
							if strings.Contains(k8sVer, "Server Version:") {
								// Show cluster version, could be useful for debugging purposes
								GinkgoWriter.Printf("K8s version on %s:\n%s\n", h, k8sVer)
							}
							return err
						}, tools.SetTimeout(1*time.Minute), 5*time.Second).Should(Not(HaveOccurred()))
					})
				}(hostName, client)
			}

			// Wait for all parallel jobs
			wg.Wait()
		}

		bootstrappedNodes = 0
		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := elemental.SetHostname(vmNameRoot, index)
			Expect(hostName).To(Not(BeEmpty()))

			// Get node information
			client, _ := GetNodeInfo(hostName)
			Expect(client).To(Not(BeNil()))

			// Execute in parallel
			wg.Add(1)
			go func(h, p string, i int, cl *tools.Client) {
				defer wg.Done()
				defer GinkgoRecover()

				By("Rebooting "+h, func() {
					// Wait a little bit to avoid starting all VMs at the same time
					misc.RandomSleep(sequential, i)

					// Execute 'reboot' in background, to avoid SSH locking
					Eventually(func() error {
						_, err := cl.RunSSH("setsid -f reboot")
						return err
					}, tools.SetTimeout(2*time.Minute), 10*time.Second).Should(Not(HaveOccurred()))
				})

				if p != "worker" {
					By("Checking cluster agent on "+h, func() {
						checkClusterAgent(cl)
					})
				}
			}(hostName, poolType, index, client)

			// Wait a bit before starting more nodes to reduce CPU and I/O load
			bootstrappedNodes = misc.WaitNodesBoot(index, vmIndex, bootstrappedNodes, numberOfNodesMax)
		}

		// Wait for all parallel jobs
		wg.Wait()

		By("Checking cluster state after reboot", func() {
			WaitCluster(clusterNS, clusterName)
		})
	})
})
