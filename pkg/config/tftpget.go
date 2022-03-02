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

package config

import (
	"bytes"
	"fmt"
	"net"
	"net/url"

	"gopkg.in/pin/tftp.v2"
	"sigs.k8s.io/yaml"
)

func tftpGet(tftpURL string) (map[string]interface{}, error) {
	u, err := url.Parse(tftpURL)
	if err != nil {
		return nil, err
	}

	host, _, err := net.SplitHostPort(u.Host)
	if err != nil {
		host = u.Host + ":69"
	}

	fmt.Printf("Downloading config from host %s, file %s\n", host, u.Path)
	client, err := tftp.NewClient(host)
	if err != nil {
		return nil, err
	}
	writerTo, err := client.Receive(u.Path, "octet")
	if err != nil {
		return nil, err
	}

	buf := &bytes.Buffer{}
	if _, err := writerTo.WriteTo(buf); err != nil {
		return nil, err
	}

	result := map[string]interface{}{}
	return result, yaml.Unmarshal(buf.Bytes(), &result)
}
