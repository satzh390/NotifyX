# Managing Secrets in Kubernetes Pods

This guide explains how to load secrets from secret managers as files in Kubernetes pods for the NotifyX application.

## Overview

There are several approaches to managing secrets in Kubernetes:

1. **Kubernetes Secrets** (Native) - Recommended for most use cases
2. **External Secrets Operator** - For AWS Secrets Manager, HashiCorp Vault, etc.
3. **Sealed Secrets** - Encrypted secrets that can be committed to git
4. **Cloud Provider Secret Managers** - AWS Secrets Manager, Azure Key Vault, GCP Secret Manager

## Method 1: Kubernetes Secrets (Native)

### Step 1: Create a Kubernetes Secret

#### Option A: From a file (Recommended)
```bash
# Create secret from APNS key file
kubectl create secret generic notifyx-worker-push-apns-secret \
  --from-file=apns-key.p8=/path/to/AuthKey_XXXXXXXXXX.p8 \
  --namespace=notifyx

# Verify secret
kubectl get secret notifyx-worker-push-apns-secret -n notifyx
kubectl describe secret notifyx-worker-push-apns-secret -n notifyx
```

#### Option B: From literal values
```bash
# Create secret with base64 encoded key
kubectl create secret generic notifyx-worker-push-apns-secret \
  --from-literal=apns-key.p8="$(base64 -i /path/to/AuthKey_XXXXXXXXXX.p8)" \
  --namespace=notifyx
```

#### Option C: Using YAML manifest
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: notifyx-worker-push-apns-secret
  namespace: notifyx
type: Opaque
data:
  # Base64 encoded content
  apns-key.p8: <base64-encoded-key-content>
stringData:
  # Plain text (Kubernetes will encode it)
  apns-key.p8: |
    -----BEGIN PRIVATE KEY-----
    ...
    -----END PRIVATE KEY-----
```

### Step 2: Mount Secret as Volume in Deployment

The Helm chart already includes secret mounting. Enable it in `values.yaml`:

```yaml
workers:
  push:
    secrets:
      enabled: true
      apnsKey: ""  # Leave empty if creating secret manually
```

Or update the deployment manually:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: notifyx-worker-push
spec:
  template:
    spec:
      containers:
      - name: worker-push
        volumeMounts:
        - name: apns-secrets
          mountPath: /etc/secrets/apns
          readOnly: true
      volumes:
      - name: apns-secrets
        secret:
          secretName: notifyx-worker-push-apns-secret
          defaultMode: 0400  # Read-only, owner only
```

### Step 3: Configure Application via AppConfig API

**Note:** Push providers are now configured via AppConfig API (multi-app support). Each app can have its own provider configuration.

Create an AppConfig using the API:

```bash
curl -X POST http://localhost:8080/api/v1/app-configs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "id": "my-app-ios",
    "name": "My iOS App",
    "provider": "apns",
    "apns": {
      "keyId": "YOUR_KEY_ID",
      "teamId": "YOUR_TEAM_ID",
      "bundleId": "com.yourcompany.app",
      "keyPath": "/etc/secrets/apns/apns-key.p8",
      "production": false
    }
  }'
```

The `keyPath` should point to the mounted secret path. For multiple apps with different keys, mount multiple secrets:

```yaml
volumeMounts:
- name: apns-secrets-app1
  mountPath: /etc/secrets/apns/app1
- name: apns-secrets-app2
  mountPath: /etc/secrets/apns/app2
```

Then reference them in AppConfig:
- App 1: `keyPath: "/etc/secrets/apns/app1/apns-key.p8"`
- App 2: `keyPath: "/etc/secrets/apns/app2/apns-key.p8"`

## Method 2: External Secrets Operator (AWS Secrets Manager)

### Prerequisites

Install External Secrets Operator:
```bash
helm repo add external-secrets https://charts.external-secrets.io
helm install external-secrets external-secrets/external-secrets -n external-secrets-system --create-namespace
```

### Step 1: Store Secret in AWS Secrets Manager

```bash
# Store APNS key in AWS Secrets Manager
aws secretsmanager create-secret \
  --name notifyx/apns/key \
  --secret-string file://AuthKey_XXXXXXXXXX.p8 \
  --region us-east-1

# Store other APNS config
aws secretsmanager create-secret \
  --name notifyx/apns/config \
  --secret-string '{"keyId":"YOUR_KEY_ID","teamId":"YOUR_TEAM_ID","bundleId":"com.yourcompany.app"}' \
  --region us-east-1
```

