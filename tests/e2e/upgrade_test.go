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
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/install"
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
)

var _ = Describe("E2E - Upgrading Elemental Operator", Label("upgrade-operator"), func() {
	// Create kubectl context
	// Default timeout is too small, so New() cannot be used
	k := &kubectl.Kubectl{
		Namespace:    "",
		PollTimeout:  misc.SetTimeout(300 * time.Second),
		PollInterval: 500 * time.Millisecond,
	}

	It("Upgrade operator", func() {
		upgradeOrder := []string{"elemental-operator-crds", "elemental-operator"}

		// Check if CRDs chart is already installed (not always the case in older versions)
		chartList, err := exec.Command("helm",
			"list",
			"--no-headers",
			"--namespace", "cattle-elemental-system",
		).CombinedOutput()
		Expect(err).To(Not(HaveOccurred()))

		if !strings.Contains(string(chartList), "-crds") {
			upgradeOrder = []string{"elemental-operator", "elemental-operator-crds"}
		}

		for _, chart := range upgradeOrder {
			err := kubectl.RunHelmBinaryWithCustomErr("upgrade", "--install", chart,
				operatorUpgrade+"/"+chart+"-chart",
				"--namespace", "cattle-elemental-system",
				"--create-namespace",
			)
			Expect(err).To(Not(HaveOccurred()))
		}

		// Delay few seconds before checking, needed because we may have 2 pods at the same time
		time.Sleep(misc.SetTimeout(30 * time.Second))

		// Wait for all pods to be started
		misc.CheckPod(k, [][]string{{"cattle-elemental-system", "app=elemental-operator"}})
	})
})

var _ = Describe("E2E - Upgrading Rancher Manager", Label("upgrade-rancher-manager"), func() {
	// Create kubectl context
	// Default timeout is too small, so New() cannot be used
	k := &kubectl.Kubectl{
		Namespace:    "",
		PollTimeout:  misc.SetTimeout(300 * time.Second),
		PollInterval: 500 * time.Millisecond,
	}

	It("Upgrade Rancher Manager", func() {
		// Get before-upgrade Rancher Manager version
		getImageVersion := []string{
			"get", "pod",
			"--namespace", "cattle-system",
			"-l", "app=rancher",
			"-o", "jsonpath={.items[*].status.containerStatuses[*].image}",
		}
		versionBeforeUpgrade, err := kubectl.Run(getImageVersion...)
		Expect(err).To(Not(HaveOccurred()))

		// Upgrade Rancher Manager
		install.DeployRancherManager(rancherHostname, rancherUpgradeChannel, rancherUpgradeVersion, caType, proxy)

		// Wait for Rancher Manager to be running
		checkList := [][]string{
			{"cattle-system", "app=rancher"},
			{"cattle-fleet-local-system", "app=fleet-agent"},
			{"cattle-system", "app=rancher-webhook"},
		}
		misc.CheckPod(k, checkList)

		// Check that all pods are using the same version
		Eventually(func() int {
			out, _ := kubectl.Run(getImageVersion...)
			return len(strings.Fields(out))
		}, misc.SetTimeout(3*time.Minute), 5*time.Second).Should(Equal(1))

		// Get after-upgrade Rancher Manager version
		// and check that it's different to the before-upgrade version
		versionAfterUpgrade, err := kubectl.Run(getImageVersion...)
		Expect(err).To(Not(HaveOccurred()))
		Expect(versionAfterUpgrade).To(Not(Equal(versionBeforeUpgrade)))
	})
})

