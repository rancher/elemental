/*
Copyright © 2025 SUSE LLC

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
	"fmt"
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

var _ = Describe("E2E - Install a simple application", Label("install-app"), func() {
	It("Install HelloWorld application", func() {
		// Report to Qase
		testCaseID = 31

		kubeConfig, err := rancher.SetClientKubeConfig(clusterNS, clusterName)
		defer os.Remove(kubeConfig)
		Expect(err).To(Not(HaveOccurred()))
		Expect(kubeConfig).To(Not(BeEmpty()))

		if strings.Contains(k8sDownstreamVersion, "rke2") {
			// Create kubectl context
			// Default timeout is too small, so New() cannot be used
			k := &kubectl.Kubectl{
				Namespace:    "",
				PollTimeout:  tools.SetTimeout(300 * time.Second),
				PollInterval: 500 * time.Millisecond,
			}

			By("Installing local-path-provisionner", func() {
				localPathNS := "kube-system"
				kubectl.Apply(localPathNS, localStorageYaml)

				// Wait for all pods to be started
				checkList := [][]string{
					{localPathNS, "app=local-path-provisioner"},
				}
				Eventually(func() error {
					return rancher.CheckPod(k, checkList)
				}, tools.SetTimeout(2*time.Minute), 30*time.Second).Should(Not(HaveOccurred()))
			})

			By("Installing MetalLB", func() {
				metallbNS := "metallb-system"

				RunHelmCmdWithRetry("repo", "add", "metallb", "https://metallb.github.io/metallb")
				RunHelmCmdWithRetry("repo", "update")

				flags := []string{
					"upgrade", "--install", "metallb", "metallb/metallb",
					"--namespace", metallbNS,
					"--create-namespace",
					"--wait", "--wait-for-jobs",
				}
				RunHelmCmdWithRetry(flags...)

				// Wait for all pods to be started
				checkList := [][]string{
					{metallbNS, "app.kubernetes.io/component=speaker"},
					{metallbNS, "app.kubernetes.io/component=controller"},
					{metallbNS, "app.kubernetes.io/instance=metallb"},
					{metallbNS, "app.kubernetes.io/name=metallb"},
				}
				Eventually(func() error {
					return rancher.CheckPod(k, checkList)
				}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(Not(HaveOccurred()))

				err := kubectl.Apply(metallbNS, metallbRscYaml)
				Expect(err).NotTo(HaveOccurred())
			})

			By("Installing Traefik", func() {
				traefikNS := "traefik-system"

				RunHelmCmdWithRetry("repo", "add", "traefik", "https://traefik.github.io/charts")
				RunHelmCmdWithRetry("repo", "update")

				flags := []string{
					"upgrade", "--install", "traefik", "traefik/traefik",
					"--namespace", traefikNS,
					"--create-namespace",
					"--set", "ports.web.redirections.entryPoint.to=websecure",
					"--set", "ports.web.redirections.entryPoint.scheme=https",
					"--set", "ingressClass.enabled=true",
					"--set", "ingressClass.isDefaultClass=true",
					"--wait", "--wait-for-jobs",
				}
				RunHelmCmdWithRetry(flags...)

				// Wait for all pods to be started
				checkList := [][]string{
					{traefikNS, "app.kubernetes.io/name=traefik"},
				}
				Eventually(func() error {
					return rancher.CheckPod(k, checkList)
				}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(Not(HaveOccurred()))
			})

			By("Checking LoadBalancer IP", func() {
				traefikNS := "traefik-system"

				// Ensure that Traefik LB is not in Pending state anymore, could take time
				Eventually(func() string {
					out, _ := kubectl.RunWithoutErr("get", "svc", "--namespace", traefikNS, "traefik")
					return out
				}, tools.SetTimeout(4*time.Minute), 4*time.Second).Should(Not(ContainSubstring("<pending>")))

				// Check that an IP address for LB is configured
				lbIP, err := kubectl.Run("get", "svc", "--namespace", traefikNS, "traefik", "-o", "jsonpath={.status.loadBalancer.ingress[0].ip}")
				Expect(err).NotTo(HaveOccurred())
				Expect(lbIP).To(Not(BeEmpty()))
			})
		}

		By("Installing application", func() {
			err := kubectl.Apply("default", appYaml)
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})

var _ = Describe("E2E - Checking a simple application", Label("check-app"), func() {
	It("Check HelloWorld application", func() {
		// Report to Qase
		testCaseID = 63

		appName := "hello-world"

		// File where to host client cluster kubeconfig
		kubeConfig, err := rancher.SetClientKubeConfig(clusterNS, clusterName)
		defer os.Remove(kubeConfig)
		Expect(err).To(Not(HaveOccurred()))
		Expect(kubeConfig).To(Not(BeEmpty()))

		By("Scaling the deployment to the number of nodes", func() {
			var nodeList string
			Eventually(func() string {
				nodeList, _ = kubectl.RunWithoutErr("get", "nodes", "-o", "jsonpath={.items[*].metadata.name}")
				return nodeList
			}, tools.SetTimeout(2*time.Minute), 30*time.Second).Should(Not(BeEmpty()))

			nodeNumber := len(strings.Fields(nodeList))
			Expect(nodeNumber).To(Not(BeZero()))

			out, err := kubectl.RunWithoutErr("scale", "--replicas="+fmt.Sprint(nodeNumber), "deployment/"+appName)
			Expect(err).To(Not(HaveOccurred()), out)
			Expect(out).To(ContainSubstring("deployment.apps/" + appName + " scaled"))
		})

		By("Waiting for deployment to be rollout", func() {
			// Wait for application to be started
			// NOTE: 1st or 2nd rollout command can sporadically fail, so better to use Eventually here
			Eventually(func() string {
				status, _ := kubectl.RunWithoutErr("rollout", "status", "deployment/"+appName)
				return status
			}, tools.SetTimeout(2*time.Minute), 30*time.Second).Should(ContainSubstring("successfully rolled out"))
		})

		By("Checking application", func() {
			// Ensure that LB is not in Pending state anymore, could take time
			Eventually(func() string {
				out, _ := kubectl.RunWithoutErr("get", "svc", appName+"-loadbalancer")
				return out
			}, tools.SetTimeout(4*time.Minute), 4*time.Second).Should(Not(ContainSubstring("<pending>")))

			// Wait until at least an IP address is returned
			cmd := []string{
				"get", "svc",
				appName + "-loadbalancer",
				"-o", "jsonpath={.status.loadBalancer.ingress[*].ip}",
			}

			Eventually(func() bool {
				ip, _ := kubectl.RunWithoutErr(cmd...)
				return tools.IsIPv4(strings.Fields(ip)[0])
			}, tools.SetTimeout(2*time.Minute), 5*time.Second).Should(BeTrue())

			// Get load balancer IPs
			appIPs, err := kubectl.RunWithoutErr(cmd...)
			Expect(err).To(Not(HaveOccurred()))
			Expect(appIPs).To(Not(BeEmpty()))

			// Loop on each IP to check the application
			for _, ip := range strings.Fields(appIPs) {
				if tools.IsIPv4(ip) {
					GinkgoWriter.Printf("Checking node with IP %s...\n", ip)

					// Retry if needed, could take some times if a pod is restarted for example
					var htmlPage []byte
					Eventually(func() error {
						htmlPage, err = exec.Command("curl", "http://"+ip+":8080").CombinedOutput()
						return err
					}, tools.SetTimeout(2*time.Minute), 5*time.Second).Should(Not(HaveOccurred()))

					// Check HTML page content
					Expect(string(htmlPage)).To(And(
						ContainSubstring("Hello world!"),
						ContainSubstring("My hostname is hello-world-"),
						ContainSubstring(ip+":8080"),
					), string(htmlPage))
				}
			}
		})
	})
})
