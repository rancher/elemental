/*
Copyright © 2022 - 2024 SUSE LLC

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
	"strconv"
	"strings"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/rancher"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	. "github.com/rancher-sandbox/qase-ginkgo"
	"github.com/rancher/elemental/tests/e2e/helpers/elemental"
)

const (
	airgapBuildScript     = "../scripts/build-airgap"
	appYaml               = "../assets/hello-world_app.yaml"
	backupYaml            = "../assets/backup.yaml"
	ciTokenYaml           = "../assets/local-kubeconfig-token-skel.yaml"
	configPrivateCAScript = "../scripts/config-private-ca"
	configRKE2Yaml        = "../assets/config_rke2.yaml"
	dumbRegistrationYaml  = "../assets/dumb_machineRegistration.yaml"
	emulateTPMYaml        = "../assets/emulateTPM.yaml"
	getOSScript           = "../scripts/get-name-from-managedosversion"
	httpSrv               = "http://192.168.122.1:8000"
	installConfigYaml     = "../../install-config.yaml"
	installHardenedScript = "../scripts/config-hardened"
	installVMScript       = "../scripts/install-vm"
	localKubeconfigYaml   = "../assets/local-kubeconfig-skel.yaml"
	localStorageYaml      = "../assets/local-storage.yaml"
	metallbRscYaml        = "../assets/metallb_rsc.yaml"
	numberOfNodesMax      = 30
	resetMachineInv       = "../assets/reset_machine_inventory.yaml"
	restoreYaml           = "../assets/restore.yaml"
	upgradeSkelYaml       = "../assets/upgrade_skel.yaml"
	userName              = "root"
	userPassword          = "r0s@pwd1"
	vmNameRoot            = "node"
)

var (
	backupRestoreVersion      string
	caType                    string
	certManagerVersion        string
	clusterName               string
	clusterNS                 string
	clusterType               string
	clusterYaml               string
	elementalSupport          string
	emulateTPM                bool
	forceDowngrade            bool
	isoBoot                   bool
	k8sUpstreamVersion        string
	k8sDownstreamVersion      string
	netDefaultFileName        string
	numberOfClusters          int
	numberOfVMs               int
	operatorInstallType       string
	operatorRepo              string
	operatorUpgrade           string
	os2Test                   string
	poolType                  string
	proxy                     string
	rancherChannel            string
	rancherHeadVersion        string
	rancherHostname           string
	rancherLogCollector       string
	rancherVersion            string
	rancherUpgrade            string
	rancherUpgradeChannel     string
	rancherUpgradeHeadVersion string
	rancherUpgradeVersion     string
	rawBoot                   bool
	registrationYaml          string
	seedImageYaml             string
	selectorYaml              string
	selinux                   bool
	sequential                bool
	snapType                  string
	sshdConfigFile            string
	testCaseID                int64
	testType                  string
	upgradeImage              string
	upgradeOSChannel          string
	upgradeType               string
	usedNodes                 int
	vmIndex                   int
	vmName                    string
)

func CheckBackupRestore(v string) {
	Eventually(func() string {
		out, _ := kubectl.RunWithoutErr("logs", "-l app.kubernetes.io/name=rancher-backup",
			"--tail=-1", "--since=5m",
			"--namespace", "cattle-resources-system")
		return out
	}, tools.SetTimeout(5*time.Minute), 10*time.Second).Should(ContainSubstring(v))
}

/*
Check that Cluster resource has been correctly created
  - @param ns Namespace where the cluster is deployed
  - @param cn Cluster resource name
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func CheckCreatedCluster(ns, cn string) {
	// Check that the cluster is correctly created
	Eventually(func() string {
		out, _ := kubectl.RunWithoutErr("get", "cluster.v1.provisioning.cattle.io",
			"--namespace", ns,
			cn, "-o", "jsonpath={.metadata.name}")
		return out
	}, tools.SetTimeout(3*time.Minute), 5*time.Second).Should(Equal(cn))
}

/*
Check that Registration resource has been correctly created
  - @param ns Namespace where the cluster is deployed
  - @param rn MachineRegistration resource name
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func CheckCreatedRegistration(ns, rn string) {
	Eventually(func() string {
		out, _ := kubectl.RunWithoutErr("get", "MachineRegistration",
			"--namespace", ns,
			"-o", "jsonpath={.items[*].metadata.name}")
		return out
	}, tools.SetTimeout(3*time.Minute), 5*time.Second).Should(ContainSubstring(rn))
}

/*
Check that a SelectorTemplate resource has been correctly created
  - @param ns Namespace where the cluster is deployed
  - @param sn Selector name
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func CheckCreatedSelectorTemplate(ns, sn string) {
	Eventually(func() string {
		out, _ := kubectl.RunWithoutErr("get", "MachineInventorySelectorTemplate",
			"--namespace", ns,
			"-o", "jsonpath={.items[*].metadata.name}")
		return out
	}, tools.SetTimeout(3*time.Minute), 5*time.Second).Should(ContainSubstring(sn))
}

/*
Check SSH connection
  - @param cl Client (node) informations
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func CheckSSH(cl *tools.Client) {
	Eventually(func() string {
		out, _ := cl.RunSSH("echo SSH_OK")
		return strings.Trim(out, "\n")
	}, tools.SetTimeout(10*time.Minute), 5*time.Second).Should(Equal("SSH_OK"))
}

/*
Download ISO built with SeedImage
  - @param ns Namespace where the cluster is deployed
  - @param seedName Name of the used SeedImage resource
  - @param filename Path and name of the file where to store the ISO
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func DownloadBuiltISO(ns, seedName, filename string) {
	// Set minimal ISO file to 250MB
	const minimalISOSize = 250 * 1024 * 1024

	By("Waiting for image to be generated", func() {
		// Check that the seed image is correctly created
		Eventually(func() string {
			out, _ := kubectl.RunWithoutErr("get", "SeedImage",
				"--namespace", ns,
				seedName,
				"-o", "jsonpath={.status}")
			return out
		}, tools.SetTimeout(3*time.Minute), 5*time.Second).Should(ContainSubstring("downloadURL"))
	})

	By("Downloading image", func() {
		// Get URL
		seedImageURL, err := kubectl.RunWithoutErr("get", "SeedImage",
			"--namespace", ns,
			seedName,
			"-o", "jsonpath={.status.downloadURL}")
		Expect(err).To(Not(HaveOccurred()))

		// ISO file size should be greater than 500MB
		Eventually(func() int64 {
			// No need to check download status, file size at the end is enough
			_ = tools.GetFileFromURL(seedImageURL, filename, false)
			file, _ := os.Stat(filename)
			return file.Size()
		}, tools.SetTimeout(2*time.Minute), 10*time.Second).Should(BeNumerically(">", minimalISOSize))
	})

	// Only supported in Dev version for now, not Stable (1.5.x) and Staging (1.6.4)
	if strings.Contains(os2Test, "dev") {
		By("Checking checksum", func() {
			// Get checksum URL
			checksumURL, err := kubectl.RunWithoutErr("get", "SeedImage",
				"--namespace", ns,
				seedName,
				"-o", "jsonpath={.status.checksumURL}")
			Expect(err).To(Not(HaveOccurred()))

			// Download checksum file
			checksumFile := filename + ".sha256"
			_ = tools.GetFileFromURL(checksumURL, checksumFile, false)

			// Check the checksum of downloaded image
			err = exec.Command("bash", "-c", "sed -i 's; .*\\.iso; "+filename+";' "+checksumFile).Run()
			Expect(err).To(Not(HaveOccurred()))
			err = exec.Command("sha256sum", "--check", checksumFile).Run()
			Expect(err).To(Not(HaveOccurred()))
		})
	}
}

/*
Get configured backup directory
  - @returns Configured backup directory
*/
func GetBackupDir() string {
	claimName, err := kubectl.RunWithoutErr("get", "pod", "-l", "app.kubernetes.io/name=rancher-backup",
		"--namespace", "cattle-resources-system",
		"-o", "jsonpath={.items[*].spec.volumes[?(@.name==\"pv-storage\")].persistentVolumeClaim.claimName}")
	Expect(err).To(Not(HaveOccurred()))

	out, err := kubectl.RunWithoutErr("get", "pv",
		"--namespace", "cattle-resources-system",
		"-o", "jsonpath={.items[?(@.spec.claimRef.name==\""+claimName+"\")].spec.local.path}")
	Expect(err).To(Not(HaveOccurred()))

	return out
}

