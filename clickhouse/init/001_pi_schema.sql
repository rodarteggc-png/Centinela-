CREATE DATABASE IF NOT EXISTS centinela_pi;

CREATE TABLE IF NOT EXISTS centinela_pi.instances
(
    instance_id String,
    instance_name LowCardinality(String),
    host String,
    port UInt16,
    engine Enum8('sqlserver' = 1, 'oracle' = 2, 'postgres' = 3),
    cluster_name LowCardinality(String),
    environment LowCardinality(String),
    logical_cpus Nullable(UInt16),
    visible_schedulers Nullable(UInt16),
    updated_at DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree(updated_at)
ORDER BY instance_id;

CREATE TABLE IF NOT EXISTS centinela_pi.sample_runs
(
    sample_id UUID,
    instance_id String,
    sample_time DateTime64(3, 'UTC'),
    sample_status Enum8('success' = 1, 'timeout' = 2, 'error' = 3),
    collection_duration_ms UInt32,
    observed_session_count UInt32,
    error_code UInt16 DEFAULT 0
)
ENGINE = MergeTree
PARTITION BY toYYYYMMDD(sample_time)
ORDER BY (instance_id, sample_time, sample_id)
TTL toDateTime(sample_time) + INTERVAL 30 DAY;

CREATE TABLE IF NOT EXISTS centinela_pi.active_session_samples
(
    sample_id UUID,
    instance_id String,
    sample_time DateTime64(3, 'UTC'),
    session_id UInt16,
    database_name LowCardinality(String),
    login_name LowCardinality(String),
    program_name LowCardinality(String),
    host_name LowCardinality(String),
    command LowCardinality(String),
    request_status LowCardinality(String),
    wait_category Enum8(
        'CPU' = 1,
        'IO' = 2,
        'Lock' = 3,
        'Memory' = 4,
        'Network' = 5,
        'Other' = 6
    ),
    effective_wait_type LowCardinality(String),
    wait_duration_ms UInt64,
    cpu_time_ms UInt32,
    logical_reads UInt64,
    physical_writes UInt32,
    duration_ms UInt32,
    blocking_session_id Int32 DEFAULT 0,
    wait_resource String DEFAULT '',
    statement_key UInt64 DEFAULT 0,
    plan_key UInt64 DEFAULT 0
)
ENGINE = MergeTree
PARTITION BY toYYYYMMDD(sample_time)
ORDER BY (instance_id, sample_time, sample_id, session_id)
TTL toDateTime(sample_time) + INTERVAL 30 DAY;

CREATE TABLE IF NOT EXISTS centinela_pi.sql_text
(
    instance_id String,
    statement_key UInt64,
    sql_text String CODEC(ZSTD(3)),
    first_seen DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree(first_seen)
ORDER BY (instance_id, statement_key)
TTL toDateTime(first_seen) + INTERVAL 30 DAY;

CREATE TABLE IF NOT EXISTS centinela_pi.execution_plans
(
    instance_id String,
    plan_key UInt64,
    plan_xml String CODEC(ZSTD(3)),
    first_seen DateTime64(3, 'UTC')
)
ENGINE = ReplacingMergeTree(first_seen)
ORDER BY (instance_id, plan_key)
TTL toDateTime(first_seen) + INTERVAL 30 DAY;