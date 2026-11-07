# Operación del Foglight Local para SQL Server

## Alcance actual

La primera entrega monitorea los dos nodos SQL Server del clúster `SQL-CLUSTER-02`. Grafana, Prometheus y SQL Exporter forman el runtime base; ClickHouse y el sampler se activan con el perfil opcional `pi`.

## Crear un usuario local del portal sin SMTP

Para crear cuentas desde el portal, inicia sesion con una cuenta que tenga el permiso **Grafana server administrator** y abre `/admin/users/create`. El Executive Overview incluye los enlaces **Usuarios locales** (`/admin/users`) y **Crear usuario**.

La pantalla nativa permite definir login, nombre y contrasena de una cuenta local. No requiere cuenta externa ni envio de correo. No uses **Invite user** de la organizacion para este flujo. Ser **Admin** de una organizacion no concede automaticamente permisos de administrador del servidor.

Reserva el permiso de administrador del servidor para el responsable de cuentas. Asigna **Editor** a quienes editaran dashboards y **Viewer** a quienes solo consultaran. Estos roles no implementan todavia el alta de instancias de monitoreo desde el portal; esa capacidad sigue pendiente.

Los enlaces pueden verse en el dashboard, pero Grafana comprueba permisos al abrir la pagina y al ejecutar operaciones. No habilites acceso anonimo ni autorregistro para resolver un error de permisos.

Como alternativa por terminal, un DBA autorizado puede conectarse al servidor por MobaXterm como `pggr` y ejecutar:

```bash
cd /opt/centinela
python3 scripts/create-grafana-local-user.py
```

El script lee la cuenta administradora desde `.env`, solicita el nuevo login y contraseña sin mostrarlos, crea el usuario mediante la API local `127.0.0.1:3000` y no envía correo. Usa un email reservado `.invalid` solo porque Grafana lo requiere como campo; no es una invitación ni una cuenta externa. El rol inicial es el rol normal de la organización (Viewer); un administrador puede elevarlo si el DBA necesita editar dashboards.

No ejecutes este helper desde otra máquina ni publiques `.env`; el uso local evita transmitir la clave administradora por la red.

## Incorporar una instancia SQL Server

1. Agrega la instancia a `inventory/databases_fleet.csv` con su host, puerto, ambiente, clúster y base inicial.
2. No guardes contraseñas en archivos versionados. Usa el mecanismo de credenciales local aprobado para el entorno antes de generar la configuración.
3. Genera y revisa `exporters/sql-exporter/config.yml`:

   ```powershell
   powershell -ExecutionPolicy Bypass -File scripts/sync-fleet-inventory.ps1
   ```

4. Verifica la configuración generada y reinicia únicamente `centinela-sql-exporter` cuando se apruebe el cambio.

El inventario actual usa el puerto 5000 para `SQL-CLUSTER-02`; confirma el puerto real de cada instancia, no asumas 1433.

## Acceso de lectura

Un DBA debe revisar y aplicar los permisos mínimos requeridos por las consultas de `exporters/sql-exporter/config.yml` y por el futuro sampler. No otorgues `sysadmin`, `db_owner` ni permisos de escritura al usuario monitor. Valida permisos contra la versión exacta de SQL Server antes de habilitar una consulta nueva.

## Validación

- Confirma que el target de cada instancia esté `UP` en Prometheus.
- Compara Health Strip y Always On con consultas de referencia aprobadas por el DBA.
- No presentes la estimación legacy como AAS por sesiones muestreadas.
- Registra evidencia y limitaciones en `docs/MATRIZ_PARIDAD_FOGLIGHT_BD.md`.