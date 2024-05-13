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
	"strings"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/elemental"
)

var _ = Describe("E2E - Upgrading Elemental Operator", Label("upgrade-operator"), func() {
	// Create kubectl context
	// Default timeout is too small, so New() cannot be used
	k := &kubectl.Kubectl{
		Namespace:    "",
		PollTimeout:  tools.SetTimeout(300 * time.Second),
		PollInterval: 500 * time.Millisecond,
	}

	It("Upgrade operator", func() {
		// Report to Qase
		testCaseID = 71

		// Check if CRDs chart is already installed (not always the case in older versions)
		chartList, err := exec.Command("helm",
			"list",
			"--no-headers",
			"--namespace", "cattle-elemental-system",
		).CombinedOutput()
		Expect(err).To(Not(HaveOccurred()))

		upgradeOrder := []string{"elemental-operator-crds", "elemental-operator"}
		if !strings.Contains(string(chartList), "-crds") {
			upgradeOrder = []string{"elemental-operator", "elemental-operator-crds"}
		}

		for _, chart := range upgradeOrder {
			RunHelmCmdWithRetry(
				"upgrade", "--install", chart,
				operatorUpgrade+"/"+chart+"-chart",
				"--namespace", "cattle-elemental-system",
				"--create-namespace",
				"--wait", "--wait-for-jobs",
			)

			// Delay few seconds for all to be installed
			time.Sleep(tools.SetTimeout(20 * time.Second))
		}

		// Wait for all pods to be started
		Eventually(func() error {
			return rancher.CheckPod(k, [][]string{{"cattle-elemental-system", "app=elemental-operator"}})
		}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(BeNil())
	})
})

var _ = Describe("E2E - Upgrading Rancher Manager", Label("upgrade-rancher-manager"), func() {
	// Create kubectl context
	// Default timeout is too small, so New() cannot be used
	k := &kubectl.Kubectl{
		Namespace:    "",
		PollTimeout:  tools.SetTimeout(300 * time.Second),
		PollInterval: 500 * time.Millisecond,
	}

	It("Upgrade Rancher Manager", func() {
		// Report to Qase
		testCaseID = 72

		// Get before-upgrade Rancher Manager version
		getImageVersion := []string{
			"get", "pod",
			"--namespace", "cattle-system",
			"-l", "app=rancher",
			"-o", "jsonpath={.items[*].status.containerStatuses[*].image}",
		}
		versionBeforeUpgrade, err := kubectl.RunWithoutErr(getImageVersion...)
		Expect(err).To(Not(HaveOccurred()))

		// Upgrade Rancher Manager
		// NOTE: Don't check the status, we can have false-positive here...
		//       Better to check the rollout after the upgrade, it will fail if the upgrade failed
		_ = rancher.DeployRancherManager(
			rancherHostname,
			rancherUpgradeChannel,
			rancherUpgradeVersion,
			rancherUpgradeHeadVersion,
			caType,
			proxy,
		)

		// Wait for Rancher Manager to be restarted
		// NOTE: 1st or 2nd rollout command can sporadically fail, so better to use Eventually here
		Eventually(func() string {
			status, _ := kubectl.RunWithoutErr(
				"rollout",
				"--namespace", "cattle-system",
				"status", "deployment/rancher",
			)
			return status
		}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(ContainSubstring("successfully rolled out"))

		// Check that all Rancher Manager pods are running
		Eventually(func() error {
			checkList := [][]string{
				{"cattle-system", "app=rancher"},
				{"cattle-fleet-local-system", "app=fleet-agent"},
				{"cattle-system", "app=rancher-webhook"},
			}
			return rancher.CheckPod(k, checkList)
		}, tools.SetTimeout(3*time.Minute), 10*time.Second).Should(Not(HaveOccurred()))

		// A bit dirty be better to wait a little here for all to be correctly started
		time.Sleep(2 * time.Minute)

		// Check that all pods are using the same version
		Eventually(func() int {
			out, _ := kubectl.RunWithoutErr(getImageVersion...)
			return len(strings.Fields(out))
		}, tools.SetTimeout(3*time.Minute), 10*time.Second).Should(Equal(1))

		// Get after-upgrade Rancher Manager version
		// and check that it's different to the before-upgrade version
		versionAfterUpgrade, err := kubectl.RunWithoutErr(getImageVersion...)
		Expect(err).To(Not(HaveOccurred()))
		Expect(versionAfterUpgrade).To(Not(Equal(versionBeforeUpgrade)))
	})
})

var _ = Describe("E2E - Upgrading node", Label("upgrade-node"), func() {
	var (
		value        string
		valueToCheck string
		wg           sync.WaitGroup
	)

	It("Upgrade node", func() {
		// Report to Qase
		testCaseID = 73

		By("Checking if upgrade type is set", func() {
			Expect(upgradeType).To(Not(BeEmpty()))
		})

		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := elemental.SetHostname(vmNameRoot, index)
			Expect(hostName).To(Not(BeEmpty()))

			// Get node information
			client, _ := GetNodeInfo(hostName)
			Expect(client).To(Not(BeNil()))

			// Execute node deployment in parallel
			wg.Add(1)
			go func(h string, cl *tools.Client) {
				defer wg.Done()
				defer GinkgoRecover()

				By("Checking OS version on "+h+" before upgrade", func() {
					out := RunSSHWithRetry(cl, "cat /etc/os-release")
					GinkgoWriter.Printf("OS Version on %s:\n%s\n", h, out)
				})
			}(hostName, client)
		}

		// Wait for all parallel jobs
		wg.Wait()

		By("Triggering Upgrade in Rancher with "+upgradeType, func() {
			// Set temporary file
			upgradeTmp, err := tools.CreateTemp("upgrade")
			Expect(err).To(Not(HaveOccurred()))
			defer os.Remove(upgradeTmp)

			if upgradeType == "managedOSVersionName" {
				// Get OSVersion name
				OSVersion, err := exec.Command(getOSScript, upgradeOSChannel).Output()
				Expect(err).To(Not(HaveOccurred()))

				// In case of sync failure OSVersion can be empty,
				// so try to force the sync before aborting
				if string(OSVersion) == "" {
					const channel = "elemental-channel"

					// Log the workaround, could be useful
					GinkgoWriter.Printf("!! ManagedOSVersionChannel not synced !! Triggering a re-sync!\n")

					// Get current syncInterval
					syncValue, err := kubectl.RunWithoutErr("get", "managedOSVersionChannel",
						"--namespace", clusterNS, channel,
						"-o", "jsonpath={.spec.syncInterval}")
					Expect(err).To(Not(HaveOccurred()))
					Expect(syncValue).To(Not(BeEmpty()))

					// Reduce syncInterval to force an update
					_, err = kubectl.RunWithoutErr("patch", "managedOSVersionChannel",
						"--namespace", clusterNS, channel,
						"--type", "merge",
						"--patch", "{\"spec\":{\"syncInterval\":\"1m\"}}")
					Expect(err).To(Not(HaveOccurred()))

					// Loop until sync is done
					Eventually(func() string {
						value, _ := exec.Command(getOSScript, upgradeOSChannel).Output()

						return string(value)
					}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(Not(BeEmpty()))

					// We should now have an OS version!
					OSVersion, err = exec.Command(getOSScript, upgradeOSChannel).Output()
					Expect(err).To(Not(HaveOccurred()))
					Expect(OSVersion).To(Not(BeEmpty()))

					// Re-patch syncInterval to the initial value
					_, err = kubectl.RunWithoutErr("patch", "managedOSVersionChannel",
						"--namespace", clusterNS, channel,
						"--type", "merge",
						"--patch", "{\"spec\":{\"syncInterval\":\""+syncValue+"\"}}")
					Expect(err).To(Not(HaveOccurred()))
				}

				// Set OS image to use for upgrade
				value = string(OSVersion)

				// Extract the value to check after the upgrade
				out, err := kubectl.RunWithoutErr("get", "ManagedOSVersion",
					"--namespace", clusterNS, value,
					"-o", "jsonpath={.spec.metadata.upgradeImage}")
				Expect(err).To(Not(HaveOccurred()))
				valueToCheck = tools.TrimStringFromChar(out, ":")
			} else if upgradeType == "osImage" {
				// Set OS image to use for upgrade
				value = upgradeImage

				// Extract the value to check after the upgrade
				valueToCheck = tools.TrimStringFromChar(upgradeImage, ":")
			}

			// Add a nodeSelector if needed
			if usedNodes == 1 {
				// Set node hostname
				hostName := elemental.SetHostname(vmNameRoot, vmIndex)
				Expect(hostName).To(Not(BeEmpty()))

				// Get node information
				client, _ := GetNodeInfo(hostName)
				Expect(client).To(Not(BeNil()))

				// Get *REAL* hostname
				hostname := RunSSHWithRetry(client, "hostname")
				hostname = strings.Trim(hostname, "\n")

				label := "kubernetes.io/hostname"
				selector, err := elemental.AddSelector(label, hostname)
				Expect(err).To(Not(HaveOccurred()), selector)

				// Create new file for this specific upgrade
				err = tools.AddDataToFile(upgradeSkelYaml, upgradeTmp, selector)
				Expect(err).To(Not(HaveOccurred()))
			} else {
				// Use original file as-is
				err := tools.CopyFile(upgradeSkelYaml, upgradeTmp)
				Expect(err).To(Not(HaveOccurred()))
			}

			// Patterns to replace
			patterns := []YamlPattern{
				{
					key:   "with-%UPGRADE_TYPE%",
					value: strings.ToLower(upgradeType),
				},
				{
					key:   "%UPGRADE_TYPE%",
					value: upgradeType + ": " + value,
				},
				{
					key:   "%CLUSTER_NAME%",
					value: clusterName,
				},
				{
					key:   "%FORCE_DOWNGRADE%",
					value: strconv.FormatBool(forceDowngrade),
				},
			}

			// Create Yaml file
			for _, p := range patterns {
				err := tools.Sed(p.key, p.value, upgradeTmp)
				Expect(err).To(Not(HaveOccurred()))
			}

			// Apply the generated file
			err = kubectl.Apply(clusterNS, upgradeTmp)
			Expect(err).To(Not(HaveOccurred()))
		})

		for index := vmIndex; index <= numberOfVMs; index++ {
			// Set node hostname
			hostName := elemental.SetHostname(vmNameRoot, index)
			Expect(hostName).To(Not(BeEmpty()))

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
						out, _ := cl.RunSSH("eval $(grep -v ^# /etc/os-release) && echo ${IMAGE}")

						// This remove the version and keep only the repo, as in the file
						// we have the exact version and we don't know it before the upgrade
						return tools.TrimStringFromChar(strings.Trim(out, "\n"), ":")
					}, tools.SetTimeout(5*time.Minute), 30*time.Second).Should(Equal(valueToCheck))
				})

				By("Checking OS version on "+h+" after upgrade", func() {
					out := RunSSHWithRetry(cl, "cat /etc/os-release")
					GinkgoWriter.Printf("OS Version on %s:\n%s\n", h, out)
				})
			}(hostName, client)
		}

		// Wait for all parallel jobs
		wg.Wait()

		By("Checking cluster state after upgrade", func() {
			WaitCluster(clusterNS, clusterName)
		})
	})
})
