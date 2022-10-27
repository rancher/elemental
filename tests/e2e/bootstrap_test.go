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
	"fmt"
	"os/exec"
	"strings"
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

var _ = Describe("E2E - Bootstrapping node", Label("bootstrap"), func() {
	var (
		client  *tools.Client
		macAdrs string
	)

	BeforeEach(func() {
		hostData, err := tools.GetHostNetConfig(".*name='"+vmName+"'.*", netDefaultFileName)
		Expect(err).To(Not(HaveOccurred()))

		client = &tools.Client{
			Host:     string(hostData.IP) + ":22",
			Username: userName,
			Password: userPassword,
		}
		macAdrs = hostData.Mac
	})

	It("Install node and add it in Rancher", func() {
		By("Checking if VM name is set", func() {
			Expect(vmName).To(Not(BeEmpty()))
		})

		By("Setting emulated TPM to "+emulateTPM, func() {
			// Set correct value for TPM emulation
			value := "false"
			if emulateTPM == "true" {
				value = "true"
			}

			// Patch the yaml file
			err := tools.Sed("emulate-tpm:.*", "emulate-tpm: "+value, emulatedTPMYaml)
			Expect(err).To(Not(HaveOccurred()))

			out, err := kubectl.Run("patch", "MachineRegistration",
				"--namespace", clusterNS, "machine-registration",
				"--type", "merge", "--patch-file", emulatedTPMYaml,
			)
			Expect(err).To(Not(HaveOccurred()), out)
		})

		By("Downloading installation config file", func() {
			// Download the new YAML installation config file
			tokenURL, err := kubectl.Run("get", "MachineRegistration",
				"--namespace", clusterNS,
				"machine-registration", "-o", "jsonpath={.status.registrationURL}")
			Expect(err).To(Not(HaveOccurred()))

			fileName := "../../install-config.yaml"
			err = tools.GetFileFromURL(tokenURL, fileName, false)
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
						"../../.github/elemental-iso-add-registration ../../install-config.yaml ../../build/elemental-*.iso",
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

		By("Creating and installing VM", func() {
			cmd := exec.Command("../scripts/install-vm", vmName, macAdrs)
			out, err := cmd.CombinedOutput()
			GinkgoWriter.Printf("%s\n", out)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Checking that the VM is available in Rancher", func() {
			id, err := misc.GetServerId(clusterNS, vmIndex)
			Expect(err).To(Not(HaveOccurred()))
			Expect(id).To(Not(BeEmpty()))
		})

		By("Ensuring that the cluster is in healthy state (not on 1st node)", func() {
			if vmIndex > 1 {
				checkClusterState()
			}
		})

		By("Increasing 'quantity' node of predefined cluster", func() {
			// Patch the already-created yaml file directly
			err := tools.Sed("quantity:.*", "quantity: "+fmt.Sprint(vmIndex), clusterYaml)
			Expect(err).To(Not(HaveOccurred()))

			out, err := kubectl.Run("patch", "cluster",
				"--namespace", clusterNS, clusterName,
				"--type", "merge", "--patch-file", clusterYaml,
			)
			Expect(err).To(Not(HaveOccurred()), out)
		})

		By("Waiting for known cluster state before adding the node", func() {
			if vmIndex > 1 {
				waitForKnownState(".status.conditions[?(@.type==\"Updated\")].message", "WaitingForBootstrapReason")
			} else {
				waitForKnownState(".status.conditions[?(@.type==\"Provisioned\")].message", "waiting for viable init node")
			}
		})

		By("Restarting the VM to add it in the cluster", func() {
			err := exec.Command("sudo", "virsh", "start", vmName).Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Checking VM connection", func() {
			id, err := misc.GetServerId(clusterNS, vmIndex)
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
			}, misc.SetTimeout(3*time.Minute), 5*time.Second).Should(ContainSubstring(k8sVersion))
		})

		By("Checking cluster state", func() {
			// Check agent and cluster state
			checkClusterAgent(client)
			checkClusterState()
		})

		By("Checking cluster version", func() {
			// Show cluster version, could be useful for debugging purposes
			version := getClusterVersion(client)
			GinkgoWriter.Printf("K8s version:\n%s\n", version)
		})

		By("Rebooting the VM and checking that cluster is still healthy after", func() {
			// Execute 'reboot' in background, to avoid ssh locking
			_, err := client.RunSSH("setsid -f reboot")
			Expect(err).To(Not(HaveOccurred()))

			// Wait a little bit for the cluster to be in an unstable state (yes!)
			time.Sleep(misc.SetTimeout(2 * time.Minute))

			// Check agent and cluster state
			checkClusterAgent(client)
			checkClusterState()
		})
	})
})
