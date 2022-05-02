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

package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/os2/tests/e2e/helpers/tools"
)

func getServerId(clusterNS string) string {
	serverId, err := kubectl.Run("get", "MachineInventories",
		"--namespace", clusterNS,
		"-o", "jsonpath={.items[0].metadata.name}")
	Expect(err).NotTo(HaveOccurred())
	Expect(serverId).ToNot(Equal(""))

	return serverId
}

var _ = Describe("E2E - Bootstrapping node with Rancher", Label("bootstrapping"), func() {
	const (
		vmName = "ros-node"
	)

	var (
		clusterName = os.Getenv("CLUSTER_NAME")
		clusterNS   = os.Getenv("CLUSTER_NS")
	)

	// Create kubectl context
	// Default timeout is too small, so New() cannot be used
	k := &kubectl.Kubectl{
		Namespace:    "",
		PollTimeout:  600 * time.Second,
		PollInterval: 500 * time.Millisecond,
	}

	It("Install Rancher", func() {
		By("Installing K3s", func() {
			// Get K3s installation script
			fileName := "k3s-install.sh"
			err := tools.GetFileFromURL("https://get.k3s.io", fileName, true)
			Expect(err).NotTo(HaveOccurred())

			// Execute K3s installation
			out, err := exec.Command("sh", fileName).CombinedOutput()
			Expect(err).NotTo(HaveOccurred())
			GinkgoWriter.Printf("%s\n", out)
		})

		By("Starting K3s", func() {
			// Start in background
			err := exec.Command("/usr/local/bin/k3s", "server", "--snapshotter=native", ">/tmp/k3s.log", "2>&1").Start()
			Expect(err).NotTo(HaveOccurred())

			// Delay few seconds before checking
			time.Sleep(5 * time.Second)
		})

		By("Waiting for K3s to be started", func() {
			// Wait for all pods to be started
			err := k.WaitForPod("kube-system", "app=local-path-provisioner", "local-path-provisioner")
			Expect(err).NotTo(HaveOccurred())
			err = k.WaitForPod("kube-system", "k8s-app=kube-dns", "coredns")
			Expect(err).NotTo(HaveOccurred())
			err = k.WaitForPod("kube-system", "k8s-app=metrics-server", "metrics-server")
			Expect(err).NotTo(HaveOccurred())
			err = k.WaitForPod("kube-system", "app.kubernetes.io/name=traefik", "traefik")
			Expect(err).NotTo(HaveOccurred())
			err = k.WaitForPod("kube-system", "app=svclb-traefik", "svclb-traefik")
			Expect(err).NotTo(HaveOccurred())
		})

		By("Installing CertManager", func() {
			err := kubectl.RunHelmBinaryWithCustomErr("repo", "add", "jetstack", "https://charts.jetstack.io")
			Expect(err).NotTo(HaveOccurred())

			err = kubectl.RunHelmBinaryWithCustomErr("repo", "update")
			Expect(err).NotTo(HaveOccurred())

			err = kubectl.RunHelmBinaryWithCustomErr("install", "cert-manager", "jetstack/cert-manager",
				"--namespace", "cert-manager",
				"--create-namespace",
				"--set", "installCRDs=true",
			)
			Expect(err).ToNot(HaveOccurred())

			err = k.WaitForPod("cert-manager", "app.kubernetes.io/instance=cert-manager", "cert-manager-cainjector")
			Expect(err).ToNot(HaveOccurred())

			err = k.WaitForNamespaceWithPod("cert-manager", "app.kubernetes.io/instance=cert-manager")
			Expect(err).ToNot(HaveOccurred())
		})

		By("Installing Rancher", func() {
			err := kubectl.RunHelmBinaryWithCustomErr("repo", "add", "rancher-stable", "https://releases.rancher.com/server-charts/stable")
			Expect(err).NotTo(HaveOccurred())

			err = kubectl.RunHelmBinaryWithCustomErr("repo", "update")
			Expect(err).NotTo(HaveOccurred())

			hostname := os.Getenv("HOSTNAME")
			err = kubectl.RunHelmBinaryWithCustomErr("install", "rancher", "rancher-stable/rancher",
				"--namespace", "cattle-system",
				"--create-namespace",
				"--set", "hostname="+hostname,
				"--set", "extraEnv[0].name=CATTLE_SERVER_URL",
				"--set", "extraEnv[0].value=https://"+hostname,
				"--set", "extraEnv[1].name=CATTLE_BOOTSTRAP_PASSWORD",
				"--set", "extraEnv[1].value=rancherpassword",
			)
			Expect(err).NotTo(HaveOccurred())

			err = k.WaitForPod("cattle-system", "app=rancher", "rancher")
			Expect(err).ToNot(HaveOccurred())

			err = k.WaitForNamespaceWithPod("cattle-system", "app=rancher")
			Expect(err).ToNot(HaveOccurred())

			err = k.WaitForNamespaceWithPod("cattle-fleet-local-system", "app=fleet-agent")
			Expect(err).ToNot(HaveOccurred())
		})
	})

	It("Configure Rancher (step 01)", func() {
		By("Installing RancherOS Operator", func() {
			err := kubectl.RunHelmBinaryWithCustomErr("repo", "add",
				"rancheros-operator",
				"https://rancher-sandbox.github.io/rancheros-operator",
			)
			Expect(err).NotTo(HaveOccurred())

			err = kubectl.RunHelmBinaryWithCustomErr("repo", "update")
			Expect(err).NotTo(HaveOccurred())

			err = kubectl.RunHelmBinaryWithCustomErr("install", "rancheros-operator", "rancheros-operator/rancheros-operator",
				"--version", ">0.0.0-0",
				"--namespace", "cattle-rancheros-operator-system",
				"--create-namespace",
			)
			Expect(err).NotTo(HaveOccurred())

			err = k.WaitForPod("cattle-rancheros-operator-system", "app=rancheros-operator", "rancheros-operator")
			Expect(err).ToNot(HaveOccurred())

			k.WaitForNamespaceWithPod("cattle-rancheros-operator-system", "app=rancheros-operator")
			Expect(err).NotTo(HaveOccurred())
		})

		By("Adding MachineRegistration in Rancher", func() {
			err := kubectl.Apply(clusterNS, "../assets/machineregistration.yaml")
			Expect(err).NotTo(HaveOccurred())
		})

		By("Creating a new cluster", func() {
			addClusterYaml := "../assets/add_cluster.yaml"
			err := tools.Sed("%CLUSTER_NAME%", clusterName, addClusterYaml)
			Expect(err).NotTo(HaveOccurred())

			err = kubectl.Apply(clusterNS, addClusterYaml)
			Expect(err).NotTo(HaveOccurred())

			tokenURL, err := kubectl.Run("get", "MachineRegistration",
				"--namespace", clusterNS,
				"machine-registration", "-o", "jsonpath={.status.registrationURL}")
			Expect(err).NotTo(HaveOccurred())

			createdCluster, err := kubectl.Run("get", "cluster",
				"--namespace", clusterNS,
				clusterName, "-o", "jsonpath={.metadata.name}")
			Expect(err).NotTo(HaveOccurred())

			// Check that's the created cluster is the good one
			Expect(createdCluster).To(Equal(clusterName))

			// Get the YAML config file
			fileName := "../../install-config.yaml"
			err = tools.GetFileFromURL(tokenURL, fileName, false)
			Expect(err).NotTo(HaveOccurred())
		})
	})

	It("Install RancherOS node", func() {
		netDefaultFileName := "../assets/net-default.xml"

		By("Configuring iPXE boot script for CI", func() {
			ipxeScript, err := tools.GetFiles("..", "*.ipxe")
			Expect(err).NotTo(HaveOccurred())
			Expect(len(ipxeScript)).To(BeNumerically("==", 1))
			err = tools.Sed("set url.*", "set url http://192.168.122.1:8000", ipxeScript[0])
			Expect(err).NotTo(HaveOccurred())
			err = tools.Sed("set config.*", "set config $${url}/install-config.yaml", ipxeScript[0])
			Expect(err).NotTo(HaveOccurred())

			scriptName := filepath.Base(ipxeScript[0])
			err = tools.Sed("%IPXE_SCRIPT%", "<bootp file='http://192.168.122.1:8000/"+scriptName+"'/>", netDefaultFileName)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Starting HTTP server for network installation", func() {
			// TODO: improve it to run in background!
			// err := tools.HTTpShare("../..", 8000)
			// Expect(err).NotTo(HaveOccurred())

			// Use Python for now...
			err := exec.Command("../scripts/start-httpd").Run()
			Expect(err).NotTo(HaveOccurred())
		})

		By("Starting libvirtd", func() {
			cmds := []string{"libvirtd", "virtlogd"}
			for _, c := range cmds {
				err := exec.Command(c, "--daemon").Run()
				Expect(err).NotTo(HaveOccurred())
			}
		})

		By("Starting default network", func() {
			err := exec.Command("virsh", "net-undefine", "default").Run()
			Expect(err).NotTo(HaveOccurred())
			err = exec.Command("virsh", "net-create", netDefaultFileName).Run()
			Expect(err).NotTo(HaveOccurred())
		})

		By("Creating and installing VM", func() {
			/*
				cmd := exec.Command("virt-install",
					"--name", vmName,
					"--os-type", "Linux",
					"--os-variant", "opensuse-unknown",
					"--virt-type", "kvm",
					"--machine", "q35",
					"--boot", "bios.useserial=on",
					"--ram", "2048",
					"--vcpus", "2",
					"--cpu", "host",
					"--disk", "path=hdd.img,bus=virtio,size=35",
					"--check", "disk_size=off",
					"--graphics", "none",
					"--serial", "pty",
					"--console", "pty,target_type=virtio",
					"--rng", "random",
					"--tpm", "emulator,model=tpm-crb,version=2.0",
					"--noreboot",
					"--pxe",
					"--network", "network=default,bridge=virbr0,model=virtio,mac=52:54:00:00:00:01",
				)
			*/
			cmd := exec.Command("../scripts/install-vm", vmName)
			out, err := cmd.CombinedOutput()
			GinkgoWriter.Printf("%s\n", out)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Checking that the VM is available in Rancher", func() {
			getServerId(clusterNS)
		})
	})

	It("Add server in "+clusterName, func() {
		By("Adding server role to predefined cluster", func() {
			serverId := getServerId(clusterNS)
			patchCmd := `{"spec":{"clusterName":"` + clusterName + `","config":{"role":"server"}}}`
			_, err := kubectl.Run("patch", "MachineInventories",
				"--namespace", clusterNS, serverId,
				"--type", "merge", "--patch", patchCmd,
			)
			Expect(err).NotTo(HaveOccurred())
		})

		By("Restarting the VM", func() {
			err := exec.Command("virsh", "start", vmName).Run()
			Expect(err).NotTo(HaveOccurred())

			// Waiting for node to be added to the cluster (maybe can be wrote purely in Go?)
			err = exec.Command("../scripts/wait-for-node").Run()
			Expect(err).NotTo(HaveOccurred())
		})

		By("Checking that the VM is added in the cluster", func() {
			serverId, err := kubectl.Run("get", "MachineInventories",
				"--namespace", clusterNS,
				"-o", "jsonpath={.items[0].metadata.name}")
			Expect(err).NotTo(HaveOccurred())
			Expect(serverId).ToNot(Equal(""))

			internalClusterName, err := kubectl.Run("get", "cluster",
				"--namespace", clusterNS, clusterName,
				"-o", "jsonpath='.status.clusterName'")
			Expect(err).NotTo(HaveOccurred())
			Expect(internalClusterName).ToNot(Equal(""))

			internalClusterToken, err := kubectl.Run("get", "MachineInventories",
				"--namespace", clusterNS, serverId,
				"-o", "jsonpath='.status.clusterRegistrationTokenNamespace'")
			Expect(err).NotTo(HaveOccurred())
			Expect(internalClusterToken).ToNot(Equal(""))

			// Check that the VM is added
			Expect(internalClusterName).To(Equal(internalClusterToken))
		})
	})
})
