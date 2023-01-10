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
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
)

func getClusterVersion(client *tools.Client) string {
	out, err := client.RunSSH("kubectl version")
	Expect(err).To(Not(HaveOccurred()))

	return out
}

func checkClusterAgent(client *tools.Client) {
	// cluster-agent is the pod that communicates to Rancher, wait for it before continuing
	Eventually(func() string {
		out, _ := client.RunSSH("kubectl get pod -n cattle-system -l app=cattle-cluster-agent")
		return out
	}, misc.SetTimeout(3*time.Minute), 10*time.Second).Should(ContainSubstring("Running"))
}

func checkClusterState() {
	// Check that a 'type' property named 'Ready' is set to true
	Eventually(func() string {
		clusterStatus, _ := kubectl.Run("get", "cluster",
			"--namespace", clusterNS, clusterName,
			"-o", "jsonpath={.status.conditions[?(@.type==\"Ready\")].status}")
		return clusterStatus
	}, misc.SetTimeout(2*time.Minute), 10*time.Second).Should(Equal("True"))

	// Wait a little bit for the cluster to be in a stable state
	// NOTE: not SetTimeout needed here!
	time.Sleep(30 * time.Second)

	// There should be no 'reason' property set in a clean cluster
	Eventually(func() string {
		reason, _ := kubectl.Run("get", "cluster",
			"--namespace", clusterNS, clusterName,
			"-o", "jsonpath={.status.conditions[*].reason}")
		return reason
	}, misc.SetTimeout(3*time.Minute), 10*time.Second).Should(BeEmpty())
}

func waitForKnownState(condition, msg string) {
	Eventually(func() string {
		clusterMsg, _ := kubectl.Run("get", "cluster",
			"--namespace", clusterNS, clusterName,
			"-o", "jsonpath={"+condition+"}")
		return clusterMsg
	}, misc.SetTimeout(5*time.Minute), 10*time.Second).Should(ContainSubstring(msg))
}

func selectPool(index int) string {
	// Set pool type
	if index < 4 {
		// First third nodes are in Master pool
		return "master"
	} else {
		// The others are in Worker pool
		return "worker"
	}
}

func getNodeInfo(hostName string, index int) (*tools.Client, string) {
	// Get VM network data
	hostData, err := tools.GetHostNetConfig(".*name=\""+hostName+"\".*", netDefaultFileName)
	Expect(err).To(Not(HaveOccurred()))

	// Set 'client' to be able to access the node through SSH
	c := &tools.Client{
		Host:     string(hostData.IP) + ":22",
		Username: userName,
		Password: userPassword,
	}

	return c, hostData.Mac
}

func deployNode(hostName string, macAdrs string, wg *sync.WaitGroup) {
	defer wg.Done()

	out, err := exec.Command(installVMScript, hostName, macAdrs).CombinedOutput()
	GinkgoWriter.Printf("Output from deployment of node '%s':\n%s\n", hostName, out)
	Expect(err).To(Not(HaveOccurred()))
}

