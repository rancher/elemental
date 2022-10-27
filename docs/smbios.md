---
sidebar_label: Smbios
title: ''
---

import Registration from "!!raw-loader!../examples/quickstart/registration.yaml"

## Introduction

The System Management BIOS (SMBIOS) specification defines data structures (and access methods) that can be used to read management information produced by the BIOS of a computer.

This allows us to gather hardware information about the running system and use that as part of our labels.

## How does elemental uses SMBIOS data?

The registration client tries to gather SMBIOS data by running `dmidecode` during the initial registration of the node and that data is
sent to the registration controller to use on interpolating different fields in the inventory that we create for that node.

Currently, we support interpolating that data into the `machineName` and the `machineInventoryLabels` of a [machineRegistration](machineregistration-reference.md)

The interpolation format is as follows:

`{$KEY/VALUE}` and `${KEY/SUBKEY/VALUE}`

This can be mixed with normal strings so `my-prefix-${KEY/VALUE}` would result into the resolved values with `my-prefix-` prefixed


For example, having the following SMBIOS data:

```console showLineNumbers
System Information
	Manufacturer: My manufacturer
	Product Name: Awesome PC
	Version: Not Specified
	Serial Number: THX1138
	Family: Toretto
```

And setting the `machineName` to `serial-${System Information/Serial Number}` would result in the final value of `serial-THX1138`

This is useful to generate automatic names for machines based on their hardware values, for example using the UUID or the Product name.
Our default `machineName` when the registration values are empty is `"m-${System Information/UUID}"`.

:::warning warning
All non-valid characters will be changed into `-` automatically on parse. Valid characters for labels are alphanumeric and `-`,`_` and `.`
For machineName the constraints are stricter as that value is used for the hostname so valid values are lowercase alphanumeric and `-` only.
:::

A good use of SMBIOS data is to set up different labels for all your machines and get those values from the hardware directly.

Having your `machineInventoryLabels` on the [machineRegistration](machineregistration-reference.md) set to SMBIOS data would allow 
you to use selectors down the line to select similar machines.

For example using the following label `cpuFamily: "${Processor Information/Family}` would allow you to use a selector to search for i7 cpus in your machine fleet.

<CodeBlock language="yaml" title="registration example with smbios labels" showLineNumbers>{Registration}</CodeBlock>