/*
Get Elemental node information
  - @param hn Node hostname
  - @returns Client structure and MAC address
*/
func GetNodeInfo(hn string) (*tools.Client, string) {
	// Get network data
	data, err := rancher.GetHostNetConfig(".*name=\""+hn+"\".*", netDefaultFileName)
	Expect(err).To(Not(HaveOccurred()))

	// Set 'client' to be able to access the node through SSH
	c := &tools.Client{
		Host:     string(data.IP) + ":22",
		Username: userName,
		Password: userPassword,
	}

	return c, data.Mac
}

/*
Get Elemental node IP address
  - @param hn Node hostname
  - @returns IP address
*/
func GetNodeIP(hn string) string {
	// Get network data
	data, err := rancher.GetHostNetConfig(".*name=\""+hn+"\".*", netDefaultFileName)
	Expect(err).To(Not(HaveOccurred()))

	return data.IP
}

/*
Install rancher-backup operator
  - @param k kubectl structure
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func InstallBackupOperator(k *kubectl.Kubectl) {
	// Default chart
	chartRepo := "rancher-chart"

	// Set specific operator version if defined
	if backupRestoreVersion != "" {
		chartRepo = "https://github.com/rancher/backup-restore-operator/releases/download/" + backupRestoreVersion
	} else {
		RunHelmCmdWithRetry("repo", "add", chartRepo, "https://charts.rancher.io")
		RunHelmCmdWithRetry("repo", "update")
	}

	for _, chart := range []string{"rancher-backup-crd", "rancher-backup"} {
		// Set the filename in chart if a custom version is defined
		chartName := chart
		if backupRestoreVersion != "" {
			chartName = chart + "-" + strings.Trim(backupRestoreVersion, "v") + ".tgz"
		}

		// Global installation flags
		flags := []string{
			"upgrade", "--install", chart, chartRepo + "/" + chartName,
			"--namespace", "cattle-resources-system",
			"--create-namespace",
			"--wait", "--wait-for-jobs",
		}

		// Add specific options for the rancher-backup chart
		if chart == "rancher-backup" {
			flags = append(flags,
				"--set", "persistence.enabled=true",
				"--set", "persistence.storageClass=local-path",
			)
		}

		RunHelmCmdWithRetry(flags...)

		Eventually(func() error {
			return rancher.CheckPod(k, [][]string{{"cattle-resources-system", "app.kubernetes.io/name=rancher-backup"}})
		}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(Not(HaveOccurred()))
	}
}

/*
Install CertManager
  - @param k kubectl structure
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func InstallCertManager(k *kubectl.Kubectl) {
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
		flags = append(flags, "--version", certManagerVersion)
	}

	RunHelmCmdWithRetry(flags...)

	checkList := [][]string{
		{"cert-manager", "app.kubernetes.io/component=controller"},
		{"cert-manager", "app.kubernetes.io/component=webhook"},
		{"cert-manager", "app.kubernetes.io/component=cainjector"},
	}
	Eventually(func() error {
		return rancher.CheckPod(k, checkList)
	}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(Not(HaveOccurred()))
}

/*
Install Elemental operator
  - @param k kubectl structure
  - @param order Order of the chart installation, mainly useful for older versions
  - @param repo Chart repository to use
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func InstallElementalOperator(k *kubectl.Kubectl, order []string, repo string) {
	for _, chart := range order {
		// Set flags for installation
		flags := []string{"upgrade", "--install", chart,
			repo + "/" + chart + "-chart",
			"--namespace", "cattle-elemental-system",
			"--create-namespace",
			"--wait", "--wait-for-jobs",
		}

		// Dev and Staging versions need a specific treatment
		if strings.Contains(repo, "/dev/") || strings.Contains(repo, "/staging/") {
			flags = append(flags, "--devel")
		}

		RunHelmCmdWithRetry(flags...)
	}

	// Wait for pod to be started
	Eventually(func() error {
		return rancher.CheckPod(k, [][]string{{"cattle-elemental-system", "app=elemental-operator"}})
	}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(Not(HaveOccurred()))
}

/*
Install local storage
  - @param k kubectl structure
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func InstallLocalStorage(k *kubectl.Kubectl) {
	localPathNS := "kube-system"
	kubectl.Apply(localPathNS, localStorageYaml)

	// Wait for all pods to be started
	checkList := [][]string{
		{localPathNS, "app=local-path-provisioner"},
	}
	Eventually(func() error {
		return rancher.CheckPod(k, checkList)
	}, tools.SetTimeout(2*time.Minute), 30*time.Second).Should(Not(HaveOccurred()))
}

/*
Install K3s
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func InstallK3s() {
	// Get K3s installation script
	fileName := "k3s-install.sh"
	Eventually(func() error {
		return tools.GetFileFromURL("https://get.k3s.io", fileName, true)
	}, tools.SetTimeout(2*time.Minute), 10*time.Second).ShouldNot(HaveOccurred())

	// Set command and arguments
	installCmd := exec.Command("sh", fileName)
	installCmd.Env = append(os.Environ(), "INSTALL_K3S_EXEC=--disable metrics-server")

	// Retry in case of (sporadic) failure...
	count := 1
	Eventually(func() error {
		// Execute K3s installation
		out, err := installCmd.CombinedOutput()
		GinkgoWriter.Printf("K3s installation loop %d:\n%s\n", count, out)
		count++
		return err
	}, tools.SetTimeout(2*time.Minute), 5*time.Second).Should(Not(HaveOccurred()))
}

/*
Install Rancher Manager
  - @param k kubectl structure
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func InstallRancher(k *kubectl.Kubectl) {
	err := rancher.DeployRancherManager(rancherHostname, rancherChannel, rancherVersion, rancherHeadVersion, caType, proxy)
	Expect(err).To(Not(HaveOccurred()))

	checkList := [][]string{
		{"cattle-system", "app=rancher"},
		{"cattle-system", "app=rancher-webhook"},
		{"cattle-fleet-local-system", "app=fleet-agent"},
		{"cattle-provisioning-capi-system", "control-plane=controller-manager"},
	}
	Eventually(func() error {
		return rancher.CheckPod(k, checkList)
	}, tools.SetTimeout(10*time.Minute), 30*time.Second).Should(Not(HaveOccurred()))
}

/*
Install RKE2
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func InstallRKE2() {
	// Get RKE2 installation script
	fileName := "rke2-install.sh"
	Eventually(func() error {
		return tools.GetFileFromURL("https://get.rke2.io", fileName, true)
	}, tools.SetTimeout(2*time.Minute), 10*time.Second).ShouldNot(HaveOccurred())

	// Retry in case of (sporadic) failure...
	count := 1
	Eventually(func() error {
		// Execute RKE2 installation
		out, err := exec.Command("sudo", "--preserve-env=INSTALL_RKE2_VERSION", "sh", fileName).CombinedOutput()
		GinkgoWriter.Printf("RKE2 installation loop %d:\n%s\n", count, out)
		count++
		return err
	}, tools.SetTimeout(2*time.Minute), 5*time.Second).Should(Not(HaveOccurred()))
}

/*
Execute RunHelmBinaryWithCustomErr within a loop with timeout
  - @param s options to pass to RunHelmBinaryWithCustomErr command
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func RunHelmCmdWithRetry(s ...string) {
	Eventually(func() error {
		return kubectl.RunHelmBinaryWithCustomErr(s...)
	}, tools.SetTimeout(2*time.Minute), 20*time.Second).Should(Not(HaveOccurred()))
}

/*
Execute SSH command with retry
  - @param cl Client (node) informations
  - @param cmd Command to execute
  - @returns result of the executed command
*/
func RunSSHWithRetry(cl *tools.Client, cmd string) string {
	var err error
	var out string

	Eventually(func() error {
		out, err = cl.RunSSH(cmd)
		return err
	}, tools.SetTimeout(2*time.Minute), 20*time.Second).Should(Not(HaveOccurred()))

	return out
}

