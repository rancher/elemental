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

package install

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/rancher-sandbox/os2/pkg/config"
	"github.com/rancher-sandbox/os2/pkg/questions"
	"sigs.k8s.io/yaml"
)

func Run(automatic bool, configFile string, powerOff bool, reboot bool, noRebootAutomatic bool, silent bool, ejectCD bool) error {
	cfg, err := config.ReadConfig(context.Background(), configFile, automatic)
	if err != nil {
		return err
	}

	if ejectCD {
		cfg.RancherOS.Install.EjectCD = true
	}
	if powerOff {
		cfg.RancherOS.Install.PowerOff = true
	}

	if reboot {
		cfg.RancherOS.Install.Reboot = true
	}

	if silent {
		cfg.RancherOS.Install.Automatic = true
	}

	// If we set the installation to automatic, reboot is set to true unless we set no reboot
	if automatic && !noRebootAutomatic {
		cfg.RancherOS.Install.Reboot = true
	}

	if automatic && !cfg.RancherOS.Install.Automatic {
		return nil
	}

	err = Ask(&cfg)
	if err != nil {
		return err
	}

	tempFile, err := ioutil.TempFile("", "ros-install")
	if err != nil {
		return err
	}
	if err := tempFile.Close(); err != nil {
		return err
	}

	return runInstall(cfg, tempFile.Name())
}

func runInstall(cfg config.Config, output string) error {
	installBytes, err := config.PrintInstall(cfg)
	if err != nil {
		return err
	}

	if !cfg.RancherOS.Install.Automatic {
		val, err := questions.PromptBool("\nConfiguration\n"+"-------------\n\n"+
			string(installBytes)+
			"\nYour disk will be formatted and installed with the above configuration.\nContinue?", false)
		if err != nil || !val {
			return err
		}
	}

	if cfg.RancherOS.Install.ConfigURL == "" && !cfg.RancherOS.Install.Automatic {
		yip := config.YipConfig{
			Rancherd: config.Rancherd{
				Server: cfg.RancherOS.Install.ServerURL,
				Token:  cfg.RancherOS.Install.Token,
				Role:   cfg.RancherOS.Install.Role,
			},
		}
		if cfg.RancherOS.Install.Password != "" || len(cfg.SSHAuthorizedKeys) > 0 {
			yip.Stages = map[string][]config.Stage{
				"network": {{
					Users: map[string]config.User{
						"root": {
							Name:              "root",
							PasswordHash:      cfg.RancherOS.Install.Password,
							SSHAuthorizedKeys: cfg.SSHAuthorizedKeys,
						},
					}},
				}}
			cfg.RancherOS.Install.Password = ""
		}

		data, err := yaml.Marshal(yip)
		if err != nil {
			return err
		}

		if err := ioutil.WriteFile(output+".yip", data, 0600); err != nil {
			return err
		}
		cfg.RancherOS.Install.ConfigURL = output + ".yip"
	} else {
		if err := config.ToFile(cfg, output); err != nil {
			return err
		}
		cfg.RancherOS.Install.ConfigURL = output
	}

	ev, err := config.ToEnv(cfg)
	if err != nil {
		return err
	}

	printEnv(cfg)

	installerOpts := []string{"elemental", "install", "--no-verify"}

	cmd := exec.Command("elemental")
	cmd.Env = append(os.Environ(), ev...)
	cmd.Stdout = os.Stdout
	cmd.Args = installerOpts
	cmd.Stdin = os.Stdin
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func printEnv(cfg config.Config) {
	if cfg.RancherOS.Install.Password != "" {
		cfg.RancherOS.Install.Password = "<removed>"
	}

	ev2, err := config.ToEnv(cfg)
	if err != nil {
		return
	}

	fmt.Println("Install environment:")
	for _, ev := range ev2 {
		fmt.Println(ev)
	}
}
