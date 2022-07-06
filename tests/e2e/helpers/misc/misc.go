package misc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

func GetServerId(clusterNS string, index int) (string, error) {
	serverId, err := kubectl.Run("get", "MachineInventories",
		"--namespace", clusterNS,
		"-o", "jsonpath={.items["+fmt.Sprint(index-1)+"].metadata.name}")
	if err != nil {
		return "", err
	}

	return serverId, nil
}

func ConfigureiPXE() (int, error) {
	ipxeScript, err := tools.GetFiles("../..", "*.ipxe")
	if err != nil {
		return 0, err
	}

	// NOTE: always use the first ipxe file found!
	if len(ipxeScript) >= 1 {
		err = tools.Sed("set url.*", "set url http://192.168.122.1:8000", ipxeScript[0])
		if err != nil {
			return 0, err
		}

		err = tools.Sed(".*set config.*", "set config $${url}/install-config.yaml", ipxeScript[0])
		if err != nil {
			return 0, err
		}

		// Delete the previous symlink, use RemoveAll to avoid error if the file doesn't exist
		symLink := "../../elemental.ipxe"
		err = os.RemoveAll(symLink)
		if err != nil {
			return 0, err
		}

		// Create a symlink to the needed ipxe file
		scriptName := filepath.Base(ipxeScript[0])
		err = os.Symlink(scriptName, symLink)
		if err != nil {
			return 0, err
		}
	}

	// Returns the number of ipxe files found
	return len(ipxeScript), nil
}
