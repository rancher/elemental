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

package tools

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

func GetFileFromURL(url string, fileName string, skipVerify bool) error {
	if !skipVerify {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}

	data, err := http.Get(url)
	if err != nil {
		return err
	}
	defer data.Body.Close()

	// Create file
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	// Save data in file
	_, err = io.Copy(file, data.Body)
	return err
}

func GetFiles(dir string, pattern string) ([]string, error) {
	files, err := filepath.Glob(dir + "/" + pattern)
	if err != nil {
		return nil, err
	}

	if files != nil {
		return files, nil
	}

	return nil, err
}

// Sed code partially from https://forum.golangbridge.org/t/using-sed-in-golang/23526/16
func Sed(oldValue, newValue, filePath string) error {
	fileData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}

	// Get file permissions
	info, err := os.Stat(filePath)
	if err != nil {
		return err
	}
	mode := info.Mode()

	// Regex is in the old value var
	regex := regexp.MustCompile(oldValue)
	fileString := string(fileData)
	fileString = regex.ReplaceAllString(fileString, newValue)
	fileData = []byte(fileString)

	err = ioutil.WriteFile(filePath, fileData, mode)
	return err
}

func HTTPShare(dir string, port int) error {
	// TODO: improve it to run in background!
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	sPort := strconv.Itoa(port)
	err := http.ListenAndServe(":"+sPort, nil)
	return err
}
