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
	"math/rand"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/elemental"
	"github.com/rancher/elemental/tests/e2e/helpers/network"
)

func checkClusterAgent(client *tools.Client) {
	// cluster-agent is the pod that communicates to Rancher, wait for it before continuing
	Eventually(func() string {
		out, _ := client.RunSSH("kubectl get pod -n cattle-system -l app=cattle-cluster-agent")
		return out
	}, tools.SetTimeout(3*time.Duration(usedNodes)*time.Minute), 10*time.Second).Should(ContainSubstring("Running"))
}

func getClusterState(ns, cluster, condition string) string {
	out, err := kubectl.Run("get", "cluster", "--namespace", ns, cluster, "-o", "jsonpath="+condition)
	Expect(err).To(Not(HaveOccurred()))
	return out
}

func randomSleep(index int) {
	// Only useful in parallel mode
	if sequential == true {
		return
	}

	// Initialize the seed
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Get a pseudo-random value
	timeMax := 240000
	value := r.Intn(timeMax + (timeMax % index))

	// Wait until value is reached
	time.Sleep(time.Duration(value) * time.Millisecond)
}

func waitNodesBoot(i, v, b int) int {
	if (i - v - b) == numberOfNodesMax {
		// Save the number of nodes already bootstrapped for the next round
		b = (i - v)

		// Wait a little
		time.Sleep(4 * time.Minute)
	}

	return b
}

