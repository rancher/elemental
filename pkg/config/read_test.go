package config_test

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/rancher-sandbox/os2/pkg/config"
)

var _ = Describe("os2 config unit tests", func() {

	var c Config

	BeforeEach(func() {
		c = Config{}
	})

	Context("convert to environment configuration", func() {
		It("handle empty config", func() {
			c = Config{}
			e, err := ToEnv(c)
			Expect(err).ToNot(HaveOccurred())
			Expect(e).To(BeEmpty())
		})

		It("converts to env slice installation parameters", func() {
			c = Config{
				Data: map[string]interface{}{
					"random": "data",
				},
				SSHAuthorizedKeys: []string{"github:mudler"},
				RancherOS: RancherOS{
					Install: Install{
						Automatic:       true,
						ForceEFI:        true,
						RegistrationURL: "Foo",
						ISOURL:          "http://foo.bar",
					},
				},
			}
			e, err := ToEnv(c)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(e)).To(Equal(5))
			Expect(e).To(
				ContainElements(
					"SSH_AUTHORIZED_KEYS=[github:mudler]",
					"_COS_INSTALL_AUTOMATIC=true",
					"_COS_INSTALL_REGISTRATION_URL=Foo",
					"_COS_INSTALL_ISO_URL=http://foo.bar",
					"_COS_INSTALL_FORCE_EFI=true",
				),
			)
		})
	})

	Context("reading config file", func() {
		It("reads iso_url and registrationUrl", func() {
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  install:
    registrationUrl: "foobaz"
    iso_url: "foo_bar"
`), os.ModePerm)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			c, err := ReadConfig(ctx, f.Name(), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RancherOS.Install.RegistrationURL).To(Equal("foobaz"))
			Expect(c.RancherOS.Install.ISOURL).To(Equal("foo_bar"))
		})

		It("reads iso_url only, without contacting a registrationUrl server", func() {
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  install:
    iso_url: "foo_bar"
`), os.ModePerm)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			c, err := ReadConfig(ctx, f.Name(), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RancherOS.Install.ISOURL).To(Equal("foo_bar"))
		})

		It("reads isoUrl instead of iso_url", func() {
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  install:
    isoUrl: "foo_bar"
`), os.ModePerm)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			c, err := ReadConfig(ctx, f.Name(), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RancherOS.Install.ISOURL).To(Equal("foo_bar"))
		})

		It("reads ssh_authorized_keys", func() {
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			ioutil.WriteFile(f.Name(), []byte(`
ssh_authorized_keys:
- foo
`), os.ModePerm)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			c, err := ReadConfig(ctx, f.Name(), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.SSHAuthorizedKeys).To(Equal([]string{"foo"}))
		})
	})

	Context("writing config", func() {
		It("uses cloud-init format, but if data is present, takes over", func() {
			c = Config{
				Data: map[string]interface{}{
					"users": []struct {
						User string `json:"user"`
						Pass string `json:"pass"`
					}{{"foo", "Bar"}},
				},
				SSHAuthorizedKeys: []string{"github:mudler"},
				RancherOS: RancherOS{
					Install: Install{
						Automatic:       true,
						ForceEFI:        true,
						RegistrationURL: "Foo",
						ISOURL:          "http://foo.bar",
					},
				},
			}

			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			err = ToFile(c, f.Name())
			Expect(err).ToNot(HaveOccurred())

			ff, _ := ioutil.ReadFile(f.Name())
			Expect(string(ff)).To(Equal("#cloud-config\nusers:\n- pass: Bar\n  user: foo\n"))
		})

		It("writes cloud-init files", func() {
			c = Config{
				SSHAuthorizedKeys: []string{"github:mudler"},
				RancherOS: RancherOS{
					Install: Install{
						Automatic:       true,
						ForceEFI:        true,
						RegistrationURL: "Foo",
						ISOURL:          "http://foo.bar",
					},
				},
			}

			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			err = ToFile(c, f.Name())
			Expect(err).ToNot(HaveOccurred())

			ff, _ := ioutil.ReadFile(f.Name())
			Expect(string(ff)).To(Equal("#cloud-config\nrancheros: {}\nssh_authorized_keys:\n- github:mudler\n"))
		})
	})
})
