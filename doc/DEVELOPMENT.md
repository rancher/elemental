# Elemental Stack Development

## Pre-requisites

Clone the following repositories in your development environment:

- <https://github.com/rancher/elemental>
  
  This repo is needed to build a development Elemental image that includes the `elemental-register` and it's ready to be used with the `elemental-operator`.  

- <https://github.com/rancher/elemental-toolkit>
  
  This repo is needed to build a development image containing the `elemental-toolkit` binary.  
  This image can be then referenced when building a development Elemental iso.  

- <https://github.com/rancher/elemental-operator>

  This repo is needed to build a development image containing the `elemental-registry` binary.  
  It also contains a convenient test cluster that can be automatically provisioned.  
  The test cluster includes a locally built version of the `elemental-operator`, a version of Rancher, and a test registry that can be used when testing Elemental [upgrades](https://elemental.docs.rancher.com/upgrade) through a `ManagedOSImage`.

## Setup a development environment

1. Provision the test cluster

    Using the `elemental-operator` repo:  

    ```bash
    make setup-full-cluster 
    ```

    If everything succeeded, you should be able to login into Rancher at: <https://172.18.0.2.sslip.io>
    The password is `rancherpassword`.  
    Note that it uses a self-signed certificate.  

    At this stage you may want to enable Extensions and install the Elemental extension from the Rancher web UI.  

1. Build a local image containing the `elemental-register` binary

    Using the `elemental-operator` repo:  

    ```bash
    make build-docker-register 
    ```

    This will build a `docker.io/local/elemental-register:dev` image that can be referenced in the following steps.  
    Note that before building this image you are free to checkout a different version of the `elemental-operator`, for example to test compatibility issues between mismatching `elemental-operator` and `elemental-register` versions.  

1. Build a local image containing the `elemental-toolkit`

    Using the `elemental-toolkit` repo:  

    ```bash
    make VERSION=dev build
    ```

    This will build a `docker.io/local/elemental-toolkit:dev` image that can be referenced in the following steps.  
    Note that before building this image you are free to checkout a different version of the `elemental-toolkit`, in order to generate *old* or *new* OS images that can be used to test downgrade/upgrade scenarios.  

1. Build and load the Elemental Dev ISO into the test cluster

    Using the `elemental` repo:

    ```bash
    make kind-load-dev-iso 
    ```

    By default this will use the previously built `docker.io/local/elemental-register:dev` and `docker.io/local/elemental-toolkit:dev` to generate a `docker.io/local/elemental-iso:dev` that can be used as a base Elemental image.  
    Since the image is also loaded into the test cluster, you can easily reference it in your `SeedImage` definition, for example:  

    ```yaml
    apiVersion: elemental.cattle.io/v1beta1
    kind: SeedImage
    metadata:
    name: fire-img
    namespace: fleet-default
    spec:
    type: iso
    baseImage: docker.io/local/elemental-iso:dev
    registrationRef:
        apiVersion: elemental.cattle.io/v1beta1
        kind: MachineRegistration
        name: fire-nodes
        namespace: fleet-default
    ```

1. Apply a test Elemental manifest and download the Dev ISO

    Using the `elemental` repo:

    ```bash
    kubectl apply -f tests/manifests/elemental-dev-example.yaml
    ```

    ```bash
    kubectl wait --for=condition=ready pod -n fleet-default fire-img
    ```

    ```bash
    wget --no-check-certificate `kubectl get seedimage -n fleet-default fire-img -o jsonpath="{.status.downloadURL}"` -O elemental-dev.x86_64.iso
    ```

    You can now use this ISO to provision Elemental machines, for example using an hypervisor on your dev environment.  
    The machines must be able to connect to the test Rancher environment `172.18.0.2:443`, and to the test registry when testing upgrade/downgrade scenarios `172.18.0.2:30000`.  

## Testing an Elemental upgrade scenario

Given that the development environment is ready and an Elemental machine was already provisioned, you can prepare a test OS version and use it for upgrades.  
The steps are equivalent for downgrades, by just checking out older versions of the components.  

1. Build local images containing the `elemental-register` and/or the `elemental-toolkit` on your next feature branch  

    Using the `elemental-operator` repo:  

    ```bash
    git checkout my-next-feature-branch
    make build-docker-register 
    ```

    Using the `elemental-toolkit` repo:  

    ```bash
    git checkout my-next-feature-branch
    make VERSION=dev GIT_COMMIT=test-upgrade build
    ```

1. Build a local OS image and push it to the test registry  

    Using the `elemental` repo:

    ```bash
    ELEMENTAL_OS_IMAGE="172.18.0.2:30000/elemental-os:dev-next" make build-dev-os
    ```

    In order to push this image to the test registry, you have add this in your Docker config: `/etc/docker/daemon.json`

    ```json
    { "insecure-registries":["172.18.0.2:30000"] } 
    ```

    Then restart docker:  

    ```bash
    sudo systemctl restart docker
    ```

    Finally push the OS image to the test registry:  

    ```bash
    docker push 172.18.0.2:30000/elemental-os:dev-next
    ```

1. Trigger the upgrade (on a Cluster level)

    Using the `elemental` repo:

    ```bash
    kubectl apply -f tests/manifests/elemental-dev-upgrade-example.yaml
    ```

1. Test the `elemental version` on the upgraded machine

    On the Elemental machine that has just been upgraded

    ```bash
    elemental version
    ```

    The version should include the `GIT_COMMIT` value that was set in the steps just above.  
    You can override the `GIT_COMMIT` variable when building the `elemental-toolkit` to test upgrades without actual code changes.  

1. Troubleshoot eventual issues

    In case of errors, refer to the [upgrade troubleshooting documentation](https://elemental.docs.rancher.com/troubleshooting-upgrade).
