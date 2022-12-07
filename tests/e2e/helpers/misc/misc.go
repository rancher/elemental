package misc

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/rancher-sandbox/ele-testhelpers/kubectl"
	"github.com/rancher-sandbox/ele-testhelpers/tools"
	"gopkg.in/yaml.v3"
	libvirtxml "libvirt.org/libvirt-go-xml"
)

// Cluster is the definition of a K8s cluster
type Cluster struct {
	APIVersion string        `yaml:"apiVersion"`
	Kind       string        `yaml:"kind,omitempty"`
	Metadata   Metadata      `yaml:"metadata"`
	Spec       ClusterSpec   `yaml:"spec"`
	Status     ClusterStatus `yaml:"status,omitempty"`
}

// Metadata is metadata attached to any object
type Metadata struct {
	Annotations     interface{} `yaml:"annotations"`
	Labels          interface{} `yaml:"labels,omitempty"`
	Finalizers      interface{} `yaml:"finalizers,omitempty"`
	ManagedFields   interface{} `yaml:"managedFields,omitempty"`
	Name            string      `yaml:"name"`
	Namespace       string      `yaml:"namespace"`
	ResourceVersion string      `yaml:"resourceVersion"`
	UID             string      `yaml:"uid"`
}

// ClusterSpec is a description of a cluster
type ClusterSpec struct {
	KubernetesVersion        string      `yaml:"kubernetesVersion"`
	LocalClusterAuthEndpoint interface{} `yaml:"localClusterAuthEndpoint"`
	RkeConfig                RKEConfig   `yaml:"rkeConfig"`
}

// RKEConfig has all RKE/K3s cluster information
type RKEConfig struct {
	Etcd                interface{}    `yaml:"etcd,omitempty"`
	ChartValues         interface{}    `yaml:"chartValues"`
	MachineGlobalConfig interface{}    `yaml:"machineGlobalConfig"`
	MachinePools        []MachinePools `yaml:"machinePools"`
	UpgradeStrategy     interface{}    `yaml:"upgradeStrategy,omitempty"`
}

// MachinePools has all pools information
type MachinePools struct {
	ControlPlaneRole bool             `yaml:"controlPlaneRole,omitempty"`
	EtcdRole         bool             `yaml:"etcdRole,omitempty"`
	MachineConfigRef MachineConfigRef `yaml:"machineConfigRef"`
	Name             string           `yaml:"name"`
	Quantity         int              `yaml:"quantity"`
	WorkerRole       bool             `yaml:"workerRole,omitempty"`
}

// MachineConfigRef makes the link between the cluster, pool and the Elemental nodes
type MachineConfigRef struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Name       string `yaml:"name"`
}

// ClusterStatus has all the cluster status information
type ClusterStatus struct {
	AgentDeployed    bool               `yaml:"agentDeployed,omitempty"`
	ClientSecretName string             `yaml:"clientSecretName"`
	ClusterName      string             `yaml:"clusterName"`
	Conditions       []ClusterCondition `yaml:"conditions,omitempty"`
	Ready            bool               `yaml:"ready,omitempty"`
}

// ClusterCondition is the cluster condition status
type ClusterCondition struct {
	LastUpdateTime string `yaml:"lastUpdateTime"`
	Message        string `yaml:"message,omitempty"`
	Reason         string `yaml:"reason,omitempty"`
	Status         string `yaml:"status"`
	Type           string `yaml:"type"`
}

