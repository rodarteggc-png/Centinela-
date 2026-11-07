# Paridad con Foglight — en porcentaje

Regla acordada con Tavo, 2026-11-07: después de cada cambio, decir en qué porcentaje estamos y cuánto avanzó ese cambio.

## Cómo se calcula (para que no sea un número inventado)

- El 100% es: todos los motores, todas las funciones Foglight, validados en operación, para el equipo de DBA.
- Peso por motor, según la flota de 420: SQL Server 71%, PostgreSQL 12%, Oracle 7%, Azure 7%, MySQL 3%.
- Por motor, 8 funciones, cada una vale lo mismo dentro de ese motor:
  Inventario, Health Strip, Alta disponibilidad, AAS / PI, Top SQL, Bloqueos, Alertas, Resumen ejecutivo + Storage/Advisory.
- Puntaje por función: 0 = no está, 0.5 = está preparado en código pero sin validar, 1 = validado en operación con evidencia.
- El porcentaje total es la suma ponderada por motor.

## Estado actual — 2026-11-07, después del push be4fa8a

- SQL Server: ~25% de su parte. Inventario y dashboards están preparados, AAS está preparado pero no validado, Top SQL es acumulado no por muestras, bloqueos pendiente, alertas sin probar.
- PostgreSQL: 0%
- Oracle: 0%
- Azure: 0%
- MySQL: 0%

Total: **~18%**

Cálculo: 71% x 25% = ~18%, resto 0%.

## Qué cambió con el push be4fa8a

- Quitar la alerta de saturación y relabelar Resource Time: **+0 puntos**.
  No es una función nueva, es corrección. Evita que el 18% se lea como si el AAS ya fuera real. El porcentaje se mantiene honesto.

## Qué falta para subir, en orden

- Validar el AAS en servidor con carga controlada: SQL Server sube a ~35%, total a ~25% (+7 puntos).
- Top SQL por muestras: +4 puntos totales.
- Bloqueos con bloqueador raíz y víctimas: +3 puntos totales.
- Alertas probadas en disparo y recuperación: +3 puntos totales.
- SQL Server completo y validado: total a ~71%.
- PostgreSQL completo: total a ~83%.
- Oracle completo: total a ~90%.
- Azure completo: total a ~97%.
- MySQL completo: total a 100%.

Actualizar este archivo después de cada cambio, con la fecha y el motivo del cambio de porcentaje.

## Actualización — 2026-11-07, batch antes del push

Cambios preparados en local, aún sin subir:
- Top SQL por muestras: consulta preparada, pero Top por texto SQL sigue pendiente por falta de sql_handle en el sampler. Top SQL pasa de 0 a 0.5 en SQL Server.
- Bloqueos: consulta preparada por bloqueador raíz y víctimas. Bloqueos pasa de 0 a 0.5 en SQL Server.
- Sampler resiliente: un target malo ya no tumba todo el sampler. Es robustez, no una función Foglight nueva, no suma puntos directos, pero desbloquea escalar a 300.

Total: **~20%** (antes 18%, +2 puntos).

El AAS sigue sin validar en servidor, así que ese +7 puntos sigue pendiente.

## Actualización — 2026-11-07, Top SQL por texto

- Sampler ya captura sql_handle y guarda statement_key en ClickHouse.
- Consulta Top SQL por texto preparada en clickhouse/queries/top_sql_text_by_samples.sql.
- Falta poblar sql_text de forma asíncrona y validar en servidor.
- Top SQL pasa de 0.5 a 0.75 en SQL Server.

Total: **~22%** (antes 20%, +2 puntos).

## Actualización — 2026-11-07, paneles Grafana PI

- Datasource ClickHouse provisionado en Grafana, plugin instalado vía docker-compose.pi.yml.
- Paneles nuevos en el dashboard PI (Linux, no portable): Top SQL por muestras y Bloqueos por minuto.
- Los paneles aún no se han visto con datos reales; requieren perfil pi en el servidor.

Total: **~23%** (antes 22%, +1 punto).
