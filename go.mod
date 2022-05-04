module github.com/rancher-sandbox/os2

go 1.16

replace (
	k8s.io/api => k8s.io/api v0.22.2
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.2
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.2
	k8s.io/apiserver => k8s.io/apiserver v0.22.2
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.2
	k8s.io/client-go => k8s.io/client-go v0.22.2
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.2
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.2
	k8s.io/code-generator => k8s.io/code-generator v0.22.2
	k8s.io/component-base => k8s.io/component-base v0.22.2
	k8s.io/component-helpers => k8s.io/component-helpers v0.22.2
	k8s.io/controller-manager => k8s.io/controller-manager v0.22.2
	k8s.io/cri-api => k8s.io/cri-api v0.22.2
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.2
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.2
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.2
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.2
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.2
	k8s.io/kubectl => k8s.io/kubectl v0.22.2
	k8s.io/kubelet => k8s.io/kubelet v0.22.2
	k8s.io/kubernetes => k8s.io/kubernetes v1.22.2
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.2
	k8s.io/metrics => k8s.io/metrics v0.22.2
	k8s.io/mount-utils => k8s.io/mount-utils v0.22.2
	k8s.io/pod-security-admission => k8s.io/pod-security-admission v0.22.2
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.2
)

require (
	github.com/bramvdbogaerde/go-scp v1.2.0
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/gorilla/websocket v1.5.0
	github.com/mattn/go-isatty v0.0.12
	github.com/onsi/ginkgo/v2 v2.1.3
	github.com/onsi/gomega v1.18.1
	github.com/pin/tftp v2.1.0+incompatible // indirect
	github.com/pkg/errors v0.9.1
	github.com/rancher-sandbox/ele-testhelpers v0.0.0-20220503081716-e20a5bbd9065 // indirect
	github.com/rancher-sandbox/go-tpm v0.0.0-20220217133323-a7b15ad1f8f7
	github.com/rancher/wrangler v0.8.11-0.20220217210408-3ecd23dfea3b
	github.com/sirupsen/logrus v1.8.1
	github.com/tredoe/osutil v1.0.5
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519
	golang.org/x/term v0.0.0-20210220032956-6a3ed077a48d
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/pin/tftp.v2 v2.1.0
	gotest.tools v2.2.0+incompatible
	sigs.k8s.io/yaml v1.2.0
)
