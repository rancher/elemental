package misc

import (
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
)

func GetServerId(clusterNS string) string {
	serverId, err := kubectl.Run("get", "MachineInventories",
		"--namespace", clusterNS,
		"-o", "jsonpath={.items[0].metadata.name}")
	Expect(err).NotTo(HaveOccurred())
	Expect(serverId).ToNot(Equal(""))

	return serverId
}
