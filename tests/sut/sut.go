package sut

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/bramvdbogaerde/go-scp"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	ssh "golang.org/x/crypto/ssh"
)

const (
	grubSwapOnce = "grub2-editenv /oem/grubenv set next_entry=%s"
	grubSwap     = "grub2-editenv /oem/grubenv set saved_entry=%s"

	Passive     = 0
	Active      = iota
	Recovery    = iota
	LiveCD      = iota
	UnknownBoot = iota

	TimeoutRawDiskTest = 600 // Timeout to connect for recovery_raw_disk_test

	Ext2 = "ext2"
	Ext3 = "ext3"
	Ext4 = "ext4"
)

// DiskLayout is the struct that contains the disk output from lsblk
type DiskLayout struct {
	BlockDevices []PartitionEntry `json:"blockdevices"`
}

// PartitionEntry represents a partition entry
type PartitionEntry struct {
	Label  string `json:"label,omitempty"`
	Size   int    `json:"size,omitempty"`
	FsType string `json:"fstype,omitempty"`
}

func (d DiskLayout) GetPartition(label string) (PartitionEntry, error) {
	for _, device := range d.BlockDevices {
		if device.Label == label {
			return device, nil
		}
	}
	return PartitionEntry{}, nil
}

type SUT struct {
	Host        string
	Username    string
	Password    string
	Timeout     int
	GreenRepo   string
	TestVersion string
	CDLocation  string
}

func NewSUT() *SUT {

	user := os.Getenv("COS_USER")
	if user == "" {
		user = "root"
	}
	pass := os.Getenv("COS_PASS")
	if pass == "" {
		pass = "cos"
	}

	host := os.Getenv("COS_HOST")
	if host == "" {
		host = "127.0.0.1:2222"
	}

	var timeout = 180
	valueStr := os.Getenv("COS_TIMEOUT")
	value, err := strconv.Atoi(valueStr)
	if err == nil {
		timeout = value
	}

	return &SUT{
		Host:        host,
		Username:    user,
		Password:    pass,
		Timeout:     timeout,
		GreenRepo:   "quay.io/costoolkit/releases-green",
		TestVersion: "0.7.11-5",
		CDLocation:  "",
	}
}

func (s *SUT) ChangeBoot(b int) error {

	var bootEntry string

	switch b {
	case Active:
		bootEntry = "cos"
	case Passive:
		bootEntry = "fallback"
	case Recovery:
		bootEntry = "recovery"
	}

	_, err := s.command(fmt.Sprintf(grubSwap, bootEntry), false)
	Expect(err).ToNot(HaveOccurred())

	return nil
}

func (s *SUT) ChangeBootOnce(b int) error {

	var bootEntry string

	switch b {
	case Active:
		bootEntry = "cos"
	case Passive:
		bootEntry = "fallback"
	case Recovery:
		bootEntry = "recovery"
	}

	_, err := s.command(fmt.Sprintf(grubSwapOnce, bootEntry), false)
	Expect(err).ToNot(HaveOccurred())

	return nil
}

// Reset runs reboots cOS into Recovery and runs cos-reset.
// It will boot back the system from the Active partition afterwards
func (s *SUT) Reset() {
	if s.BootFrom() != Recovery {
		By("Reboot to recovery before reset")
		err := s.ChangeBootOnce(Recovery)
		Expect(err).ToNot(HaveOccurred())
		s.Reboot()
		Expect(s.BootFrom()).To(Equal(Recovery))
	}

	By("Running cos-reset")
	out, err := s.command("cos-reset", false)
	Expect(err).ToNot(HaveOccurred())
	Expect(out).Should(ContainSubstring("Installing"))

	By("Reboot to active after cos-reset")
	s.Reboot()
	ExpectWithOffset(1, s.BootFrom()).To(Equal(Active))
}

// BootFrom returns the booting partition of the SUT
func (s *SUT) BootFrom() int {
	out, err := s.command("cat /proc/cmdline", false)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	switch {
	case strings.Contains(out, "COS_ACTIVE"):
		return Active
	case strings.Contains(out, "COS_PASSIVE"):
		return Passive
	case strings.Contains(out, "COS_RECOVERY"), strings.Contains(out, "COS_SYSTEM"):
		return Recovery
	case strings.Contains(out, "live:CDLABEL"):
		return LiveCD
	default:
		return UnknownBoot
	}
}

// SquashFSRecovery returns true if we are in recovery mode and booting from squashfs
func (s *SUT) SquashFSRecovery() bool {
	out, err := s.command("cat /proc/cmdline", false)
	ExpectWithOffset(1, err).ToNot(HaveOccurred())

	return strings.Contains(out, "rd.live.squashimg")
}

