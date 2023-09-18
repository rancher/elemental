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
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

var _ = Describe("E2E - Configure test", Label("configure"), func() {
	It("Configure Rancher and libvirt", func() {
		type pattern struct {
			key   string
			value string
		}

		// Patterns to replace
		basePatterns := []pattern{
			{
				key:   "%CLUSTER_NAME%",
				value: clusterName,
			},
			{
				key:   "%K8S_VERSION%",
				value: k8sVersion,
			},
		}

		By("Creating a new cluster", func() {
			// Create Yaml file
			for _, p := range basePatterns {
				err := tools.Sed(p.key, p.value, clusterYaml)
				Expect(err).To(Not(HaveOccurred()))
			}

			// Apply to k8s
			err := kubectl.Apply(clusterNS, clusterYaml)
			Expect(err).To(Not(HaveOccurred()))

			// Check that the cluster is correctly created
			CheckCreatedCluster(clusterNS, clusterName)
		})

		By("Creating cluster selectors", func() {
			// Set temporary file
			selectorTmp, err := tools.CreateTemp("selector")
			Expect(err).To(Not(HaveOccurred()))
			defer os.Remove(selectorTmp)

			for _, pool := range []string{"master", "worker"} {
				// Patterns to replace
				addPatterns := []pattern{
					{
						key:   "%POOL_TYPE%",
						value: pool,
					},
				}
				patterns := append(basePatterns, addPatterns...)

				// Save original file as it will have to be modified twice
				err := tools.CopyFile(selectorYaml, selectorTmp)
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
				CheckCreatedSelectorTemplate(clusterNS, "selector-"+pool+"-"+clusterName)
			}
		})

		By("Adding MachineRegistration", func() {
			// Set temporary file
			registrationTmp, err := tools.CreateTemp("machineRegistration")
			Expect(err).To(Not(HaveOccurred()))
			defer os.Remove(registrationTmp)

			for _, pool := range []string{"master", "worker"} {
				// Patterns to replace
				addPatterns := []pattern{
					{
						key:   "%PASSWORD%",
						value: userPassword,
					},
					{
						key:   "%POOL_TYPE%",
						value: pool,
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
				patterns := append(basePatterns, addPatterns...)

				// Save original file as it will have to be modified twice
				err := tools.CopyFile(registrationYaml, registrationTmp)
				Expect(err).To(Not(HaveOccurred()))

				// Create Yaml file
				for _, p := range patterns {
					err := tools.Sed(p.key, p.value, registrationTmp)
					Expect(err).To(Not(HaveOccurred()))
				}

				// Apply to k8s
				err = kubectl.Apply(clusterNS, registrationTmp)
				Expect(err).To(Not(HaveOccurred()))

				// Check that the machine registration is correctly created
				CheckCreatedRegistration(clusterNS, "machine-registration-"+pool+"-"+clusterName)
			}
		})

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
})