var _ = Describe("E2E - Upgrading node", Label("upgrade-node"), func() {
	var (
		wg           sync.WaitGroup
		value        string
		valueToCheck string
	)

	It("Upgrade node", func() {
		By("Checking if upgrade type is set", func() {
			Expect(upgradeType).To(Not(BeEmpty()))
		})

		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := misc.SetHostname(vmNameRoot, index)
			Expect(hostName).To(Not(BeNil()))

			// Get node information
			client, _ := GetNodeInfo(hostName)
			Expect(client).To(Not(BeNil()))

			// Execute node deployment in parallel
			wg.Add(1)
			go func(h string, cl *tools.Client) {
				defer wg.Done()
				defer GinkgoRecover()

				By("Checking OS version on "+h+" before upgrade", func() {
					out, err := client.RunSSH("cat /etc/os-release")
					Expect(err).To(Not(HaveOccurred()))
					GinkgoWriter.Printf("OS Version on %s:\n%s\n", h, out)
				})
			}(hostName, client)
		}

		// Wait for all parallel jobs
		wg.Wait()

		By("Triggering Upgrade in Rancher with "+upgradeType, func() {
			// Set temporary file
			upgradeTmp, err := misc.CreateTemp("upgrade")
			Expect(err).To(Not(HaveOccurred()))
			defer os.Remove(upgradeTmp)

			if upgradeType == "managedOSVersionName" {
				// Get elemental-operator version
				operatorVersion, err := misc.GetOperatorVersion()
				Expect(err).To(Not(HaveOccurred()))
				operatorVersionShort := strings.Split(operatorVersion, ".")

				// Remove 'syncInterval' option if needed (only supported in operator v1.1+)
				if (operatorVersionShort[0] + "." + operatorVersionShort[1]) == "1.0" {
					err := tools.Sed("syncInterval:.*", "", osListYaml)
					Expect(err).To(Not(HaveOccurred()))
				}

				// Add OS channel list
				err = tools.Sed("%UPGRADE_CHANNEL_LIST%", upgradeChannelList, osListYaml)
				Expect(err).To(Not(HaveOccurred()))

				// Apply the generated file
				err = kubectl.Apply(clusterNS, osListYaml)
				Expect(err).To(Not(HaveOccurred()))

				// Wait for ManagedOSVersion to be populated from ManagedOSVersionChannel
				Eventually(func() string {
					out, _ := kubectl.Run("get", "ManagedOSVersion",
						"--namespace", clusterNS, upgradeOsChannel)
					return out
				}, misc.SetTimeout(2*time.Minute), 10*time.Second).Should(Not(ContainSubstring("Error")))

				// Set OS image to use for upgrade
				value = upgradeOsChannel

				// Extract the value to check after the upgrade
				out, err := kubectl.Run("get", "ManagedOSVersion",
					"--namespace", clusterNS, upgradeOsChannel,
					"-o", "jsonpath={.spec.metadata.upgradeImage}")
				Expect(err).To(Not(HaveOccurred()))
				valueToCheck = misc.TrimStringFromChar(out, ":")
			} else if upgradeType == "osImage" {
				// Set OS image to use for upgrade
				value = upgradeImage

				// Extract the value to check after the upgrade
				valueToCheck = misc.TrimStringFromChar(upgradeImage, ":")
			}

			// Add a nodeSelector if needed
			if usedNodes == 1 {
				// Set node hostname
				hostName := misc.SetHostname(vmNameRoot, vmIndex)
				Expect(hostName).To(Not(BeNil()))

				// Get node information
				client, _ := GetNodeInfo(hostName)
				Expect(client).To(Not(BeNil()))

				// Get *REAL* hostname
				hostname, err := client.RunSSH("hostname")
				Expect(err).To(Not(HaveOccurred()))
				hostname = strings.Trim(hostname, "\n")

				label := "kubernetes.io/hostname"
				selector, err := misc.AddSelector(label, hostname)
				Expect(err).To(Not(HaveOccurred()), selector)

				// Create new file for this specific upgrade
				err = misc.ConcateFiles(upgradeSkelYaml, upgradeTmp, selector)
				Expect(err).To(Not(HaveOccurred()))
			} else {
				// Use original file as-is
				misc.CopyFile(upgradeSkelYaml, upgradeTmp)
			}

			// Set values
			err = tools.Sed("with-%UPGRADE_TYPE%", strings.ToLower(upgradeType), upgradeTmp)
			Expect(err).To(Not(HaveOccurred()))
			err = tools.Sed("%UPGRADE_TYPE%", upgradeType+": "+value, upgradeTmp)
			Expect(err).To(Not(HaveOccurred()))
			err = tools.Sed("%CLUSTER_NAME%", clusterName, upgradeTmp)
			Expect(err).To(Not(HaveOccurred()))

			// Apply the generated file
			err = kubectl.Apply(clusterNS, upgradeTmp)
			Expect(err).To(Not(HaveOccurred()))
		})

		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := misc.SetHostname(vmNameRoot, index)
			Expect(hostName).To(Not(BeNil()))

			// Get node information
			client, _ := GetNodeInfo(hostName)
			Expect(client).To(Not(BeNil()))

			// Execute node deployment in parallel
			wg.Add(1)
			go func(h string, cl *tools.Client) {
				defer wg.Done()
				defer GinkgoRecover()

				By("Checking VM upgrade on "+h, func() {
					Eventually(func() string {
						// Use grep here in case of comment in the file!
						out, _ := client.RunSSH("eval $(grep -v ^# /etc/os-release) && echo ${IMAGE}")

						// This remove the version and keep only the repo, as in the file
						// we have the exact version and we don't know it before the upgrade
						return misc.TrimStringFromChar(strings.Trim(out, "\n"), ":")
					}, misc.SetTimeout(5*time.Minute), 30*time.Second).Should(Equal(valueToCheck))
				})

				By("Checking OS version on "+h+" after upgrade", func() {
					out, err := client.RunSSH("cat /etc/os-release")
					Expect(err).To(Not(HaveOccurred()))
					GinkgoWriter.Printf("OS Version on %s:\n%s\n", h, out)
				})
			}(hostName, client)
		}

		// Wait for all parallel jobs
		wg.Wait()

		By("Checking cluster state after upgrade", func() {
			CheckClusterState(clusterNS, clusterName)
		})
	})
})
