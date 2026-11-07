WITH
    toDateTime64({from:DateTime64(3)}, 3, 'UTC') AS from_time,
    toDateTime64({to:DateTime64(3)}, 3, 'UTC') AS to_time,
    {instance_id:String} AS requested_instance,
    categories AS
    (
        SELECT arrayJoin(['CPU', 'IO', 'Lock', 'Memory', 'Network', 'Other']) AS wait_category
    ),
    attempts AS
    (
        SELECT
            toStartOfInterval(sample_time, INTERVAL 1 MINUTE) AS minute,
            count() AS attempt_count,
            countIf(sample_status = 'success') AS successful_attempt_count
        FROM centinela_pi.sample_runs
        WHERE instance_id = requested_instance
            AND sample_time >= from_time
            AND sample_time < to_time
        GROUP BY minute
    ),
    successful_samples AS
    (
        SELECT
            sample_id,
            instance_id,
            toStartOfInterval(sample_time, INTERVAL 1 MINUTE) AS minute
        FROM centinela_pi.sample_runs
        WHERE instance_id = requested_instance
            AND sample_status = 'success'
            AND sample_time >= from_time
            AND sample_time < to_time
        GROUP BY sample_id, instance_id, minute
    ),
    category_observations AS
    (
        SELECT
            s.minute,
            toString(a.wait_category) AS wait_category,
            count() AS session_observations
        FROM successful_samples AS s
        INNER JOIN centinela_pi.active_session_samples AS a
            ON a.sample_id = s.sample_id
            AND a.instance_id = s.instance_id
        GROUP BY s.minute, a.wait_category
    )
SELECT
    attempts.minute,
    categories.wait_category,
    ifNull(category_observations.session_observations, toUInt64(0)) AS session_observations,
    attempts.successful_attempt_count AS successful_sample_count,
    toFloat64(ifNull(category_observations.session_observations, toUInt64(0)))
        / nullIf(toFloat64(attempts.successful_attempt_count), 0) AS aas,
    attempts.attempt_count,
    attempts.successful_attempt_count,
    toFloat64(attempts.successful_attempt_count) / attempts.attempt_count AS attempt_success_ratio
FROM attempts
CROSS JOIN categories
LEFT JOIN category_observations
    ON category_observations.minute = attempts.minute
    AND category_observations.wait_category = categories.wait_category
ORDER BY attempts.minute, categories.wait_category;