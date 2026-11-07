# Centinela - Foglight Local para Bases de Datos

Foglight local para el equipo de DBA, con Tavo como administrador. Objetivo: vigilar toda la flota, 420 bases (SQL Server 300 servidores, PostgreSQL 50, Oracle 30, Azure 30 bases, MySQL 10), sin el límite de licencias de Foglight.

Primera entrega local para SQL Server, con dos nodos inventariados en `SQL-CLUSTER-02`. Paridad estimada con Foglight: ~23% (ver `docs/PARIDAD_PORCENTAJE.md`).

Fases por motor: 1 SQL Server, 2 PostgreSQL, 3 Oracle (RAC), 4 Azure, 5 MySQL. Ningún motor se publica sin conexión real, métricas reales y validación por DBA. Ver `docs/FASES_POR_MOTOR.md`.

## Catálogo del portal

1. `01 - SQL Server Performance Investigator`: métricas del motor, Always On y AAS por muestras cuando el perfil PI está habilitado. Incluye paneles PI de Top SQL por muestras y Bloqueos por minuto (requieren ClickHouse).
2. `02 - Executive Overview`: resumen ejecutivo SQL Server y acceso al PI.

Oracle, PostgreSQL, Azure y MySQL no se publican en el portal activo hasta contar con conexiones, métricas reales y validación por DBA. No se incluyen dashboards simulados.

Estado PI: AAS, Top SQL por muestras y bloqueos están preparados, no validados en servidor. La alerta de saturación AAS está retirada hasta tener una métrica real. El sampler ya captura `sql_handle` y guarda `statement_key`; falta poblar `sql_text` de forma asíncrona.

## Runtime base
## Runtime y primera configuración

Desde un clon limpio, crea los archivos locales de entorno e inventario a partir de las plantillas. El inventario activo y la configuración generada están ignorados por Git porque contienen DSN:

```bash
cp .env.example .env
cp inventory/databases_fleet.example.csv inventory/databases_fleet.csv
cp exporters/sql-exporter/config.example.yml exporters/sql-exporter/config.yml
```

Sustituye los placeholders del inventario y define secretos solo en el entorno local/servidor autorizado. El renderer rechaza los valores `REPLACE_WITH_*`:

```powershell
python scripts/render_sql_exporter_config.py
python -m unittest discover -s tests -v
```

Después valida con `docker compose config` e inicia el runtime base con `docker compose up -d`. El perfil PI opcional activa ClickHouse y el sampler dentro de la red privada:

```bash
docker compose -f docker-compose.yml -f docker-compose.pi.yml --profile pi config
docker compose -f docker-compose.yml -f docker-compose.pi.yml --profile pi up -d --build
```

Usa una contraseña local para ClickHouse antes de habilitar PI. Los datos se almacenan bajo `CENTINELA_DATA_DIR`, que en Rocky Linux debe apuntar a `/data/centinela`. Nunca copies `.env` al repositorio ni lo sobrescribas al desplegar.

## Despliegue al servidor

Desde Windows, con la unidad SSHFS montada, `scripts/deploy-to-server.ps1` copia los archivos del proyecto con verificación SHA256 y no reinicia servicios. Incluye los archivos nuevos de `clickhouse/queries/`, `docs/` y el sampler. En el servidor, valida con `docker compose config` antes de aplicar cambios, y con `go test ./...` en `sampler/` antes de desplegar el sampler. Para PI, el plugin de ClickHouse en Grafana se instala con `GF_INSTALL_PLUGINS` en `docker-compose.pi.yml`.

## Verificación

```powershell
python -m unittest discover -s tests -v
```

La matriz de [aceptación PI](docs/MATRIZ_PARIDAD_FOGLIGHT_BD.md) distingue código preparado de capacidades validadas en operación.
