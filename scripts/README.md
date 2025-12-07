# NotifyX Startup Scripts

This directory contains startup scripts for running NotifyX in different modes.

## Prerequisites

- Docker and Docker Compose installed and running
- Go 1.23+ (for direct mode)
- Infrastructure services running (MongoDB, Kafka, Redis, etc.)

## Usage

### Direct Mode

Start all services as local processes, connecting to Docker network for infrastructure:

**Linux/Mac:**
```bash
./scripts/start.sh direct
```

**Windows:**
```powershell
.\scripts\start.ps1 direct
```

This will:
- Start API on port 8080
- Start Processor
- Start all Workers (email, sms, push, webhook)
- All services connect to infrastructure running in Docker

**Stop services:**
```bash
./scripts/stop.sh
```

### Docker Mode

Build Docker images and start all services in containers:

**Linux/Mac:**
```bash
./scripts/start.sh docker
```

**Windows:**
```powershell
.\scripts\start.ps1 docker
```

This will:
- Build all Docker images locally
- Start all services in Docker containers
- Create proper network isolation
- Connect to infrastructure network

**Stop services:**
```bash
docker compose -f docker-compose.notifyx.yaml down
```

## Infrastructure Setup

Before starting services, ensure infrastructure is running:

```bash
docker compose -f docker-compose.local.yaml up -d
```

This starts:
- MongoDB (port 27017)
- Kafka (port 9092)
- Redis (port 6379)
- LocalStack (port 4566)
- OAuth Mock Server (port 8081)
- MailHog (ports 1025, 8025)
- Webhook Receiver (port 8888)

## Network Architecture

### Direct Mode
- Services run as local processes
- Connect to `notifyx-infra-net` Docker network for infrastructure

### Docker Mode
- Services run in `notifyx-net` network
- Connect to `notifyx-infra-net` network for infrastructure
- Proper network isolation between notification services and infrastructure

## Troubleshooting

1. **Port conflicts**: Ensure ports 8080, 9092, 27017, 6379 are available
2. **Docker not running**: Start Docker Desktop or Docker daemon
3. **Infrastructure not ready**: Wait a few seconds after starting infrastructure
4. **Build failures**: Ensure all Dockerfiles are in correct locations

