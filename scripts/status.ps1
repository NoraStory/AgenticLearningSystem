$Root = Split-Path -Parent $PSScriptRoot
Write-Host '=== Docker 基础设施 ==='
docker compose -f (Join-Path $Root 'docker-compose.yml') ps
Write-Host '=== HTTP 健康检查 ==='
try { Invoke-RestMethod 'http://127.0.0.1:8080/health' | ConvertTo-Json -Depth 4 } catch { Write-Host '后端未启动' }
try { $response=Invoke-WebRequest 'http://127.0.0.1:5000/' -UseBasicParsing; Write-Host "前端 HTTP $($response.StatusCode)" } catch { Write-Host '前端未启动' }
