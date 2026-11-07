# Centinela — Fases por motor

Definición acordada con Tavo, 2026-11-06.
Meta final: una copia de Foglight para uso laboral del equipo de DBA, con Tavo como administrador, para vigilar **todas** las bases, no solo las que cubren las licencias de Foglight.

## Flota a cubrir: 420

- SQL Server: 300 servidores, incluyendo clústeres con alta disponibilidad (Always On)
- PostgreSQL: 50
- Oracle: 30, algunas RAC
- Azure: 30 bases de datos (no servidores)
- MySQL: 10

Hoy Centinela solo inventaría 2 nodos de SQL Server (SQL-CLUSTER-02): 2 de 420.

## Regla general

Ningún motor se publica en el portal hasta tener conexión real, métricas reales y validación por DBA. Sin dashboards simulados. Es la misma regla que ya tiene el README para Oracle, PostgreSQL y Azure, aplicada a todos los motores.

Cada fase repite el mismo patrón de funciones Foglight, adaptado al motor:
PI / AAS por muestras, Top SQL, bloqueos, salud del motor, alta disponibilidad / replicación, almacenamiento, alertas y resumen ejecutivo.

## Fase 0 — Base y definición (en curso)

- Definición de "copia de Foglight para Tavo" guardada.
- Inventario real de la flota: 420 unidades.
- Diagnóstico de dónde está parado Centinela, pendiente de retomar con esta flota en mente.

Criterio de salida: definición y flota acordadas, y orden de fases confirmado por Tavo.

## Fase 1 — SQL Server: 300 servidores

Es el 71% de la flota y es el único motor con código ya preparado. Por eso va primero.

Alcance:
- Inventario a escala: 300 servidores, con identidad estable por clúster, no por nodo suelto.
- Always On: roles primario/réplica, salud de sincronización y detección de failover sin alertas falsas.
- PI validado: sampler, ClickHouse, AAS por muestras reales, con fallos excluidos del denominador y muestras vacías registradas.
- Top SQL por muestras del periodo, no por contadores acumulados de caché.
- Bloqueos: bloqueador raíz y víctimas, incluyendo sesión idle con transacción abierta.
- Alertas Prometheus probadas en disparo y recuperación: HADR, deadlocks, PLE y métricas del motor.
- Resumen ejecutivo para el equipo de DBA.

Criterio de salida: la matriz de paridad de SQL Server pasa de "preparado" a "validado en operación", con evidencia registrada en la propia matriz, empezando por un piloto con pocos servidores y luego el resto de los 300.

Riesgo principal: la escala. Lo que funciona para 2 nodos no garantiza que funcione para 300 sin revisar retención, frecuencia de muestreo, permisos mínimos e impacto de las consultas.

## Fase 2 — PostgreSQL: 50

Segundo motor por tamaño. Sirve para probar que la arquitectura multi-motor funciona de verdad, no solo para SQL Server.

Alcance:
- Exporter y sampler propios de PostgreSQL: sesiones activas, wait events, Top SQL por muestras, bloqueos, replicación y almacenamiento.
- Mismo patrón de PI que en SQL Server, adaptado a PostgreSQL.
- Alertas y resumen ejecutivo por motor.

Criterio de salida: mismo nivel de validación que Fase 1, pero para PostgreSQL, con carga controlada y evidencia.

## Fase 3 — Oracle: 30, algunas RAC

Motor más complejo por RAC, aunque son solo 30.

Alcance:
- Monitoreo por instancia/nodo dentro de cada RAC, no solo por base.
- Equivalente a PI: sesiones, esperas, Top SQL, bloqueos y Data Guard donde aplique.
- Alertas y resumen ejecutivo por motor.

Criterio de salida: validación por DBA en al menos un RAC real antes de publicar el resto.

## Fase 4 — Azure: 30 bases de datos

Es PaaS, no servidores. Se trata aparte a propósito.

Alcance:
- 30 bases de datos Azure SQL, con métricas por base, sin monitoreo de SO.
- Conexión y autenticación propias de Azure, con secretos fuera del repositorio.
- PI adaptado a Azure: sesiones, esperas, Top SQL y alertas por base.

Criterio de salida: validación por DBA en unas pocas bases piloto antes de publicar las 30.

## Fase 5 — MySQL: 10

El motor más pequeño. Va al final por tamaño, no por dificultad.

Alcance:
- Exporter y sampler propios de MySQL: sesiones, Top SQL, bloqueos, replicación y almacenamiento.
- Alertas y resumen ejecutivo por motor.

Criterio de salida: mismo nivel de validación que las fases anteriores.

Nota: si Tavo quiere una victoria rápida multi-motor antes de Oracle/Azure, MySQL puede adelantarse como prueba del patrón, sin cambiar el resto del orden.

## Temas transversales (todas las fases)

- Inventario único y secretos fuera de archivos versionados.
- Permisos mínimos por motor, medidos antes de escalar.
- Retención y capacidad de Prometheus y ClickHouse para 420 unidades.
- Roles: solo equipo de DBA, Tavo como administrador.
- Sin publicar un motor hasta validarlo, aunque el código ya exista.

## Orden propuesto

0. Base y definición
1. SQL Server (300)
2. PostgreSQL (50)
3. Oracle (30, RAC)
4. Azure (30 bases)
5. MySQL (10)

Pendiente de confirmar por Tavo.
