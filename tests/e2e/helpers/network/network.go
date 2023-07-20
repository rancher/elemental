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

package network

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/rancher-sandbox/ele-testhelpers/tools"
)

/**
 * Configure iPXE server for OS provisioning
 * @remarks An iPXE server is up and running
 * @param httpSrv IP address:port where the files are shared
 * @returns The number of .ipxe files found or an error
 */
func ConfigureiPXE(httpSrv string) (int, error) {
	ipxeScript, err := tools.GetFiles("../..", "*.ipxe")
	if err != nil {
		return 0, err
	}

	// NOTE: always use the first ipxe file found!
	if len(ipxeScript) >= 1 {
		err = tools.Sed("set url.*", "set url "+httpSrv, ipxeScript[0])
		if err != nil {
			return 0, err
		}

		err = tools.Sed(".*set config.*", "set config $${url}/install-config.yaml", ipxeScript[0])
		if err != nil {
			return 0, err
		}

		// Delete the previous symlink, use RemoveAll to avoid error if the file doesn't exist
		symLink := "../../install-elemental.ipxe"
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

/**
 * Share files through HTTP (simple way, no security at all!)
 * @remarks A HTTP server is up and running
 * @param directory The directory where is files are
 * @param listenAddr Port where to listen to
 */
func HttpShare(directory, listenAddr string) {
	fs := http.FileServer(http.Dir(directory))

	go func() {
		if err := http.ListenAndServe(listenAddr, fs); err != nil {
			fmt.Printf("Server failed: %s\n", err)
		}
	}()
}
