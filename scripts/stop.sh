#!/bin/bash

set -e

print_info() {
    echo -e "\033[0;32m[INFO]\033[0m $1"
}

print_info "Stopping NotifyX services..."

# Stop direct mode processes
if [ -f /tmp/notifyx-api.pid ]; then
    PID=$(cat /tmp/notifyx-api.pid)
    if ps -p $PID > /dev/null 2>&1; then
        kill $PID
        print_info "Stopped API (PID: $PID)"
    fi
    rm -f /tmp/notifyx-api.pid
fi

if [ -f /tmp/notifyx-processor.pid ]; then
    PID=$(cat /tmp/notifyx-processor.pid)
    if ps -p $PID > /dev/null 2>&1; then
        kill $PID
        print_info "Stopped Processor (PID: $PID)"
    fi
    rm -f /tmp/notifyx-processor.pid
fi

# Stop workers
for worker in email sms push; do
    if [ -f /tmp/notifyx-worker-$worker.pid ]; then
        PID=$(cat /tmp/notifyx-worker-$worker.pid)
        if ps -p $PID > /dev/null 2>&1; then
            kill $PID
            print_info "Stopped Worker-$worker (PID: $PID)"
        fi
        rm -f /tmp/notifyx-worker-$worker.pid
    fi
done

# Stop docker compose services
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
cd "$PROJECT_ROOT"

if docker compose -f docker-compose.notifyx.yaml ps -q > /dev/null 2>&1; then
    docker compose -f docker-compose.notifyx.yaml down
    print_info "Stopped Docker services"
fi

print_info "All services stopped"

