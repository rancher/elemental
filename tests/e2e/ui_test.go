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
	"os/exec"
	"strings"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
)

var _ = Describe("E2E - Bootstrap node for UI", Label("ui"), func() {
	var (
		macAdrs string
		client  *tools.Client
	)

	BeforeEach(func() {
		// Add node in network configuration if needed
		if macAdrs == "" {
			err := misc.AddNode(vmName, vmIndex, netDefaultFileName)
			Expect(err).To(Not(HaveOccurred()))
		}

		hostData, err := tools.GetHostNetConfig(".*name=\""+vmName+"\".*", netDefaultFileName)
		Expect(err).To(Not(HaveOccurred()))

		client = &tools.Client{
			Host:     string(hostData.IP) + ":22",
			Username: userName,
			Password: userPassword,
		}

		macAdrs = hostData.Mac
	})

	It("Configure libvirt and bootstrap a node", func() {
		By("Adding MachineRegistration", func() {
			registrationYaml := "../assets/machineregistration.yaml"

			err := tools.Sed("%VM_NAME%", vmNameRoot, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = tools.Sed("%USER%", userName, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = tools.Sed("%PASSWORD%", userPassword, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = tools.Sed("%CLUSTER_NAME%", clusterName, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = kubectl.Apply(clusterNS, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))

			tokenURL, err := kubectl.Run("get", "MachineRegistration",
				"--namespace", clusterNS,
				"machine-registration", "-o", "jsonpath={.status.registrationURL}")
			Expect(err).To(Not(HaveOccurred()))

			// Get the YAML config file
			fileName := "../../install-config.yaml"
			err = tools.GetFileFromURL(tokenURL, fileName, false)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Starting HTTP server for network installation", func() {
			// TODO: improve it to run in background!
			// err := tools.HTTPShare("../..", 8000)
			// Expect(err).To(Not(HaveOccurred()))

			// Use Python for now...
			err := exec.Command("../scripts/start-httpd").Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Starting default network", func() {
			// Don't check return code, as the default network could be already removed
			cmds := []string{"net-destroy", "net-undefine"}
			for _, c := range cmds {
				_ = exec.Command("sudo", "virsh", c, "default").Run()
			}

			err := exec.Command("sudo", "virsh", "net-create", netDefaultFileName).Run()
			Expect(err).To(Not(HaveOccurred()))
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

		By("Restarting the VM to add it in the cluster", func() {
			err := exec.Command("sudo", "virsh", "start", vmName).Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Checking VM connection", func() {
			id, err := misc.GetServerId(clusterNS, vmIndex)
			Expect(err).To(Not(HaveOccurred()))
			Expect(id).To(Not(BeEmpty()))

			// Retry the SSH connection, as it can takes time for the user to be created
			Eventually(func() string {
				out, _ := client.RunSSH("uname -n")
				out = strings.Trim(out, "\n")
				return out
			}, misc.SetTimeout(2*time.Minute), 5*time.Second)
		})
	})
})
