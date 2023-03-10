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
	"strconv"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"github.com/rancher/elemental/tests/e2e/helpers/misc"
)

const (
	clusterYaml           = "../assets/cluster.yaml"
	emulateTPMYaml        = "../assets/emulateTPM.yaml"
	ciTokenYaml           = "../assets/local-kubeconfig-token-skel.yaml"
	configPrivateCAScript = "../scripts/config-private-ca"
	installConfigYaml     = "../../install-config.yaml"
	installHardenedScript = "../scripts/config-hardened"
	installVMScript       = "../scripts/install-vm"
	localKubeconfigYaml   = "../assets/local-kubeconfig-skel.yaml"
	netDefaultFileName    = "../assets/net-default.xml"
	osListYaml            = "../assets/managedOSVersionChannel.yaml"
	registrationYaml      = "../assets/machineRegistration.yaml"
	selectorYaml          = "../assets/selector.yaml"
	upgradeSkelYaml       = "../assets/upgrade_skel.yaml"
	userName              = "root"
	userPassword          = "r0s@pwd1"
	vmNameRoot            = "node"
)

var (
	arch                string
	caType              string
	CertManagerVersion  string
	clusterName         string
	clusterNS           string
	clusterType         string
	elementalSupport    string
	emulateTPM          bool
	rancherHostname     string
	imageVersion        string
	isoBoot             string
	k8sVersion          string
	numberOfVMs         int
	osImage             string
	poolType            string
	proxy               string
	rancherChannel      string
	rancherLogCollector string
	rancherVersion      string
	sequential          bool
	testType            string
	upgradeChannelList  string
	upgradeImage        string
	upgradeOperator     string
	upgradeOsChannel    string
	upgradeType         string
	usedNodes           int
	vmIndex             int
	vmName              string
)

func CheckClusterState(ns, cluster string) {
	// Check that a 'type' property named 'Ready' is set to true
	Eventually(func() string {
		clusterStatus, _ := kubectl.Run("get", "cluster",
			"--namespace", ns, cluster,
			"-o", "jsonpath={.status.conditions[?(@.type==\"Ready\")].status}")
		return clusterStatus
	}, misc.SetTimeout(2*time.Duration(usedNodes)*time.Minute), 10*time.Second).Should(Equal("True"))

	// Wait a little bit for the cluster to be in a stable state
	// Because if we do the next test too quickly it can be a false positive!
	// NOTE: no SetTimeout needed here!
	time.Sleep(30 * time.Second)

	// There should be no 'reason' property set in a clean cluster
	Eventually(func() string {
		reason, _ := kubectl.Run("get", "cluster",
			"--namespace", ns, cluster,
			"-o", "jsonpath={.status.conditions[*].reason}")
		return reason
	}, misc.SetTimeout(3*time.Duration(usedNodes)*time.Minute), 10*time.Second).Should(BeEmpty())
}

func GetNodeInfo(hostName string) (*tools.Client, string) {
	// Get network data
	hostData, err := tools.GetHostNetConfig(".*name=\""+hostName+"\".*", netDefaultFileName)
	Expect(err).To(Not(HaveOccurred()))

	// Set 'client' to be able to access the node through SSH
	c := &tools.Client{
		Host:     string(hostData.IP) + ":22",
		Username: userName,
		Password: userPassword,
	}

	return c, hostData.Mac
}

func FailWithReport(message string, callerSkip ...int) {
	// Ensures the correct line numbers are reported
	Fail(message, callerSkip[0]+1)
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(FailWithReport)
	RunSpecs(t, "Elemental End-To-End Test Suite")
}

var _ = BeforeSuite(func() {
	arch = os.Getenv("ARCH")
	caType = os.Getenv("CA_TYPE")
	CertManagerVersion = os.Getenv("CERT-MANAGER_VERSION")
	clusterName = os.Getenv("CLUSTER_NAME")
	clusterNS = os.Getenv("CLUSTER_NS")
	clusterType = os.Getenv("CLUSTER_TYPE")
	elementalSupport = os.Getenv("ELEMENTAL_SUPPORT")
	eTPM := os.Getenv("EMULATE_TPM")
	rancherHostname = os.Getenv("HOSTNAME")
	index := os.Getenv("VM_INDEX")
	isoBoot = os.Getenv("ISO_BOOT")
	k8sVersion = os.Getenv("K8S_VERSION_TO_PROVISION")
	number := os.Getenv("VM_NUMBERS")
	poolType = os.Getenv("POOL")
	proxy = os.Getenv("PROXY")
	rancherChannel = os.Getenv("RANCHER_CHANNEL")
	rancherLogCollector = os.Getenv("RANCHER_LOG_COLLECTOR")
	rancherVersion = os.Getenv("RANCHER_VERSION")
	seqString := os.Getenv("SEQUENTIAL")
	testType = os.Getenv("TEST_TYPE")
	upgradeChannelList = os.Getenv("UPGRADE_CHANNEL_LIST")
	upgradeImage = os.Getenv("UPGRADE_IMAGE")
	upgradeOperator = os.Getenv("UPGRADE_OPERATOR")
	upgradeOsChannel = os.Getenv("UPGRADE_OS_CHANNEL")
	upgradeType = os.Getenv("UPGRADE_TYPE")

	// Only if VM_INDEX is set
	if index != "" {
		var err error
		vmIndex, err = strconv.Atoi(index)
		Expect(err).To(Not(HaveOccurred()))

		// Set default hostname
		vmName = misc.SetHostname(vmNameRoot, vmIndex)
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

	// Set number of "used" nodes
	// NOTE: could be the number added nodes or the number of nodes to use/upgrade
	usedNodes = (numberOfVMs - vmIndex) + 1

	// Force a correct value for emulateTPM
	switch eTPM {
	case "true":
		emulateTPM = true
	default:
		emulateTPM = false
	}

	// Same for sequential
	switch seqString {
	case "true":
		sequential = true
	default:
		sequential = false
	}

	// Start HTTP server
	misc.FileShare("../..", ":8000")
})
