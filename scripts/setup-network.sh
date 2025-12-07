#!/bin/bash

# Setup script to create Docker network for infrastructure

NETWORK_NAME="notifyx-infra-net"

if docker network ls | grep -q "$NETWORK_NAME"; then
    echo "Network $NETWORK_NAME already exists"
else
    echo "Creating network $NETWORK_NAME..."
    docker network create "$NETWORK_NAME"
    echo "Network $NETWORK_NAME created successfully"
fi

