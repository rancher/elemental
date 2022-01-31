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
	automatic   = flag.Bool("automatic", false, "Check for and run automatic installation")
	printConfig = flag.Bool("print-config", false, "Print effective configuration and exit")
	configFile  = flag.String("config-file", "", "Config file to use, local file or http/tftp URL")
	powerOff    = flag.Bool("power-off", false, "Power off after installation")
	yes         = flag.Bool("y", false, "Do not prompt for questions")
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

	if err := install.Run(*automatic, *configFile, *powerOff, *yes); err != nil {
		logrus.Fatal(err)
	}
}
