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
	"os"
	"strconv"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

const (
	vmNameRoot         = "node"
	userName           = "root"
	userPassword       = "r0s@pwd1"
	netDefaultFileName = "../assets/net-default.xml"
)

var (
	k3sVersion  string
	clusterName string
	clusterNS   string
	osImage     string
	vmName      string
	upgradeType string
	vmIndex     int
)

func FailWithReport(message string, callerSkip ...int) {
	// Ensures the correct line numbers are reported
	Fail(message, callerSkip[0]+1)
}

func TestE2E(t *testing.T) {
	RegisterFailHandler(FailWithReport)
	RunSpecs(t, "End-To-End Test Suite")
}

var _ = BeforeSuite(func() {
	k3sVersion = os.Getenv("INSTALL_K3S_VERSION")
	clusterName = os.Getenv("CLUSTER_NAME")
	clusterNS = os.Getenv("CLUSTER_NS")
	osImage = os.Getenv("CONTAINER_IMAGE")
	upgradeType = os.Getenv("UPGRADE_TYPE")
	index, set := os.LookupEnv("VM_INDEX")

	// Only if VM_INDEX is set
	if set {
		var err error
		vmIndex, err = strconv.Atoi(index)
		Expect(err).To(Not(HaveOccurred()))

		// Now we can set the VM name
		vmName = vmNameRoot + "-" + fmt.Sprint(vmIndex)
	}
})
