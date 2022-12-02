---
sidebar_label: Upgrade
title: ''
---

<head>
  <meta charset="utf-8" />
  <title>Redirecting to https://elemental.docs.rancher.com</title>
  <meta http-equiv="refresh" content="0; URL=https://elemental.docs.rancher.com/" />
  <link rel="canonical" href="https://elemental.docs.rancher.com/" />
</head>

import ClusterTarget from "!!raw-loader!../examples/upgrade/upgrade-cluster-target.yaml"
import NodeSelector from "!!raw-loader!../examples/upgrade/upgrade-node-selector.yaml"
import ManagedOSVersion from "!!raw-loader!../examples/upgrade/upgrade-managedos-version.yaml"
import MangedOSVersionChannelJson from "!!raw-loader!../examples/upgrade/managed-os-version-channel-json.yaml"
import ManagedOSVersionChannelCustom from "!!raw-loader!../examples/upgrade/managed-os-version-channel-custom.yaml"
import Versions from "../examples/upgrade/versions.raw!=!raw-loader!../examples/upgrade/versions.json"

# Upgrade

All components in Elemental are managed using Kubernetes. Below is how
to use Kubernetes approaches to upgrade the components.

## Elemental Teal node upgrade

Elemental Teal is upgraded with the <Vars name="elemental_operator_name" />. Refer to the
[<Vars name="elemental_operator_name" />](elementaloperatorchart-reference.md) documentation for complete information.

There are two ways of selecting nodes for upgrading. Via a cluster target, which will match ALL nodes in a cluster that matches our
selector or via node selector, which will match nodes based on the node labels. Node selector allows us to be more targeted with the upgrade
while cluster selector just selects all the nodes in a matched cluster.

<Tabs>
<TabItem value="clusterTarget" label="With 'clusterTarget'" default>
You can target nodes for an upgrade via a `clusterTarget` by setting it to the cluster name that you want to upgrade.
All nodes in a cluster that matches that name will match and be upgraded.

<CodeBlock language="yaml" title="upgrade-cluster-target.yaml" showLineNumbers>{ClusterTarget}</CodeBlock>

</TabItem>
<TabItem value="nodeSelector" label="With nodeSelector">
You can target nodes for an upgrade via a `nodeSelector` by setting it to the label and value that you want to match.
Any nodes containing that key with the value will match and be upgraded.

<CodeBlock language="yaml" title="upgrade-node-selector.yaml" showLineNumbers>{NodeSelector}</CodeBlock>

</TabItem>
</Tabs>


### Selecting source for upgrade

<Tabs>
<TabItem value="osImage" label="Via 'osImage'">

Just specify an OCI image on the `osImage` field

<CodeBlock language="yaml" title="upgrade-cluster-target.yaml" showLineNumbers>{ClusterTarget}</CodeBlock>

</TabItem>
<TabItem value="managedOSVersion" label="Via 'ManagedOSVersion'">

In this case we use the auto populated `ManagedOSVersion` resources to set the wanted `managedOSVersionName` field.
See section [Managing available versions](#managing-available-versions) to understand how the `ManagedOSVersion` are managed.

<CodeBlock language="yaml" title="upgrade-managedos-version.yaml" showLineNumbers>{ManagedOSVersion}</CodeBlock>

</TabItem>
</Tabs>

:::warning Warning
If both `osImage` and `ManagedOSVersion` are defined in the same `ManagedOSImage` be aware that `osImage` takes precedence.
:::

### Managing available versions

An `ManagedOSVersionChannel` resource can be created in a Kubernetes cluster where the Elemental operator is installed to synchronize available versions for upgrades.

It has a syncer in order to generate `ManagedOSVersion` automatically. Currently, we provide a json syncer and a custom one.

<Tabs>
<TabItem value="jsonSyncer" label="Json syncer">

This syncer will fetch a json from url and parse it into valid `ManagedOSVersion` resources.

<CodeBlock language="yaml" title="managed-os-version-json" showLineNumbers>{MangedOSVersionChannelJson}</CodeBlock>

</TabItem>
<TabItem value="customSyncer" label="Custom syncer">

A custom syncer allows more flexibility on how to gather `ManagedOSVersion` by allowing custom commands with custom images.

This type of syncer allows to run a given command with arguments and env vars in a custom image and output a json file to `/data/output`.
The generated data is then automounted by the syncer and then parsed so it can gather create the proper versions.

:::info
The only requirement to make your own custom syncer is to make it output a json file to `/data/output` and keep the correct json structure.
:::

See below for an example use of our [discovery plugin](https://github.com/rancher-sandbox/upgradechannel-discovery), 
which gathers versions from either git or github releases.

<CodeBlock language="yaml" title="managed-os-version-channel-json.yaml" showLineNumbers>{ManagedOSVersionChannelCustom}</CodeBlock>

</TabItem>
</Tabs>

In both cases the file that the operator expects to parse is a json file with the versions on it as follows

<CodeBlock language="json" title="versions.json" showLineNumbers>{Versions}</CodeBlock>
