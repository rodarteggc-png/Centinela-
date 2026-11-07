Write-Host "Stopping Centinela SQL Server PI..." -ForegroundColor Yellow

foreach ($serviceName in @("grafana-server", "prometheus", "sql_exporter")) {
    Get-Process -Name $serviceName -ErrorAction SilentlyContinue | Stop-Process -Force
}

Write-Host "Centinela SQL Server PI stopped." -ForegroundColor Green