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

package config_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	gotpm "github.com/rancher-sandbox/go-tpm"
	values "github.com/rancher/wrangler/pkg/data"

	"github.com/gorilla/websocket"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "github.com/rancher-sandbox/os2/pkg/config"
)

func writeRead(conn *websocket.Conn, input []byte) ([]byte, error) {
	writer, err := conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return nil, err
	}

	if _, err := writer.Write(input); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	_, reader, err := conn.NextReader()
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(reader)
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Mimics a WS server which accepts TPM Bearer token
func WSServer(ctx context.Context, data map[string]interface{}) {
	s := http.Server{
		Addr:         "127.0.0.1:9980",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	m := http.NewServeMux()
	m.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		conn, _ := upgrader.Upgrade(w, r, nil) // error ignored for sake of simplicity

		for {

			token := r.Header.Get("Authorization")
			ek, at, err := gotpm.GetAttestationData(token)
			if err != nil {
				fmt.Println("error", err.Error())
				return
			}

			secret, challenge, err := gotpm.GenerateChallenge(ek, at)
			if err != nil {
				fmt.Println("error", err.Error())
				return
			}

			resp, _ := writeRead(conn, challenge)

			if err := gotpm.ValidateChallenge(secret, resp); err != nil {
				fmt.Println(string(resp))
				fmt.Println("error validating challenge", err.Error())
				return
			}

			writer, _ := conn.NextWriter(websocket.BinaryMessage)
			_ = json.NewEncoder(writer).Encode(data)
		}
	})

	s.Handler = m

	go s.ListenAndServe()
	go func() {
		<-ctx.Done()
		_ = s.Shutdown(ctx)
	}()
}

