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
	"regexp"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

var _ = Describe("E2E - Build the airgap archive", Label("prepare-archive"), func() {
	It("Execute the script to build the archive", func() {
		// Force to latest if nothing is defined
		if certManagerVersion == "" {
			certManagerVersion = "latest"
		}

		// Could be useful for manual debugging!
		GinkgoWriter.Printf("Executed command: %s %s %s %s %s %s\n", airgapBuildScript, k8sUpstreamVersion, certManagerVersion, rancherChannel, k8sDownstreamVersion, operatorRepo)
		out, err := exec.Command(airgapBuildScript, k8sUpstreamVersion, certManagerVersion, rancherChannel, k8sDownstreamVersion, operatorRepo).CombinedOutput()
		Expect(err).To(Not(HaveOccurred()), string(out))
	})
})

var _ = Describe("E2E - Deploy K3S/Rancher in airgap environment", Label("airgap-rancher"), func() {
	It("Create the rancher-manager machine", func() {
		By("Updating the default network configuration", func() {
			// Don't check return code, as the default network could be already removed
			for _, c := range []string{"net-destroy", "net-undefine"} {
				_ = exec.Command("sudo", "virsh", c, "default").Run()
			}

			// Wait a bit between virsh commands
			time.Sleep(30 * time.Second)
			err := exec.Command("sudo", "virsh", "net-create", netDefaultFileName).Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Creating the Rancher Manager VM", func() {
			err := exec.Command("sudo", "virt-install",
				"--name", "rancher-manager",
				"--memory", "16384",
				"--vcpus", "4",
				"--disk", "path="+os.Getenv("HOME")+"/rancher-image.qcow2,bus=sata",
				"--import",
				"--os-variant", "opensuse-unknown",
				"--network=default,mac=52:54:00:00:00:10",
				"--noautoconsole").Run()
			Expect(err).To(Not(HaveOccurred()))
		})
	})

	It("Install K3S/Rancher in the rancher-manager machine", func() {
		airgapRepo := os.Getenv("HOME") + "/airgap_rancher"
		archiveFile := "airgap_rancher.zst"
		certPath := "/quay.io/jetstack/"
		optRancher := "/opt/rancher"
		password := "root"
		rancherManager := "rancher-manager.test"
		rancherPath := "/rancher/"
		repoServer := rancherManager + ":5000"
		userName := "root"

		// For ssh access
		client := &tools.Client{
			Host:     "192.168.122.102:22",
			Username: userName,
			Password: password,
		}

		// Create kubectl context
		// Default timeout is too small, so New() cannot be used
		k := &kubectl.Kubectl{
			Namespace:    "",
			PollTimeout:  tools.SetTimeout(300 * time.Second),
			PollInterval: 500 * time.Millisecond,
		}

		By("Sending the archive file into the rancher server", func() {
			// Destination archive file
			destFile := optRancher + "/" + archiveFile

			// Make sure SSH is available
			CheckSSH(client)

			// Create the destination repository
			_, err := client.RunSSH("mkdir -p " + optRancher)
			Expect(err).To(Not(HaveOccurred()))

			// Send the airgap archive
			err = client.SendFile(os.Getenv("HOME")+"/"+archiveFile, destFile, "0644")
			Expect(err).To(Not(HaveOccurred()))

			// Extract the airgap archive
			_, err = client.RunSSH("tar -I pzstd -vxf " + destFile + " -C " + optRancher)
			Expect(err).To(Not(HaveOccurred()))

			// Delete the archive file, not needed anymore, this will save some storage too!
			_, err = client.RunSSH("rm -f " + destFile)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Deploying airgap infrastructure by executing the deploy script", func() {
			value := regexp.MustCompile(`v(.*)\+.*`).FindStringSubmatch(k8sUpstreamVersion)
			cmd := optRancher + "/k3s_" + value[1] + "/deploy-airgap " + k8sUpstreamVersion

			// Could be useful for manual debugging!
			GinkgoWriter.Printf("Executed command: %s\n", cmd)
			out, err := client.RunSSH(cmd)
			Expect(err).To(Not(HaveOccurred()), string(out))
		})

		By("Getting the kubeconfig file of the airgap cluster", func() {
			// Define local Kubeconfig file
			localKubeconfig := os.Getenv("HOME") + "/.kube/config"

			err := os.Mkdir(os.Getenv("HOME")+"/.kube", 0755)
			Expect(err).To(Not(HaveOccurred()))

			err = client.GetFile(localKubeconfig, "/etc/rancher/k3s/k3s.yaml", 0644)
			Expect(err).To(Not(HaveOccurred()))
			// NOTE: not sure that this is need because we have the config file in ~/.kube/

			err = os.Setenv("KUBECONFIG", localKubeconfig)
			Expect(err).To(Not(HaveOccurred()))

			// Replace localhost with the IP of the VM
			err = tools.Sed("127.0.0.1", "192.168.122.102", localKubeconfig)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Installing kubectl", func() {
			// TODO: Variable for kubectl version
			err := exec.Command("curl", "-sLO", "https://dl.k8s.io/release/v1.28.2/bin/linux/amd64/kubectl").Run()
			Expect(err).To(Not(HaveOccurred()))
			err = exec.Command("chmod", "+x", "kubectl").Run()
			Expect(err).To(Not(HaveOccurred()))
			err = exec.Command("sudo", "mv", "kubectl", "/usr/local/bin/").Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Installing CertManager", func() {
			// Get the version
			certManagerChart, err := exec.Command("bash", "-c", "ls "+airgapRepo+"/helm/cert-manager-*.tgz").Output()
			Expect(err).To(Not(HaveOccurred()))

			// Set flags for cert-manager installation
			flags := []string{
				"upgrade", "--install", "cert-manager", string(certManagerChart),
				"--namespace", "cert-manager",
				"--create-namespace",
				"--set", "image.repository=" + repoServer + certPath + "cert-manager-controller",
				"--set", "webhook.image.repository=" + repoServer + certPath + "cert-manager-webhook",
				"--set", "cainjector.image.repository=" + repoServer + certPath + "cert-manager-cainjector",
				"--set", "startupapicheck.image.repository=" + repoServer + certPath + "cert-manager-startupapicheck",
				"--set", "installCRDs=true",
				"--wait", "--wait-for-jobs",
			}

			RunHelmCmdWithRetry(flags...)

			checkList := [][]string{
				{"cert-manager", "app.kubernetes.io/component=controller"},
				{"cert-manager", "app.kubernetes.io/component=webhook"},
				{"cert-manager", "app.kubernetes.io/component=cainjector"},
			}
			err = rancher.CheckPod(k, checkList)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Installing Rancher", func() {
			// TODO: Use the DeployRancherManager function from install.go
			rancherManagerChart, err := exec.Command("bash", "-c", "ls "+airgapRepo+"/helm/rancher-*.tgz").Output()
			Expect(err).To(Not(HaveOccurred()))

			// Set flags for Rancher Manager installation
			flags := []string{
				"upgrade", "--install", "rancher", string(rancherManagerChart),
				"--namespace", "cattle-system",
				"--create-namespace",
				"--set", "hostname=" + rancherManager,
				"--set", "bootstrapPassword=rancherpassword",
				"--set", "extraEnv[0].name=CATTLE_SERVER_URL",
				"--set", "extraEnv[0].value=https://" + rancherManager,
				"--set", "replicas=1",
				"--set", "useBundledSystemChart=true",
				"--set", "rancherImage=" + repoServer + rancherPath + "rancher",
				"--set", "systemDefaultRegistry=" + repoServer,
				"--wait", "--wait-for-jobs",
			}

			RunHelmCmdWithRetry(flags...)

			// Wait for all pods to be started
			checkList := [][]string{
				{"cattle-system", "app=rancher"},
				{"cattle-fleet-local-system", "app=fleet-agent"},
				{"cattle-system", "app=rancher-webhook"},
			}
			err = rancher.CheckPod(k, checkList)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Installing Elemental Operator", func() {
			// Install Elemental Operator CRDs first
			// Set flags for Elemental Operator CRDs installation
			elementalCrdsChart, err := exec.Command("bash", "-c", "ls "+airgapRepo+"/helm/elemental-operator-crds-chart-*.tgz").Output()
			Expect(err).To(Not(HaveOccurred()))

			flags := []string{
				"upgrade", "--install", "elemental-crds", string(elementalCrdsChart),
				"--namespace", "cattle-elemental-system",
				"--create-namespace",
			}

			RunHelmCmdWithRetry(flags...)
			time.Sleep(20 * time.Second)

			// Set flags for Elemental Operator installation
			elementalChart, err := exec.Command("bash", "-c", "ls "+airgapRepo+"/helm/elemental-operator-chart-*.tgz").Output()
			Expect(err).To(Not(HaveOccurred()))

			flags = []string{
				"upgrade", "--install", "elemental", string(elementalChart),
				"--namespace", "cattle-elemental-system",
				"--create-namespace",
				"--set", "image.repository=" + repoServer + rancherPath + "elemental-operator",
				"--set", "seedImage.repository=" + repoServer + rancherPath + "seedimage-builder",
				"--set", "channel.image=" + repoServer + rancherPath + "elemental-channel-" + rancherManager,
				"--set", "registryUrl=",
				"--wait", "--wait-for-jobs",
			}

			RunHelmCmdWithRetry(flags...)

			// Wait for pod to be started
			err = rancher.CheckPod(k, [][]string{{"cattle-elemental-system", "app=elemental-operator"}})
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})
