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

var _ = Describe("E2E - Build the airgap archive", Label("prepare-archive"), func() {
	It("Execute the script to build the archive", func() {
		err := exec.Command("sudo", airgapBuildScript, k8sUpstreamVersion, certManagerVersion, rancherChannel, k8sVersion, operatorRepo).Run()
		Expect(err).To(Not(HaveOccurred()))
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
			time.Sleep(1 * time.Minute)
			err := exec.Command("sudo", "virsh", "net-create", netDefaultAirgapFileName).Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Downloading the qcow2 image from GCP storage", func() {
			err := exec.Command("/opt/google-cloud-sdk/bin/gcloud", "storage", "cp", "gs://elemental-airgap-image/rancher-image.qcow2", os.Getenv("HOME")+"/rancher-image.qcow2").Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Creating the Rancher Manager VM", func() {
			err := exec.Command("sudo", "virt-install",
				"--name", "rancher-manager",
				"--memory", "4096",
				"--vcpus", "2",
				"--disk", "path="+os.Getenv("HOME")+"/rancher-image.qcow2,bus=sata",
				"--import",
				"--os-variant", "opensuse-unknown",
				"--network=default,mac=52:54:00:00:00:10",
				"--noautoconsole").Run()
			Expect(err).To(Not(HaveOccurred()))
		})
	})

	It("Install K3S/Rancher in the rancher-manager machine", func() {
		userName := "root"
		password := "root"
		client := &tools.Client{
			Host:     "192.168.122.102:22",
			Username: userName,
			Password: password,
		}

		// Get the version of the Elemental Operator
		out, err := exec.Command("bash", "-c", "ls /opt/rancher/helm/elemental-operator-chart-*.tgz | cut -d '-' -f 4").Output()
		Expect(err).To(Not(HaveOccurred()))
		elementalVersion := strings.Trim(string(out), "\n")

		// Create kubectl context
		// Default timeout is too small, so New() cannot be used
		k := &kubectl.Kubectl{
			Namespace:    "",
			PollTimeout:  tools.SetTimeout(300 * time.Second),
			PollInterval: 500 * time.Millisecond,
		}

		By("Sending the archive file into the rancher server", func() {
			// Make sure SSH is available
			Eventually(func() string {
				out, _ := client.RunSSH("echo SSH_OK")
				out = strings.Trim(out, "\n")
				return out
			}, tools.SetTimeout(10*time.Minute), 5*time.Second).Should(Equal("SSH_OK"))

			// Send the airgap archive
			err := client.SendFile("/opt/airgap_rancher.zst", "/opt/airgap_rancher.zst", "0644")
			Expect(err).To(Not(HaveOccurred()))

			// Extract the airgap archive
			_, err = client.RunSSH("mkdir /opt/rancher; tar -I zstd -vxf /opt/airgap_rancher.zst -C /opt/rancher")
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Deploying airgap infrastructure by executing the deploy script", func() {
			_, err := client.RunSSH("/opt/rancher/k3s_" + k8sUpstreamVersion + "/deploy-airgap " + k8sUpstreamVersion + " " + certManagerVersion)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Getting the kubeconfig file of the airgap cluster", func() {
			// Define local Kubeconfig file
			localKubeconfig := os.Getenv("HOME") + "/.kube/config"
			Expect(err).To(Not(HaveOccurred()))
			err := os.Mkdir(os.Getenv("HOME")+"/.kube", 0755)
			Expect(err).To(Not(HaveOccurred()))
			err = client.GetFile(localKubeconfig, "/etc/rancher/k3s/k3s.yaml", 0644)
			Expect(err).To(Not(HaveOccurred()))
			err = os.Setenv("KUBECONFIG", localKubeconfig)
			Expect(err).To(Not(HaveOccurred()))

			// Replace localhost with the IP of the VM
			err = tools.Sed("127.0.0.1", "192.168.122.102", localKubeconfig)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Installing kubectl", func() {
			// TODO: Variable for kubectl version
			err := exec.Command("curl", "-LO", "https://dl.k8s.io/release/v1.28.2/bin/linux/amd64/kubectl").Run()
			Expect(err).To(Not(HaveOccurred()))
			err = exec.Command("chmod", "+x", "kubectl").Run()
			Expect(err).To(Not(HaveOccurred()))
			err = exec.Command("sudo", "mv", "kubectl", "/usr/local/bin/").Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Installing CertManager", func() {
			// Set flags for cert-manager installation
			flags := []string{
				"upgrade", "--install", "cert-manager", "/opt/rancher/helm/cert-manager-" + certManagerVersion + ".tgz",
				"--namespace", "cert-manager",
				"--create-namespace",
				"--set", "image.repository=rancher-manager.test:5000/cert/cert-manager-controller",
				"--set", "webhook.image.repository=rancher-manager.test:5000/cert/cert-manager-webhook",
				"--set", "cainjector.image.repository=rancher-manager.test:5000/cert/cert-manager-cainjector",
				"--set", "startupapicheck.image.repository=rancher-manager.test:5000/cert/cert-manager-ctl",
				"--set", "installCRDs=true",
				"--wait", "--wait-for-jobs",
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

		By("Installing Rancher", func() {
			// TODO: Use the DeployRancherManager function from install.go
			rancherAirgapVersion, err := exec.Command("bash", "-c", "ls /opt/rancher/helm/rancher-*.tgz").Output()
			Expect(err).To(Not(HaveOccurred()))

			// Set flags for Rancher Manager installation
			flags := []string{
				"upgrade", "--install", "rancher", string(rancherAirgapVersion),
				"--namespace", "cattle-system",
				"--create-namespace",
				"--set", "hostname=rancher-manager.test",
				"--set", "extraEnv[0].name=CATTLE_SERVER_URL",
				"--set", "extraEnv[0].value=https://rancher-manager.test",
				"--set", "extraEnv[1].name=CATTLE_BOOTSTRAP_PASSWORD",
				"--set", "extraEnv[1].value=rancherpassword",
				"--set", "replicas=1",
				"--set", "useBundledSystemChart=true",
				"--set", "rancherImage=rancher-manager.test:5000/rancher/rancher",
				"--set", "systemDefaultRegistry=rancher-manager.test:5000",
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
			flags := []string{
				"upgrade", "--install", "elemental-crds", "/opt/rancher/helm/elemental-operator-crds-chart-" + elementalVersion,
				"--namespace", "cattle-elemental-system",
				"--create-namespace",
			}

			RunHelmCmdWithRetry(flags...)
			time.Sleep(20 * time.Second)

			// Set flags for Elemental Operator installation
			flags = []string{
				"upgrade", "--install", "elemental", "/opt/rancher/helm/elemental-operator-chart-" + elementalVersion,
				"--namespace", "cattle-elemental-system",
				"--create-namespace",
				"--set", "image.repository=rancher-manager.test:5000/elemental/elemental-operator",
				"--set", "registryUrl=",
				"--set", "seedImage.repository=rancher-manager.test:5000/elemental/seedimage-builder",
				"--set", "channel.repository=rancher-manager.test:5000/elemental/elemental-channel",
				"--wait", "--wait-for-jobs",
			}

			RunHelmCmdWithRetry(flags...)

			// Wait for pod to be started
			err := rancher.CheckPod(k, [][]string{{"cattle-elemental-system", "app=elemental-operator"}})
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})
