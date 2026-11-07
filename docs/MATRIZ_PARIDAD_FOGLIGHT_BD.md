# Matriz de paridad: SQL Server PI

La primera entrega se limita a SQL Server local. El inventario contiene dos nodos del clúster `SQL-CLUSTER-02`. La presencia de consultas o paneles no equivale a validación operativa.

| Capacidad | Fuente | Estado actual | Criterio de aceptación |
| --- | --- | --- | --- |
| Inventario y cobertura | CSV, renderer y SQL Exporter | Dos nodos configurados; generación offline probada. | Targets únicos, `UP`, identidad estable y secretos fuera de archivos versionados. |
| Health Strip | PLE, Buffer Cache, throughput | Consultas presentes; semántica y umbrales pendientes de prueba controlada. | Unidades y resets comparados con valores del motor. |
| Always On | Rol, conectividad y salud de sincronización | Consultas presentes; estados de ambos nodos pendientes de prueba registrada. | Detectar roles y fallos sin alertas falsas. |
| AAS por sesiones activas | `sampler/internal/pi`, query SQL Server y esquema ClickHouse | Núcleo, query y contrato de almacenamiento preparados; faltan daemon, conexiones, ClickHouse en Compose y datos reales. El panel existente es una estimación, no AAS. | Registrar cada intento y cada muestra exitosa vacía; excluir fallos del denominador; comparar categorías, cobertura y AAS con una carga controlada. |
| AAS por sesiones activas | Daemon `sampler/`, ClickHouse y Prometheus | Prototipo ejecutable en perfil Compose `pi`; integración con SQL Server y ClickHouse aún no validada en servidor. | Registrar cada intento y cada muestra exitosa vacía; excluir fallos del denominador; comparar categorías, cobertura y AAS con una carga controlada. |
La regla de saturación usa AAS muestreado y cores detectados; falta validar disparo/recuperación en SQL Server.
| Top SQL | Muestras y dimensión de texto | Pendiente; no presentar los contadores acumulados de caché como impacto del intervalo actual. | Clasificar por apariciones en muestras del periodo; revisar redacción y acceso al texto SQL. |
| Bloqueos | Muestreo de bloqueadores | Pendiente. | Detectar bloqueador raíz y víctimas en una prueba reproducible, incluyendo sesión idle con transacción abierta. |
| Alertas | Reglas Prometheus SQL Server | Reglas de HADR, deadlocks, PLE y métricas SQL presentes; pruebas de disparo/recuperación pendientes. La alerta de saturación AAS está retirada hasta tener una métrica real. | Prueba positiva y negativa por regla, con duración, severidad y labels útiles. |
| Top SQL histórico / advisory | DMVs y configuración SQL Exporter | Parcial; acumulados y recomendaciones no sustituyen análisis por muestras. | Validar unidades, deduplicación, ventana y revisión DBA antes de aplicar recomendaciones. |

## Secuencia de aceptación

1. Completar sampler SQL Server y persistencia de muestras exitosas, vacías y fallidas.
2. Validar AAS y cobertura contra carga controlada antes de cambiar dashboards o alertas.
3. Probar permisos mínimos, impacto de consultas, reinicios, recuperación y retención; registrar evidencia aquí.
