-- Centinela — Top por muestras de sesiones activas (no por caché)
-- Estado: PREPARADO, no validado. Requiere el perfil PI con datos reales
-- en centinela_pi.active_session_samples.
--
-- Limitación conocida: el sampler actual NO guarda sql_handle ni
-- statement_key por muestra, así que este Top es por dimensión de la
-- muestra (base, programa, login, categoría de espera), no por texto SQL.
-- El Top SQL por statement requiere enriquecer el sampler con sql_handle
-- y poblar centinela_pi.sql_text de forma asíncrona, según
-- docs/PI_STORAGE_CONTRACT.md. No usar el panel actual de Top SQL
-- (CPU acumulado en caché) como si fuera este Top por muestras.
--
-- Uso: reemplazar {instance_id:String}, {from:DateTime64(3)}, {to:DateTime64(3)}
-- y {limit:UInt32} con los parámetros del periodo a investigar.
WITH
    toDateTime64({from:DateTime64(3)}, 3, 'UTC') AS from_time,
    toDateTime64({to:DateTime64(3)}, 3, 'UTC') AS to_time,
    {instance_id:String} AS requested_instance,
    successful_samples AS (
        SELECT sample_id, instance_id
        FROM centinela_pi.sample_runs
        WHERE instance_id = requested_instance
            AND sample_status = 'success'
            AND sample_time >= from_time
            AND sample_time < to_time
    ),
    total_successful AS (
        SELECT count() AS n FROM successful_samples
    )
SELECT
    a.database_name,
    a.program_name,
    a.login_name,
    toString(a.wait_category) AS wait_category,
    count() AS session_observations,
    toFloat64(count()) / nullIf((SELECT n FROM total_successful), 0) AS aas,
    (SELECT n FROM total_successful) AS successful_sample_count
FROM successful_samples AS s
INNER JOIN centinela_pi.active_session_samples AS a
    ON a.sample_id = s.sample_id
    AND a.instance_id = s.instance_id
GROUP BY a.database_name, a.program_name, a.login_name, a.wait_category
ORDER BY session_observations DESC
LIMIT {limit:UInt32};
