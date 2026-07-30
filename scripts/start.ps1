$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $PSScriptRoot
$Backend = Join-Path $Root 'backend'
$Frontend = Join-Path $Root 'frontend'
$RunDir = Join-Path $Root '.run'
$LogDir = Join-Path $Root 'logs'
New-Item -ItemType Directory -Path $RunDir,$LogDir -Force | Out-Null
docker compose -f (Join-Path $Root 'docker-compose.yml') up -d
$env:GOMODCACHE = Join-Path $Backend '.gomodcache'
$env:GOCACHE = Join-Path $Backend '.gocache'
$env:GOPATH = Join-Path $Backend '.gopath'
$env:GOPROXY = 'https://goproxy.cn,direct'
$backendProcess = Start-Process -FilePath 'go' -ArgumentList @('run','./cmd/server') -WorkingDirectory $Backend -RedirectStandardOutput (Join-Path $LogDir 'backend.out.log') -RedirectStandardError (Join-Path $LogDir 'backend.err.log') -WindowStyle Hidden -PassThru
Set-Content -LiteralPath (Join-Path $RunDir 'backend.pid') -Value $backendProcess.Id
$env:COREPACK_HOME = Join-Path $Frontend '.corepack'
$env:PNPM_HOME = Join-Path $Frontend '.pnpm-home'
$frontendProcess = Start-Process -FilePath 'corepack' -ArgumentList @('pnpm','exec','next','dev','-p','5000') -WorkingDirectory $Frontend -RedirectStandardOutput (Join-Path $LogDir 'frontend.out.log') -RedirectStandardError (Join-Path $LogDir 'frontend.err.log') -WindowStyle Hidden -PassThru
Set-Content -LiteralPath (Join-Path $RunDir 'frontend.pid') -Value $frontendProcess.Id
Write-Host "后端 PID: $($backendProcess.Id)  http://localhost:8080"
Write-Host "前端 PID: $($frontendProcess.Id)  http://localhost:5000"
Write-Host 'MinIO Console: http://localhost:9001'
Write-Host '日志目录: .\logs'
