# Centinela: instrucciones de desarrollo

## Alcance vigente

Construir un Foglight local para el equipo de DBA, con Tavo como administrador, para vigilar todas las bases de la flota (420: SQL Server 300 servidores, PostgreSQL 50, Oracle 30, Azure 30 bases, MySQL 10). Fase 1: SQL Server, inicialmente limitado a los dos nodos inventariados de `SQL-CLUSTER-02` y al clúster 10 ya arriba. El runtime base es Grafana, Prometheus y SQL Exporter; sampler y ClickHouse son opt-in con perfil `pi` y aún pendientes de validación operacional.

Fases por motor: 1 SQL Server, 2 PostgreSQL, 3 Oracle (RAC), 4 Azure, 5 MySQL. Ningún motor se publica sin conexión real, métricas reales y validación por DBA. No añadir APM ni demo-app. No afirmar paridad completa con Foglight antes de validar criterios en cada motor.

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