var _ = Describe("E2E - Bootstrapping node", Label("bootstrap"), func() {
	var (
		bootstrappedNodes int
		wg                sync.WaitGroup
	)

	It("Provision the node", func() {
		type pattern struct {
			key   string
			value string
		}

		// Set MachineRegistration name based on hostname
		machineRegName := "machine-registration-" + poolType + "-" + clusterName
		seedImageName := "seed-image-" + poolType + "-" + clusterName
		baseImageURL := "http://192.168.122.1:8000/base-image.iso"

		// Patterns to replace
		patterns := []pattern{
			{
				key:   "%CLUSTER_NAME%",
				value: clusterName,
			},
			{
				key:   "%BASE_IMAGE%",
				value: baseImageURL,
			},
			{
				key:   "%POOL_TYPE%",
				value: poolType,
			},
		}

		By("Setting emulated TPM to "+strconv.FormatBool(emulateTPM), func() {
			// Set temporary file
			emulatedTmp, err := tools.CreateTemp("emulatedTPM")
			Expect(err).To(Not(HaveOccurred()))
			defer os.Remove(emulatedTmp)

			// Save original file as it can be modified multiple time
			err = tools.CopyFile(emulateTPMYaml, emulatedTmp)
			Expect(err).To(Not(HaveOccurred()))

			// Patch the yaml file
			err = tools.Sed("%EMULATE_TPM%", strconv.FormatBool(emulateTPM), emulatedTmp)
			Expect(err).To(Not(HaveOccurred()))

			// And apply it
			_, err = kubectl.Run("patch", "MachineRegistration",
				"--namespace", clusterNS, machineRegName,
				"--type", "merge", "--patch-file", emulatedTmp,
			)
			Expect(err).To(Not(HaveOccurred()))
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

		if isoBoot == "true" {
			By("Adding SeedImage", func() {
				// Set temporary file
				seedimageTmp, err := tools.CreateTemp("seedimage")
				Expect(err).To(Not(HaveOccurred()))
				defer os.Remove(seedimageTmp)

				// Set poweroff to false for master pool to have time to check SeedImage cloud-config
				if poolType == "master" {
					_, err := kubectl.Run("patch", "MachineRegistration",
						"--namespace", clusterNS, machineRegName,
						"--type", "merge", "--patch",
						"{\"spec\":{\"config\":{\"elemental\":{\"install\":{\"poweroff\":false}}}}}")
					Expect(err).To(Not(HaveOccurred()))
				}

				// Save original file as it will have to be modified twice
				err = tools.CopyFile(seedimageYaml, seedimageTmp)
				Expect(err).To(Not(HaveOccurred()))

				// Create Yaml file
				for _, p := range patterns {
					err := tools.Sed(p.key, p.value, seedimageTmp)
					Expect(err).To(Not(HaveOccurred()))
				}

				// Apply to k8s
				err = kubectl.Apply(clusterNS, seedimageTmp)
				Expect(err).To(Not(HaveOccurred()))

				// Check that the seed image is correctly created
				Eventually(func() string {
					out, _ := kubectl.Run("get", "SeedImage",
						"--namespace", clusterNS,
						seedImageName,
						"-o", "jsonpath={.status}")
					return out
				}, tools.SetTimeout(3*time.Minute), 5*time.Second).Should(ContainSubstring("downloadURL"))
			})

			By("Downloading ISO built by SeedImage", func() {
				seedImageURL, err := kubectl.Run("get", "SeedImage",
					"--namespace", clusterNS,
					seedImageName,
					"-o", "jsonpath={.status.downloadURL}")
				Expect(err).To(Not(HaveOccurred()))

				// ISO file size should be greater than 500MB
				Eventually(func() int64 {
					// No need to check download status, file size at the end is enough
					filename := "../../elemental-" + poolType + ".iso"
					_ = tools.GetFileFromURL(seedImageURL, filename, false)
					file, _ := os.Stat(filename)
					return file.Size()
				}, tools.SetTimeout(2*time.Minute), 10*time.Second).Should(BeNumerically(">", 500*1024*1024))
			})
		}

		if isoBoot != "true" {
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
			Expect(hostName).To(Not(BeNil()))

			// Add node in network configuration
			err := rancher.AddNode(netDefaultFileName, hostName, index)
			Expect(err).To(Not(HaveOccurred()))

			// Get generated MAC address
			_, macAdrs := GetNodeInfo(hostName)
			Expect(macAdrs).To(Not(BeNil()))

			wg.Add(1)
			go func(s, h, m string, i int) {
				defer wg.Done()
				defer GinkgoRecover()

				By("Installing node "+h, func() {
					// Wait a little bit to avoid starting all VMs at the same time
					randomSleep(i)

					// Execute node deployment in parallel
					err := exec.Command(s, h, m).Run()
					Expect(err).To(Not(HaveOccurred()))
				})
			}(installVMScript, hostName, macAdrs, index)

			// Wait a bit before starting more nodes to reduce CPU and I/O load
			bootstrappedNodes = waitNodesBoot(index, vmIndex, bootstrappedNodes)
		}

		// Wait for all parallel jobs
		wg.Wait()

		// Loop on nodes to check that SeedImage cloud-config is correctly applied
		// Only for master pool
		if poolType == "master" && isoBoot == "true" {
			for index := vmIndex; index <= numberOfVMs; index++ {
				hostName := elemental.SetHostname(vmNameRoot, index)
				Expect(hostName).To(Not(BeNil()))

				client, _ := GetNodeInfo(hostName)
				Expect(client).To(Not(BeNil()))

				wg.Add(1)
				go func(h string, cl *tools.Client) {
					defer wg.Done()
					defer GinkgoRecover()

					By("Checking SeedImage cloud-config on "+h, func() {
						// Wait for SSH to be available
						// NOTE: this also checks that the root password was correctly set by cloud-config
						Eventually(func() string {
							out, _ := cl.RunSSH("echo SSH_OK")
							out = strings.Trim(out, "\n")
							return out
						}, tools.SetTimeout(10*time.Minute), 5*time.Second).Should(Equal("SSH_OK"))

						// Check that the cloud-config is correctly applied by checking the presence of a file
						_, err := cl.RunSSH("ls /etc/elemental-test")
						Expect(err).To(Not(HaveOccurred()))

						// Check that the installation is completed before halting the VM
						Eventually(func() error {
							// A little bit dirty but this is temporary to keep compatibility with older Stable versions
							_, err := cl.RunSSH("(journalctl --no-pager -u elemental-register.service ; journalctl --no-pager -u elemental-register-install.service) | grep -Eiq 'elemental install.* completed'")
							return err
						}, tools.SetTimeout(8*time.Minute), 10*time.Second).Should(Not(HaveOccurred()))

						// Halt the VM
						_, err = cl.RunSSH("setsid -f init 0")
						Expect(err).To(Not(HaveOccurred()))
					})
				}(hostName, client)
			}
			wg.Wait()
		}
	})

	It("Add the nodes in Rancher Manager", func() {
		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := elemental.SetHostname(vmNameRoot, index)
			Expect(hostName).To(Not(BeNil()))

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
				CheckClusterState(clusterNS, clusterName)
			})
		}

		By("Incrementing number of nodes in "+poolType+" pool", func() {
			// Increase 'quantity' field
			value, err := rancher.SetNodeQuantity(clusterNS,
				clusterName,
				"pool-"+poolType+"-"+clusterName, usedNodes)
			Expect(err).To(Not(HaveOccurred()))
			Expect(value).To(BeNumerically(">=", 1))

			// Check that the selector has been correctly created
			Eventually(func() string {
				out, _ := kubectl.Run("get", "MachineInventorySelector",
					"--namespace", clusterNS,
					"-o", "jsonpath={.items[*].metadata.name}")
				return out
			}, tools.SetTimeout(3*time.Minute), 5*time.Second).Should(ContainSubstring("selector-" + poolType + "-" + clusterName))
		})

		By("Waiting for known cluster state before adding the node(s)", func() {
			msg := `(configuring .* node\(s\)|waiting for viable init node)`
			Eventually(func() string {
				clusterMsg := getClusterState(clusterNS, clusterName,
					"{.status.conditions[?(@.type==\"Updated\")].message}")

				if clusterMsg == "" {
					clusterMsg = getClusterState(clusterNS, clusterName,
						"{.status.conditions[?(@.type==\"Provisioned\")].message}")
				}

				return clusterMsg
			}, tools.SetTimeout(5*time.Duration(usedNodes)*time.Minute), 10*time.Second).Should(MatchRegexp(msg))
		})

		bootstrappedNodes = 0
		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := elemental.SetHostname(vmNameRoot, index)
			Expect(hostName).To(Not(BeNil()))

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
					randomSleep(i)

					err := exec.Command("sudo", "virsh", "start", h).Run()
					Expect(err).To(Not(HaveOccurred()))
				})

				By("Checking "+h+" SSH connection", func() {
					// Retry the SSH connection, as it can takes time for the user to be created
					Eventually(func() string {
						out, _ := cl.RunSSH("echo SSH_OK")
						out = strings.Trim(out, "\n")
						return out
					}, tools.SetTimeout(10*time.Minute), 5*time.Second).Should(Equal("SSH_OK"))
				})

				By("Checking that TPM is correctly configured on "+h, func() {
					testValue := "-c"
					if t == true {
						testValue = "! -e"
					}
					Eventually(func() error {
						_, err := cl.RunSSH("[[ " + testValue + " /dev/tpm0 ]]")
						return err
					}, tools.SetTimeout(1*time.Minute), 5*time.Second).Should(Not(HaveOccurred()))
				})

				By("Checking OS version on "+h, func() {
					out, err := cl.RunSSH("cat /etc/os-release")
					Expect(err).To(Not(HaveOccurred()))
					GinkgoWriter.Printf("OS Version on %s:\n%s\n", h, out)
				})
			}(clusterNS, hostName, index, emulateTPM, client)

			// Wait a bit before starting more nodes to reduce CPU and I/O load
			bootstrappedNodes = waitNodesBoot(index, vmIndex, bootstrappedNodes)
		}

		// Wait for all parallel jobs
		wg.Wait()

		if poolType != "worker" {
			for index := vmIndex; index <= numberOfVMs; index++ {
				// Set node hostname
				hostName := elemental.SetHostname(vmNameRoot, index)
				Expect(hostName).To(Not(BeNil()))

				// Get node information
				client, _ := GetNodeInfo(hostName)
				Expect(client).To(Not(BeNil()))

				// Execute in parallel
				wg.Add(1)
				go func(h string, cl *tools.Client) {
					defer wg.Done()
					defer GinkgoRecover()

					if strings.Contains(k8sVersion, "rke2") {
						By("Configuring kubectl command on node "+h, func() {
							dir := "/var/lib/rancher/rke2/bin"
							kubeCfg := "export KUBECONFIG=/etc/rancher/rke2/rke2.yaml"

							// Wait a little to be sure that RKE2 installation has started
							// Otherwise the directory is not available!
							Eventually(func() error {
								_, err := cl.RunSSH("[[ -d " + dir + " ]]")
								return err
							}, tools.SetTimeout(3*time.Minute), 5*time.Second).Should(Not(HaveOccurred()))

							// Configure kubectl
							_, err := cl.RunSSH("I=" + dir + "/kubectl; if [[ -x ${I} ]]; then ln -s ${I} bin/; echo " + kubeCfg + " >> .bashrc; fi")
							Expect(err).To(Not(HaveOccurred()))
						})
					}

					By("Checking kubectl command on "+h, func() {
						// Check if kubectl works
						Eventually(func() string {
							out, _ := cl.RunSSH("kubectl version 2>/dev/null | grep 'Server Version:'")
							return out
						}, tools.SetTimeout(5*time.Minute), 5*time.Second).Should(ContainSubstring(k8sVersion))
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
			CheckClusterState(clusterNS, clusterName)
		})

		if poolType != "worker" {
			for index := vmIndex; index <= numberOfVMs; index++ {
				// Set node hostname
				hostName := elemental.SetHostname(vmNameRoot, index)
				Expect(hostName).To(Not(BeNil()))

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
			Expect(hostName).To(Not(BeNil()))

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
					randomSleep(i)

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
			bootstrappedNodes = waitNodesBoot(index, vmIndex, bootstrappedNodes)
		}

		// Wait for all parallel jobs
		wg.Wait()

		By("Checking cluster state after reboot", func() {
			CheckClusterState(clusterNS, clusterName)
		})

		if isoBoot == "true" {
			By("Removing the ISO", func() {
				err := exec.Command("bash", "-c", "rm -f ../../elemental-*.iso").Run()
				Expect(err).To(Not(HaveOccurred()))
			})
		}
	})
})
