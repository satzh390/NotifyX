# Quick Start: Mounting Secrets as Files in Kubernetes Pods

## Method 1: Kubernetes Secrets (Simplest)

### Step 1: Create Secret from File
```bash
# Create secret from APNS key file
kubectl create secret generic notifyx-worker-push-apns-secret \
  --from-file=apns-key.p8=/path/to/AuthKey_XXXXXXXXXX.p8 \
  --namespace=notifyx
```

### Step 2: Enable Secret Mounting in Helm
```yaml
# values.yaml
workers:
  push:
    secrets:
      enabled: true
```

### Step 3: Deploy
```bash
helm upgrade --install notifyx ./helm/notifyx \
  --set workers.push.secrets.enabled=true \
  --namespace notifyx
```

### Step 4: Verify Secret is Mounted
```bash
# Check if file exists in pod
kubectl exec -it notifyx-worker-push-xxxxx -n notifyx -- ls -la /etc/secrets/apns

# View file (should show your key)
kubectl exec -it notifyx-worker-push-xxxxx -n notifyx -- cat /etc/secrets/apns/apns-key.p8
```

## Method 2: Using Helm Template (If you want to include in Helm)

The Helm chart includes a template at `helm/notifyx/templates/secret-apns.yaml` that creates the secret automatically.

**Note**: For production, it's recommended to create secrets manually using `kubectl` rather than storing them in Helm values.

## Configuration via AppConfig API

**Note:** Push providers are now configured via AppConfig API (multi-app support). The push worker no longer uses static config files.

After mounting the secret, create an AppConfig via the API:

```bash
curl -X POST "http://localhost:8080/api/v1/app-configs" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "id": "my-ios-app",
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

The `keyPath` should point to the mounted secret path. For multiple apps with different keys, mount multiple secrets and reference them in separate AppConfig entries.

## Security Notes

1. **File Permissions**: The secret is mounted with `0400` (read-only, owner only)
2. **Read-Only**: Volume is mounted as `readOnly: true`
3. **Never commit**: Never commit `.p8` files to git (already in `.gitignore`)

## Troubleshooting

**Secret not found:**
```bash
kubectl get secret notifyx-worker-push-apns-secret -n notifyx
```

**Pod can't read file:**
```bash
kubectl exec -it notifyx-worker-push-xxxxx -n notifyx -- ls -la /etc/secrets/apns
kubectl describe pod notifyx-worker-push-xxxxx -n notifyx
```

**Application error:**
```bash
kubectl logs notifyx-worker-push-xxxxx -n notifyx
```

