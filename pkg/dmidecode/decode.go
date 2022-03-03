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

package dmidecode

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	values "github.com/rancher/wrangler/pkg/data"
	"github.com/rancher/wrangler/pkg/kv"
)

func Decode() (map[string]interface{}, error) {
	buf := &bytes.Buffer{}
	cmd := exec.Command("dmidecode")
	cmd.Stdout = buf

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("looking up SMBIOS tables (using dmidecode): %w", err)
	}

	return dmiOutputToMap(buf), nil
}

func dmiOutputToMap(buf io.Reader) map[string]interface{} {
	var (
		result    = map[string]interface{}{}
		scanner   = bufio.NewScanner(buf)
		start     = false
		lastKey   []string
		stopLines = map[string]bool{
			"OEM-specific Type": true,
			"End Of Table":      true,
		}
	)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "Handle ") {
			start = true
			continue
		} else if strings.TrimSpace(line) == "" || !start || stopLines[line] {
			start = false
			continue
		}

		var key []string
		for strings.HasPrefix(line, "\t") {
			line = strings.TrimPrefix(line, "\t")
			if len(lastKey) > len(key) {
				key = append(key, lastKey[len(key)])
			}
		}
		name, value := kv.Split(line, ": ")
		key = append(key, name)

		if strings.TrimSpace(value) != "" || strings.Contains(line, ":") {
			values.PutValue(result, value, key...)
		} else if len(key) > 1 {
			parentKey := key[:len(key)-1]
			parentValue := values.GetValueN(result, parentKey...)
			if parentSlice, ok := parentValue.([]interface{}); ok {
				parentValue = append(parentSlice, name)
			} else {
				parentValue = []interface{}{name}
			}
			values.PutValue(result, parentValue, parentKey...)
		} else {
			values.PutValue(result, map[string]interface{}{}, key...)
		}

		lastKey = key
	}

	return result
}
