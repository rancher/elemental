---
sidebar_label: Backup
title: ''
---

# Backup

Follow this guide to create backup for Elemental configuration installed together with Rancher.

## Install rancher-backup operator for Rancher

Go to official [Rancher documentation](https://docs.ranchermanager.rancher.io/how-to-guides/new-user-guides/backup-restore-and-disaster-recovery/back-up-rancher) and install rancher-bakup operator from there.

:::warning warning
For Rancher v2.7 and below it is needed to edit `ResourceSet` for rancher-backup operator.
For Rancher v2.7.1+ backup will be done automatically by rancher-backup operator and no further operation are needed.
:::

## Backup Elemental with rancher-backup operator (only for Rancher v2.7 and below)

Fetch `rancher-resource-set` object from Kubernetes cluster

```shell showLineNumbers
kubectl get ResourceSet rancher-resource-set -o yaml > rancher-resource-set.yaml
```

<Tabs>
<TabItem value="manualEdit" label="Manually editing the resource set yaml">

At the end of `rancher-resource-set.yaml` file add the definition of Elemental resources

```yaml showLineNumbers
- apiVersion: apiextensions.k8s.io/v1
kindsRegexp: .
resourceNameRegexp: elemental.cattle.io$
- apiVersion: apps/v1
kindsRegexp: ^deployments$
namespaces:
- cattle-elemental-system
resourceNames:
- elemental-operator
- apiVersion: rbac.authorization.k8s.io/v1
kindsRegexp: ^clusterroles$
resourceNames:
- elemental-operator
- apiVersion: rbac.authorization.k8s.io/v1
kindsRegexp: ^clusterrolebindings$
resourceNames:
- elemental-operator
- apiVersion: v1
kindsRegexp: ^serviceaccounts$
namespaces:
- cattle-elemental-system
resourceNames:
- elemental-operator
- apiVersion: management.cattle.io/v3
kindsRegexp: ^globalrole$
resourceNames:
- elemental-operator
- apiVersion: management.cattle.io/v3
kindsRegexp: ^apiservice$
resourceNameRegexp: elemental.cattle.io$
- apiVersion: elemental.cattle.io/v1beta1
kindsRegexp: .
namespaceRegexp: ^cattle-fleet-|^fleet-|^cluster-fleet-
- apiVersion: rbac.authorization.k8s.io/v1
kindsRegexp: ^roles$|^rolebindings$
labelSelectors:
  matchExpressions:
  - key: elemental.cattle.io/managed
    operator: In
    values:
    - "true"
namespaceRegexp: ^cattle-fleet-|^fleet-|^cluster-fleet-
- apiVersion: v1
kindsRegexp: ^secrets$|^serviceaccounts$
labelSelectors:
  matchExpressions:
  - key: elemental.cattle.io/managed
    operator: In
    values:
    - "true"
namespaceRegexp: ^cattle-fleet-|^fleet-|^cluster-fleet-
```

</TabItem>
<TabItem value="yqMerge" label="Using yq to auto merge yaml files">

You can use yq to auto merge `rancher-resource-set.yaml` and `elemental-resource-set.yaml`. Please go and install [yq v4.x](https://github.com/mikefarah/yq/#install) version

Create `elemental-resource-set.yaml` file

```yaml showLineNumbers
apiVersion: resources.cattle.io/v1
kind: ResourceSet
metadata:
  name: rancher-resource-set
resourceSelectors:
- apiVersion: apiextensions.k8s.io/v1
  kindsRegexp: .
  resourceNameRegexp: elemental.cattle.io$
- apiVersion: apps/v1
  kindsRegexp: ^deployments$
  namespaces:
  - cattle-elemental-system
  resourceNames:
  - elemental-operator
- apiVersion: rbac.authorization.k8s.io/v1
  kindsRegexp: ^clusterroles$
  resourceNames:
  - elemental-operator
- apiVersion: rbac.authorization.k8s.io/v1
  kindsRegexp: ^clusterrolebindings$
  resourceNames:
  - elemental-operator
- apiVersion: v1
  kindsRegexp: ^serviceaccounts$
  namespaces:
  - cattle-elemental-system
  resourceNames:
  - elemental-operator
- apiVersion: management.cattle.io/v3
  kindsRegexp: ^globalrole$
  resourceNames:
  - elemental-operator
- apiVersion: management.cattle.io/v3
  kindsRegexp: ^apiservice$
  resourceNameRegexp: elemental.cattle.io$
- apiVersion: elemental.cattle.io/v1beta1
  kindsRegexp: .
  namespaceRegexp: ^cattle-fleet-|^fleet-|^cluster-fleet-
- apiVersion: rbac.authorization.k8s.io/v1
  kindsRegexp: ^roles$|^rolebindings$
  labelSelectors:
    matchExpressions:
    - key: elemental.cattle.io/managed
      operator: In
      values:
      - "true"
  namespaceRegexp: ^cattle-fleet-|^fleet-|^cluster-fleet-
- apiVersion: v1
  kindsRegexp: ^secrets$|^serviceaccounts$
  labelSelectors:
    matchExpressions:
    - key: elemental.cattle.io/managed
      operator: In
      values:
      - "true"
  namespaceRegexp: ^cattle-fleet-|^fleet-|^cluster-fleet-
```

To merge both files, use `yq` command

```shell showLineNumbers
yq ea --inplace '. as $item ireduce ({}; . *+ $item )' rancher-resource-set.yaml elemental-resource-set.yaml
```

</TabItem>
</Tabs>

Then apply changes to Kubernetes cluster

```shell showLineNumbers
kubectl apply -f rancher-resource-set.yaml
```

Create backup with creating Backup object

```yaml showLineNumbers
apiVersion: resources.cattle.io/v1
kind: Backup
metadata:
  name: elemental-backup
spec:
  resourceSetName: rancher-resource-set
  schedule: "10 3 * * *"
  retentionCount: 10
```

Check logs from rancher-backup operator

```shell showLineNumbers
kubectl logs -n cattle-resources-system -l app.kubernetes.io/name=rancher-backup -f
```

Verify if backup file was created on Persistent Volume.

```shell showLineNumbers
...
INFO[2022/10/17 07:45:04] Finding files starting with /var/lib/backups/rancher-backup-430169aa-edde-4a61-85e8-858f625a755b*.tar.gz 
INFO[2022/10/17 07:45:04] File rancher-backup-430169aa-edde-4a61-85e8-858f625a755b-2022-10-17T05-15-00Z.tar.gz was created at 2022-10-17 0
...
```
