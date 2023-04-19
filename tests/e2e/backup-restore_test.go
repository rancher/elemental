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
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
)

var _ = Describe("E2E - Install Backup/Restore Operator", Label("install-backup-restore"), func() {
	// Create kubectl context
	// Default timeout is too small, so New() cannot be used
	k := &kubectl.Kubectl{
		Namespace:    "",
		PollTimeout:  misc.SetTimeout(300 * time.Second),
		PollInterval: 500 * time.Millisecond,
	}

	It("Install Backup/Restore Operator", func() {
		// Default chart
		chartRepo := "rancher-chart"

		By("Configuring Chart repository", func() {
			// Set specific operator version if defined
			if backupRestoreVersion != "" {
				chartRepo = "https://github.com/rancher/backup-restore-operator/releases/download/" + backupRestoreVersion
			} else {
				err := kubectl.RunHelmBinaryWithCustomErr("repo", "add", chartRepo, "https://charts.rancher.io")
				Expect(err).To(Not(HaveOccurred()))

				err = kubectl.RunHelmBinaryWithCustomErr("repo", "update")
				Expect(err).To(Not(HaveOccurred()))
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
				}

				// Add specific options for the rancher-backup chart
				if chart == "rancher-backup" {
					flags = append(flags,
						"--set", "persistence.enabled=true",
						"--set", "persistence.storageClass=local-path",
					)
				}

				// Install through Helm
				err := kubectl.RunHelmBinaryWithCustomErr(flags...)
				Expect(err).To(Not(HaveOccurred()))
			}
		})

		By("Waiting for rancher-backup-operator pod", func() {
			// Wait for Pod to run
			err := k.WaitForNamespaceWithPod("cattle-resources-system", "app.kubernetes.io/name=rancher-backup")
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})

var _ = Describe("E2E - Test Backup/Restore", Label("test-backup-restore"), func() {
	// Variable(s)
	backupResourceName := "elemental-backup"
	restoreResourceName := "elemental-restore"

	It("Do a backup", func() {
		By("Adding a backup resource", func() {
			err := kubectl.Apply(clusterNS, backupYaml)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Checking that the backup has been done", func() {
			out, err := kubectl.Run("get", "backup", backupResourceName,
				"-o", "jsonpath={.metadata.name}")
			Expect(err).To(Not(HaveOccurred()))
			Expect(out).To(ContainSubstring(backupResourceName))

			// Check operator logs
			Eventually(func() string {
				out, _ := kubectl.Run("logs", "-l app.kubernetes.io/name=rancher-backup",
					"--tail=-1", "--since=5m",
					"--namespace", "cattle-resources-system")
				return out
			}, misc.SetTimeout(5*time.Minute), 10*time.Second).Should(ContainSubstring("Done with backup"))
		})
	})

	It("Do a restore", func() {
		By("Deleting some Elemental resources", func() {
			for _, obj := range []string{"MachineRegistration", "MachineInventorySelectorTemplate"} {
				// List the resources
				list, err := kubectl.Run("get", obj,
					"--namespace", clusterNS,
					"-o", "jsonpath={.items[*].metadata.name}")
				Expect(err).To(Not(HaveOccurred()))

				// Delete the resources
				for _, rsc := range strings.Split(list, " ") {
					_, err := kubectl.Run("delete", obj, "--namespace", clusterNS, rsc)
					Expect(err).To(Not(HaveOccurred()))
				}
			}
		})

		By("Adding a restore resource", func() {
			// Get the backup file from the previous backup
			backupFile, err := kubectl.Run("get", "backup", backupResourceName, "-o", "jsonpath={.status.filename}")
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
				out, _ := kubectl.Run("get", "restore", restoreResourceName,
					"-o", "jsonpath={.metadata.name}")
				return out
			}, misc.SetTimeout(5*time.Minute), 10*time.Second).Should(ContainSubstring(restoreResourceName))

			// Check operator logs
			Eventually(func() string {
				out, _ := kubectl.Run("logs", "-l app.kubernetes.io/name=rancher-backup",
					"--tail=-1", "--since=5m",
					"--namespace", "cattle-resources-system")
				return out
			}, misc.SetTimeout(5*time.Minute), 10*time.Second).Should(ContainSubstring("Done restoring"))
		})

		By("Checking cluster state after restore", func() {
			CheckClusterState(clusterNS, clusterName)
		})
	})
})
