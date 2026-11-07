-- Centinela — Top SQL por texto, por muestras de sesiones activas
-- Estado: PREPARADO, no validado. Requiere el perfil PI con datos reales
-- en centinela_pi.active_session_samples, con statement_key <> 0.
--
-- statement_key es un hash FNV-1a del sql_handle + offsets, calculado
-- en el sampler. No es el texto SQL. El texto se resuelve en
-- centinela_pi.sql_text, que se puebla de forma asíncrona con
-- sys.dm_exec_sql_text. Hasta que esa población exista, esta consulta
-- devuelve el Top por statement_key, sin texto, y el texto aparece
-- como '(sin texto aún)'.
-- Nunca exponer el texto SQL como etiqueta Prometheus.
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
    a.statement_key,
    count() AS session_observations,
    toFloat64(count()) / nullIf((SELECT n FROM total_successful), 0) AS aas,
    countDistinct(a.session_id) AS distinct_sessions,
    countDistinct(a.database_name) AS databases,
    anyLast(t.sql_text) AS sql_text,
    (SELECT n FROM total_successful) AS successful_sample_count
FROM successful_samples AS s
INNER JOIN centinela_pi.active_session_samples AS a
    ON a.sample_id = s.sample_id
    AND a.instance_id = s.instance_id
LEFT JOIN centinela_pi.sql_text AS t
    ON t.instance_id = a.instance_id
    AND t.statement_key = a.statement_key
WHERE a.statement_key <> 0
GROUP BY a.statement_key
ORDER BY session_observations DESC
LIMIT {limit:UInt32};
