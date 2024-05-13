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
	"os"
	"os/exec"
	"strconv"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/elemental"
)

var _ = Describe("E2E - Bootstrapping nodes", Label("multi-cluster"), func() {
	// Define some variables
	const seedImageName = "seed-image-multi"
	const machineRegName = "machine-registration-multi"

	var (
		basePatterns []YamlPattern
		globalNodeID int
		wg           sync.WaitGroup
	)

	BeforeEach(func() {

		// Patterns to replace
		basePatterns = []YamlPattern{
			{
				key:   "%K8S_VERSION%",
				value: k8sDownstreamVersion,
			},
			{
				key:   "%SNAP_TYPE%",
				value: snapType,
			},
			{
				key:   "%PASSWORD%",
				value: userPassword,
			},
			{
				key:   "%USER%",
				value: userName,
			},
			{
				key:   "%VM_NAME%",
				value: vmNameRoot,
			},
		}
	})

	It("Configure Libvirt", func() {
		// Report to Qase
		testCaseID = 68

		By("Starting default network", func() {
			// Don't check return code, as the default network could be already removed
			for _, c := range []string{"net-destroy", "net-undefine"} {
				_ = exec.Command("sudo", "virsh", c, "default").Run()
			}

			// Wait a bit between virsh commands
			time.Sleep(1 * time.Minute)
			err := exec.Command("sudo", "virsh", "net-create", netDefaultFileName).Run()
			Expect(err).To(Not(HaveOccurred()))
		})
	})

	It("Configure and create ISO image", func() {
		// Report to Qase
		testCaseID = 38

		By("Adding MachineRegistration", func() {
			// Set temporary file
			registrationTmp, err := tools.CreateTemp("machineRegistration")
			Expect(err).To(Not(HaveOccurred()))
			defer os.Remove(registrationTmp)

			// Save original file as it may have to be modified twice
			err = tools.CopyFile(registrationYaml, registrationTmp)
			Expect(err).To(Not(HaveOccurred()))

			// Create Yaml file
			for _, p := range basePatterns {
				err := tools.Sed(p.key, p.value, registrationTmp)
				Expect(err).To(Not(HaveOccurred()))
			}

			// Apply to k8s
			Eventually(func() error {
				return kubectl.Apply(clusterNS, registrationTmp)
			}, tools.SetTimeout(2*time.Minute), 10*time.Second).ShouldNot(HaveOccurred())

			// Check that the machine registration is correctly created
			CheckCreatedRegistration(clusterNS, "machine-registration-multi")
		})

		By("Downloading MachineRegistration file", func() {
			// Download the new YAML installation config file
			tokenURL, err := kubectl.RunWithoutErr("get", "MachineRegistration",
				"--namespace", clusterNS, machineRegName,
				"-o", "jsonpath={.status.registrationURL}")
			Expect(err).To(Not(HaveOccurred()))

			Eventually(func() error {
				return tools.GetFileFromURL(tokenURL, installConfigYaml, false)
			}, tools.SetTimeout(2*time.Minute), 10*time.Second).ShouldNot(HaveOccurred())
		})

		By("Creating ISO from SeedImage", func() {
			// Wait for list of OS versions to be populated
			WaitForOSVersion(clusterNS)

			// Get OSVersion name
			OSVersion, err := exec.Command(getOSScript, os2Test, "true").Output()
			Expect(err).To(Not(HaveOccurred()))
			Expect(OSVersion).To(Not(BeEmpty()))

			// Extract container image URL
			baseImageURL, err := elemental.GetImageURI(clusterNS, string(OSVersion))
			Expect(err).To(Not(HaveOccurred()))
			Expect(baseImageURL).To(Not(BeEmpty()))

			// Set temporary file
			seedImageTmp, err := tools.CreateTemp("seedimage")
			Expect(err).To(Not(HaveOccurred()))
			defer os.Remove(seedImageTmp)

			// Save original file as it may have to be modified twice
			err = tools.CopyFile(seedImageYaml, seedImageTmp)
			Expect(err).To(Not(HaveOccurred()))

			seedImagePatterns := []YamlPattern{
				{
					key:   "%BASE_IMAGE%",
					value: baseImageURL,
				},
			}
			patterns := append(basePatterns, seedImagePatterns...)

			// Create Yaml file
			for _, p := range patterns {
				err := tools.Sed(p.key, p.value, seedImageTmp)
				Expect(err).To(Not(HaveOccurred()))
			}

			// Apply to k8s
			err = kubectl.Apply(clusterNS, seedImageTmp)
			Expect(err).To(Not(HaveOccurred()))
		})
	})

	It("Downloading ISO built by SeedImage", func() {
		// Report to Qase
		testCaseID = 38

		DownloadBuiltISO(clusterNS, seedImageName, "../../elemental-multi.iso")
	})

	It("Create clusters and deploy nodes", func() {
		// Report to Qase
		testCaseID = 9

		// Loop on all clusters to create
		for clusterIndex := 1; clusterIndex <= numberOfClusters; clusterIndex++ {
			createdClusterName := clusterName + "-" + strconv.Itoa(clusterIndex)

			// Patterns to replace
			addPatterns := []YamlPattern{
				{
					key:   "%CLUSTER_NAME%",
					value: createdClusterName,
				},
			}
			patterns := append(basePatterns, addPatterns...)

			By("Creating cluster "+createdClusterName, func() {
				// Set temporary file
				clusterTmp, err := tools.CreateTemp(createdClusterName)
				Expect(err).To(Not(HaveOccurred()))
				defer os.Remove(clusterTmp)

				// Save original file as it may have to be modified twice
				err = tools.CopyFile(clusterYaml, clusterTmp)
				Expect(err).To(Not(HaveOccurred()))

				// Create Yaml file
				for _, p := range patterns {
					err := tools.Sed(p.key, p.value, clusterTmp)
					Expect(err).To(Not(HaveOccurred()))
				}

				// Apply to k8s
				err = kubectl.Apply(clusterNS, clusterTmp)
				Expect(err).To(Not(HaveOccurred()))

				// Check that the cluster is correctly created
				CheckCreatedCluster(clusterNS, createdClusterName)
			})

			By("Creating cluster selector for cluster "+createdClusterName, func() {
				// Set temporary file
				selectorTmp, err := tools.CreateTemp("selector")
				Expect(err).To(Not(HaveOccurred()))
				defer os.Remove(selectorTmp)

				// Save original file as it may have to be modified twice
				err = tools.CopyFile(selectorYaml, selectorTmp)
				Expect(err).To(Not(HaveOccurred()))

				// Create Yaml file
				for _, p := range patterns {
					err := tools.Sed(p.key, p.value, selectorTmp)
					Expect(err).To(Not(HaveOccurred()))
				}

				// Apply to k8s
				err = kubectl.Apply(clusterNS, selectorTmp)
				Expect(err).To(Not(HaveOccurred()))

				// Check that the selector template is correctly created
				CheckCreatedSelectorTemplate(clusterNS, "selector-"+createdClusterName)
			})

			// Loop on node provisionning
			for nodeIndex := 1; nodeIndex <= 3; nodeIndex++ {
				// Incremente global node index
				globalNodeID++

				// Set node hostname
				hostName := elemental.SetHostname(vmNameRoot+"-"+createdClusterName, nodeIndex)
				Expect(hostName).To(Not(BeNil()))

				// Add node in network configuration
				err := rancher.AddNode(netDefaultFileName, hostName, globalNodeID)
				Expect(err).To(Not(HaveOccurred()))

				// Get generated MAC address
				_, macAdrs := GetNodeInfo(hostName)
				Expect(macAdrs).To(Not(BeNil()))

				wg.Add(1)
				go func(s, h, m string) {
					defer wg.Done()
					defer GinkgoRecover()

					By("Installing node "+h+" on cluster "+createdClusterName, func() {
						// Execute node deployment in parallel
						err := exec.Command(s, h, m).Run()
						Expect(err).To(Not(HaveOccurred()))
					})
				}(installVMScript, hostName, macAdrs)
			}

			// Wait for all parallel jobs
			wg.Wait()

			// Add needed label on provisionned nodes
			for nodeIndex := 1; nodeIndex <= 3; nodeIndex++ {
				// Set node hostname
				hostName := elemental.SetHostname(vmNameRoot+"-"+createdClusterName, nodeIndex)
				Expect(hostName).To(Not(BeNil()))

				// Get node's IP
				ip := GetNodeIP(hostName)

				// Get MachineInventory name
				nodeName, err := kubectl.RunWithoutErr("get", "MachineInventory",
					"--namespace", clusterNS,
					"-o", "jsonpath={.items[?(@.metadata.annotations.elemental\\.cattle\\.io/registration-ip==\""+ip+"\")].metadata.name}")
				Expect(err).To(Not(HaveOccurred()))

				// Add label
				elemental.SetMachineInventoryLabel(clusterNS, nodeName, "clusterName", createdClusterName)

				// Get node information
				client, _ := GetNodeInfo(hostName)

				// Restart node(s)
				wg.Add(1)
				go func(h string, cl *tools.Client) {
					defer wg.Done()
					defer GinkgoRecover()

					By("Restarting "+h+" to add it in cluster "+createdClusterName, func() {
						err := exec.Command("sudo", "virsh", "start", h).Run()
						Expect(err).To(Not(HaveOccurred()))
					})

					By("Checking "+h+" SSH connection", func() {
						CheckSSH(cl)
					})
				}(hostName, client)
			}

			// Wait for all parallel jobs
			wg.Wait()

			By("Waiting for cluster "+createdClusterName+" to be Active", func() {
				WaitCluster(clusterNS, createdClusterName)
			})
		}

		// Loop on all clusters to check
		for clusterIndex := 1; clusterIndex <= numberOfClusters; clusterIndex++ {
			createdClusterName := clusterName + "-" + strconv.Itoa(clusterIndex)

			// Do a final check on all created clusters to validate them
			// NOTE: do it in parallel to speed-up the checking process
			wg.Add(1)
			go func(ns, c string) {
				defer wg.Done()
				defer GinkgoRecover()
				By("Waiting for cluster "+c+" to be Active", func() {
					WaitCluster(ns, c)
				})
			}(clusterNS, createdClusterName)

			// Wait for all parallel jobs
			wg.Wait()
		}
	})
})
