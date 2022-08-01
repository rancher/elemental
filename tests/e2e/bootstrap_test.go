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
	"fmt"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
)

func checkClusterAgent(client *tools.Client) {
	// cluster-agent is the pod that communicates to Rancher, wait for it before continuing
	Eventually(func() string {
		out, _ := client.RunSSH("kubectl get pod -n cattle-system -l app=cattle-cluster-agent")
		return out
	}, "5m", "30s").Should(ContainSubstring("Running"))
}

func checkClusterState() {
	// Check that a 'type' property named 'Ready' is set to true
	Eventually(func() string {
		clusterStatus, _ := kubectl.Run("get", "cluster",
			"--namespace", clusterNS, clusterName,
			"-o", "jsonpath={.status.conditions[?(@.type==\"Ready\")].status}")
		return clusterStatus
	}, "5m", "10s").Should(Equal("True"))

	// Wait a little bit for the cluster to be in a stable state
	time.Sleep(2 * time.Minute)

	// There should be no 'reason' property set in a clean cluster
	reason, err := kubectl.Run("get", "cluster",
		"--namespace", clusterNS, clusterName,
		"-o", "jsonpath={.status.conditions[*].reason}")
	Expect(err).To(Not(HaveOccurred()))
	Expect(reason).To(BeEmpty())
}

var _ = Describe("E2E - Bootstrapping node", Label("bootstrap"), func() {
	var (
		client  *tools.Client
		macAdrs string
	)

	BeforeEach(func() {
		hostData, err := tools.GetHostNetConfig(".*name='"+vmName+"'.*", netDefaultFileName)
		Expect(err).To(Not(HaveOccurred()))

		client = &tools.Client{
			Host:     string(hostData.IP) + ":22",
			Username: userName,
			Password: userPassword,
		}
		macAdrs = hostData.Mac
	})

	It("Install node and add it in Rancher", func() {
		By("Checking if VM name is set", func() {
			Expect(vmName).To(Not(BeEmpty()))
		})

		By("Configuring iPXE boot script for network installation", func() {
			numberOfFile, err := misc.ConfigureiPXE()
			Expect(err).To(Not(HaveOccurred()))
			Expect(numberOfFile).To(BeNumerically(">=", 1))
		})

		By("Creating and installing VM", func() {
			// Install VM
			cmd := exec.Command("../scripts/install-vm", vmName, macAdrs)
			out, err := cmd.CombinedOutput()
			GinkgoWriter.Printf("%s\n", out)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Checking that the VM is available in Rancher", func() {
			id, err := misc.GetServerId(clusterNS, vmIndex)
			Expect(err).To(Not(HaveOccurred()))
			Expect(id).To(Not(BeEmpty()))
		})

		By("Increasing 'quantity' node to predefined cluster", func() {
			// Patch the already-created yaml file directly
			err := tools.Sed("quantity:.*", "quantity: "+fmt.Sprint(vmIndex), clusterYaml)
			Expect(err).To(Not(HaveOccurred()))

			//kubectl patch cluster cluster-k3s -n fleet-default --type merge --patch-file patch.yaml
			out, err := kubectl.Run("patch", "cluster",
				"--namespace", clusterNS, clusterName,
				"--type", "merge", "--patch-file", clusterYaml,
			)
			Expect(err).To(Not(HaveOccurred()), out)
		})

		By("Restarting the VM to add it in the cluster", func() {
			err := exec.Command("virsh", "start", vmName).Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Checking VM connection and cluster state", func() {
			/* Disable this check for now, until https://github.com/rancher/elemental-operator/issues/90 is fixed!
			// Retry the SSH connection, as it can takes time for the user to be created
			Eventually(func() string {
				out, _ := client.RunSSH("uname -n")
				return out
			}, "5m", "5s").Should(ContainSubstring(vmNameRoot))
			*/

			// Check agent and cluster state
			checkClusterAgent(client)
			checkClusterState()
		})

		By("Rebooting the VM and checking that cluster is still healthy after", func() {
			// Execute 'reboot' in background, to avoid ssh locking
			_, err := client.RunSSH("setsid -f reboot")
			Expect(err).To(Not(HaveOccurred()))

			// Wait a little bit for the cluster to be in an unstable state (yes!)
			time.Sleep(2 * time.Minute)

			// Check agent and cluster state
			checkClusterAgent(client)
			checkClusterState()
		})
	})
})
