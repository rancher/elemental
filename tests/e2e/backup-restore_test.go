/*
Copyright © 2024 SUSE LLC

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
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

var _ = Describe("E2E - Install Backup/Restore Operator", Label("install-backup-restore"), func() {
	// Create kubectl context
	// Default timeout is too small, so New() cannot be used
	k := &kubectl.Kubectl{
		Namespace:    "",
		PollTimeout:  tools.SetTimeout(300 * time.Second),
		PollInterval: 500 * time.Millisecond,
	}

	It("Install Backup/Restore Operator", func() {
		// Report to Qase
		testCaseID = 64

		// Default chart
		chartRepo := "rancher-chart"

		By("Configuring Chart repository", func() {
			// Set specific operator version if defined
			if backupRestoreVersion != "" {
				chartRepo = "https://github.com/rancher/backup-restore-operator/releases/download/" + backupRestoreVersion
			} else {
				RunHelmCmdWithRetry("repo", "add", chartRepo, "https://charts.rancher.io")
				RunHelmCmdWithRetry("repo", "update")
			}
		})

		By("Installing rancher-backup-operator", func() {
			for _, chart := range []string{"rancher-backup-crd", "rancher-backup"} {
				// Set the filename in chart if a custom version is defined
				chartName := chart
				if backupRestoreVersion != "" {
					chartName = chart + "-" + strings.Trim(backupRestoreVersion, "v") + ".tgz"
				}

				// Global installation flags
				flags := []string{
					"upgrade", "--install", chart, chartRepo + "/" + chartName,
					"--namespace", "cattle-resources-system",
					"--create-namespace",
					"--wait", "--wait-for-jobs",
				}

				// Add specific options for the rancher-backup chart
				if chart == "rancher-backup" {
					flags = append(flags,
						"--set", "persistence.enabled=true",
						"--set", "persistence.storageClass=local-path",
					)
				}

				// Install through Helm
				RunHelmCmdWithRetry(flags...)

				// Delay few seconds for all to be installed
				time.Sleep(tools.SetTimeout(20 * time.Second))
			}
		})

		By("Waiting for rancher-backup-operator pod", func() {
			// Wait for pod to be started
			Eventually(func() error {
				return rancher.CheckPod(k, [][]string{{"cattle-resources-system", "app.kubernetes.io/name=rancher-backup"}})
			}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(BeNil())
		})
	})
})

var _ = Describe("E2E - Test Backup/Restore", Label("test-backup-restore"), func() {
	backupResourceName := "elemental-backup"
	restoreResourceName := "elemental-restore"

	It("Do a backup", func() {
		// Report to Qase
		testCaseID = 65

		By("Adding a backup resource", func() {
			err := kubectl.Apply(clusterNS, backupYaml)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Checking that the backup has been done", func() {
			out, err := kubectl.RunWithoutErr("get", "backup", backupResourceName,
				"-o", "jsonpath={.metadata.name}")
			Expect(err).To(Not(HaveOccurred()))
			Expect(out).To(ContainSubstring(backupResourceName))

			// Check operator logs
			Eventually(func() string {
				out, _ := kubectl.RunWithoutErr("logs", "-l app.kubernetes.io/name=rancher-backup",
					"--tail=-1", "--since=5m",
					"--namespace", "cattle-resources-system")
				return out
			}, tools.SetTimeout(5*time.Minute), 10*time.Second).Should(ContainSubstring("Done with backup"))
		})
	})

	It("Do a restore", func() {
		// Report to Qase
		testCaseID = 66

		By("Deleting some Elemental resources", func() {
			for _, obj := range []string{"MachineRegistration", "MachineInventorySelectorTemplate"} {
				// List the resources
				list, err := kubectl.RunWithoutErr("get", obj,
					"--namespace", clusterNS,
					"-o", "jsonpath={.items[*].metadata.name}")
				Expect(err).To(Not(HaveOccurred()))

				// Delete the resources
				for _, rsc := range strings.Split(list, " ") {
					_, err := kubectl.RunWithoutErr("delete", obj, "--namespace", clusterNS, rsc)
					Expect(err).To(Not(HaveOccurred()))
				}
			}
		})

		By("Adding a restore resource", func() {
			// Get the backup file from the previous backup
			backupFile, err := kubectl.RunWithoutErr("get", "backup", backupResourceName, "-o", "jsonpath={.status.filename}")
			Expect(err).To(Not(HaveOccurred()))

			// Set the backup file in the restore resource
			err = tools.Sed("%BACKUP_FILE%", backupFile, restoreYaml)
			Expect(err).To(Not(HaveOccurred()))

			// And apply
			err = kubectl.Apply(clusterNS, restoreYaml)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Checking that the restore has been done", func() {
			// Wait until resources are available again
			Eventually(func() string {
				out, _ := kubectl.RunWithoutErr("get", "restore", restoreResourceName,
					"-o", "jsonpath={.metadata.name}")
				return out
			}, tools.SetTimeout(5*time.Minute), 10*time.Second).Should(ContainSubstring(restoreResourceName))

			// Check operator logs
			Eventually(func() string {
				out, _ := kubectl.RunWithoutErr("logs", "-l app.kubernetes.io/name=rancher-backup",
					"--tail=-1", "--since=5m",
					"--namespace", "cattle-resources-system")
				return out
			}, tools.SetTimeout(5*time.Minute), 10*time.Second).Should(ContainSubstring("Done restoring"))
		})

		By("Checking cluster state after restore", func() {
			WaitCluster(clusterNS, clusterName)
		})
	})
})
