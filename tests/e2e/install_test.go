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
	"os"
	"os/exec"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

var _ = Describe("E2E - Install Rancher", Label("install"), func() {
	// Create kubectl context
	// Default timeout is too small, so New() cannot be used
	k := &kubectl.Kubectl{
		Namespace:    "",
		PollTimeout:  1200 * time.Second,
		PollInterval: 500 * time.Millisecond,
	}

	It("Install and configure Rancher and libvirt", func() {
		By("Installing K3s", func() {
			// Get K3s installation script
			fileName := "k3s-install.sh"
			err := tools.GetFileFromURL("https://get.k3s.io", fileName, true)
			Expect(err).To(Not(HaveOccurred()))

			// Execute K3s installation
			out, err := exec.Command("sh", fileName).CombinedOutput()
			GinkgoWriter.Printf("%s\n", out)
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Starting K3s", func() {
			// Start in background
			// TODO: use something better to start in background, because this code seems to not write k3s.log file...
			err := exec.Command("/usr/local/bin/k3s", "server", "--snapshotter=native", ">/tmp/k3s.log", "2>&1").Start()
			Expect(err).To(Not(HaveOccurred()))

			// Delay few seconds before checking
			time.Sleep(20 * time.Second)
		})

		By("Waiting for K3s to be started", func() {
			// Wait for all pods to be started
			err := k.WaitForPod("kube-system", "app=local-path-provisioner", "local-path-provisioner")
			Expect(err).To(Not(HaveOccurred()))

			err = k.WaitForPod("kube-system", "k8s-app=kube-dns", "coredns")
			Expect(err).To(Not(HaveOccurred()))

			err = k.WaitForPod("kube-system", "k8s-app=metrics-server", "metrics-server")
			Expect(err).To(Not(HaveOccurred()))

			err = k.WaitForPod("kube-system", "app.kubernetes.io/name=traefik", "traefik")
			Expect(err).To(Not(HaveOccurred()))

			err = k.WaitForPod("kube-system", "app=svclb-traefik", "svclb-traefik")
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Installing CertManager", func() {
			err := kubectl.RunHelmBinaryWithCustomErr("repo", "add", "jetstack", "https://charts.jetstack.io")
			Expect(err).To(Not(HaveOccurred()))

			err = kubectl.RunHelmBinaryWithCustomErr("repo", "update")
			Expect(err).To(Not(HaveOccurred()))

			err = kubectl.RunHelmBinaryWithCustomErr("install", "cert-manager", "jetstack/cert-manager",
				"--namespace", "cert-manager",
				"--create-namespace",
				"--set", "installCRDs=true",
			)
			Expect(err).To(Not(HaveOccurred()))

			err = k.WaitForNamespaceWithPod("cert-manager", "app.kubernetes.io/component=controller")
			Expect(err).To(Not(HaveOccurred()))

			err = k.WaitForNamespaceWithPod("cert-manager", "app.kubernetes.io/component=webhook")
			Expect(err).To(Not(HaveOccurred()))

			err = k.WaitForNamespaceWithPod("cert-manager", "app.kubernetes.io/component=cainjector")
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Installing Rancher", func() {
			err := kubectl.RunHelmBinaryWithCustomErr("repo", "add", "rancher-stable", "https://releases.rancher.com/server-charts/stable")
			Expect(err).To(Not(HaveOccurred()))

			err = kubectl.RunHelmBinaryWithCustomErr("repo", "update")
			Expect(err).To(Not(HaveOccurred()))

			hostname := os.Getenv("HOSTNAME")
			err = kubectl.RunHelmBinaryWithCustomErr("install", "rancher", "rancher-stable/rancher",
				"--namespace", "cattle-system",
				"--create-namespace",
				"--set", "hostname="+hostname,
				"--set", "extraEnv[0].name=CATTLE_SERVER_URL",
				"--set", "extraEnv[0].value=https://"+hostname,
				"--set", "extraEnv[1].name=CATTLE_BOOTSTRAP_PASSWORD",
				"--set", "extraEnv[1].value=rancherpassword",
				"--set", "replicas=1",
			)
			Expect(err).To(Not(HaveOccurred()))

			err = k.WaitForNamespaceWithPod("cattle-system", "app=rancher")
			Expect(err).To(Not(HaveOccurred()))

			err = k.WaitForNamespaceWithPod("cattle-fleet-local-system", "app=fleet-agent")
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Installing Elemental Operator", func() {
			err := kubectl.RunHelmBinaryWithCustomErr("repo", "add",
				"elemental-operator",
				"https://rancher.github.io/elemental-operator",
			)
			Expect(err).To(Not(HaveOccurred()))

			err = kubectl.RunHelmBinaryWithCustomErr("repo", "update")
			Expect(err).To(Not(HaveOccurred()))

			err = kubectl.RunHelmBinaryWithCustomErr("install", "elemental-operator", "elemental-operator/elemental-operator",
				"--devel",
				"--namespace", "cattle-elemental-system",
				"--create-namespace",
			)
			Expect(err).To(Not(HaveOccurred()))

			k.WaitForNamespaceWithPod("cattle-elemental-system", "app=elemental-operator")
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Creating a new cluster", func() {
			addClusterYaml := "../assets/add_cluster.yaml"
			err := tools.Sed("%CLUSTER_NAME%", clusterName, addClusterYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = kubectl.Apply(clusterNS, addClusterYaml)
			Expect(err).To(Not(HaveOccurred()))

			createdCluster, err := kubectl.Run("get", "cluster",
				"--namespace", clusterNS,
				clusterName, "-o", "jsonpath={.metadata.name}")
			Expect(err).To(Not(HaveOccurred()))

			// Check that's the created cluster is the good one
			Expect(createdCluster).To(Equal(clusterName))
		})

		By("Adding MachineRegistration", func() {
			registrationYaml := "../assets/machineregistration.yaml"

			err := tools.Sed("%VM_NAME%", vmNameRoot, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = tools.Sed("%USER%", userName, registrationYaml)
			Expect(err).To(Not(HaveOccurred()))

			err = tools.Sed("%PASSWORD%", userPassword, registrationYaml)
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
			// err := tools.HTTpShare("../..", 8000)
			// Expect(err).To(Not(HaveOccurred()))

			// Use Python for now...
			err := exec.Command("../scripts/start-httpd").Run()
			Expect(err).To(Not(HaveOccurred()))
		})

		By("Starting libvirtd", func() {
			cmds := []string{"libvirtd", "virtlogd"}
			for _, c := range cmds {
				err := exec.Command(c, "--daemon").Run()
				Expect(err).To(Not(HaveOccurred()))
			}
		})

		By("Starting default network", func() {
			err := exec.Command("virsh", "net-undefine", "default").Run()
			Expect(err).To(Not(HaveOccurred()))

			err = exec.Command("virsh", "net-create", netDefaultFileName).Run()
			Expect(err).To(Not(HaveOccurred()))
		})
	})
})
