$ErrorActionPreference = "Stop"
$sourceRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path
$targetRoot = "Z:\opt\centinela"

if (-not (Test-Path "Z:\")) {
  Write-Host "Connecting SSHFS drive Z:. Enter credentials in the Windows prompt." -ForegroundColor Cyan
  net use Z: \\sshfs.r\pggr@192.168.220.20 /persistent:no
}
if (-not (Test-Path $targetRoot)) {
  throw "Remote project path is unavailable: $targetRoot"
}

function Copy-ProjectFile {
  param([string]$RelativePath)

  $sourcePath = Join-Path $sourceRoot $RelativePath
  $destinationPath = Join-Path $targetRoot $RelativePath
  if (-not (Test-Path $sourcePath -PathType Leaf)) {
    throw "Source file not found: $RelativePath"
  }

  $destinationDirectory = Split-Path $destinationPath -Parent
  if (-not (Test-Path $destinationDirectory)) {
    New-Item -ItemType Directory -Path $destinationDirectory -Force | Out-Null
  }
  Copy-Item -LiteralPath $sourcePath -Destination $destinationPath -Force

  $sourceHash = (Get-FileHash -LiteralPath $sourcePath -Algorithm SHA256).Hash
  $destinationHash = (Get-FileHash -LiteralPath $destinationPath -Algorithm SHA256).Hash
  if ($sourceHash -ne $destinationHash) {
    throw "Hash mismatch after copy: $RelativePath"
  }
  Write-Host "Verified: $RelativePath" -ForegroundColor Green
}

$files = @(
  "docker-compose.yml",
  "README.md",
  ".env.example",
  ".gitignore",
  ".github\copilot-instructions.md",
  "inventory\databases_fleet.csv",
  "exporters\sql-exporter\config.yml",
  "prometheus\prometheus.yml",
  "prometheus\prometheus-pi.yml",
  "docker-compose.pi.yml",
  "prometheus\alerts\foglight_rules.yml",
  "grafana\provisioning\dashboards\dashboard.yml",
  "grafana\provisioning\datasources\datasources.yml",
  "scripts\sync-fleet-inventory.sh",
  "scripts\sync-fleet-inventory.ps1",
  "scripts\render_sql_exporter_config.py",
  "scripts\create-grafana-local-user.py",
  "docs\MATRIZ_PARIDAD_FOGLIGHT_BD.md",
  "docs\PI_STORAGE_CONTRACT.md",
  "docs\GUIA_SQLSERVER_PI.md",
  "docs\VALIDACION_AAS.md",
  "docs\PREFLIGHT_AAS.md",
  "docs\PRUEBA_CARGA_AAS.sql",
  "docs\FASES_POR_MOTOR.md",
  "docs\PARIDAD_PORCENTAJE.md",
  "clickhouse\init\001_pi_schema.sql",
  "clickhouse\queries\aas_by_minute.sql",
  "clickhouse\queries\top_sql_by_samples.sql",
  "clickhouse\queries\top_sql_text_by_samples.sql",
  "clickhouse\queries\blocking_by_minute.sql",
  "sampler\go.sum",
  "sampler\Dockerfile",
  "sampler\go.mod",
  "sampler\README.md",
  "sampler\internal\pi\aas.go",
  "sampler\internal\pi\aas_test.go",
  "sampler\sqlserver\active_sessions.sql",
  "sampler\sqlserver\query.go",
  "sampler\internal\config\targets.go",
  "sampler\internal\config\targets_test.go",
  "sampler\internal\clickhouse\writer.go",
  "sampler\internal\clickhouse\writer_test.go",
  "sampler\internal\metrics\prometheus.go",
  "sampler\internal\metrics\prometheus_test.go",
  "sampler\internal\model\samples.go",
  "sampler\internal\sqlserver\sampler.go",
  "sampler\internal\sqlserver\sampler_test.go",
  "sampler\cmd\sampler\main.go"
)

foreach ($file in $files) {
  Copy-ProjectFile -RelativePath $file
}

Get-ChildItem -Path (Join-Path $sourceRoot "grafana\dashboards") -Filter "*.json" -File | ForEach-Object {
  Copy-ProjectFile -RelativePath (Join-Path "grafana\dashboards" $_.Name)
}
$currentDashboardNames = @(Get-ChildItem -Path (Join-Path $sourceRoot "grafana\dashboards") -Filter "centinela-*.json" -File | ForEach-Object { $_.Name })
$remoteDashboardDirectory = Join-Path $targetRoot "grafana\dashboards"
Get-ChildItem -Path $remoteDashboardDirectory -Filter "centinela-*.json" -File -ErrorAction SilentlyContinue | Where-Object { $currentDashboardNames -notcontains $_.Name } | ForEach-Object {
  Remove-Item -LiteralPath $_.FullName -Force
  Write-Host "Removed obsolete Centinela dashboard: $($_.Name)" -ForegroundColor Yellow
}
$localDashboardNames = @(Get-ChildItem -Path (Join-Path $sourceRoot "grafana\dashboards") -Filter "centinela-*.json" -File | ForEach-Object { $_.Name })
$remoteDashboardDirectory = Join-Path $targetRoot "grafana\dashboards"
Get-ChildItem -Path $remoteDashboardDirectory -Filter "centinela-*.json" -File -ErrorAction SilentlyContinue | Where-Object { $localDashboardNames -notcontains $_.Name } | ForEach-Object {
  Remove-Item -LiteralPath $_.FullName -Force
  Write-Host "Removed obsolete dashboard: $($_.Name)" -ForegroundColor Yellow
}
Get-ChildItem -Path (Join-Path $sourceRoot "grafana\branding") -File | ForEach-Object {
  Copy-ProjectFile -RelativePath (Join-Path "grafana\branding" $_.Name)
}

Write-Host "All SQL Server PI files copied and hash-verified. No services were restarted." -ForegroundColor Green
Write-Host "On the server, validate with 'docker compose config' before applying changes." -ForegroundColor Yellow