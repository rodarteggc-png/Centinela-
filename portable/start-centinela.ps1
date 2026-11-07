$ErrorActionPreference = "Stop"
$portableDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$rootDir = Split-Path -Parent $portableDir
$binDir = Join-Path $portableDir "bin"
$configDir = Join-Path $portableDir "config"
$dataDir = Join-Path $portableDir "data"
$grafanaHome = Join-Path $binDir "grafana"
$requiredFiles = @(
  (Join-Path $binDir "prometheus.exe"),
  (Join-Path $binDir "sql_exporter.exe"),
  (Join-Path $grafanaHome "bin\grafana-server.exe"),
  (Join-Path $configDir "prometheus.yml"),
  (Join-Path $configDir "grafana.ini"),
  (Join-Path $rootDir "exporters\sql-exporter\config.yml")
)

foreach ($requiredFile in $requiredFiles) {
  if (-not (Test-Path $requiredFile)) {
    throw "Required file not found: $requiredFile"
  }
}

foreach ($port in @(9090, 9399, 3000)) {
  if (Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue) {
    throw "Port $port is already in use; stop its owner before starting Centinela."
  }
}

if (-not $env:GF_SECURITY_ADMIN_PASSWORD) {
  $envFile = Join-Path $rootDir ".env"
  if (Test-Path $envFile) {
    $passwordLine = Get-Content $envFile | Where-Object { $_ -match '^GF_SECURITY_ADMIN_PASSWORD=' } | Select-Object -Last 1
    if ($passwordLine) {
      $env:GF_SECURITY_ADMIN_PASSWORD = $passwordLine.Substring($passwordLine.IndexOf('=') + 1)
    }
  }
}
if ([string]::IsNullOrWhiteSpace($env:GF_SECURITY_ADMIN_PASSWORD)) {
  throw "Set GF_SECURITY_ADMIN_PASSWORD in the process environment or .env before starting Grafana."
}

New-Item -ItemType Directory -Force -Path (Join-Path $dataDir "prometheus"), (Join-Path $dataDir "grafana") | Out-Null
$env:GF_PATHS_DATA = Join-Path $dataDir "grafana"
$env:GF_PATHS_LOGS = Join-Path $dataDir "grafana\log"
$env:GF_PATHS_PLUGINS = Join-Path $grafanaHome "data\plugins"
$env:GF_PATHS_PROVISIONING = Join-Path $configDir "grafana-provisioning"
$env:GF_DASHBOARDS_DEFAULT_HOME_DASHBOARD_PATH = Join-Path $portableDir "dashboards\centinela-executive-overview.json"
$env:CENTINELA_DASHBOARDS_PATH = Join-Path $portableDir "dashboards"

$prometheusArgs = @(
  "--config.file=$(Join-Path $configDir 'prometheus.yml')",
  "--storage.tsdb.path=$(Join-Path $dataDir 'prometheus')",
  "--web.listen-address=127.0.0.1:9090"
)
Start-Process -FilePath (Join-Path $binDir "prometheus.exe") -ArgumentList $prometheusArgs -WorkingDirectory $rootDir -WindowStyle Hidden

$sqlConfig = Join-Path $rootDir "exporters\sql-exporter\config.yml"
$sqlArgs = @("-config.file=`"$sqlConfig`"", "-web.listen-address=127.0.0.1:9399")
Start-Process -FilePath (Join-Path $binDir "sql_exporter.exe") -ArgumentList $sqlArgs -WorkingDirectory $rootDir -WindowStyle Hidden

$grafanaArgs = @("--homepath", $grafanaHome, "--config", (Join-Path $configDir "grafana.ini"))
Start-Process -FilePath (Join-Path $grafanaHome "bin\grafana-server.exe") -ArgumentList $grafanaArgs -WorkingDirectory $rootDir -WindowStyle Hidden

Write-Host "Centinela SQL Server PI is starting."
Write-Host "Grafana: http://localhost:3000/d/centinela-sqlserver-cluster"
Write-Host "Prometheus: http://localhost:9090"
Write-Host "SQL Exporter: http://localhost:9399/metrics"
Start-Process "http://localhost:3000/d/centinela-executive-overview"