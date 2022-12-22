/*
Copyright © 2022 SUSE LLC

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

var _ = Describe("E2E - Upgrading node", Label("upgrade"), func() {
	It("Upgrade node", func() {
		hostData, err := tools.GetHostNetConfig(".*name=\""+vmName+"\".*", netDefaultFileName)
		Expect(err).To(Not(HaveOccurred()))

		client := &tools.Client{
			Host:     string(hostData.IP) + ":22",
			Username: userName,
			Password: userPassword,
		}

		By("Checking if upgrade type is set", func() {
			Expect(upgradeType).To(Not(BeEmpty()))
		})

		By("Showing OS version before upgrade", func() {
			out, err := client.RunSSH("cat /etc/os-release")
			Expect(err).To(Not(HaveOccurred()))
			GinkgoWriter.Printf("OS Version:\n%s\n", out)
		})

		if upgradeType != "manual" {
			By("Triggering Upgrade in Rancher with "+upgradeType, func() {
				upgradeOsYaml := upgradeClusterTargetsYaml
				upgradeTypeValue := osImage // Default to osImage
				if upgradeType == "managedOSVersionName" {
					upgradeTypeValue = imageVersion
				}

				// We should have a version defined
				Expect(upgradeTypeValue).NotTo(BeNil())

				// We don't know what is the previous type of upgrade, so easier to replace all here
				// as there is only one in the yaml file anyway
				for _, p := range []string{"%OS_IMAGE%", "osImage:.*", "managedOSVersionName:.*"} {
					err := tools.Sed(p, upgradeType+": "+upgradeTypeValue, upgradeOsYaml)
					Expect(err).To(Not(HaveOccurred()))
				}

				err := tools.Sed("%CLUSTER_NAME%", clusterName, upgradeClusterTargetsYaml)
				Expect(err).To(Not(HaveOccurred()))

				if upgradeType == "managedOSVersionName" {
					// Get elemental-operator version
					operatorVersion, err := misc.GetOperatorVersion()
					Expect(err).To(Not(HaveOccurred()))
					operatorVersionShort := strings.Split(operatorVersion, ".")

					// Remove 'syncInterval' option if needed (only supported in operator v1.1+)
					if (operatorVersionShort[0] + "." + operatorVersionShort[1]) == "1.0" {
						err = tools.Sed("syncInterval:.*", "", osListYaml)
						Expect(err).To(Not(HaveOccurred()))
					}

					// Add OS list
					err = kubectl.Apply(clusterNS, osListYaml)
					Expect(err).To(Not(HaveOccurred()))

					// Wait for ManagedOSVersion to be populated from ManagedOSVersionChannel
					Eventually(func() string {
						out, _ := kubectl.Run("get", "ManagedOSVersion",
							"--namespace", clusterNS, imageVersion)
						return out
					}, misc.SetTimeout(2*time.Minute), 10*time.Second).Should(Not(ContainSubstring("Error")))

					// Get *REAL* hostname
					hostname, err := client.RunSSH("hostname")
					Expect(err).To(Not(HaveOccurred()))
					hostname = strings.Trim(hostname, "\n")

					label := "kubernetes.io/hostname"
					selector, err := misc.AddSelector(label, hostname)
					Expect(err).To(Not(HaveOccurred()), selector)

					// Create new file for this specific upgrade
					err = misc.ConcateFiles(upgradeClusterTargetsYaml, upgradeOSVersionNameYaml, selector)
					Expect(err).To(Not(HaveOccurred()), selector)

					// Swap yaml file
					upgradeOsYaml = upgradeOSVersionNameYaml

					// Set correct value for os osImage
					out, err := kubectl.Run("get", "ManagedOSVersion",
						"--namespace", clusterNS, imageVersion,
						"-o", "jsonpath={.spec.metadata.upgradeImage}")
					Expect(err).To(Not(HaveOccurred()))
					osImage = misc.TrimStringFromChar(out, ":")
				}

				err = kubectl.Apply(clusterNS, upgradeOsYaml)
				Expect(err).To(Not(HaveOccurred()))
			})
		}

		if upgradeType == "manual" {
			By("Triggering Manual Upgrade", func() {
				out, err := client.RunSSH("elemental upgrade --system.uri docker:" + osImage)
				Expect(err).To(Not(HaveOccurred()), out)
				Expect(out).To((ContainSubstring("Upgrade completed")))

				// Execute 'reboot' in background, to avoid ssh locking
				_, err = client.RunSSH("setsid -f reboot")
				Expect(err).To(Not(HaveOccurred()))
			})
		}

		By("Checking VM upgrade", func() {
			Eventually(func() string {
				// Use grep here in case of comment in the file!
				out, _ := client.RunSSH("eval $(grep -v ^# /etc/os-release) && echo ${IMAGE}")
				out = strings.Trim(out, "\n")

				// Re-format the output if needed
				if upgradeType == "managedOSVersionName" {
					// NOTE: this remove the version and keep only the repo,
					// as 'latest' is set and in the file we have the exact version
					out = misc.TrimStringFromChar(out, ":")
				}

				return out
			}, misc.SetTimeout(5*time.Minute), 30*time.Second).Should(Equal(osImage))
		})

		By("Showing OS version after upgrade", func() {
			out, err := client.RunSSH("cat /etc/os-release")
			Expect(err).To(Not(HaveOccurred()))
			GinkgoWriter.Printf("OS Version:\n%s\n", out)
		})

		if upgradeType != "manual" {
			By("Cleaning upgrade orders", func() {
				if upgradeType == "managedOSVersionName" {
					err := kubectl.DeleteResource(clusterNS, "ManagedOSVersionChannel", "os-versions")
					Expect(err).To(Not(HaveOccurred()))
				}

				err := kubectl.DeleteResource(clusterNS, "ManagedOSImage", "default-os-image")
				Expect(err).To(Not(HaveOccurred()))
			})
		}
	})
})
