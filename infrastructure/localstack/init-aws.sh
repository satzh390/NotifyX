#!/bin/bash
set -e

echo "Creating S3 bucket for templates..."
awslocal s3 mb s3://notifyx-templates || true
echo "S3 bucket 'notifyx-templates' ready"