/*
Start K3s
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func StartK3s() {
	err := exec.Command("sudo", "systemctl", "start", "k3s").Run()
	Expect(err).To(Not(HaveOccurred()))
}

/*
Start RKE2
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func StartRKE2() {
	// Copy config file, this allows custom configuration for RKE2 installation
	// NOTE: CopyFile cannot be used, as we need root permissions for this file
	err := exec.Command("sudo", "mkdir", "-p", "/etc/rancher/rke2").Run()
	Expect(err).To(Not(HaveOccurred()))
	err = exec.Command("sudo", "cp", configRKE2Yaml, "/etc/rancher/rke2/config.yaml").Run()
	Expect(err).To(Not(HaveOccurred()))

	// Activate and start RKE2
	err = exec.Command("sudo", "systemctl", "enable", "--now", "rke2-server.service").Run()
	Expect(err).To(Not(HaveOccurred()))

	err = exec.Command("sudo", "ln", "-s", "/var/lib/rancher/rke2/bin/kubectl", "/usr/local/bin/kubectl").Run()
	Expect(err).To(Not(HaveOccurred()))
}

/*
Wait for cluster to be in a stable state
  - @param ns Namespace where the cluster is deployed
  - @param cn Cluster resource name
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func WaitCluster(ns, cn string) {
	type state struct {
		conditionStatus string
		conditionType   string
	}

	// List of conditions to check
	states := []state{
		{
			conditionStatus: "True",
			conditionType:   "AgentDeployed",
		},
		{
			conditionStatus: "True",
			conditionType:   "NoDiskPressure",
		},
		{
			conditionStatus: "True",
			conditionType:   "NoMemoryPressure",
		},
		{
			conditionStatus: "True",
			conditionType:   "Provisioned",
		},
		{
			conditionStatus: "True",
			conditionType:   "Ready",
		},
		{
			conditionStatus: "False",
			conditionType:   "Reconciling",
		},
		{
			conditionStatus: "False",
			conditionType:   "Stalled",
		},
		{
			conditionStatus: "True",
			conditionType:   "Updated",
		},
		{
			conditionStatus: "True",
			conditionType:   "Waiting",
		},
	}

	// Check that the cluster is in Ready state (this means that it has been created)
	Eventually(func() string {
		status, _ := kubectl.RunWithoutErr("get", "cluster.v1.provisioning.cattle.io",
			"--namespace", ns, cn,
			"-o", "jsonpath={.status.ready}")
		return status
	}, tools.SetTimeout(2*time.Duration(usedNodes)*time.Minute), 10*time.Second).Should(Equal("true"))

	// Check that all needed conditions are in the good state
	for _, s := range states {
		counter := 0

		Eventually(func() string {
			status, _ := kubectl.RunWithoutErr("get", "cluster.v1.provisioning.cattle.io",
				"--namespace", ns, cn,
				"-o", "jsonpath={.status.conditions[?(@.type==\""+s.conditionType+"\")].status}")

			if status != s.conditionStatus {
				// Show the status in case of issue, easier to debug (but log after 10 different issues)
				// NOTE: it's not perfect but it's mainly a way to inform that the cluster took time to came up
				counter++
				if counter > 10 {
					GinkgoWriter.Printf("!! Cluster status issue !! %s is %s instead of %s\n",
						s.conditionType, status, s.conditionStatus)

					// Reset counter
					counter = 0
				}

				// Check if rancher-system-agent.service has some issue
				if s.conditionType == "Provisioned" || s.conditionType == "Ready" || s.conditionType == "Updated" {
					msg := "error applying plan -- check rancher-system-agent.service logs on node for more information"

					// Extract the list of failed nodes
					listIP, _ := kubectl.RunWithoutErr("get", "machine",
						"--namespace", ns,
						"-o", "jsonpath={.items[?(@.status.conditions[*].message==\""+msg+"\")].status.addresses[?(@.type==\"InternalIP\")].address}")

					// We can try to restart the rancher-system-agent service on the failing node
					// because sometimes it can fail just because of a sporadic/timeout issue and a restart can fix it!
					for _, ip := range strings.Fields(listIP) {
						if tools.IsIPv4(ip) {
							// Set 'client' to be able to access the node through SSH
							cl := &tools.Client{
								Host:     ip + ":22",
								Username: userName,
								Password: userPassword,
							}

							// Log the workaround, could be useful
							GinkgoWriter.Printf("!! rancher-system-agent issue !! Service has been restarted on node with IP %s\n", ip)

							// Restart rancher-system-agent service on the node
							// NOTE: wait a little to be sure that all is restarted before continuing
							RunSSHWithRetry(cl, "systemctl restart rancher-system-agent.service")
							time.Sleep(tools.SetTimeout(15 * time.Second))
						}
					}
				}
			}

			return status
		}, tools.SetTimeout(2*time.Duration(usedNodes)*time.Minute), 10*time.Second).Should(Equal(s.conditionStatus))
	}
}

/*
Wait for K3s to start
  - @param k kubectl structure
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func WaitForK3s(k *kubectl.Kubectl) {
	checkList := [][]string{
		{"kube-system", "app=local-path-provisioner"},
		{"kube-system", "k8s-app=kube-dns"},
		{"kube-system", "app.kubernetes.io/name=traefik"},
		{"kube-system", "svccontroller.k3s.cattle.io/svcname=traefik"},
	}
	Eventually(func() error {
		return rancher.CheckPod(k, checkList)
	}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(Not(HaveOccurred()))
}

/*
Wait for RKE2 to start
  - @param k kubectl structure
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func WaitForRKE2(k *kubectl.Kubectl) {
	err := os.Setenv("KUBECONFIG", "/etc/rancher/rke2/rke2.yaml")
	Expect(err).To(Not(HaveOccurred()))

	checkList := [][]string{
		{"kube-system", "k8s-app=kube-dns"},
		{"kube-system", "app.kubernetes.io/name=rke2-ingress-nginx"},
	}
	Eventually(func() error {
		return rancher.CheckPod(k, checkList)
	}, tools.SetTimeout(4*time.Minute), 30*time.Second).Should(Not(HaveOccurred()))

	err = k.WaitLabelFilter("kube-system", "Ready", "rke2-ingress-nginx-controller", "app.kubernetes.io/name=rke2-ingress-nginx")
	Expect(err).To(Not(HaveOccurred()))
}

/*
Wait for OSVersion to be populated
  - @param ns Namespace where the cluster is deployed
  - @returns Nothing, the function will fail through Ginkgo in case of issue
*/
func WaitForOSVersion(ns string) {
	Eventually(func() string {
		out, _ := kubectl.RunWithoutErr("get", "ManagedOSVersion",
			"--namespace", ns,
			"-o", "jsonpath={.items[*].metadata.name}")
		return out
	}, tools.SetTimeout(2*time.Minute), 5*time.Second).Should(Not(BeEmpty()))
}

