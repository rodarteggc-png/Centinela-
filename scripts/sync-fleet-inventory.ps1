param(
    [string]$InventoryPath = "$PSScriptRoot\..\inventory\databases_fleet.csv",
    [string]$ConfigPath = "$PSScriptRoot\..\exporters\sql-exporter\config.yml"
)

$rendererPath = Join-Path $PSScriptRoot "render_sql_exporter_config.py"
& python $rendererPath --inventory $InventoryPath --config $ConfigPath
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
