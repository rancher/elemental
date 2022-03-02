package smoke_test

import (
	"fmt"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rancher-sandbox/os2/tests/sut"
)

func systemdUnitIsStarted(s string, st *sut.SUT) {
	out, _ := st.Command(fmt.Sprintf("systemctl status %s", s))

	Expect(out).To(And(
		ContainSubstring(fmt.Sprintf("%s.service; enabled", s)),
		ContainSubstring("status=0/SUCCESS"),
	))
}

var _ = Describe("os2 Smoke tests", func() {
	var s *sut.SUT
	BeforeEach(func() {
		s = sut.NewSUT()
		s.EventuallyConnects()
	})

	AfterEach(func() {
		if CurrentGinkgoTestDescription().Failed {
			s.Command("k3s kubectl get pods -A -o json > /tmp/pods.json")
			s.Command("k3s kubectl get events -A -o json > /tmp/events.json")
			s.Command("k3s kubectl get helmcharts -A -o json > /tmp/helm.json")
			s.Command("k3s kubectl get ingress -A -o json > /tmp/ingress.json")
			s.Command("df -h > /tmp/disk")
			s.Command("mount > /tmp/mounts")
			s.Command("blkid > /tmp/blkid")

			s.GatherAllLogs(
				[]string{
					"ros-installer",
					"cos-setup-boot",
					"cos-setup-network",
					"rancherd",
					"k3s",
				},
				[]string{
					"/tmp/pods.json",
					"/tmp/disk",
					"/tmp/mounts",
					"/tmp/blkid",
					"/tmp/events.json",
					"/tmp/helm.json",
					"/tmp/ingress.json",
				})
		}
	})

	Context("First boot", func() {
		for _, unit := range []string{"ros-installer", "rancherd", "cos-setup-rootfs"} {
			It(fmt.Sprintf("starts successfully %s on boot", unit), func() {
				systemdUnitIsStarted(unit, s)
			})
		}

		It("has default mounts", func() {
			out, _ := s.Command("mount")
			Expect(out).To(And(
				ContainSubstring("/var/lib/rancher"),
				ContainSubstring("/etc/ssh"),
				ContainSubstring("/etc/rancher"),
			))
		})

		It("has default cmdline", func() {
			out, _ := s.Command("cat /proc/cmdline")
			Expect(out).To(And(
				ContainSubstring("selinux=1"),
			))
		})

		// This test is flaky as it relies on dmesg output
		PIt("correctly starts cos services", func() {
			out, _ := s.Command("dmesg | grep cos")
			Expect(out).To(And(
				ContainSubstring("cos-immutable-rootfs.service: Succeeded"),
				ContainSubstring("cos-setup-rootfs.service: Succeeded"),
			))
		})

		// Added user via cloud-init is functional
		It("has the user added via cloud-init", func() {
			out, _ := s.Command(`su - vagrant -c 'id -un'`)
			Expect(out).To(Equal("vagrant\n"))

			out, _ = s.Command(`cat /run/vagrant/.ssh/authorized_keys`)
			Expect(out).To(ContainSubstring("ssh-rsa"))
			out, _ = s.Command(`sudo cat /root/.ssh/authorized_keys`)
			Expect(out).To(ContainSubstring("ssh-rsa"))
		})
	})

	Context("rancherd", func() {
		It("starts a single-node cluster", func() {
			err := s.SendFile("../assets/rancherd.yaml", "/oem/99_custom.yaml", "0770")
			Expect(err).ToNot(HaveOccurred())
			s.Command("systemctl restart --no-block rancherd")
			Eventually(func() string {
				out, _ := s.Command("ps aux")
				return out
			}, 230*time.Second, 1*time.Second).Should(ContainSubstring("k3s server"))

			systemdUnitIsStarted("k3s", s)

			Eventually(func() string {
				out, _ := s.Command("k3s kubectl get nodes -o wide")
				return out
			}, 230*time.Second, 1*time.Second).Should(ContainSubstring("Ready"))
		})
	})

	Context("ros-operator", func() {
		It("installs and create a machine registration resource", func() {
			chart := os.Getenv("ROS_CHART")
			if chart == "" {
				Skip("No chart provided, skipping tests")
			}
			err := s.SendFile(chart, "/usr/local/ros.tgz", "0770")
			Expect(err).ToNot(HaveOccurred())

			err = s.SendFile("../assets/machineregistration.yaml", "/usr/local/machine.yaml", "0770")
			Expect(err).ToNot(HaveOccurred())

			err = s.SendFile("../assets/external_charts.yaml", "/usr/local/charts.yaml", "0770")
			Expect(err).ToNot(HaveOccurred())

			err = s.SendFile("../assets/setting.yaml", "/usr/local/setting.yaml", "0770")
			Expect(err).ToNot(HaveOccurred())

			By("Having running pods", func() {
				Eventually(func() string {
					out, _ := s.Command("k3s kubectl get pods --all-namespaces")
					return out
				}, 5*time.Minute, 2*time.Second).Should(ContainSubstring("Running"))
			})

			By("Applying charts", func() {
				Eventually(func() string {
					out, _ := s.Command("k3s kubectl apply -f /usr/local/charts.yaml")
					return out
				}, 5*time.Minute, 2*time.Second).Should(
					Or(
						ContainSubstring("unchanged"),
						ContainSubstring("configured"),
					),
				)
			})

			By("having cert-manager", func() {
				Eventually(func() string {
					out, _ := s.Command("k3s kubectl get pods --all-namespaces")
					return out
				}, 6*time.Minute, 2*time.Second).Should(ContainSubstring("cert-manager-cainjector"))
			})

			By("having rancher", func() {
				Eventually(func() string {
					out, _ := s.Command("k3s kubectl get svc --all-namespaces")
					return out
				}, 6*time.Minute, 2*time.Second).Should(ContainSubstring("rancher"))
			})

			By("installing the ros-operator chart", func() {

				// Applies the setting required to generate a correct registrationURL.
				// normally this is generated by the user from the graphical interface, but we don't.
				Eventually(func() string {
					out, _ := s.Command("k3s kubectl apply -f /usr/local/setting.yaml")
					return out
				}, 15*time.Minute, 30*time.Second).Should(
					Or(
						ContainSubstring("unchanged"),
						ContainSubstring("configured"),
					),
				)

				Eventually(func() string {
					out, _ := s.Command("KUBECONFIG=/etc/rancher/k3s/k3s.yaml helm -n cattle-rancheros-operator-system install --create-namespace rancheros-operator /usr/local/ros.tgz")
					return out
				}, 15*time.Minute, 2*time.Second).Should(ContainSubstring("STATUS: deployed"))

				Eventually(func() string {
					out, _ := s.Command("k3s kubectl get pods --all-namespaces")
					return out
				}, 15*time.Minute, 2*time.Second).Should(ContainSubstring("rancheros-operator-"))
			})

			By("adding a machine registration", func() {
				Eventually(func() string {
					out, _ := s.Command("k3s kubectl apply -f /usr/local/machine.yaml")
					return out
				}, 30*time.Minute, 1*time.Second).Should(
					Or(
						ContainSubstring("unchanged"),
						ContainSubstring("configured"),
					),
				)

				var url string
				Eventually(func() string {
					out, _ := s.Command("k3s kubectl get machineregistration -n fleet-default machine-registration -o json | jq '.status.registrationURL' -r")
					url = out
					return out
				}, 15*time.Minute, 1*time.Second).Should(ContainSubstring("127.0.0.1.nip.io/v1-rancheros/registration"))

				Eventually(func() string {
					out, _ := s.Command("curl -k -L " + url)
					return out
				}, 5*time.Minute, 1*time.Second).Should(And(ContainSubstring("BEGIN CERTIFICATE"), ContainSubstring("127.0.0.1.nip.io/v1-rancheros/registration")))
			})
		})
	})
})
