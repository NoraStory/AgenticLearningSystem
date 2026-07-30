param([switch]$SkipFrontendInstall)
$ErrorActionPreference = 'Stop'
$Root = Split-Path -Parent $PSScriptRoot
$Backend = Join-Path $Root 'backend'
$Frontend = Join-Path $Root 'frontend'
Write-Host '[1/4] 启动 PostgreSQL、Redis、MinIO...'
docker compose -f (Join-Path $Root 'docker-compose.yml') up -d
Write-Host '[2/4] 创建项目内 Python venv...'
$VenvPython = Join-Path $Backend '.venv\Scripts\python.exe'
if (-not (Test-Path -LiteralPath $VenvPython)) { python -m venv (Join-Path $Backend '.venv') }
Write-Host '[3/4] 下载 Go 依赖到项目目录...'
$env:GOMODCACHE = Join-Path $Backend '.gomodcache'
$env:GOCACHE = Join-Path $Backend '.gocache'
$env:GOPATH = Join-Path $Backend '.gopath'
$env:GOPROXY = 'https://goproxy.cn,direct'
Push-Location $Backend
try { go mod download } finally { Pop-Location }
if (-not $SkipFrontendInstall) {
  Write-Host '[4/4] 下载前端依赖到项目目录...'
  $env:COREPACK_HOME = Join-Path $Frontend '.corepack'
  $env:PNPM_HOME = Join-Path $Frontend '.pnpm-home'
  Push-Location $Frontend
  try { corepack pnpm install --frozen-lockfile --store-dir .pnpm-store } finally { Pop-Location }
} else { Write-Host '[4/4] 已跳过前端依赖安装。' }
Write-Host '初始化完成。运行 .\scripts\start.ps1 启动系统。'
