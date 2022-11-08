---
sidebar_label: Restore
title: ''
---

# Troubleshooting restore

:::warning warning
When a restore is performed, do not restart the `rancher-system-agent` on elemental nodes as it can stale and end with the following error:

`panic: error while connecting to Kubernetes cluster: the server has asked for the client to provide credentials`

If you face this problem, please follow the procedure below.
:::

Before you initiate a restore, you need to copy `/var/lib/rancher/agent/rancher2_connection_info.json` from the elemental node to a place where you have access with Rancher UI.

Once the file is copied, download the `rancher-agent-token-update.sh` script from the [Elemental repository](https://github.com/rancher/elemental):

```shell showLineNumbers
wget -q https://raw.githubusercontent.com/rancher/elemental/main/scripts/rancher-agent-token-update && chmod +x rancher-agent-token-update
```

Execute the script without any additional options:

```shell showLineNumbers
./rancher-agent-token-update
```

After the restore successfully completed, copy `rancher2_connection_info.json` back to the elemental node to the path
`/var/lib/rancher/agent/rancher2_connection_info.json`. Finally, restart the `rancher-system-agent` service:

```shell showLineNumbers
systemctl restart rancher-system-agent
```
