# PowerShell startup script for NotifyX

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("direct", "docker")]
    [string]$Mode
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$ProjectRoot = Split-Path -Parent $ScriptDir

function Write-Info {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warn {
    param([string]$Message)
    Write-Host "[WARN] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

function Test-Docker {
    try {
        docker info | Out-Null
    } catch {
        Write-Error "Docker is not running. Please start Docker and try again."
        exit 1
    }
}

function Test-Infrastructure {
    Write-Info "Checking infrastructure services..."
    
    $mongoRunning = docker ps --filter "name=mongo" --format "{{.Names}}" | Select-String "mongo"
    
    if (-not $mongoRunning) {
        Write-Warn "MongoDB is not running. Starting infrastructure..."
        Set-Location $ProjectRoot
        docker compose -f docker-compose.local.yaml up -d
        Write-Info "Waiting for services to be ready..."
        Start-Sleep -Seconds 5
    }
    
    Write-Info "Infrastructure services are running"
}

function Start-Direct {
    Write-Info "Starting services in DIRECT mode..."
    
    Test-Docker
    Test-Infrastructure
    
    Set-Location $ProjectRoot
    
    # Start API
    Write-Info "Starting API..."
    Set-Location "$ProjectRoot\app\api"
    $env:NOTIFYX_API_CONFIG = "config\config.yaml"
    Start-Process -FilePath "go" -ArgumentList "run", "./cmd/main.go" -WindowStyle Hidden
    Start-Sleep -Seconds 2
    
    # Start Processor
    Write-Info "Starting Processor..."
    Set-Location "$ProjectRoot\app\processor"
    $env:NOTIFYX_PROCESSOR_CONFIG = "config\config.yaml"
    Start-Process -FilePath "go" -ArgumentList "run", "./cmd/main.go" -WindowStyle Hidden
    Start-Sleep -Seconds 2
    
    # Start Workers
    Write-Info "Starting Workers..."
    
    Set-Location "$ProjectRoot\app\workers\worker-email"
    $env:NOTIFYX_WORKER_EMAIL_CONFIG = "config\config.yaml"
    Start-Process -FilePath "go" -ArgumentList "run", "./cmd/main.go" -WindowStyle Hidden
    
    Set-Location "$ProjectRoot\app\workers\worker-sms"
    $env:NOTIFYX_WORKER_SMS_CONFIG = "config\config.yaml"
    Start-Process -FilePath "go" -ArgumentList "run", "./cmd/main.go" -WindowStyle Hidden
    
    Set-Location "$ProjectRoot\app\workers\worker-push"
    $env:NOTIFYX_WORKER_PUSH_CONFIG = "config\config.yaml"
    Start-Process -FilePath "go" -ArgumentList "run", "./cmd/main.go" -WindowStyle Hidden
    
    Write-Info "All services started in direct mode"
    Write-Info "API: http://localhost:8080"
    Write-Info "Check Task Manager to stop processes"
}

function Build-Images {
    Write-Info "Building Docker images..."
    
    Set-Location $ProjectRoot
    
    # Build API
    Write-Info "Building API image..."
    docker build -f app/api/Dockerfile -t notifyx-api:latest .
    
    # Build Processor
    Write-Info "Building Processor image..."
    docker build -f app/processor/Dockerfile -t notifyx-processor:latest .
    
    # Build Workers
    Write-Info "Building Worker images..."
    docker build -f app/workers/worker-email/Dockerfile -t notifyx-worker-email:latest .
    docker build -f app/workers/worker-sms/Dockerfile -t notifyx-worker-sms:latest .
    docker build -f app/workers/worker-push/Dockerfile -t notifyx-worker-push:latest .
    
    Write-Info "All images built successfully"
}

function Start-Docker {
    Write-Info "Starting services in DOCKER mode..."
    
    Test-Docker
    Test-Infrastructure
    
    Build-Images
    
    Set-Location $ProjectRoot
    
    # Start notification services
    Write-Info "Starting notification services in Docker..."
    docker compose -f docker-compose.notifyx.yaml up -d
    
    Write-Info "All services started in Docker mode"
    Write-Info "API: http://localhost:8080"
    Write-Info "Use 'docker compose -f docker-compose.notifyx.yaml down' to stop services"
}

# Main execution
switch ($Mode) {
    "direct" { Start-Direct }
    "docker" { Start-Docker }
    default {
        Write-Host "Usage: .\start.ps1 [direct|docker]"
        Write-Host ""
        Write-Host "Options:"
        Write-Host "  direct  - Start services as local processes (connect to Docker network for infra)"
        Write-Host "  docker  - Build images and start all services in Docker"
        exit 1
    }
}