const (
	httpSrv = "http://192.168.122.1:8000"
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

func GetOperatorImage() (string, error) {
	operatorImage, err := kubectl.Run("get", "pod",
		"--namespace", "cattle-elemental-system",
		"-l", "app=elemental-operator", "-o", "jsonpath={.items[*].status.containerStatuses[*].image}")
	if err != nil {
		return "", err
	}

	return operatorImage, nil
}

func GetOperatorVersion() (string, error) {
	operatorImage, err := GetOperatorImage()
	if err != nil {
		return "", err
	}

	// Extract version
	operatorVersion := strings.Split(operatorImage, ":")

	return operatorVersion[1], nil
}

func ConfigureiPXE() (int, error) {
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

// Don't return error, in the worst case return the initial value
// Otherwise an additional step will be needed for some commands (like Eventually)
func SetTimeout(timeout time.Duration) time.Duration {
	s, set := os.LookupEnv("TIMEOUT_SCALE")

	// Only if TIMEOUT_SCALE is set
	if set {
		scale, err := strconv.Atoi(s)
		if err != nil {
			return timeout
		}

		// Return the scaled timeout
		return timeout * time.Duration(scale)
	}

	// Nothing to do
	return timeout
}

func IncreaseQuantity(clusterName, clusterNS, pool string) (int, error) {
	// Quantity set
	quantitySet := 0

	// Data to store
	c := &Cluster{}

	// Get content
	out, err := kubectl.Run("get", "cluster",
		"--namespace", clusterNS, clusterName,
		"-o", "yaml")
	if err != nil {
		return 0, err
	}

	// Convert string to []byte
	clusterConf := []byte(out)

	// Decode content
	if err := yaml.Unmarshal(clusterConf, &c); err != nil {
		return 0, err
	}

	// Increase quantity field
	for i := range c.Spec.RkeConfig.MachinePools {
		// Only on selected pool
		if c.Spec.RkeConfig.MachinePools[i].Name == pool {
			c.Spec.RkeConfig.MachinePools[i].Quantity++
			quantitySet = c.Spec.RkeConfig.MachinePools[i].Quantity
		}
	}

	// Encode content
	clusterConf, err = yaml.Marshal(&c)
	if err != nil {
		return 0, err
	}

	// Write temporary file
	tempFile := "updated-" + pool + ".yaml"
	if err = os.WriteFile(tempFile, clusterConf, 0644); err != nil {
		return 0, err
	}

	// Apply the new configuration
	err = kubectl.Apply(clusterNS, tempFile)
	if err != nil {
		return 0, err
	}

	// All good!
	return quantitySet, nil
}

func AddSelector(key, value string) ([]byte, error) {
	type selectorYaml struct {
		MatchLabels map[string]string `yaml:"matchLabels,omitempty"`
	}

	type selector struct {
		SelectorYaml selectorYaml `yaml:"nodeSelector,omitempty"`
	}

	v := selectorYaml{map[string]string{key: value}}
	s := selector{v}
	out, err := yaml.Marshal(s)
	if err != nil {
		return nil, err
	}

	// Add indent at the beginning
	out = append([]byte("  "), out...)

	return out, nil
}

func ConcateFiles(srcfile, dstfile string, data []byte) error {
	// Open source file
	f, err := os.Open(srcfile)
	if err != nil {
		return err
	}
	defer f.Close()

	// Open/create destination file
	d, err := os.OpenFile(dstfile, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer d.Close()

	// Copy source to dest
	if _, err = io.Copy(d, f); err != nil {
		return err
	}

	// Add data to dest
	if _, err = d.Write([]byte(data)); err != nil {
		return err
	}

	// All good!
	return nil
}

func CopyFile(srcFile, dstFile string) error {
	// Concate files without adding data is in fact a copy
	return (ConcateFiles(srcFile, dstFile, []byte("")))
}

func TrimStringFromChar(s, c string) string {
	if idx := strings.Index(s, c); idx != -1 {
		return s[:idx]
	}
	return s
}

func AddNode(name string, index int, file string) error {
	// Read live XML configuration
	fileContent, err := exec.Command("sudo", "virsh", "net-dumpxml", "default").Output()
	if err != nil {
		return err
	}

	// Unmarshal fileContent
	netcfg := &libvirtxml.Network{}
	if err := netcfg.Unmarshal(string(fileContent)); err != nil {
		return err
	}

	// Add new host
	// NOTE: we only use one network (IPs[0])
	host := libvirtxml.NetworkDHCPHost{
		Name: name,
		MAC:  "52:54:00:00:00:" + fmt.Sprintf("%02x", index),
		IP:   "192.168.122." + strconv.Itoa(index+1),
	}
	netcfg.IPs[0].DHCP.Hosts = append(netcfg.IPs[0].DHCP.Hosts, host)

	// Marshal new content
	newFileContent, err := netcfg.Marshal()
	if err != nil {
		return err
	}

	// Re-write XML file
	if err := os.WriteFile(file, []byte(newFileContent), 0644); err != nil {
		return err
	}

	// Update live network configuration
	xmlValue, err := host.Marshal()
	if err != nil {
		return err
	}
	if err := exec.Command("sudo", "virsh", "net-update",
		"default", "add", "ip-dhcp-host", "--live", "--xml", xmlValue).Run(); err != nil {
		return err
	}

	// All good!
	return nil
}
