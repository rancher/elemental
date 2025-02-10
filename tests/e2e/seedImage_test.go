/*
Copyright © 2022 - 2025 SUSE LLC

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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/elemental"
)

var _ = Describe("E2E - Creating ISO image", Label("iso-image"), func() {
	var (
		machineRegName string
		seedImageName  string
	)

	BeforeEach(func() {
		machineRegName = "machine-registration-" + poolType + "-" + clusterName
		seedImageName = "seed-image-" + poolType + "-" + clusterName
	})

	It("Configure and create ISO image", func() {
		// Report to Qase
		testCaseID = 38

		// If registry keyword is found this means that we have to test a specific OS channel
		if strings.Contains(os2Test, "registry") {
			// Get default channel image
			defChannel, err := kubectl.RunWithoutErr("get", "managedOSVersionChannel",
				"--namespace", clusterNS,
				"-o", "jsonpath={.items[0].spec.options.image}")
			Expect(err).To(Not(HaveOccurred()))
			Expect(defChannel).To(Not(BeEmpty()))

			// Add channel to test if needed
			if !strings.Contains(defChannel, os2Test) {
				By("Adding OSChannel to test", func() {
					// Get channel name (to be able to remove it later)
					channelName, err := kubectl.RunWithoutErr("get", "managedOSVersionChannel",
						"--namespace", clusterNS,
						"-o", "jsonpath={.items[0].metadata.name}")
					Expect(err).To(Not(HaveOccurred()))
					Expect(channelName).To(Not(BeEmpty()))

					// Set temporary file
					osChannelTmp, err := tools.CreateTemp("osChannel")
					Expect(err).To(Not(HaveOccurred()))
					defer os.Remove(osChannelTmp)

					// Save original file as it can be modified multiple time
					err = tools.CopyFile(osChannelYaml, osChannelTmp)
					Expect(err).To(Not(HaveOccurred()))

					// Set the OS channel to test
					err = tools.Sed("%OS_CHANNEL%", os2Test, osChannelTmp)
					Expect(err).To(Not(HaveOccurred()))

					// Apply to k8s
					err = kubectl.Apply(clusterNS, osChannelTmp)
					Expect(err).To(Not(HaveOccurred()))

					// Check that the OS channel to test has been added
					const channel = "os-channel-to-test"
					newChannel, err := kubectl.RunWithoutErr("get", "managedOSVersionChannel",
						"--namespace", clusterNS, channel,
						"-o", "jsonpath={.spec.options.image}")
					Expect(err).To(Not(HaveOccurred()))
					Expect(newChannel).To(Equal(os2Test))

					// Delete the default channel (to be sure that it can't be used)
					out, err := kubectl.RunWithoutErr("delete", "managedOSVersionChannel",
						"--namespace", clusterNS, channelName)
					Expect(err).To(Not(HaveOccurred()))
					Expect(out).To(ContainSubstring("deleted"))
				})
			}

			// Clear OS_TO_TEST, as Staging and Dev channels manually added do not contain "unstable" tag
			os2Test = ""
		}

		By("Adding SeedImage", func() {
			var (
				baseImageURL string
				err          error
				OSVersion    []byte
			)

			if selinux {
				// For now SELinux images are only for testing purposes and not added to any channel, so we should force the value here
				baseImageURL = os2Test
			} else {
				// Wait for list of OS versions to be populated
				WaitForOSVersion(clusterNS)

				// Get OSVersion name
				Eventually(func() string {
					OSVersion, _ = exec.Command(getOSScript, os2Test, "true").Output()
					return string(OSVersion)
				}, tools.SetTimeout(2*time.Minute), 30*time.Second).Should(Not(BeEmpty()))

				// Extract container image URL
				baseImageURL, err = elemental.GetImageURI(clusterNS, string(OSVersion))
				Expect(err).To(Not(HaveOccurred()))
			}

			// We should be sure that we have an image to use
			Expect(baseImageURL).To(Not(BeEmpty()))

			// Set poweroff to false for master pool to have time to check SeedImage cloud-config
			if poolType == "master" && isoBoot {
				_, err := kubectl.RunWithoutErr("patch", "MachineRegistration",
					"--namespace", clusterNS, machineRegName,
					"--type", "merge", "--patch",
					"{\"spec\":{\"config\":{\"elemental\":{\"install\":{\"poweroff\":false}}}}}")
				Expect(err).To(Not(HaveOccurred()))
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
				_, err = kubectl.RunWithoutErr("patch", "MachineRegistration",
					"--namespace", clusterNS, machineRegName,
					"--type", "merge", "--patch-file", emulatedTmp,
				)
				Expect(err).To(Not(HaveOccurred()))
			})

			// Patterns to replace
			patterns := []YamlPattern{
				{
					key:   "%BASE_IMAGE%",
					value: baseImageURL,
				},
				{
					key:   "%CLUSTER_NAME%",
					value: clusterName,
				},
				{
					key:   "%POOL_TYPE%",
					value: poolType,
				},
				{
					key:   "%SSHD_CONFIG_FILE%",
					value: sshdConfigFile,
				},
			}

			// Set temporary file
			seedImageTmp, err := tools.CreateTemp("seedImage")
			Expect(err).To(Not(HaveOccurred()))
			defer os.Remove(seedImageTmp)

			// Save original file as it will have to be modified twice
			err = tools.CopyFile(seedImageYaml, seedImageTmp)
			Expect(err).To(Not(HaveOccurred()))

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

	It("Download ISO built by SeedImage", func() {
		// Report to Qase
		testCaseID = 39

		DownloadBuiltISO(clusterNS, seedImageName, "../../elemental-"+poolType+".iso")
	})
})
