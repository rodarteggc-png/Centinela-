# Ficha de infraestructura: Foglight local SQL Server

## Alcance

Primera entrega on-premise para dos nodos SQL Server de `SQL-CLUSTER-02`. El runtime actual es Grafana, Prometheus y SQL Exporter. El sampler PI y ClickHouse siguen pendientes de integración y dimensionamiento con mediciones reales.

## Plataforma

- Rocky Linux 9.4 o superior, x86_64.
- Docker Engine y Docker Compose Plugin.
- Recomendación inicial: 4 vCPU y 8-16 GB RAM. Ajustar tras medir Prometheus y el futuro PI de muestras.
- Disco de sistema separado del disco de datos.
- Persistencia en `/data/centinela`; no almacenar telemetría en `/`.

## Red

| Dirección | Puerto | Uso |
| --- | --- | --- |
| Entrante | TCP 3000 | Consola Grafana; limitar a la red DBA. |
| Entrante | TCP 9090 | UI/API Prometheus; limitar a administradores. |
| Entrante | TCP 9399 | SQL Exporter; preferir acceso interno/restringido. |
| Entrante | TCP 22 | Administración SSH. |
| Saliente | TCP 5000 | SQL Server `SQL-CLUSTER-02`; confirmar puerto por instancia. |

No abrir puertos OTLP, Tempo, Loki, Oracle o PostgreSQL para esta entrega. ClickHouse deberá permanecer en la red privada de contenedores cuando se integre; no publicar su puerto sin un requisito y controles explícitos.

## Almacenamiento y acceso

Mantener Prometheus y Grafana bajo `CENTINELA_DATA_DIR=/data/centinela`. Antes de agregar ClickHouse, revisar consumo esperado, límites, retención y recuperación. El usuario SQL monitor debe tener solo los permisos de lectura aprobados por un DBA para la versión exacta del motor.