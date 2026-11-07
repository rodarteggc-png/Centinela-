# Validar el AAS real del PI — SQL Server

Objetivo: que el AAS de Centinela sea una medición por muestras de sesiones activas, no una estimación. Estado al 2026-11-06: código revisado, validación en servidor pendiente.

## Lo que ya revisé en código y pasa

- `sampler/internal/pi/aas.go` — `CalculateAAS` es correcto: denominador solo muestras exitosas, muestra exitosa vacía cuenta como cero, fallos/timeouts se excluyen, rechaza IDs y sesiones duplicadas, calcula cobertura.
- `sampler/internal/pi/aas_test.go` — 4 tests cubren esos casos: vacío + fallo, conteo por sesión, duplicados, y cero éxitos.
- `sampler/internal/metrics/prometheus.go` — ventana móvil de 1 minuto, exporta `centinela_aas_total` por categoría, `centinela_sampler_success_ratio`, latencia, errores y `centinela_sql_visible_schedulers`. No exporta AAS si no hay muestras exitosas.
- `clickhouse/queries/aas_by_minute.sql` — AAS por minuto = observaciones / muestras exitosas, NULL si el denominador es cero, con cobertura por intento.
- `sampler/cmd/sampler/main.go` — daemon real, 1 goroutine por target, intervalo 3s, timeout 2s, registra cada intento en `sample_runs` aunque la muestra sea vacía.
- `sampler/cmd/check-targets/main.go` — herramienta de preflight que ya existe y hay que usar: valida `VIEW SERVER STATE`, `VIEW SERVER PERFORMANCE STATE` en SQL 2022+, detecta CPUs/schedulers, hace una muestra y reporta sesiones activas y duración.
- Pruebas Python del inventario: 5/5 pasan. No cubren AAS.

## Hallazgos que hay que corregir antes de darlo por válido

1. La alerta de saturación AAS sí está en `prometheus/alerts/foglight_rules.yml`, aunque la matriz y el contrato dicen que está retirada hasta tener métrica real. Quitarla o desactivarla hasta validar.
2. El panel "Resource Time Distribution" del dashboard usa `rate(sqlserver_wait_category_time_ms)`, que son contadores acumulados de `sys.dm_os_wait_stats`, no sesiones activas. No es AAS y no debe presentarse como tal.
3. En `main.go`, si un target falla al abrir la conexión, todo el sampler falla. Para 300 servidores, un target malo no debe tumbar a los demás.
4. Si `DetectCapacity` falla, el sampler sigue sin schedulers visibles y la comparación de saturación no tiene con qué compararse. Hay que registrarlo como fallo de cobertura, no seguir en silencio.
5. No se corrieron los tests Go en este entorno porque no hay toolchain Go. Hay que correr `go test ./...` en `sampler/` antes de desplegar.

## Plan de validación en servidor — en este orden

### Paso 1 — Preflight, sin levantar el PI
En el servidor, con el inventario piloto (CLUSTER-02, y luego clúster 10):
```
go run ./cmd/check-targets -inventory inventory/databases_fleet.csv -targets <nodo1>,<nodo2>
```
Debe dar OK con versión, permisos, CPUs, schedulers, sesiones activas y duración de muestra. Si falla por permisos, no seguir: el usuario monitor necesita solo lectura aprobada por DBA, nunca sysadmin.

### Paso 2 — Levantar el perfil PI y comprobar que hay datos
```
docker compose -f docker-compose.yml -f docker-compose.pi.yml --profile pi config
docker compose -f docker-compose.yml -f docker-compose.pi.yml --profile pi up -d --build
```
Comprobar: ClickHouse sano, sampler en logs sin errores, Prometheus con el job `centinela-pi-sampler` en UP, y las métricas `centinela_aas_total`, `centinela_sampler_success_ratio` y `centinela_sql_visible_schedulers` presentes.

### Paso 3 — Carga controlada, la prueba clave
Generar una carga conocida y medible durante varios minutos, por ejemplo sesiones que corran CPU, esperen IO y se bloqueen entre sí, con un número conocido de sesiones activas.
Comparar, minuto a minuto:
- AAS por categoría en Grafana / Prometheus contra el conteo manual en `sys.dm_exec_requests` en ese mismo minuto.
- AAS por categoría en ClickHouse con `aas_by_minute.sql` contra el AAS de Prometheus.
- Cobertura: intentos vs exitosos. Debe ser alta en condiciones normales.

Criterio: si la carga controlada es, por ejemplo, 4 sesiones CPU constantes, el AAS CPU debe ser cercano a 4, no a 0 ni a 20.

### Paso 4 — Probar los casos que no son carga
- Periodo sin carga: AAS debe ser 0, no ausente, y la cobertura debe seguir alta.
- Fallo inducido: parar una instancia o usar credenciales malas unos minutos. El AAS no debe caer a cero por el fallo, la cobertura debe caer, y al recuperar debe volver sin huecos falsos.
- Reinicio del sampler y de ClickHouse: no debe duplicar muestras ni perder el registro de intentos.

### Paso 5 — Medir impacto y volumen
- Duración de `active_sessions.sql` por muestra y CPU que consume en SQL Server.
- Volumen en ClickHouse por día para 2 nodos, y extrapolar a 300 servidores antes de escalar. Con 3s por muestra, 300 servidores son 6.000 muestras por minuto.
- Retención de 30 días: ¿cuánto disco ocupa y cuánto tarda `aas_by_minute.sql`?

### Paso 6 — Solo entonces, tocar dashboards y alertas
- Corregir el panel de Resource Time o etiquetarlo como acumulado, no como AAS.
- Re-activar la alerta de saturación solo si el AAS y los schedulers son reales y la cobertura es alta.
- Registrar la evidencia en `docs/MATRIZ_PARIDAD_FOGLIGHT_BD.md`: fecha, carga usada, AAS esperado vs obtenido, cobertura, permisos e impacto.

## Criterio de aceptación

AAS validado cuando, con carga controlada, el AAS por categoría coincide con el conteo manual, la cobertura es alta y estable, los fallos no contaminan el denominador, y ClickHouse y Prometheus dan el mismo AAS por minuto.