### Step 2: Create IAM Role/ServiceAccount

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: notifyx-worker-push
  namespace: notifyx
  annotations:
    eks.amazonaws.com/role-arn: arn:aws:iam::ACCOUNT_ID:role/notifyx-secrets-role
```

### Step 3: Create ExternalSecret Resource

```yaml
apiVersion: external-secrets.io/v1beta1
kind: ExternalSecret
metadata:
  name: notifyx-apns-secret
  namespace: notifyx
spec:
  refreshInterval: 1h
  secretStoreRef:
    name: aws-secrets-manager
    kind: SecretStore
  target:
    name: notifyx-worker-push-apns-secret
    creationPolicy: Owner
  data:
  - secretKey: apns-key.p8
    remoteRef:
      key: notifyx/apns/key
```

### Step 4: Create SecretStore

```yaml
apiVersion: external-secrets.io/v1beta1
kind: SecretStore
metadata:
  name: aws-secrets-manager
  namespace: notifyx
spec:
  provider:
    aws:
      service: SecretsManager
      region: us-east-1
      auth:
        jwt:
          serviceAccountRef:
            name: notifyx-worker-push
```

## Method 3: Using Environment Variables (Alternative)

Instead of mounting as files, you can inject secrets as environment variables:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: notifyx-worker-push
spec:
  template:
    spec:
      containers:
      - name: worker-push
        env:
        - name: APNS_KEY_ID
          valueFrom:
            secretKeyRef:
              name: notifyx-worker-push-apns-secret
              key: keyId
        - name: APNS_TEAM_ID
          valueFrom:
            secretKeyRef:
              name: notifyx-worker-push-apns-secret
              key: teamId
        - name: APNS_BUNDLE_ID
          valueFrom:
            secretKeyRef:
              name: notifyx-worker-push-apns-secret
              key: bundleId
        - name: APNS_KEY_PATH
          value: "/etc/secrets/apns/apns-key.p8"
```

Then modify the application to read the key content from environment variable and write it to a temp file, or use a base64-encoded key in an env var.

## Security Best Practices

1. **File Permissions**: Always set `defaultMode: 0400` (read-only, owner only)
2. **Read-Only Mounts**: Use `readOnly: true` in volumeMounts
3. **RBAC**: Limit access to secrets using Kubernetes RBAC
4. **Encryption**: Enable encryption at rest for etcd (where Kubernetes stores secrets)
5. **Rotation**: Implement secret rotation policies
6. **Audit**: Enable audit logging for secret access

## Example: Complete Setup for APNS

### 1. Create Secret
```bash
kubectl create secret generic notifyx-worker-push-apns-secret \
  --from-file=apns-key.p8=./AuthKey_XXXXXXXXXX.p8 \
  --from-literal=keyId=YOUR_KEY_ID \
  --from-literal=teamId=YOUR_TEAM_ID \
  --from-literal=bundleId=com.yourcompany.app \
  --namespace=notifyx
```

### 2. Deploy with Helm
```bash
helm upgrade --install notifyx ./helm/notifyx \
  --set workers.push.secrets.enabled=true \
  --namespace notifyx
```

### 3. Create AppConfig via API
```bash
# After deployment, create AppConfig for your app
curl -X POST http://notifyx-api/api/v1/app-configs \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "id": "my-app",
    "name": "My App",
    "provider": "apns",
    "apns": {
      "keyId": "YOUR_KEY_ID",
      "teamId": "YOUR_TEAM_ID",
      "bundleId": "com.yourcompany.app",
      "keyPath": "/etc/secrets/apns/apns-key.p8",
      "production": false
    }
  }'
```

**Note:** The push worker no longer uses static config. All provider configurations are managed via AppConfig API endpoints (`/api/v1/app-configs`).

## Troubleshooting

### Check if secret is mounted:
```bash
kubectl exec -it notifyx-worker-push-xxxxx -- ls -la /etc/secrets/apns
kubectl exec -it notifyx-worker-push-xxxxx -- cat /etc/secrets/apns/apns-key.p8
```

### Verify secret exists:
```bash
kubectl get secret notifyx-worker-push-apns-secret -n notifyx
kubectl describe secret notifyx-worker-push-apns-secret -n notifyx
```

### Check pod logs:
```bash
kubectl logs notifyx-worker-push-xxxxx -n notifyx
```

## References

- [Kubernetes Secrets Documentation](https://kubernetes.io/docs/concepts/configuration/secret/)
- [External Secrets Operator](https://external-secrets.io/)
- [AWS Secrets Manager](https://aws.amazon.com/secrets-manager/)
- [HashiCorp Vault](https://www.vaultproject.io/)

