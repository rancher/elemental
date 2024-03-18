/*
Copyright © 2024 SUSE LLC

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

		By("Scaling the deployment to the number of nodes", func() {
			var nodeList string
			Eventually(func() string {
				nodeList, _ = kubectl.RunWithoutErr("get", "nodes", "-o", "jsonpath={.items[*].metadata.name}")
				return nodeList
			}, tools.SetTimeout(2*time.Minute), 30*time.Second).Should(Not(BeEmpty()))

			nodeNumber := len(strings.Fields(nodeList))
			Expect(nodeNumber).To(Not(BeNil()))

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
			cmd := []string{
				"get", "svc",
				appName + "-loadbalancer",
				"-o", "jsonpath={.status.loadBalancer.ingress[*].ip}",
			}

			// Wait until at least an IP address is returned
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
