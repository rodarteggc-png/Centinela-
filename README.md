# Centinela - Foglight Local para Bases de Datos

Primera entrega local para SQL Server, con dos nodos inventariados en `SQL-CLUSTER-02`.

## Catálogo del portal

1. `01 - SQL Server Performance Investigator`: métricas del motor, Always On y AAS por muestras cuando el perfil PI está habilitado.
2. `02 - Executive Overview`: resumen ejecutivo SQL Server y acceso al PI.

Oracle, PostgreSQL y Azure no se publican en el portal activo hasta contar con conexiones, métricas reales y validación por DBA. No se incluyen dashboards simulados.

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

## Verificación

```powershell
python -m unittest discover -s tests -v
```

La matriz de [aceptación PI](docs/MATRIZ_PARIDAD_FOGLIGHT_BD.md) distingue código preparado de capacidades validadas en operación.
