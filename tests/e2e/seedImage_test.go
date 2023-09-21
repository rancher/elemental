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
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/elemental"
)

var _ = Describe("E2E - Creating ISO image", Label("iso-image"), func() {
	It("Configure and create ISO image", func() {
		// Globales variables
		machineRegName := "machine-registration-" + poolType + "-" + clusterName
		seedImageName := "seed-image-" + poolType + "-" + clusterName

		By("Adding SeedImage", func() {
			type pattern struct {
				key   string
				value string
			}

			// Wait for list of OS versions to be populated
			Eventually(func() string {
				out, _ := kubectl.Run("get", "ManagedOSVersion",
					"--namespace", clusterNS,
					"-o", "jsonpath={.items[*].metadata.name}")
				return out
			}, tools.SetTimeout(2*time.Minute), 5*time.Second).Should(Not(BeEmpty()))

			// Get OSVersion name
			OSVersion, err := exec.Command(getOSScript, os2Test, "true").Output()
			Expect(err).To(Not(HaveOccurred()))
			Expect(OSVersion).To(Not(BeEmpty()))

			// Extract container image URL
			baseImageURL, err := elemental.GetImageURI(clusterNS, string(OSVersion))
			Expect(err).To(Not(HaveOccurred()))
			Expect(baseImageURL).To(Not(BeEmpty()))

			// Set poweroff to false for master pool to have time to check SeedImage cloud-config
			if poolType == "master" && isoBoot == "true" {
				_, err := kubectl.Run("patch", "MachineRegistration",
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
				_, err = kubectl.Run("patch", "MachineRegistration",
					"--namespace", clusterNS, machineRegName,
					"--type", "merge", "--patch-file", emulatedTmp,
				)
				Expect(err).To(Not(HaveOccurred()))
			})

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
	})
})
