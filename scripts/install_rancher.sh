#!/bin/bash

set -e

# Add stable Rancher Helm chart repo
helm repo add rancher-latest https://releases.rancher.com/server-charts/latest

# Add stable CertManager Helm chart repo
helm repo add jetstack https://charts.jetstack.io

# Mandatory! Otherwise Helm repos are not seen...
helm repo update

# Cert Manager has to be installed before Rancher
kubectl create namespace cattle-system
helm upgrade --install cert-manager jetstack/cert-manager \
  --namespace cert-manager \
  --create-namespace \
  --set installCRDs=true \
  --set extraArgs[0]=--enable-certificate-owner-ref=true \
  --wait

# Wait for cert-manager deployment to be ready
kubectl rollout status deployment cert-manager -n cert-manager --timeout=120s

# Install Rancher
helm upgrade --install rancher rancher-latest/rancher \
  --namespace cattle-system \
  --create-namespace \
  --set hostname=${MY_HOSTNAME} \
  --set bootstrapPassword=rancherpassword \
  --set "extraEnv[0].name=CATTLE_UI_DASHBOARD_INDEX" \
  --set "extraEnv[0].value=https://releases.rancher.com/dashboard/${DASHBOARD_VERSION}/index.html" \
  --set "extraEnv[1].name=CATTLE_UI_OFFLINE_PREFERRED" \
  --set "extraEnv[1].value=Remote" \
  --set "extraEnv[2].name=CATTLE_SERVER_URL" \
  --set "extraEnv[2].value=https://${MY_HOSTNAME}" \
  --wait

# Wait for rancher deployment to be ready
kubectl rollout status deployment rancher -n cattle-system --timeout=300s
