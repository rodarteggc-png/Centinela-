#!/usr/bin/env python3
"""Render SQL Server exporter targets from the fleet inventory."""

import argparse
import csv
import json
import os
import re
import sys
import tempfile
from pathlib import Path
from urllib.parse import quote


REQUIRED_FIELDS = {
    "engine",
    "instance_name",
    "host_or_ip",
    "port",
    "environment",
    "cluster_name",
    "username",
    "password",
    "database_name",
}
SUPPORTED_ENGINES = {"sqlserver", "postgres", "oracle"}


def load_inventory(inventory_path):
    with inventory_path.open("r", encoding="utf-8-sig", newline="") as source:
        rows = (line for line in source if line.strip() and not line.lstrip().startswith("#"))
        reader = csv.DictReader(rows)
        missing_fields = REQUIRED_FIELDS.difference(reader.fieldnames or [])
        if missing_fields:
            raise ValueError("Missing inventory columns: " + ", ".join(sorted(missing_fields)))

        instances = []
        seen_names = set()
        for row_number, row in enumerate(reader, start=2):
            engine = (row.get("engine") or "").strip().lower()
            if engine not in SUPPORTED_ENGINES:
                raise ValueError(f"Unsupported engine on inventory row {row_number}: {engine!r}")
            if engine != "sqlserver":
                continue

            values = {key: (row.get(key) or "").strip() for key in REQUIRED_FIELDS}
            missing_values = [key for key, value in values.items() if not value]
            if missing_values:
                raise ValueError(
                    f"SQL Server inventory row {row_number} has empty fields: "
                    + ", ".join(sorted(missing_values))
                )
            placeholders = [
                key for key, value in values.items() if value.upper().startswith("REPLACE_WITH_")
            ]
            if placeholders:
                raise ValueError(
                    f"SQL Server inventory row {row_number} contains example placeholders: "
                    + ", ".join(sorted(placeholders))
                )
            if values["instance_name"] in seen_names:
                raise ValueError(f"Duplicate SQL Server instance_name: {values['instance_name']}")
            seen_names.add(values["instance_name"])

            try:
                port = int(values["port"])
            except ValueError as error:
                raise ValueError(f"Invalid port on inventory row {row_number}") from error
            if not 1 <= port <= 65535:
                raise ValueError(f"Port out of range on inventory row {row_number}")

            host = values["host_or_ip"]
            if ":" in host and not host.startswith("["):
                host = f"[{host}]"
            dsn = (
                f"sqlserver://{quote(values['username'], safe='')}:"
                f"{quote(values['password'], safe='')}@{host}:{port}"
                f"?database={quote(values['database_name'], safe='')}&encrypt=disable"
            )
            instances.append(
                {
                    "name": values["instance_name"],
                    "dsn": dsn,
                    "cluster": values["cluster_name"],
                    "environment": values["environment"],
                }
            )

    if not instances:
        raise ValueError("No SQL Server instances found; exporter configuration was not changed")
    return instances


def render_configuration(inventory_path, config_path):
    instances = load_inventory(inventory_path)
    template_path = config_path
    if not template_path.exists():
        template_path = config_path.with_name("config.example.yml")
    if not template_path.exists():
        raise ValueError(f"No exporter config or example template found for {config_path}")
    template = template_path.read_text(encoding="utf-8-sig")
    collectors_match = re.search(r"(?m)^collectors:\s*$", template)
    if collectors_match is None:
        raise ValueError(f"No top-level collectors section found in {config_path}")

    target_lines = []
    for instance in instances:
        target_lines.extend(
            [
                "      - targets:",
                f"          {json.dumps(instance['name'])}: {json.dumps(instance['dsn'])}",
                "        labels:",
                f"          cluster: {json.dumps(instance['cluster'])}",
                f"          environment: {json.dumps(instance['environment'])}",
                '          db_engine: "sqlserver"',
            ]
        )

    generated_header = "\n".join(
        [
            "# SQL Server exporter targets generated from inventory/databases_fleet.csv",
            f"# SQL Server instances: {len(instances)}",
            "global:",
            "  scrape_timeout_offset: 500ms",
            "  min_interval: 0s",
            "  max_connections: 3",
            "  max_idle_connections: 3",
            "jobs:",
            "  - job_name: sqlserver_fleet",
            "    collectors:",
            "      - mssql_standard",
            "      - mssql_alwayson",
            "      - mssql_foglight",
            "    static_configs:",
            *target_lines,
            "",
        ]
    )
    output = generated_header + template[collectors_match.start() :]

    temporary_path = None
    try:
        with tempfile.NamedTemporaryFile(
            "w", encoding="utf-8", newline="\n", dir=config_path.parent, delete=False
        ) as temporary_file:
            temporary_file.write(output)
            temporary_path = Path(temporary_file.name)
        os.replace(temporary_path, config_path)
    finally:
        if temporary_path is not None and temporary_path.exists():
            temporary_path.unlink()

    return len(instances)


def main(argv=None):
    parser = argparse.ArgumentParser(description=__doc__)
    project_root = Path(__file__).resolve().parents[1]
    parser.add_argument(
        "--inventory", type=Path, default=project_root / "inventory" / "databases_fleet.csv"
    )
    parser.add_argument(
        "--config", type=Path, default=project_root / "exporters" / "sql-exporter" / "config.yml"
    )
    arguments = parser.parse_args(argv)

    try:
        count = render_configuration(arguments.inventory, arguments.config)
    except (OSError, ValueError) as error:
        print(f"ERROR: {error}", file=sys.stderr)
        return 1

    print(f"Rendered SQL Server exporter configuration for {count} instance(s).")
    print("PostgreSQL and Oracle targets are not rendered by this SQL Server exporter.")
    print("Review the generated file, then restart the exporter explicitly when ready.")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())