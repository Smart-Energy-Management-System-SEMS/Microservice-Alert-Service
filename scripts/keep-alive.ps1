$ErrorActionPreference = "Stop"

$targetUrl = $env:TARGET_URL
$pingPath = if ($env:PING_PATH) { $env:PING_PATH } else { "/api/v1/alerts" }
$intervalSeconds = if ($env:INTERVAL_SECONDS) { [int]$env:INTERVAL_SECONDS } else { 600 }

if ([string]::IsNullOrWhiteSpace($targetUrl)) {
    Write-Error "Define TARGET_URL, por ejemplo https://tu-servicio.onrender.com"
}

$targetUrl = $targetUrl.TrimEnd("/")
$url = "$targetUrl$pingPath"

Write-Host "Keep-alive iniciado para: $url"
Write-Host "Intervalo: $intervalSeconds s"

while ($true) {
    try {
        Invoke-WebRequest -Uri $url -Method Get -TimeoutSec 20 | Out-Null
        Write-Host "$(Get-Date -Format o) ping ok"
    }
    catch {
        Write-Warning "$(Get-Date -Format o) ping error: $($_.Exception.Message)"
    }

    Start-Sleep -Seconds $intervalSeconds
}
