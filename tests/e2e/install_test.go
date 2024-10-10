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
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

func rolloutDeployment(ns, d string) {
	// NOTE: 1st or 2nd rollout command can sporadically fail, so better to use Eventually here
	Eventually(func() string {
		status, _ := kubectl.RunWithoutErr("rollout", "restart", "deployment/"+d,
			"--namespace", ns)
		return status
	}, tools.SetTimeout(1*time.Minute), 20*time.Second).Should(ContainSubstring("restarted"))

	// Wait for deployment to be restarted
	Eventually(func() string {
		status, _ := kubectl.RunWithoutErr("rollout", "status", "deployment/"+d,
			"--namespace", ns)
		return status
	}, tools.SetTimeout(2*time.Minute), 30*time.Second).Should(ContainSubstring("successfully rolled out"))
}

var _ = Describe("E2E - Install Rancher Manager", Label("install"), func() {
	// Create kubectl context
	// Default timeout is too small, so New() cannot be used
	k := &kubectl.Kubectl{
		Namespace:    "",
		PollTimeout:  tools.SetTimeout(300 * time.Second),
		PollInterval: 500 * time.Millisecond,
	}

	// Define local Kubeconfig file
	localKubeconfig := os.Getenv("HOME") + "/.kube/config"

	It("Install upstream K8s cluster", func() {
		if strings.Contains(k8sUpstreamVersion, "rke2") {
			// Report to Qase
			testCaseID = 60

			By("Installing RKE2", func() {
				InstallRKE2()
			})

			if clusterType == "hardened" {
				By("Configuring hardened cluster", func() {
					err := exec.Command("sudo", installHardenedScript).Run()
					Expect(err).To(Not(HaveOccurred()))
				})
			}

			By("Starting RKE2", func() {
				StartRKE2()
			})

			By("Waiting for RKE2 to be started", func() {
				WaitForRKE2(k)
			})

			By("Installing local-path-provisionner", func() {
				InstallLocalStorage(k)
			})
		} else {
			// Report to Qase
			testCaseID = 59

			By("Installing K3s", func() {
				InstallK3s()
			})

			if clusterType == "hardened" {
				By("Configuring hardened cluster", func() {
					err := exec.Command("sudo", installHardenedScript).Run()
					Expect(err).To(Not(HaveOccurred()))
				})
			}

			By("Starting K3s", func() {
				StartK3s()
			})

			By("Waiting for K3s to be started", func() {
				WaitForK3s(k)
			})
		}

		By("Configuring Kubeconfig file", func() {
			// Copy K3s file in ~/.kube/config
			// NOTE: don't check for error, as it will happen anyway (only K3s or RKE2 is installed at a time)
			file, _ := exec.Command("bash", "-c", "ls /etc/rancher/{k3s,rke2}/{k3s,rke2}.yaml").Output()
			Expect(file).To(Not(BeEmpty()))
			err := tools.CopyFile(strings.Trim(string(file), "\n"), localKubeconfig)
			Expect(err).To(Not(HaveOccurred()))

			err = os.Setenv("KUBECONFIG", localKubeconfig)
			Expect(err).To(Not(HaveOccurred()))
		})

		if caType == "private" {
			By("Configuring Private CA", func() {
				out, err := exec.Command(configPrivateCAScript).CombinedOutput()
				GinkgoWriter.Printf("%s\n", out)
				Expect(err).To(Not(HaveOccurred()))
			})
		} else {
			By("Installing CertManager", func() {
				InstallCertManager(k)
			})
		}
	})

	It("Install Rancher Manager", func() {
		// Report to Qase
		testCaseID = 61

		// Inject secret for Private CA
		if caType == "private" {
			// The namespace must exist before adding secret
			err := exec.Command("kubectl", "create", "namespace", "cattle-system").Run()
			Expect(err).To(Not(HaveOccurred()))

			_, err = kubectl.RunWithoutErr("create", "secret",
				"--namespace", "cattle-system",
				"tls", "tls-rancher-ingress",
				"--cert=tls.crt",
				"--key=tls.key",
			)
			Expect(err).To(Not(HaveOccurred()))

			_, err = kubectl.RunWithoutErr("create", "secret",
				"--namespace", "cattle-system",
				"generic", "tls-ca",
				"--from-file=cacerts.pem=./cacerts.pem",
			)
			Expect(err).To(Not(HaveOccurred()))
		}

		InstallRancher(k)

		// Check issuer for Private CA
		if caType == "private" {
			Eventually(func() error {
				out, err := exec.Command("curl", "-vk", "https://"+rancherHostname).CombinedOutput()
				if err != nil {
					// Show only if there's no error
					GinkgoWriter.Printf("%s\n", out)
				}
				return err
			}, tools.SetTimeout(2*time.Minute), 5*time.Second).Should(Not(HaveOccurred()))
		}

		By("Configuring kubectl to use Rancher admin user", func() {
			// Getting internal username for admin
			internalUsername, err := kubectl.RunWithoutErr("get", "user",
				"-o", "jsonpath={.items[?(@.username==\"admin\")].metadata.name}",
			)
			Expect(err).To(Not(HaveOccurred()))
			Expect(internalUsername).To(Not(BeEmpty()))

			// Add token in Rancher Manager
			err = tools.Sed("%ADMIN_USER%", internalUsername, ciTokenYaml)
			Expect(err).To(Not(HaveOccurred()))
			err = kubectl.Apply("default", ciTokenYaml)
			Expect(err).To(Not(HaveOccurred()))

			// Getting Rancher Manager local cluster CA
			// NOTE: loop until the cmd return something, it could take some time
			var rancherCA string
			Eventually(func() error {
				rancherCA, err = kubectl.RunWithoutErr("get", "secret",
					"--namespace", "cattle-system",
					"tls-rancher-ingress",
					"-o", "jsonpath={.data.tls\\.crt}",
				)
				return err
			}, tools.SetTimeout(2*time.Minute), 5*time.Second).Should(Not(HaveOccurred()))

			// Copy skel file for ~/.kube/config
			err = tools.CopyFile(localKubeconfigYaml, localKubeconfig)
			Expect(err).To(Not(HaveOccurred()))

			// Create kubeconfig for local cluster
			err = tools.Sed("%RANCHER_URL%", rancherHostname, localKubeconfig)
			Expect(err).To(Not(HaveOccurred()))
			err = tools.Sed("%RANCHER_CA%", rancherCA, localKubeconfig)
			Expect(err).To(Not(HaveOccurred()))

			// Set correct file permissions
			_ = exec.Command("chmod", "0600", localKubeconfig).Run()

			// Remove the "old" kubeconfig file to force the use of the new one
			// NOTE: in fact move it, just to keep it in case of issue
			// Also don't check the returned error, as it will always not equal 0
			_ = exec.Command("bash", "-c", "sudo mv -f /etc/rancher/{k3s,rke2}/{k3s,rke2}.yaml ~/").Run()
		})
	})

	// Deploy operator in CLI test
	It("Install Elemental Operator if needed", func() {
		if operatorInstallType == "cli" {
			By("Installing Operator with CLI", func() {
				// Report to Qase
				testCaseID = 62

				installOrder := []string{"elemental-operator-crds", "elemental-operator"}
				InstallElementalOperator(k, installOrder, operatorRepo)
			})
		}
	})
})
