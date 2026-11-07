#!/usr/bin/env bash
set -euo pipefail

DATA_DIR="/data/centinela"

echo "Preparing Centinela SQL Server PI on Rocky Linux..."
if mountpoint -q /data; then
    echo "Dedicated /data mount detected."
else
    echo "Warning: /data is not a separate mount; verify storage before production use."
fi

mkdir -p "$DATA_DIR/prometheus" "$DATA_DIR/grafana" "$DATA_DIR/backups"
chown -R 65534:65534 "$DATA_DIR/prometheus"
chown -R 472:0 "$DATA_DIR/grafana"
chmod -R 775 "$DATA_DIR/prometheus" "$DATA_DIR/grafana" "$DATA_DIR/backups"

if command -v getenforce >/dev/null 2>&1 && [ "$(getenforce)" != "Disabled" ]; then
    chcon -Rt container_file_t "$DATA_DIR"
fi

if systemctl is-active --quiet firewalld; then
    firewall-cmd --permanent --add-port=3000/tcp
    firewall-cmd --reload
    echo "Grafana UI port 3000 opened. Restrict its source network with the DBA firewall policy."
else
    echo "Firewalld is inactive; allow only the required Grafana UI access in the host/network firewall."
fi

echo "Preparation complete. Review .env, then run 'docker compose config' and 'docker compose up -d'."
echo "Grafana: http://$(hostname -I | awk '{print $1}'):3000"