var _ = Describe("E2E - Bootstrapping node", Label("bootstrap"), func() {
	It("Install node and add it in Rancher", func() {
		// indexInPool is 1 by default
		indexInPool := 1

		// Get pool data
		poolType := selectPool(vmIndex)

		// Set MachineRegistration name based on VM hostname
		machineRegName := "machine-registration-" + poolType + "-" + clusterName

		By("Checking if parallel deployment is set and authorized", func() {
			if numberOfVMs > vmIndex {
				// Parallel deployment set, only on worker pool
				Expect(poolType).To(Equal("worker"))
			}
		})

		By("Setting emulated TPM to "+strconv.FormatBool(emulateTPM), func() {
			// Set temporary file
			tmp, err := os.CreateTemp("", "emulatedTPM")
			Expect(err).To(Not(HaveOccurred()))
			emulatedTmp := tmp.Name()
			defer os.Remove(emulatedTmp)

			// Save original file as it can be modified multiple time
			misc.CopyFile(emulateTPMYaml, emulatedTmp)

			// Patch the yaml file
			err = tools.Sed("emulate-tpm:.*", "emulate-tpm: "+strconv.FormatBool(emulateTPM), emulatedTmp)
			Expect(err).To(Not(HaveOccurred()))

			// And apply it
			out, err := kubectl.Run("patch", "MachineRegistration",
				"--namespace", clusterNS, machineRegName,
				"--type", "merge", "--patch-file", emulatedTmp,
			)
			Expect(err).To(Not(HaveOccurred()), out)
		})

		By("Downloading installation config file", func() {
			// Download the new YAML installation config file
			tokenURL, err := kubectl.Run("get", "MachineRegistration",
				"--namespace", clusterNS, machineRegName,
				"-o", "jsonpath={.status.registrationURL}")
			Expect(err).To(Not(HaveOccurred()))

			err = tools.GetFileFromURL(tokenURL, installConfigYaml, false)
			Expect(err).To(Not(HaveOccurred()))
		})

		if isoBoot != "true" {
			By("Configuring iPXE boot script for network installation", func() {
				numberOfFile, err := misc.ConfigureiPXE()
				Expect(err).To(Not(HaveOccurred()))
				Expect(numberOfFile).To(BeNumerically(">=", 1))
			})
		}

		if isoBoot == "true" {
			By("Adding registration file to ISO", func() {
				// Check if generated ISO is already here
				isIso, _ := exec.Command("bash", "-c", "ls ../../elemental-*.iso").Output()

				// No need to recreate the ISO twice
				if len(isIso) == 0 {
					cmd := exec.Command(
						"bash", "-c",
						"../../.github/elemental-iso-add-registration "+installConfigYaml+" ../../build/elemental-*.iso",
					)
					out, err := cmd.CombinedOutput()
					GinkgoWriter.Printf("%s\n", out)
					Expect(err).To(Not(HaveOccurred()))

					// Move generated ISO to the destination directory
					err = exec.Command("bash", "-c", "mv -f elemental-*.iso ../..").Run()
					Expect(err).To(Not(HaveOccurred()))
				}
			})
		}

		// Loop on node provisionning
		// NOTE: if numberOfVMs == vmIndex then only one node will be provisionned
		var wg sync.WaitGroup
		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := misc.SetHostname(vmNameRoot, index)

			// Add node in network configuration
			err := misc.AddNode(netDefaultFileName, hostName, index)
			Expect(err).To(Not(HaveOccurred()))

			// Get generated MAC address
			_, macAdrs := getNodeInfo(hostName, index)
			Expect(macAdrs).To(Not(BeNil()))

			By("Creating and installing VM "+hostName, func() {
				// Execute node deployment in parallel
				wg.Add(1)
				go deployNode(hostName, macAdrs, &wg)
			})
		}
		// Wait for all node to be deployed
		wg.Wait()

		for index := vmIndex; index <= numberOfVMs; index++ {
			By("Checking that the VM(s) is/are available in Rancher", func() {
				id, err := misc.GetServerId(clusterNS, index)
				Expect(err).To(Not(HaveOccurred()))
				Expect(id).To(Not(BeEmpty()))
			})
		}

		if vmIndex > 1 {
			By("Ensuring that the cluster is in healthy state", func() {
				checkClusterState()
			})

			By("Increasing 'quantity' node of predefined cluster", func() {
				// Increase 'quantity' field
				var err error
				indexInPool, err = misc.IncreaseQuantity(clusterNS,
					clusterName,
					"pool-"+poolType+"-"+clusterName,
					(numberOfVMs - vmIndex + 1))
				Expect(err).To(Not(HaveOccurred()))
			})
		}

		By("Waiting for known cluster state before adding the node", func() {
			// Get elemental-operator version
			operatorVersion, err := misc.GetOperatorVersion()
			Expect(err).To(Not(HaveOccurred()))
			operatorVersionShort := strings.Split(operatorVersion, ".")

			if (operatorVersionShort[0] + "." + operatorVersionShort[1]) == "1.0" {
				// Only for elemental-operator v1.0.x
				if vmIndex > 1 {
					if indexInPool == 1 {
						waitForKnownState(".status.conditions[?(@.type==\"Updated\")].message",
							"waiting for agent to check in and apply initial plan")
					} else {
						waitForKnownState(".status.conditions[?(@.type==\"Updated\")].message",
							"WaitingForBootstrapReason")
					}
				} else {
					waitForKnownState(".status.conditions[?(@.type==\"Provisioned\")].message",
						"waiting for viable init node")
				}
			} else {
				// For newer elemental-operator versions
				var state string
				if vmIndex < 4 {
					state = "waiting for agent to check in and apply initial plan"
				} else {
					state = "configuring control plane node(s)"
				}
				waitForKnownState(".status.conditions[?(@.type==\"Updated\")].message", state)
			}
		})

		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := misc.SetHostname(vmNameRoot, index)

			// Restart the VM(s)
			By("Restarting the VM(s) to add it/them in the cluster", func() {
				err := exec.Command("sudo", "virsh", "start", hostName).Run()
				Expect(err).To(Not(HaveOccurred()))
			})

			// Get node information
			client, _ := getNodeInfo(hostName, index)
			Expect(client).To(Not(BeNil()))

			By("Checking VM connection", func() {
				id, err := misc.GetServerId(clusterNS, index)
				Expect(err).To(Not(HaveOccurred()))
				Expect(id).To(Not(BeEmpty()))

				// Retry the SSH connection, as it can takes time for the user to be created
				Eventually(func() string {
					out, _ := client.RunSSH("uname -n")
					out = strings.Trim(out, "\n")
					return out
				}, misc.SetTimeout(2*time.Minute), 5*time.Second).Should(Equal(id))
			})

			By("Showing OS version", func() {
				out, err := client.RunSSH("cat /etc/os-release")
				Expect(err).To(Not(HaveOccurred()))
				GinkgoWriter.Printf("OS Version:\n%s\n", out)
			})
		}

		// No need to check on multiple VMs for now, as only worker pool can be bootstrapped in parallel for now
		if poolType != "worker" {
			// Get node information
			client, _ := getNodeInfo(vmName, vmIndex)

			By("Configuring kubectl command on the VM", func() {
				if strings.Contains(k8sVersion, "rke2") {
					dir := "/var/lib/rancher/rke2/bin"
					kubeCfg := "export KUBECONFIG=/etc/rancher/rke2/rke2.yaml"

					// Wait a little to be sure that RKE2 installation has started
					// Otherwise the directory is not available!
					Eventually(func() string {
						out, _ := client.RunSSH("[[ -d " + dir + " ]] && echo -n OK")
						return out
					}, misc.SetTimeout(3*time.Minute), 5*time.Second).Should(Equal("OK"))

					// Configure kubectl
					_, err := client.RunSSH("I=" + dir + "/kubectl; if [[ -x ${I} ]]; then ln -s ${I} bin/; echo " + kubeCfg + " >> .bashrc; fi")
					Expect(err).To(Not(HaveOccurred()))
				}

				// Check if kubectl works
				Eventually(func() string {
					out, _ := client.RunSSH("kubectl version 2>/dev/null | grep 'Server Version:'")
					return out
				}, misc.SetTimeout(5*time.Minute), 5*time.Second).Should(ContainSubstring(k8sVersion))
			})
		}

		By("Checking cluster state", func() {
			// Check agent and cluster state
			if poolType != "worker" {
				// Get node information
				client, _ := getNodeInfo(vmName, vmIndex)
				Expect(client).To(Not(BeNil()))
				checkClusterAgent(client)
			}
			checkClusterState()
		})

		// No need to check on multiple VMs for now, as only worker pool can be bootstrapped in parallel for now
		if poolType != "worker" {
			By("Checking cluster version", func() {
				// Get node information
				client, _ := getNodeInfo(vmName, vmIndex)
				Expect(client).To(Not(BeNil()))

				// Show cluster version, could be useful for debugging purposes
				version := getClusterVersion(client)
				GinkgoWriter.Printf("K8s version:\n%s\n", version)
			})
		}

		for index := vmIndex; index <= numberOfVMs; index++ {
			By("Rebooting the VM(s)", func() {
				// Set node hostname
				hostName := misc.SetHostname(vmNameRoot, index)

				// Get node information
				client, _ := getNodeInfo(hostName, index)
				Expect(client).To(Not(BeNil()))

				// Execute 'reboot' in background, to avoid SSH locking
				_, err := client.RunSSH("setsid -f reboot")
				Expect(err).To(Not(HaveOccurred()))

				// Wait a little bit for the cluster to be in an unstable state (yes!)
				time.Sleep(misc.SetTimeout(2 * time.Minute))
			})
		}

		By("Checking that cluster is still healthy after", func() {
			// Check agent and cluster state
			if poolType != "worker" {
				// Get node information
				client, _ := getNodeInfo(vmName, vmIndex)
				Expect(client).To(Not(BeNil()))
				checkClusterAgent(client)
			}
			checkClusterState()
		})
	})
})
