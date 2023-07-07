/*
Copyright © 2023 SUSE LLC

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
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
)

var _ = Describe("E2E - Install a simple application", Label("install-app"), func() {
	It("Install HelloWorld application", func() {
		kubeConfig, err := misc.SetClientKubeConfig(clusterNS, clusterName)
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
		appName := "hello-world"

		// File where to host client cluster kubeconfig
		kubeConfig, err := misc.SetClientKubeConfig(clusterNS, clusterName)
		defer os.Remove(kubeConfig)
		Expect(err).To(Not(HaveOccurred()))

		By("Waiting for deployment to be rollout", func() {
			// Wait for application to be started
			status, err := kubectl.Run("rollout", "status", "deployment/"+appName)
			Expect(err).To(Not(HaveOccurred()))
			Expect(status).To(ContainSubstring("successfully rolled out"))
		})

		By("Checking application", func() {
			// Get load balancer IPs
			appIPs, err := kubectl.Run("get", "svc",
				appName+"-loadbalancer",
				"-o", "jsonpath={.status.loadBalancer.ingress[*].ip}")
			Expect(err).To(Not(HaveOccurred()))
			Expect(appIPs).To(Not(BeEmpty()))

			// Loop on each IP to check the application
			for _, ip := range strings.Fields(appIPs) {
				GinkgoWriter.Printf("Checking node with IP %s...\n", ip)

				// Retry if needed, could take some times if a pod is restarted for example
				var htmlPage []byte
				Eventually(func() error {
					htmlPage, err = exec.Command("curl", "http://"+ip+":8080").CombinedOutput()
					return err
				}, misc.SetTimeout(2*time.Minute), 5*time.Second).Should(Not(HaveOccurred()))

				// Check HTML page content
				Expect(string(htmlPage)).To(And(
					ContainSubstring("Hello world!"),
					ContainSubstring("My hostname is hello-world-"),
					ContainSubstring(ip+":8080"),
				), string(htmlPage))
			}
		})
	})
})
