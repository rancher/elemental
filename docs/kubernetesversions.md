---
sidebar_label: Kubernetes versions
title: ''
---

<head>
  <meta charset="utf-8" />
  <title>Redirecting to https://elemental.docs.rancher.com</title>
  <meta http-equiv="refresh" content="0; URL=https://elemental.docs.rancher.com/" />
  <link rel="canonical" href="https://elemental.docs.rancher.com/" />
</head>

## Valid Versions

The list of valid versions for the `kubernetesVersion` field can be determined
from the Rancher metadata using the following commands.

__k3s:__
```bash showLineNumbers
curl -sL https://raw.githubusercontent.com/rancher/kontainer-driver-metadata/release-v2.6/data/data.json | jq -r '.k3s.releases[].version'
```
__rke2:__
```bash showLineNumbers
curl -sL https://raw.githubusercontent.com/rancher/kontainer-driver-metadata/release-v2.6/data/data.json | jq -r '.rke2.releases[].version'
```
