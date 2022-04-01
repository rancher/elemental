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

// Get file from URL
func GetFileFromUrl(url string, fileName string, skipVerify bool) error {
	if skipVerify == false {
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

// Get files' list
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

// Partially from https://forum.golangbridge.org/t/using-sed-in-golang/23526/16
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

// Share a directory through HTTP
// TODO: improve it to run in background!
func HttpShare(dir string, port int) error {
	fs := http.FileServer(http.Dir(dir))
	http.Handle("/", fs)

	sPort := strconv.Itoa(port)
	err := http.ListenAndServe(":"+sPort, nil)
	return err
}
