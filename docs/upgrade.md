# Upgrade

All components in Elemental are managed using Kubernetes. Below is how
to use Kubernetes approaches to upgrade the components.

## Elemental Teal node upgrade

Elemental Teal is upgraded with the {{elemental.operator.name}}. Refer to the
[{{elemental.operator.name}}]({{elemental.operator.url}}) documentation for complete information.

There are two ways of selecting nodes for upgrading. Via a cluster target, which will match ALL nodes in a cluster that matches our
selector or via node selector, which will match nodes based on the node labels. Node selector allows us to be more targeted with the upgrade
while cluster selector just selects all the nodes in a matched cluster.

=== "With `#!yaml clusterTarget`"
    You can target nodes for an upgrade via a `#!yaml clusterTarget` by setting it to the cluster name that you want to upgrade.
    All nodes in a cluster that matches that name will match and be upgraded.

    ```yaml title="upgrade-cluster-target.yaml"
    --8<-- "examples/upgrade/upgrade-cluster-target.yaml"
    ```

=== "With `#!yaml nodeSelector`"
    You can target nodes for an upgrade via a `#!yaml nodeSelector` by setting it to the label and value that you want to match.
    Any nodes containing that key with the value will match and be upgraded.

    ```yaml title="upgrade-node-selector.yaml"
    --8<-- "examples/upgrade/upgrade-node-selector.yaml"
    ```


### Selecting source for upgrade

=== "Via `#!yaml osImage`"
    
    Just specify an OCI image on the `#!yaml osImage` field

    ```yaml title="upgrade-cluster-target.yaml"
    --8<-- "examples/upgrade/upgrade-cluster-target.yaml"
    ```
    

=== "Via `#!yaml ManagedOSVersion`"
    
    In this case we use the auto populated `#!yaml ManagedOSVersion` resources to set the wanted `#!yaml managedOSVersionName` field.
    See section [Managing available versions](#managing-available-versions) to understand how the `#!yaml ManagedOSVersion` are managed.

    ```yaml title="upgrade-managedos-version.yaml"
    --8<-- "examples/upgrade/upgrade-managedos-version.yaml"
    ```

!!! warning
    If both `#!yaml osImage` and `#!yaml ManagedOSVersion` are defined in the same `#!yaml ManagedOSImage` be aware that `#!yaml osImage` takes precedence.

### Managing available versions

An `#!yaml ManagedOSVersionChannel` resource can be created in a Kubernetes cluster where the elemental operator is installed to synchronize available versions for upgrades.

It has a syncer in order to generate `#!yaml ManagedOSVersion` automatically. Currently, we provide a json syncer and a custom one.

=== "Json syncer"

    This syncer will fetch a json from url and parse it into valid `#!yaml ManagedOSVersion` resources.

    ```yaml title="managed-os-version-channel-json.yaml"
    --8<-- "examples/upgrade/managed-os-version-channel-json.yaml"
    ```

=== "Custom syncer"

    A custom syncer allows more flexibility on how to gather `#!yaml ManagedOSVersion` by allowing custom commands with custom images.
    
    This type of syncer allows to run a given command with arguments and env vars in a custom image and output a json file to `/data/output`
    `/data/output` is then automounted by the syncer and then parsed so it can gather create the proper versions.

    !!! info
        The only requirement to make your own custom syncer is to make it output a json file to `/data/output` and keep the correct json structure.
    
    See below for an example use of our [discovery plugin](https://github.com/rancher-sandbox/upgradechannel-discovery), 
    which gathers versions from either git or github releases.

    ```yaml title="managed-os-version-channel-json.yaml"
    --8<-- "examples/upgrade/managed-os-version-channel-custom.yaml"
    ```

In both cases the file that the operator expects to parse is a json file with the versions on it as follows

```json
[
    {
        "metadata": {
            "name": "v0.1.0"
        },
        "spec": {
            "version": "v0.1.0",
            "type": "container",
            "metadata": {
                "upgradeImage": "foo/bar:v0.1.0"
            }
        }
    },
    {
        "metadata": {
            "name": "v0.2.0"
        },
        "spec": {
            "version": "v0.2.0",
            "type": "container",
            "metadata": {
                "upgradeImage": "foo/bar:v0.2.0"
            }
        }
    }
]
```