# NotifyX Helm Chart

This Helm chart deploys NotifyX notification service to Kubernetes.

## Prerequisites

- Kubernetes cluster 1.19+
- Helm 3.0+
- Infrastructure services (MongoDB, Kafka, Redis) accessible from cluster
- Docker Hub images pushed (or use private registry)

## Installation

### 1. Add/Update Dependencies

```bash
helm dependency update helm/notifyx
```

### 2. Install Chart

```bash
# Install with default values
helm install notifyx ./helm/notifyx

# Install with custom values
helm install notifyx ./helm/notifyx -f my-values.yaml

# Install to specific namespace
helm install notifyx ./helm/notifyx -n notifyx --create-namespace
```

### 3. Verify Installation

```bash
# Check deployments
kubectl get deployments -l app.kubernetes.io/name=notifyx

# Check pods
kubectl get pods -l app.kubernetes.io/name=notifyx

# Check services
kubectl get services -l app.kubernetes.io/name=notifyx
```

## Configuration

### Image Registry

Update `values.yaml` to use your Docker Hub organization:

```yaml
imageRegistry: docker.io
api:
  image:
    repository: your-org/notifyx-api
    tag: "v1.0.0"
```

### Infrastructure Connection

Configure infrastructure endpoints:

```yaml
infrastructure:
  mongo:
    host: "mongo-service"
    port: 27017
  kafka:
    brokers:
      - "kafka-service:9092"
  redis:
    host: "redis-service"
    port: 6379
```

### Scaling

Adjust replica counts:

```yaml
api:
  replicaCount: 3

processor:
  replicaCount: 2

workers:
  email:
    replicaCount: 3
```

### Resources

Configure resource limits:

```yaml
api:
  resources:
    requests:
      cpu: 200m
      memory: 256Mi
    limits:
      cpu: 1000m
      memory: 1Gi
```

## Ingress

Enable ingress:

```yaml
ingress:
  enabled: true
  className: "nginx"
  hosts:
    - host: notifyx.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: notifyx-tls
      hosts:
        - notifyx.example.com
```

## Upgrading

```bash
helm upgrade notifyx ./helm/notifyx -f my-values.yaml
```

## Uninstallation

```bash
helm uninstall notifyx
```

## Chart Structure

```
helm/notifyx/
├── Chart.yaml              # Chart metadata
├── values.yaml             # Default values
└── templates/
    ├── deployment-api.yaml
    ├── deployment-processor.yaml
    ├── deployment-workers.yaml
    ├── service-api.yaml
    ├── configmap-api.yaml
    ├── configmap-processor.yaml
    ├── configmap-workers.yaml
    ├── ingress.yaml
    └── _helpers.tpl        # Template helpers
```

## Components

### API
- HTTP REST API for managing notifications
- Exposed via Service and optional Ingress
- Configurable replica count

### Processor
- Processes events from Kafka
- Routes to appropriate workers
- Uses Redis for caching

### Workers
- **email**: SMTP email delivery
- **sms**: AWS SNS SMS delivery
- **push**: Firebase/APNS push notifications
- **webhook**: HTTP webhook delivery

Each worker:
- Consumes from dedicated Kafka topic
- Writes results to MongoDB
- Configurable replica count

## Notes

- All images are pulled from Docker Hub (configurable)
- Infrastructure services must be accessible from cluster
- ConfigMaps are generated from values.yaml
- Health checks are configured for all deployments
- Network policies can be added for additional security