func FailWithReport(message string, callerSkip ...int) {
	// Ensures the correct line numbers are reported
	Fail(message, callerSkip[0]+1)
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(FailWithReport)
	RunSpecs(t, "Elemental End-To-End Test Suite")
}

// Use to modify yaml templates
type YamlPattern struct {
	key   string
	value string
}

var _ = BeforeSuite(func() {
	backupRestoreVersion = os.Getenv("BACKUP_RESTORE_VERSION")
	bootTypeString := os.Getenv("BOOT_TYPE")
	caType = os.Getenv("CA_TYPE")
	certManagerVersion = os.Getenv("CERT_MANAGER_VERSION")
	clusterName = os.Getenv("CLUSTER_NAME")
	clusterNS = os.Getenv("CLUSTER_NS")
	clusterType = os.Getenv("CLUSTER_TYPE")
	elementalSupport = os.Getenv("ELEMENTAL_SUPPORT")
	eTPM := os.Getenv("EMULATE_TPM")
	forceDowngradeString := os.Getenv("FORCE_DOWNGRADE")
	index := os.Getenv("VM_INDEX")
	k8sDownstreamVersion = os.Getenv("K8S_DOWNSTREAM_VERSION")
	k8sUpstreamVersion = os.Getenv("K8S_UPSTREAM_VERSION")
	number := os.Getenv("VM_NUMBERS")
	clusterNumber := os.Getenv("CLUSTER_NUMBER")
	operatorInstallType = os.Getenv("OPERATOR_INSTALL_TYPE")
	operatorRepo = os.Getenv("OPERATOR_REPO")
	operatorUpgrade = os.Getenv("OPERATOR_UPGRADE")
	os2Test = os.Getenv("OS_TO_TEST")
	poolType = os.Getenv("POOL")
	proxy = os.Getenv("PROXY")
	rancherHostname = os.Getenv("PUBLIC_FQDN")
	rancherLogCollector = os.Getenv("RANCHER_LOG_COLLECTOR")
	rancherVersion = os.Getenv("RANCHER_VERSION")
	rancherUpgrade = os.Getenv("RANCHER_UPGRADE")
	selinuxString := os.Getenv("SELINUX")
	seqString := os.Getenv("SEQUENTIAL")
	snapType = os.Getenv("SNAP_TYPE")
	testType = os.Getenv("TEST_TYPE")
	upgradeImage = os.Getenv("UPGRADE_IMAGE")
	upgradeOSChannel = os.Getenv("UPGRADE_OS_CHANNEL")
	upgradeType = os.Getenv("UPGRADE_TYPE")

	// Define boot type
	switch bootTypeString {
	case "iso":
		isoBoot = true
	case "raw":
		rawBoot = true
	}

	// Force correct value for emulateTPM
	switch eTPM {
	case "true":
		emulateTPM = true
	default:
		emulateTPM = false
	}

	// Force correct value for forceDowngrade
	switch forceDowngradeString {
	case "true":
		forceDowngrade = true
	default:
		forceDowngrade = false
	}

	// Only if VM_INDEX is set
	if index != "" {
		var err error
		vmIndex, err = strconv.Atoi(index)
		Expect(err).To(Not(HaveOccurred()))

		// Set default hostname
		vmName = elemental.SetHostname(vmNameRoot, vmIndex)
	} else {
		// Default value for vmIndex
		vmIndex = 0
	}

	// Only if VM_NUMBER is set
	if number != "" {
		var err error
		numberOfVMs, err = strconv.Atoi(number)
		Expect(err).To(Not(HaveOccurred()))
	} else {
		// By default set to vmIndex
		numberOfVMs = vmIndex
	}

	// Extract Rancher Manager channel/version to upgrade
	if rancherUpgrade != "" {
		// Split rancherUpgrade and reset it
		s := strings.Split(rancherUpgrade, "/")

		// Get needed informations
		rancherUpgradeChannel = s[0]
		if len(s) > 1 {
			rancherUpgradeVersion = s[1]
		}
		if len(s) > 2 {
			rancherUpgradeHeadVersion = s[2]
		}
	}

	// Extract Rancher Manager channel/version to install
	if rancherVersion != "" {
		// Split rancherVersion and reset it
		s := strings.Split(rancherVersion, "/")
		rancherVersion = ""

		// Get needed informations
		rancherChannel = s[0]
		if len(s) > 1 {
			rancherVersion = s[1]
		}
		if len(s) > 2 {
			rancherHeadVersion = s[2]
		}
	}

	// Force correct value for selinux
	switch selinuxString {
	case "true":
		selinux = true
	default:
		selinux = false
	}

	// Force correct value for sequential
	switch seqString {
	case "true":
		sequential = true
	default:
		sequential = false
	}

	// Depending of the OS version, the SSH config file could be at different places
	switch {
	case strings.Contains(os2Test, "dev"):
		sshdConfigFile = "/etc/ssh/sshd_config.d/root_access.conf"
	default:
		sshdConfigFile = "/etc/ssh/sshd_config"
	}

	// Define some variables depending on the type of test
	switch testType {
	case "airgap":
		// Enable airgap support
		clusterYaml = "../assets/cluster-airgap.yaml"
		netDefaultFileName = "../assets/net-default-airgap.xml"
		registrationYaml = "../assets/machineRegistration.yaml"
		seedImageYaml = "../assets/seedImage.yaml"
		selectorYaml = "../assets/selector.yaml"
	case "multi":
		// Enable multi-cluster support
		if clusterNumber != "" {
			var err error
			numberOfClusters, err = strconv.Atoi(clusterNumber)
			Expect(err).To(Not(HaveOccurred()))
		}

		clusterYaml = "../assets/cluster-multi.yaml"
		netDefaultFileName = "../assets/net-default.xml"
		registrationYaml = "../assets/machineRegistration-multi.yaml"
		seedImageYaml = "../assets/seedImage-multi.yaml"
		selectorYaml = "../assets/selector-multi.yaml"
	default:
		// Default cluster support
		clusterYaml = "../assets/cluster.yaml"
		netDefaultFileName = "../assets/net-default.xml"
		registrationYaml = "../assets/machineRegistration.yaml"
		seedImageYaml = "../assets/seedImage.yaml"
		selectorYaml = "../assets/selector.yaml"
	}

	// Set number of "used" nodes
	// NOTE: could be the number of added nodes or the number of nodes to use/upgrade
	usedNodes = (numberOfVMs - vmIndex) + 1

	// Final step: start local HTTP server
	tools.HTTPShare("../..", ":8000")
})

var _ = ReportBeforeEach(func(report SpecReport) {
	// Reset case ID
	testCaseID = -1
})

var _ = ReportAfterEach(func(report SpecReport) {
	// Add result in Qase if asked
	Qase(testCaseID, report)
})
