param([switch]$Infrastructure)
$Root = Split-Path -Parent $PSScriptRoot
$RunDir = Join-Path $Root '.run'
foreach ($name in @('frontend','backend')) {
  $pidFile = Join-Path $RunDir "$name.pid"
  if (Test-Path -LiteralPath $pidFile) {
    $processId = [int](Get-Content -LiteralPath $pidFile -Raw)
    Stop-Process -Id $processId -Force -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $pidFile -Force -ErrorAction SilentlyContinue
    Write-Host "已停止 $name (PID $processId)"
  }
}
if ($Infrastructure) { docker compose -f (Join-Path $Root 'docker-compose.yml') down }
