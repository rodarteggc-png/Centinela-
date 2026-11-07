# PI Sampling and Storage Contract

This contract is the first implementation boundary for the SQL Server PI sampler. It does not claim that the sampler, ClickHouse service, or PI Engine is operational yet.

## AAS semantics

For a time bucket and category, AAS is the arithmetic mean of the number of active sessions observed in each successful sample:

`AAS(category) = sum(active sessions in category for each successful sample) / successful sample count`

`sample_runs` must contain one row for every attempt, including a successful sample that returned zero active sessions. A failed or timed-out sample is not zero load and must not enter the denominator. Report sample coverage alongside AAS. Keep sampling cadence regular; if future scheduling is irregular, use a time-weighted calculation instead of a plain sample mean.

Do not derive this AAS from `sys.dm_os_wait_stats`, `rate()` of cumulative counters, or CPU percentage. Detect logical and physical CPU counts per instance; do not hardcode a fleet-wide core limit.

## Storage contract

The initial DDL is `clickhouse/init/001_pi_schema.sql`:

- `sample_runs` records sample identity, outcome, duration, and returned row count so zero-session samples and collection gaps remain visible.
- `active_session_samples` stores one row per observed user request and sample. The ClickHouse order is instance, time, sample, and session; retention is 30 days.
- `statement_key` and `plan_key` are opaque sampler keys scoped by `instance_id`. They are not SQL text hashes and must never be exposed as Prometheus labels.
- `sql_text` and `execution_plans` are separate dimensions populated asynchronously. Full query text and plans may contain sensitive literals or identifiers; redact or parameterize them, restrict access, and review retention before production use.
- No materialized views are defined yet. First verify sample completeness, AAS, query shapes, and storage volume with controlled data.
- `clickhouse/queries/aas_by_minute.sql` returns the observation numerator, successful-sample denominator, and attempt coverage per category. The PI Engine must return AAS as unavailable when the successful-sample denominator is zero.
- `clickhouse/queries/aas_by_minute.sql` returns AAS, the observation numerator, successful-sample denominator, and attempt coverage per category. AAS is `NULL` when the successful-sample denominator is zero.
- No materialized views are defined yet. First verify sample completeness, AAS, query shapes, and storage volume with controlled data.

## Initial rollout boundary

Start with the two inventoried SQL Server nodes only. Keep `sql-exporter` for the existing low-cardinality Health Strip and Always On metrics. Add the sampler and ClickHouse as separate services after credentials, resource limits, `/data/centinela` persistence, health checks, and recovery behavior are defined. Do not enable Oracle or PostgreSQL sampling as part of this first slice.

The legacy Grafana workload estimate remains available but is explicitly labeled as not-AAS. The invalid workload-saturation alert is removed until the sampler exports a validated `centinela_aas_total` metric.

The SQL Server sampling principal needs the least privileges verified for the target version. Do not grant permissions automatically; have a DBA review and apply them. The documented latency and CPU figures are targets to measure, not guarantees.