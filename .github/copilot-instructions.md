# Centinela: instrucciones de desarrollo

## Alcance vigente

Construir un Foglight local para SQL Server, inicialmente limitado a los dos nodos inventariados de `SQL-CLUSTER-02`. El runtime es Grafana, Prometheus y SQL Exporter. El sampler y ClickHouse son la siguiente rebanada, aún no operativa.
Construir un Foglight local para SQL Server, inicialmente limitado a los dos nodos inventariados de `SQL-CLUSTER-02`. El runtime base es Grafana, Prometheus y SQL Exporter; sampler y ClickHouse son opt-in con perfil `pi` y aún pendientes de validación operacional.

No añadir APM, demo-app, nube, Oracle ni PostgreSQL al runtime ni a la presentación actual. No afirmar paridad completa con Foglight antes de validar criterios en SQL Server.

## Reglas técnicas

- Mantener archivos YAML sin BOM.
- En SQL Exporter, métricas con varios valores requieren `value_label`.
- Evitar series duplicadas: agrupar por las etiquetas clave y agregar los valores.
- No guardar contraseñas en archivos versionados; `.env` es local y debe permanecer ignorado.
- Tratar texto SQL y planes como información potencialmente sensible. No usarlos como etiquetas Prometheus.
- Mantener SQL Exporter para Health Strip y Always On; el sampler es aditivo.

### AAS

Calcular AAS de muestras de sesiones activas, nunca de `rate()` sobre `sys.dm_os_wait_stats`, CPU porcentual ni un número fijo de cores:

$$\mathrm{AAS}_{c} = \frac{\sum_{i=1}^{N} n_{i,c}}{N}$$

Registrar intentos, contar como cero cada muestra exitosa sin sesiones, excluir fallos del denominador y reportar cobertura. Mantener cadencia regular; si es irregular, usar promedio ponderado por tiempo. Detectar cores por instancia.

## Estructura principal

- `exporters/sql-exporter/`: métricas de SQL Server y Always On.
- `inventory/databases_fleet.csv`: instancias SQL Server activas.
- `sampler/`: núcleo de cálculo y query rápida; daemon pendiente.
- `clickhouse/init/` y `clickhouse/queries/`: contrato inicial pendiente de integración.
- `grafana/dashboards/`: vista global y PI SQL Server.
- `prometheus/alerts/`: reglas de base de datos; la saturación AAS queda pendiente de una métrica real.
- `docs/MATRIZ_PARIDAD_FOGLIGHT_BD.md`: aceptación y evidencia.

## Flujo de trabajo

- Cambiar targets mediante el inventario y los scripts de sincronización, no editar a mano el YAML generado.
- Validar las consultas en la versión SQL Server soportada y medir su impacto antes de habilitarlas.
- Ejecutar `python -m unittest discover -s tests -v` para las pruebas offline.
- No desplegar ni reiniciar servicios sin validar archivos y estado del servidor.