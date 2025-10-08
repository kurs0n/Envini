#!/bin/bash

# Script to sync secrets from Google Secret Manager to Kubernetes
# Usage: ./sync-secrets.sh

PROJECT_ID="envini"
NAMESPACE="envini"

echo "Syncing secrets from Secret Manager to Kubernetes..."

# Get GitHub Client ID from Secret Manager
echo "Getting GitHub Client ID..."
GITHUB_CLIENT_ID=$(gcloud secrets versions access latest --secret="github-client-id" --project=$PROJECT_ID)

# Get Master Encryption Key from Secret Manager  
echo "Getting Master Encryption Key..."
MASTER_KEY=$(gcloud secrets versions access latest --secret="master-encryption-key" --project=$PROJECT_ID)

# Update GitHub OAuth secret
echo "Updating GitHub OAuth secret..."
kubectl create secret generic github-oauth \
  --from-literal=client_id="$GITHUB_CLIENT_ID" \
  --namespace=$NAMESPACE \
  --dry-run=client -o yaml | kubectl apply -f -

# Update Master Encryption secret
echo "Updating Master Encryption secret..."
kubectl create secret generic master-encryption \
  --from-literal=key="$MASTER_KEY" \
  --namespace=$NAMESPACE \
  --dry-run=client -o yaml | kubectl apply -f -

echo "Secrets synced successfully!"
echo "Restarting deployments to pick up new secrets..."

kubectl rollout restart deployment/authservice -n $NAMESPACE
kubectl rollout restart deployment/secretoperationservice -n $NAMESPACE

echo "Done!"
