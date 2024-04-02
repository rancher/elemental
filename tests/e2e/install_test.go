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
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
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
				// Get RKE2 installation script
				fileName := "rke2-install.sh"
				Eventually(func() error {
					return tools.GetFileFromURL("https://get.rke2.io", fileName, true)
				}, tools.SetTimeout(2*time.Minute), 10*time.Second).ShouldNot(HaveOccurred())

				// Retry in case of (sporadic) failure...
				count := 1
				Eventually(func() error {
					// Execute RKE2 installation
					out, err := exec.Command("sudo", "--preserve-env=INSTALL_RKE2_VERSION", "sh", fileName).CombinedOutput()
					GinkgoWriter.Printf("RKE2 installation loop %d:\n%s\n", count, out)
					count++
					return err
				}, tools.SetTimeout(2*time.Minute), 5*time.Second).Should(BeNil())
			})

			if clusterType == "hardened" {
				By("Configuring hardened cluster", func() {
					err := exec.Command("sudo", installHardenedScript).Run()
					Expect(err).To(Not(HaveOccurred()))
				})
			}

			By("Starting RKE2", func() {
				// Copy config file, this allows custom configuration for RKE2 installation
				// NOTE: CopyFile cannot be used, as we need root permissions for this file
				err := exec.Command("sudo", "mkdir", "-p", "/etc/rancher/rke2").Run()
				Expect(err).To(Not(HaveOccurred()))
				err = exec.Command("sudo", "cp", configRKE2Yaml, "/etc/rancher/rke2/config.yaml").Run()
				Expect(err).To(Not(HaveOccurred()))

				// Activate and start RKE2
				err = exec.Command("sudo", "systemctl", "enable", "--now", "rke2-server.service").Run()
				Expect(err).To(Not(HaveOccurred()))

				// Delay few seconds before checking
				time.Sleep(tools.SetTimeout(20 * time.Second))

				err = exec.Command("sudo", "ln", "-s", "/var/lib/rancher/rke2/bin/kubectl", "/usr/local/bin/kubectl").Run()
				Expect(err).To(Not(HaveOccurred()))
			})

			By("Waiting for RKE2 to be started", func() {
				// Wait for all pods to be started
				err := os.Setenv("KUBECONFIG", "/etc/rancher/rke2/rke2.yaml")
				Expect(err).To(Not(HaveOccurred()))

				checkList := [][]string{
					{"kube-system", "k8s-app=kube-dns"},
					{"kube-system", "app.kubernetes.io/name=rke2-ingress-nginx"},
				}
				Eventually(func() error {
					return rancher.CheckPod(k, checkList)
				}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(BeNil())

				err = k.WaitLabelFilter("kube-system", "Ready", "rke2-ingress-nginx-controller", "app.kubernetes.io/name=rke2-ingress-nginx")
				Expect(err).To(Not(HaveOccurred()))
			})
		} else {
			// Report to Qase
			testCaseID = 59

			By("Installing K3s", func() {
				// Get K3s installation script
				fileName := "k3s-install.sh"
				Eventually(func() error {
					return tools.GetFileFromURL("https://get.k3s.io", fileName, true)
				}, tools.SetTimeout(2*time.Minute), 10*time.Second).ShouldNot(HaveOccurred())

				// Set command and arguments
				installCmd := exec.Command("sh", fileName)
				installCmd.Env = append(os.Environ(), "INSTALL_K3S_EXEC=--disable metrics-server")

				// Retry in case of (sporadic) failure...
				count := 1
				Eventually(func() error {
					// Execute K3s installation
					out, err := installCmd.CombinedOutput()
					GinkgoWriter.Printf("K3s installation loop %d:\n%s\n", count, out)
					count++
					return err
				}, tools.SetTimeout(2*time.Minute), 5*time.Second).Should(BeNil())
			})

			if clusterType == "hardened" {
				By("Configuring hardened cluster", func() {
					err := exec.Command("sudo", installHardenedScript).Run()
					Expect(err).To(Not(HaveOccurred()))
				})
			}

			By("Starting K3s", func() {
				err := exec.Command("sudo", "systemctl", "start", "k3s").Run()
				Expect(err).To(Not(HaveOccurred()))

				// Delay few seconds before checking
				time.Sleep(tools.SetTimeout(20 * time.Second))
			})

			By("Waiting for K3s to be started", func() {
				// Wait for all pods to be started
				checkList := [][]string{
					{"kube-system", "app=local-path-provisioner"},
					{"kube-system", "k8s-app=kube-dns"},
					{"kube-system", "app.kubernetes.io/name=traefik"},
					{"kube-system", "svccontroller.k3s.cattle.io/svcname=traefik"},
				}
				Eventually(func() error {
					return rancher.CheckPod(k, checkList)
				}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(BeNil())
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
				RunHelmCmdWithRetry("repo", "add", "jetstack", "https://charts.jetstack.io")
				RunHelmCmdWithRetry("repo", "update")

				// Set flags for cert-manager installation
				flags := []string{
					"upgrade", "--install", "cert-manager", "jetstack/cert-manager",
					"--namespace", "cert-manager",
					"--create-namespace",
					"--set", "installCRDs=true",
					"--wait", "--wait-for-jobs",
				}

				if clusterType == "hardened" {
					flags = append(flags, "--version", certManagerVersion)
				}

				RunHelmCmdWithRetry(flags...)

				checkList := [][]string{
					{"cert-manager", "app.kubernetes.io/component=controller"},
					{"cert-manager", "app.kubernetes.io/component=webhook"},
					{"cert-manager", "app.kubernetes.io/component=cainjector"},
				}
				Eventually(func() error {
					return rancher.CheckPod(k, checkList)
				}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(BeNil())
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

		err := rancher.DeployRancherManager(rancherHostname, rancherChannel, rancherVersion, rancherHeadVersion, caType, proxy)
		Expect(err).To(Not(HaveOccurred()))

		// Wait for all pods to be started
		checkList := [][]string{
			{"cattle-system", "app=rancher"},
			{"cattle-system", "app=rancher-webhook"},
			{"cattle-fleet-local-system", "app=fleet-agent"},
			{"cattle-provisioning-capi-system", "control-plane=controller-manager"},
		}
		Eventually(func() error {
			return rancher.CheckPod(k, checkList)
		}, tools.SetTimeout(10*time.Minute), 30*time.Second).Should(BeNil())

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

		if testType == "ui" {
			By("Workaround for upgrade test, restart Fleet controller and agent", func() {
				for _, d := range [][]string{
					{"cattle-fleet-local-system", "fleet-agent"},
					{"cattle-fleet-system", "fleet-controller"},
				} {
					rolloutDeployment(d[0], d[1])
				}
			})
		}
	})

	// Deploy operator in CLI test
	It("Install Elemental Operator if needed", func() {
		if testType == "cli" || testType == "multi" {
			By("Installing Operator for CLI tests", func() {
				// Report to Qase
				testCaseID = 62

				for _, chart := range []string{"elemental-operator-crds", "elemental-operator"} {
					RunHelmCmdWithRetry("upgrade", "--install", chart,
						operatorRepo+"/"+chart+"-chart",
						"--namespace", "cattle-elemental-system",
						"--create-namespace",
						"--wait", "--wait-for-jobs",
					)

					// Delay few seconds for all to be installed
					time.Sleep(tools.SetTimeout(20 * time.Second))
				}

				// Wait for pod to be started
				Eventually(func() error {
					return rancher.CheckPod(k, [][]string{{"cattle-elemental-system", "app=elemental-operator"}})
				}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(BeNil())
			})
		}
	})
})
