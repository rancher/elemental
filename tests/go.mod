module github.com/rancher/elemental/tests

go 1.24.0

replace go.qase.io/client => github.com/rancher/qase-go/client v0.0.0-20231114201952-65195ec001fa

require (
	github.com/onsi/ginkgo/v2 v2.23.4
	github.com/onsi/gomega v1.37.0
	github.com/rancher-sandbox/ele-testhelpers v0.0.0-20250415062725-efdf8e57c793
	github.com/rancher-sandbox/qase-ginkgo v1.0.1
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/mod v0.24.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/antihax/optional v1.0.0 // indirect
	github.com/bramvdbogaerde/go-scp v1.5.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250403155104-27863c87afa6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	go.qase.io/client v0.0.0-20231114201952-65195ec001fa // indirect
	go.uber.org/automaxprocs v1.6.0 // indirect
	go.uber.org/multierr v1.11.0 // indirect
	go.uber.org/zap v1.27.0 // indirect
	golang.org/x/crypto v0.37.0 // indirect
	golang.org/x/net v0.39.0 // indirect
	golang.org/x/oauth2 v0.29.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	golang.org/x/tools v0.32.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	libvirt.org/libvirt-go-xml v7.4.0+incompatible // indirect
)
