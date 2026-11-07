$ErrorActionPreference = "Stop"
$portableDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$binDir = Join-Path $portableDir "bin"
$tempDir = Join-Path $portableDir "temp"
New-Item -ItemType Directory -Force -Path $binDir, $tempDir | Out-Null

function Download-And-Extract {
    param([string]$Name, [string]$Url, [string]$ArchiveFile, [scriptblock]$ExtractAction)
    $archivePath = Join-Path $tempDir $ArchiveFile
    Write-Host "Downloading $Name..."
    Invoke-WebRequest -Uri $Url -OutFile $archivePath -UseBasicParsing
    & $ExtractAction $archivePath
}

Download-And-Extract -Name "Prometheus" `
    -Url "https://github.com/prometheus/prometheus/releases/download/v2.54.1/prometheus-2.54.1.windows-amd64.zip" `
    -ArchiveFile "prometheus.zip" -ExtractAction {
        param($archivePath)
        $extractDir = Join-Path $tempDir "prometheus_extract"
        Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force
        $releaseDir = Get-ChildItem -Directory $extractDir | Select-Object -First 1
        Copy-Item (Join-Path $releaseDir.FullName "prometheus.exe") (Join-Path $binDir "prometheus.exe") -Force
        Remove-Item $extractDir -Recurse -Force
    }

Download-And-Extract -Name "Grafana OSS" `
    -Url "https://dl.grafana.com/oss/release/grafana-11.1.0.windows-amd64.zip" `
    -ArchiveFile "grafana.zip" -ExtractAction {
        param($archivePath)
        $extractDir = Join-Path $tempDir "grafana_extract"
        Expand-Archive -Path $archivePath -DestinationPath $extractDir -Force
        $releaseDir = Get-ChildItem -Directory $extractDir | Select-Object -First 1
        $grafanaTarget = Join-Path $binDir "grafana"
        if (Test-Path $grafanaTarget) { Remove-Item $grafanaTarget -Recurse -Force }
        Move-Item $releaseDir.FullName $grafanaTarget -Force
        Remove-Item $extractDir -Recurse -Force
    }

if (-not (Test-Path (Join-Path $binDir "sql_exporter.exe"))) {
    throw "Place a trusted sql_exporter.exe in portable/bin before starting the local SQL Server PI."
}
Write-Host "Prometheus and Grafana are ready; sql_exporter.exe is present."