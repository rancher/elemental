---
sidebar_label: Restore
title: ''
---

# Restore

Follow this guide to restore an elemental node configuration from a backup with Rancher.

## Prepare rancher-backup operator and backup files for restoring

Go to official [Rancher documentation](https://docs.ranchermanager.rancher.io/how-to-guides/new-user-guides/backup-restore-and-disaster-recovery/restore-rancher) and make sure that `rancher-bakup operator` is installed and has access to backup files.

## Restore the elemental node configuration with rancher-backup operator

Create a `restore object` to restore the backup tarball:

```yaml showLineNumbers
apiVersion: resources.cattle.io/v1
kind: Restore
metadata:
  name: restore-migration
spec:
  backupFilename: rancher-backup-430169aa-edde-4a61-85e8-858f625a755b-2022-10-17T05-15-00Z.tar.gz
```

Apply manifest on Kubernetes

```shell showLineNumbers
kubectl apply -f rancher-restore.yaml
```

Check logs from rancher-backup operator

```shell showLineNumbers
kubectl logs -n cattle-resources-system -l app.kubernetes.io/name=rancher-backup -f
```

Verify if backup file was restore successfully.

```shell showLineNumbers
...
INFO[2022/10/31 06:34:50] Processing controllerRef apps/v1/deployments/rancher 
INFO[2022/10/31 06:34:50] Done restoring
...
```

Continue with procedure from [Rancher documentation](https://docs.ranchermanager.rancher.io/how-to-guides/new-user-guides/backup-restore-and-disaster-recovery/migrate-rancher-to-new-cluster)

:::warning warning
Try to not restart `rancher-system-agent` on elemental nodes, otherwise it can stuck in following error

`panic: error while connecting to Kubernetes cluster: the server has asked for the client to provide credentials`

If this problem occurs please go with procedure below.
:::

Copy `/var/lib/rancher/agent/rancher2_connection_info.json` from elemental node to place where you have access to Kubernetes with Rancher UI

Download `rancher-agent-token-update.sh` script which from [Elemental repository](https://github.com/rancher/elemental)

```shell showLineNumbers
wget -q https://raw.githubusercontent.com/rancher/elemental/main/scripts/rancher-agent-token-update && chmod +x rancher-agent-token-update
```

Execute script without any additional commands

```shell showLineNumbers
rancher-agent-token-update
```

After it will finish successfully, copy back `rancher2_connection_info.json` to the elemental node 
`/var/lib/rancher/agent/rancher2_connection_info.json` and restart there `rancher-system-agent` service

```shell showLineNumbers
systemctl restart rancher-system-agent
```
