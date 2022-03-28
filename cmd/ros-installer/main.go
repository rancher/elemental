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

package main

import (
	"context"
	"flag"
	"os"

	"github.com/rancher-sandbox/os2/pkg/config"
	"github.com/rancher-sandbox/os2/pkg/install"
	"github.com/sirupsen/logrus"
	"sigs.k8s.io/yaml"
)

var (
	automatic         = flag.Bool("automatic", false, "Check for and run automatic installation")
	printConfig       = flag.Bool("print-config", false, "Print effective configuration and exit")
	configFile        = flag.String("config-file", "/oem/userdata", "Config file to use, local file or http/tftp URL")
	powerOff          = flag.Bool("power-off", false, "Power off after installation")
	reboot            = flag.Bool("reboot", false, "Reboot after installation")
	noRebootAutomatic = flag.Bool("no-reboot-automatic", false, "Dont reboot after installation (only for automatic installation which defaults to reboot after install)")
	yes               = flag.Bool("y", false, "Do not prompt for questions")
	ejectCD           = flag.Bool("eject-cd", false, "Ejects the CD on system reboot")
)

func main() {
	flag.Parse()
	if *printConfig {
		cfg, err := config.ReadConfig(context.Background(), *configFile, *automatic)
		if err != nil {
			logrus.Fatal(err)
		}
		data, err := yaml.Marshal(cfg)
		if err != nil {
			logrus.Fatal(err)
		}
		os.Stdout.Write(data)
		return
	}

	if err := install.Run(*automatic, *configFile, *powerOff, *reboot, *noRebootAutomatic, *yes, *ejectCD); err != nil {
		logrus.Fatal(err)
	}
}
