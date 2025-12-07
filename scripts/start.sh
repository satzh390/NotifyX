#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if docker is running
check_docker() {
    if ! docker info > /dev/null 2>&1; then
        print_error "Docker is not running. Please start Docker and try again."
        exit 1
    fi
}

# Function to check if infrastructure is running
check_infrastructure() {
    print_info "Checking infrastructure services..."
    
    if ! docker ps | grep -q "mongo"; then
        print_warn "MongoDB is not running. Starting infrastructure..."
        cd "$PROJECT_ROOT"
        docker compose -f docker-compose.local.yaml up -d
        print_info "Waiting for services to be ready..."
        sleep 5
    fi
    
    print_info "Infrastructure services are running"
}

# Function to start services in direct mode
start_direct() {
    print_info "Starting services in DIRECT mode..."
    
    check_docker
    check_infrastructure
    
    cd "$PROJECT_ROOT"
    
    # Start API
    print_info "Starting API..."
    cd "$PROJECT_ROOT/app/api"
    NOTIFYX_API_CONFIG="config/config.yaml" go run ./cmd/main.go &
    API_PID=$!
    echo $API_PID > /tmp/notifyx-api.pid
    print_info "API started with PID: $API_PID"
    sleep 2
    
    # Start Processor
    print_info "Starting Processor..."
    cd "$PROJECT_ROOT/app/processor"
    NOTIFYX_PROCESSOR_CONFIG="config/config.yaml" go run ./cmd/main.go &
    PROCESSOR_PID=$!
    echo $PROCESSOR_PID > /tmp/notifyx-processor.pid
    print_info "Processor started with PID: $PROCESSOR_PID"
    sleep 2
    
    # Start Workers
    print_info "Starting Workers..."
    
    cd "$PROJECT_ROOT/app/workers/worker-email"
    NOTIFYX_WORKER_EMAIL_CONFIG="config/config.yaml" go run ./cmd/main.go &
    echo $! > /tmp/notifyx-worker-email.pid
    
    cd "$PROJECT_ROOT/app/workers/worker-sms"
    NOTIFYX_WORKER_SMS_CONFIG="config/config.yaml" go run ./cmd/main.go &
    echo $! > /tmp/notifyx-worker-sms.pid
    
    cd "$PROJECT_ROOT/app/workers/worker-push"
    NOTIFYX_WORKER_PUSH_CONFIG="config/config.yaml" go run ./cmd/main.go &
    echo $! > /tmp/notifyx-worker-push.pid
    
    cd "$PROJECT_ROOT/app/workers/worker-webhook"
    NOTIFYX_WORKER_WEBHOOK_CONFIG="config/config.yaml" go run ./cmd/main.go &
    echo $! > /tmp/notifyx-worker-webhook.pid
    
    print_info "All services started in direct mode"
    print_info "API: http://localhost:8080"
    print_info "Use 'scripts/stop.sh' to stop all services"
}

# Function to build docker images
build_images() {
    print_info "Building Docker images..."
    
    cd "$PROJECT_ROOT"
    
    # Build API
    print_info "Building API image..."
    docker build -f app/api/Dockerfile -t notifyx-api:latest .
    
    # Build Processor
    print_info "Building Processor image..."
    docker build -f app/processor/Dockerfile -t notifyx-processor:latest .
    
    # Build Workers
    print_info "Building Worker images..."
    docker build -f app/workers/worker-email/Dockerfile -t notifyx-worker-email:latest .
    docker build -f app/workers/worker-sms/Dockerfile -t notifyx-worker-sms:latest .
    docker build -f app/workers/worker-push/Dockerfile -t notifyx-worker-push:latest .
    docker build -f app/workers/worker-webhook/Dockerfile -t notifyx-worker-webhook:latest .
    
    print_info "All images built successfully"
}

# Function to start services in docker mode
start_docker() {
    print_info "Starting services in DOCKER mode..."
    
    check_docker
    check_infrastructure
    
    build_images
    
    cd "$PROJECT_ROOT"
    
    # Start notification services
    print_info "Starting notification services in Docker..."
    docker compose -f docker-compose.notifyx.yaml up -d
    
    print_info "All services started in Docker mode"
    print_info "API: http://localhost:8080"
    print_info "Use 'docker compose -f docker-compose.notifyx.yaml down' to stop services"
}

# Main script
if [ "$1" == "direct" ]; then
    start_direct
elif [ "$1" == "docker" ]; then
    start_docker
else
    echo "Usage: $0 [direct|docker]"
    echo ""
    echo "Options:"
    echo "  direct  - Start services as local processes (connect to Docker network for infra)"
    echo "  docker  - Build images and start all services in Docker"
    exit 1
fi