func (s *SUT) GetOSRelease(ss string) string {
	out, err := s.Command(fmt.Sprintf("source /etc/os-release && echo $%s", ss))
	Expect(err).ToNot(HaveOccurred())
	Expect(out).ToNot(Equal(""))

	return strings.TrimSpace(out)
}

func (s *SUT) GetArch() string {
	out, err := s.Command("uname -p")
	Expect(err).ToNot(HaveOccurred())
	Expect(out).ToNot(Equal(""))

	return strings.TrimSpace(out)
}

func (s *SUT) EventuallyConnects(t ...int) {
	dur := s.Timeout
	if len(t) > 0 {
		dur = t[0]
	}
	Eventually(func() error {
		out, err := s.command("echo ping", true)
		if out == "ping\n" {
			return nil
		}
		return err
	}, time.Duration(time.Duration(dur)*time.Second), time.Duration(5*time.Second)).ShouldNot(HaveOccurred())
}

// Command sends a command to the SUIT and waits for reply
func (s *SUT) Command(cmd string) (string, error) {
	return s.command(cmd, false)
}

func (s *SUT) command(cmd string, timeout bool) (string, error) {
	client, err := s.connectToHost(timeout)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}

	out, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(out), errors.Wrap(err, string(out))
	}

	return string(out), err
}

// Reboot reboots the system under test
func (s *SUT) Reboot(t ...int) {
	By("Reboot")
	s.command("reboot", true)
	time.Sleep(10 * time.Second)
	s.EventuallyConnects(t...)
}

func (s *SUT) clientConfig() *ssh.ClientConfig {
	sshConfig := &ssh.ClientConfig{
		User:    s.Username,
		Auth:    []ssh.AuthMethod{ssh.Password(s.Password)},
		Timeout: 30 * time.Second, // max time to establish connection
	}
	sshConfig.HostKeyCallback = ssh.InsecureIgnoreHostKey()

	return sshConfig
}

func (s *SUT) SendFile(src, dst, permission string) error {
	sshConfig := s.clientConfig()
	scpClient := scp.NewClientWithTimeout(s.Host, sshConfig, 10*time.Second)
	defer scpClient.Close()

	if err := scpClient.Connect(); err != nil {
		return err
	}

	f, err := os.Open(src)
	if err != nil {
		return err
	}

	defer scpClient.Close()
	defer f.Close()

	if err := scpClient.CopyFile(f, dst, permission); err != nil {
		return err
	}
	return nil
}

