# Paso 1 — Preflight AAS, antes de levantar el PI

Hacer en el servidor Rocky, en `/opt/centinela`, con el inventario piloto. No levantar el perfil `pi` hasta que todo esto dé OK.

## 1.1 Inventario y configuración

- [ ] `inventory/databases_fleet.csv` existe, con CLUSTER-02 y sus 2 nodos, host, puerto real (no asumir 1433, la guía dice 5000 para CLUSTER-02), ambiente, clúster y base inicial.
- [ ] Sin placeholders `REPLACE_WITH_*` y sin nombres duplicados.
- [ ] `.env` existe en el servidor, con `CLICKHOUSE_PASSWORD` y `GF_SECURITY_ADMIN_PASSWORD`. No copiarlo al repo.
- [ ] Generar y revisar el exporter:
```
python scripts/render_sql_exporter_config.py
python -m unittest discover -s tests -v
docker compose config
```

## 1.2 Preflight por target, con check-targets

En la carpeta `sampler/`, con Go instalado:
```
go run ./cmd/check-targets -inventory ../inventory/databases_fleet.csv -targets <nodo1>,<nodo2>
```

Esperado por nodo:
```
<nodo> OK SQL_MAJOR=<versión> VIEW_SERVER_STATE=1 VIEW_SERVER_PERFORMANCE_STATE=1 LOGICAL_CPUS=<n> VISIBLE_SCHEDULERS=<n> ACTIVE_SESSIONS=<n> SAMPLE_MS=<n>
```

No-go si:
- `VIEW_SERVER_STATE=0`, o en SQL 2022+ `VIEW_SERVER_PERFORMANCE_STATE=0`.
- Falla conexión, timeout o permisos.
- `SAMPLE_MS` es muy alto o errático en varias corridas.

El usuario monitor debe tener solo lectura aprobada por DBA. Nunca sysadmin ni db_owner.

## 1.3 Levantar el perfil PI

```
docker compose -f docker-compose.yml -f docker-compose.pi.yml --profile pi config
docker compose -f docker-compose.yml -f docker-compose.pi.yml --profile pi up -d --build
docker compose ps
docker compose logs clickhouse --tail 50
docker compose logs sampler --tail 100
```

No-go si ClickHouse no está sano, el sampler falla cerrado por falta de `CLICKHOUSE_PASSWORD`, o hay errores de conexión a SQL Server en los logs.

## 1.4 Comprobar métricas en Prometheus

En `http://<servidor>:9090/targets`, el job `centinela-pi-sampler` debe estar UP.
En `http://<servidor>:9090/graph`, deben existir:
- `centinela_aas_total`
- `centinela_sampler_success_ratio`
- `centinela_sql_visible_schedulers`
- `centinela_sampler_latency_ms`
- `centinela_sampler_errors_total`

No-go si `centinela_aas_total` no aparece, si la cobertura es menor a 0.8 sostenida, o si los schedulers son 0 o ausentes.

## 1.5 Comprobar ClickHouse

```
docker compose exec clickhouse clickhouse-client --query "SELECT count(), sample_status FROM centinela_pi.sample_runs GROUP BY sample_status"
```

Debe haber filas `success` creciendo cada 3 segundos por instancia, y las filas de muestra vacía deben existir como `success` con 0 sesiones, no como ausencia de fila.

Si todo lo anterior es OK, pasar al Paso 4: carga controlada, en `VALIDACION_AAS.md`.
