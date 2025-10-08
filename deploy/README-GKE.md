# Envini backend deployment to Google Kubernetes Engine (GKE)

## Prerequisites
- gcloud CLI authenticated and a GCP project selected
- Artifact Registry enabled
- GKE cluster created
- kubectl configured to talk to the cluster

## Variables
Replace these in commands below:
- REGION: e.g. europe-central2 -> 
- PROJECT_ID: your GCP project id -> envini
- REPO: artifact registry repo name, e.g. envini

## Create Artifact Registry
```bash
gcloud artifacts repositories create $REPO \
  --repository-format=docker \
  --location=$REGION \
  --description="Envini images"
```

## Configure Docker auth
```bash
gcloud auth configure-docker $REGION-docker.pkg.dev
```

## Build and push images
From repo root:
```bash
# AuthService
docker build -f AuthService/Dockerfile -t $REGION-docker.pkg.dev/$PROJECT_ID/$REPO/authservice:latest .
docker push $REGION-docker.pkg.dev/$PROJECT_ID/$REPO/authservice:latest

# SecretOperationService
docker build -f SecretOperationService/Dockerfile -t $REGION-docker.pkg.dev/$PROJECT_ID/$REPO/secretoperationservice:latest .
docker push $REGION-docker.pkg.dev/$PROJECT_ID/$REPO/secretoperationservice:latest

# BackendGate
docker build -f BackendGate/Dockerfile -t $REGION-docker.pkg.dev/$PROJECT_ID/$REPO/backendgate:latest .
docker push $REGION-docker.pkg.dev/$PROJECT_ID/$REPO/backendgate:latest
```

## Prepare manifests
Replace REPLACE_REGISTRY in k8s manifests with your image registry path `$REGION-docker.pkg.dev/$PROJECT_ID/$REPO`.

Example:
```bash
sed -i '' "s#REPLACE_REGISTRY#europe-central2-docker.pkg.dev/envini/envini#g" deploy/k8s/*.yaml
```

## Deploy to GKE
```bash
kubectl apply -f deploy/k8s/namespace.yaml
kubectl apply -f deploy/k8s/postgres-auth.yaml
kubectl apply -f deploy/k8s/postgres-secrets.yaml
kubectl apply -f deploy/k8s/authservice.yaml
kubectl apply -f deploy/k8s/secretoperationservice.yaml
kubectl apply -f deploy/k8s/backendgate.yaml
# Optional ingress if you have GCE ingress controller
kubectl apply -f deploy/k8s/ingress.yaml
```

## Setup Google Secret Manager
Create secrets in Secret Manager:
```bash
# Create GitHub Client ID secret
echo -n "YOUR_GITHUB_CLIENT_ID" | gcloud secrets create github-client-id --data-file=-

# Create Master Encryption Key secret
openssl rand -base64 32 | gcloud secrets create master-encryption-key --data-file=-
```

Grant your GKE service account access to these secrets:
```bash
# Get your node service account
NODE_SA=$(gcloud compute project-info describe --format='value(defaultServiceAccount)')

# Grant Secret Manager access
gcloud secrets add-iam-policy-binding github-client-id \
  --member=serviceAccount:$NODE_SA \
  --role=roles/secretmanager.secretAccessor

gcloud secrets add-iam-policy-binding master-encryption-key \
  --member=serviceAccount:$NODE_SA \
  --role=roles/secretmanager.secretAccessor
```

## Sync secrets from Secret Manager
Run the sync script to populate Kubernetes secrets from Secret Manager:
```bash
./deploy/sync-secrets.sh
```

This script will:
- Fetch secrets from Google Secret Manager
- Update Kubernetes secrets
- Restart deployments to pick up new values

## Verify
```bash
kubectl -n envini get pods,svc
```