func (s *SUT) connectToHost(timeout bool) (*ssh.Client, error) {
	sshConfig := s.clientConfig()

	client, err := DialWithDeadline("tcp", s.Host, sshConfig, timeout)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// GatherAllLogs will try to gather as much info from the system as possible, including services, dmesg and os related info
func (s SUT) GatherAllLogs(services []string, logFiles []string) {
	// services
	for _, ser := range services {
		out, err := s.command(fmt.Sprintf("journalctl -u %s -o short-iso >> /tmp/%s.log", ser, ser), true)
		if err != nil {
			fmt.Printf("Error getting journal for service %s: %s\n", ser, err.Error())
			fmt.Printf("Output from command: %s\n", out)
		}
		s.GatherLog(fmt.Sprintf("/tmp/%s.log", ser))
	}

	// log files
	for _, file := range logFiles {
		s.GatherLog(file)
	}

	// dmesg
	out, err := s.command("dmesg > /tmp/dmesg", true)
	if err != nil {
		fmt.Printf("Error getting dmesg : %s\n", err.Error())
		fmt.Printf("Output from command: %s\n", out)
	}
	s.GatherLog("/tmp/dmesg")

	// grab full journal
	out, err = s.command("journalctl -o short-iso > /tmp/journal.log", true)
	if err != nil {
		fmt.Printf("Error getting full journalctl info : %s\n", err.Error())
		fmt.Printf("Output from command: %s\n", out)
	}
	s.GatherLog("/tmp/journal.log")

	// uname
	out, err = s.command("uname -a > /tmp/uname.log", true)
	if err != nil {
		fmt.Printf("Error getting uname info : %s\n", err.Error())
		fmt.Printf("Output from command: %s\n", out)
	}
	s.GatherLog("/tmp/uname.log")

	// disk info
	out, err = s.command("lsblk -a >> /tmp/disks.log", true)
	if err != nil {
		fmt.Printf("Error getting disk info : %s\n", err.Error())
		fmt.Printf("Output from command: %s\n", out)
	}
	out, err = s.command("blkid >> /tmp/disks.log", true)
	if err != nil {
		fmt.Printf("Error getting disk info : %s\n", err.Error())
		fmt.Printf("Output from command: %s\n", out)
	}
	s.GatherLog("/tmp/disks.log")

	// Grab users
	s.GatherLog("/etc/passwd")
	// Grab system info
	s.GatherLog("/etc/os-release")

}

// GatherLog will try to scp the given log from the machine to a local file
func (s SUT) GatherLog(logPath string) {
	fmt.Printf("Trying to get file: %s\n", logPath)
	sshConfig := s.clientConfig()
	scpClient := scp.NewClientWithTimeout(s.Host, sshConfig, 10*time.Second)

	err := scpClient.Connect()
	if err != nil {
		scpClient.Close()
		fmt.Println("Couldn't establish a connection to the remote server ", err)
		return
	}

	fmt.Printf("Connection to %s established!\n", s.Host)
	baseName := filepath.Base(logPath)
	_ = os.Mkdir("logs", 0755)

	f, _ := os.Create(fmt.Sprintf("logs/%s", baseName))
	// Close the file after it has been copied
	// Close client connection after the file has been copied
	defer scpClient.Close()
	defer f.Close()

	err = scpClient.CopyFromRemote(f, logPath)

	if err != nil {
		fmt.Printf("Error while copying file: %s\n", err.Error())
		return
	}
	// Change perms so its world readable
	_ = os.Chmod(fmt.Sprintf("logs/%s", baseName), 0666)
	fmt.Printf("File %s copied!\n", baseName)

}

// EmptyDisk will try to trash the disk given so on reboot the disk is empty and we are forced to use the cd to boot
// used mainly for installer testing booting from iso
func (s *SUT) EmptyDisk(disk string) {
	By(fmt.Sprintf("Trashing %s to restore VM to a blank state", disk))
	_, _ = s.Command(fmt.Sprintf("wipefs -af %s*", disk))
	_, _ = s.Command("sync")
	_, _ = s.Command("sleep 5")
}

// SetCOSCDLocation gets the location of the iso attached to the vbox vm and stores it for later remount
func (s *SUT) SetCOSCDLocation() {
	By("Store CD location")
	out, err := exec.Command("bash", "-c", "VBoxManage list dvds|grep Location|cut -d ':' -f 2|xargs").CombinedOutput()
	Expect(err).To(BeNil())
	s.CDLocation = strings.TrimSpace(string(out))
}

// EjectCOSCD force removes the DVD so we can boot from disk directly on EFI VMs
func (s *SUT) EjectCOSCD() {
	// first store the cd location
	s.SetCOSCDLocation()
	By("Ejecting the CD")
	_, err := exec.Command("bash", "-c", "VBoxManage storageattach 'test' --storagectl 'sata controller' --port 1 --device 0 --type dvddrive --medium emptydrive --forceunmount").CombinedOutput()
	Expect(err).To(BeNil())
}

// RestoreCOSCD reattaches the cOS iso to the VM
func (s *SUT) RestoreCOSCD() {
	By("Restoring the CD")
	out, err := exec.Command("bash", "-c", fmt.Sprintf("VBoxManage storageattach 'test' --storagectl 'sata controller' --port 1 --device 0 --type dvddrive --medium %s --forceunmount", s.CDLocation)).CombinedOutput()
	fmt.Printf(string(out))
	Expect(err).To(BeNil())
}

func (s SUT) GetDiskLayout(disk string) DiskLayout {
	// -b size in bytes
	// -n no headings
	// -J json output
	diskLayout := DiskLayout{}
	out, err := s.Command(fmt.Sprintf("lsblk %s -o LABEL,SIZE,FSTYPE -b -n -J", disk))
	Expect(err).To(BeNil())
	err = json.Unmarshal([]byte(strings.TrimSpace(out)), &diskLayout)
	Expect(err).To(BeNil())
	return diskLayout
}

// DialWithDeadline Dials SSH with a deadline to avoid Read timeouts
func DialWithDeadline(network string, addr string, config *ssh.ClientConfig, timeout bool) (*ssh.Client, error) {
	conn, err := net.DialTimeout(network, addr, config.Timeout)
	if err != nil {
		return nil, err
	}
	if config.Timeout > 0 {
		conn.SetReadDeadline(time.Now().Add(config.Timeout))
		conn.SetWriteDeadline(time.Now().Add(config.Timeout))
	}
	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		return nil, err
	}
	if !timeout {
		conn.SetReadDeadline(time.Time{})
		conn.SetWriteDeadline(time.Time{})
	}

	go func() {
		t := time.NewTicker(2 * time.Second)
		defer t.Stop()
		for range t.C {
			_, _, err := c.SendRequest("keepalive@golang.org", true, nil)
			if err != nil {
				return
			}
		}
	}()
	return ssh.NewClient(c, chans, reqs), nil
}

func (s *SUT) WriteInlineFile(content, path string) {
	_, err := s.Command(`cat << EOF > ` + path + `
` + content + `
EOF`)
	Expect(err).ToNot(HaveOccurred())
}
