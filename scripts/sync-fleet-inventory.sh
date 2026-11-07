#!/usr/bin/env bash
set -euo pipefail

BASE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INVENTORY_FILE="${BASE_DIR}/inventory/databases_fleet.csv"
CONFIG_FILE="${BASE_DIR}/exporters/sql-exporter/config.yml"
python3 "${BASE_DIR}/scripts/render_sql_exporter_config.py" \
    --inventory "$INVENTORY_FILE" \
    --config "$CONFIG_FILE"

