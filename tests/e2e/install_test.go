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
		status, _ := kubectl.Run("rollout", "restart", "deployment/"+d,
			"--namespace", ns)
		return status
	}, tools.SetTimeout(1*time.Minute), 20*time.Second).Should(ContainSubstring("restarted"))

	// Wait for deployment to be restarted
	Eventually(func() string {
		status, _ := kubectl.Run("rollout", "status", "deployment/"+d,
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

	It("Install Rancher Manager", func() {
		if strings.Contains(k8sUpstreamVersion, "rke2") {
			By("Installing RKE2", func() {
				// Get RKE2 installation script
				fileName := "rke2-install.sh"
				err := tools.GetFileFromURL("https://get.rke2.io", fileName, true)
				Expect(err).To(Not(HaveOccurred()))

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
				err := exec.Command("sudo", "systemctl", "enable", "--now", "rke2-server.service").Run()
				Expect(err).To(Not(HaveOccurred()))

				// Delay few seconds before checking
				time.Sleep(tools.SetTimeout(20 * time.Second))

				err = exec.Command("sudo", "chown", "gh-runner", "/etc/rancher/rke2/rke2.yaml").Run()
				Expect(err).To(Not(HaveOccurred()))

				err = exec.Command("sudo", "ln", "-s", "/var/lib/rancher/rke2/bin/kubectl", "/usr/local/bin/kubectl").Run()
				Expect(err).To(Not(HaveOccurred()))
			})

			By("Waiting for RKE2 to be started", func() {
				// Wait for all pods to be started
				err := os.Setenv("KUBECONFIG", "/etc/rancher/rke2/rke2.yaml")
				Expect(err).To(Not(HaveOccurred()))

				checkList := [][]string{
					{"kube-system", "k8s-app=kube-dns"},
					{"kube-system", "app=rke2-metrics-server"},
					{"kube-system", "app.kubernetes.io/name=rke2-ingress-nginx"},
				}
				err = rancher.CheckPod(k, checkList)
				Expect(err).To(Not(HaveOccurred()))

				err = k.WaitLabelFilter("kube-system", "Ready", "rke2-ingress-nginx-controller", "app.kubernetes.io/name=rke2-ingress-nginx")
				Expect(err).To(Not(HaveOccurred()))
			})
		} else {
			By("Installing K3s", func() {
				// Get K3s installation script
				fileName := "k3s-install.sh"
				err := tools.GetFileFromURL("https://get.k3s.io", fileName, true)
				Expect(err).To(Not(HaveOccurred()))

				// Retry in case of (sporadic) failure...
				count := 1
				Eventually(func() error {
					// Execute K3s installation
					out, err := exec.Command("sh", fileName).CombinedOutput()
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
					{"kube-system", "k8s-app=metrics-server"},
					{"kube-system", "app.kubernetes.io/name=traefik"},
					{"kube-system", "svccontroller.k3s.cattle.io/svcname=traefik"},
				}
				err := rancher.CheckPod(k, checkList)
				Expect(err).To(Not(HaveOccurred()))
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
					flags = append(flags, "--version", CertManagerVersion)
				}

				RunHelmCmdWithRetry(flags...)

				checkList := [][]string{
					{"cert-manager", "app.kubernetes.io/component=controller"},
					{"cert-manager", "app.kubernetes.io/component=webhook"},
					{"cert-manager", "app.kubernetes.io/component=cainjector"},
				}
				err := rancher.CheckPod(k, checkList)
				Expect(err).To(Not(HaveOccurred()))
			})
		}

		By("Installing Rancher Manager", func() {
			err := rancher.DeployRancherManager(rancherHostname, rancherChannel, rancherVersion, caType, proxy)
			Expect(err).To(Not(HaveOccurred()))

			// Inject secret for Private CA
			if caType == "private" {
				_, err := kubectl.Run("create", "secret",
					"--namespace", "cattle-system",
					"tls", "tls-rancher-ingress",
					"--cert=tls.crt",
					"--key=tls.key",
				)
				Expect(err).To(Not(HaveOccurred()))

				_, err = kubectl.Run("create", "secret",
					"--namespace", "cattle-system",
					"generic", "tls-ca",
					"--from-file=cacerts.pem=./cacerts.pem",
				)
				Expect(err).To(Not(HaveOccurred()))
			}

			// Wait for all pods to be started
			checkList := [][]string{
				{"cattle-system", "app=rancher"},
				{"cattle-fleet-local-system", "app=fleet-agent"},
				{"cattle-system", "app=rancher-webhook"},
			}
			err = rancher.CheckPod(k, checkList)
			Expect(err).To(Not(HaveOccurred()))

			// We have to restart Rancher Manager to be sure that Private CA is used
			if caType == "private" {
				rolloutDeployment("cattle-system", "rancher")
			}

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
		})

		By("Configuring kubectl to use Rancher admin user", func() {
			// Getting internal username for admin
			internalUsername, err := kubectl.Run("get", "user",
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
				rancherCA, err = kubectl.Run("get", "secret",
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
				// https://github.com/rancher/elemental/issues/410
				time.Sleep(tools.SetTimeout(120 * time.Second))
				_, err := kubectl.Run("scale", "deployment/fleet-agent", "-n", "cattle-fleet-local-system", "--replicas=0")
				Expect(err).To(Not(HaveOccurred()))
				_, err = kubectl.Run("scale", "deployment/fleet-controller", "-n", "cattle-fleet-system", "--replicas=0")
				Expect(err).To(Not(HaveOccurred()))
				_, err = kubectl.Run("scale", "deployment/fleet-controller", "-n", "cattle-fleet-system", "--replicas=1")
				Expect(err).To(Not(HaveOccurred()))
				_, err = kubectl.Run("scale", "deployment/fleet-agent", "-n", "cattle-fleet-local-system", "--replicas=1")
				Expect(err).To(Not(HaveOccurred()))
			})
		}

		By("Installing Elemental Operator", func() {
			for _, chart := range []string{"elemental-operator-crds", "elemental-operator"} {
				RunHelmCmdWithRetry("upgrade", "--install", chart,
					operatorRepo+"/"+chart+"-chart",
					"--namespace", "cattle-elemental-system",
					"--create-namespace",
					"--wait", "--wait-for-jobs",
				)
			}

			// Wait for pod to be started
			err := rancher.CheckPod(k, [][]string{{"cattle-elemental-system", "app=elemental-operator"}})
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})
