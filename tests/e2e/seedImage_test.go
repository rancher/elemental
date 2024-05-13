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

		By("Adding SeedImage", func() {
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
		})
	})

	It("Download ISO built by SeedImage", func() {
		// Report to Qase
		testCaseID = 39

		DownloadBuiltISO(clusterNS, seedImageName, "../../elemental-"+poolType+".iso")
	})
})
