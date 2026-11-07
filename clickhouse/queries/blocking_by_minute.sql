-- Centinela — Bloqueos por muestras: bloqueador raíz y víctimas
-- Estado: PREPARADO, no validado. Requiere el perfil PI con datos reales
-- en centinela_pi.active_session_samples.
--
-- El sampler guarda blocking_session_id por muestra. Un bloqueador raíz
-- es una sesión con blocking_session_id = 0 que bloquea a otras.
-- Una víctima es una sesión con blocking_session_id <> 0.
-- Incluye el caso de sesión idle con transacción abierta: esa sesión
-- aparece como bloqueador aunque no tenga una request activa en ese
-- momento, por eso hay que revisar también las muestras donde el
-- bloqueador no está en active_session_samples en ese minuto.
--
-- Uso: reemplazar {instance_id:String}, {from:DateTime64(3)}, {to:DateTime64(3)}
-- con los parámetros del periodo a investigar.
WITH
    toDateTime64({from:DateTime64(3)}, 3, 'UTC') AS from_time,
    toDateTime64({to:DateTime64(3)}, 3, 'UTC') AS to_time,
    {instance_id:String} AS requested_instance,
    successful_samples AS (
        SELECT sample_id, instance_id, toStartOfInterval(sample_time, INTERVAL 1 MINUTE) AS minute
        FROM centinela_pi.sample_runs
        WHERE instance_id = requested_instance
            AND sample_status = 'success'
            AND sample_time >= from_time
            AND sample_time < to_time
    )
SELECT
    s.minute,
    a.blocking_session_id AS blocker_session_id,
    count() AS victim_observations,
    countDistinct(a.session_id) AS victim_sessions,
    countDistinct(a.sample_id) AS samples_with_blocking,
    groupUniqArray(a.database_name) AS databases,
    groupUniqArray(a.wait_resource) AS wait_resources
FROM successful_samples AS s
INNER JOIN centinela_pi.active_session_samples AS a
    ON a.sample_id = s.sample_id
    AND a.instance_id = s.instance_id
WHERE a.blocking_session_id <> 0
GROUP BY s.minute, a.blocking_session_id
ORDER BY s.minute, victim_observations DESC;