var _ = Describe("os2 config unit tests", func() {

	var c Config
	var data map[string]interface{}

	BeforeEach(func() {
		c = Config{}
		data = map[string]interface{}{
			"rancheros": map[string]interface{}{
				"install": map[string]string{
					"isoUrl": "foo",
				},
			},
		}
	})

	Context("Validation", func() {
		It("fails if isoUrl and containerImage are both used at the same time", func() {
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			_ = ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  install:
    containerImage: "docker/image:test"
    isoUrl: "test"
`), os.ModePerm)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			_, err = ReadConfig(ctx, f.Name(), false)
			Expect(err).To(HaveOccurred())
		})
		It("fails if isoUrl and containerImage are both empty", func() {
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			_ = ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  tpm:
    emulated: true
    no_smbios: true
    seed: "5"
  install:
    registrationUrl: "http://127.0.0.1:9980/test"
`), os.ModePerm)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			// Empty the install key so there is no isourl nor containerImage
			values.PutValue(data, "", "rancheros", "install")
			WSServer(ctx, data)
			_, err = ReadConfig(ctx, f.Name(), false)
			Expect(err).To(HaveOccurred())
		})

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
					// Those settings below are tied to the
					// elemental installer.
					Install: Install{
						Device:          "foob",
						ConfigURL:       "fooc",
						ForceEFI:        true,
						RegistrationURL: "Foo",
						ISOURL:          "http://foo.bar",
						NoFormat:        true,
						Debug:           true,
						PowerOff:        true,
						TTY:             "foo",
						ContainerImage:  "container",
					},
				},
			}
			e, err := ToEnv(c)
			Expect(err).ToNot(HaveOccurred())
			Expect(len(e)).To(Equal(11))
			Expect(e).To(
				ContainElements(
					"SSH_AUTHORIZED_KEYS=[github:mudler]",
					"ELEMENTAL_TARGET=foob",
					"ELEMENTAL_CLOUD_INIT=fooc",
					"ELEMENTAL_FORCE_EFI=true",
					"ELEMENTAL_REGISTRATION_URL=Foo",
					"ELEMENTAL_ISO=http://foo.bar",
					"ELEMENTAL_NO_FORMAT=true",
					"ELEMENTAL_DEBUG=true",
					"ELEMENTAL_POWEROFF=true",
					"ELEMENTAL_TTY=foo",
					"ELEMENTAL_DOCKER_IMAGE=container",
				),
			)
		})
	})

	Context("reading config file", func() {
		It("reads iso_url and registrationUrl", func() {
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			_ = ioutil.WriteFile(f.Name(), []byte(`
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

			_ = ioutil.WriteFile(f.Name(), []byte(`
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

		It("reads containerImage, without contacting a registrationUrl server", func() {
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			_ = ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  install:
    containerImage: "docker/image:test"
`), os.ModePerm)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			c, err := ReadConfig(ctx, f.Name(), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RancherOS.Install.ISOURL).To(Equal(""))
			Expect(c.RancherOS.Install.ContainerImage).To(Equal("docker/image:test"))
		})
		It("reads containerImage and registrationUrl", func() {

			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			_ = ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  install:
    registrationUrl: "foobar"
    containerImage: "docker/image:test"
`), os.ModePerm)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()
			c, err := ReadConfig(ctx, f.Name(), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RancherOS.Install.ContainerImage).To(Equal("docker/image:test"))
		})

		It("reads isoUrl instead of iso_url", func() {
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			_ = ioutil.WriteFile(f.Name(), []byte(`
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

			_ = ioutil.WriteFile(f.Name(), []byte(`
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

		It("reads iso_url by contacting a registrationUrl server", func() {

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			WSServer(ctx, data)
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			_ = ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  tpm:
    emulated: true
    no_smbios: true
    seed: "5"
  install:
    registrationUrl: "http://127.0.0.1:9980/test"
`), os.ModePerm)

			c, err := ReadConfig(ctx, f.Name(), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RancherOS.Install.ISOURL).To(Equal("foo"))

			_ = ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  tpm:
    emulated: true
    no_smbios: true
  install:
    registrationUrl: "http://127.0.0.1:9980/test"
`), os.ModePerm)

			c, err = ReadConfig(ctx, f.Name(), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RancherOS.Install.ISOURL).To(Equal("foo"))
		})
		It("reads containerImage by contacting a registrationUrl server", func() {

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Override the install value on the data
			value := map[string]string{"containerImage": "test"}
			values.PutValue(data, value, "rancheros", "install")

			WSServer(ctx, data)
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			_ = ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  tpm:
    emulated: true
    no_smbios: true
    seed: "5"
  install:
    registrationUrl: "http://127.0.0.1:9980/test"
`), os.ModePerm)

			c, err := ReadConfig(ctx, f.Name(), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RancherOS.Install.ContainerImage).To(Equal("test"))

			_ = ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  tpm:
    emulated: true
    no_smbios: true
  install:
    registrationUrl: "http://127.0.0.1:9980/test"
`), os.ModePerm)

			c, err = ReadConfig(ctx, f.Name(), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RancherOS.Install.ContainerImage).To(Equal("test"))
		})

		It("doesn't error out if isoUrl or containerImage are not provided", func() {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Override the install value on the data
			value := map[string]string{}
			values.PutValue(data, value, "rancheros", "install")

			WSServer(ctx, data)
			f, err := ioutil.TempFile("", "xxxxtest")
			Expect(err).ToNot(HaveOccurred())
			defer os.Remove(f.Name())

			_ = ioutil.WriteFile(f.Name(), []byte(`
rancheros:
  tpm:
    emulated: true
    no_smbios: true
    seed: "5"
  install:
    registrationUrl: "http://127.0.0.1:9980/test"
`), os.ModePerm)

			c, err := ReadConfig(ctx, f.Name(), false)
			Expect(err).ToNot(HaveOccurred())
			Expect(c.RancherOS.Install.ContainerImage).To(Equal(""))
			Expect(c.RancherOS.Install.ISOURL).To(Equal(""))
		})
	})
})
