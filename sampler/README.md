# SQL Server PI Sampler

The SQL Server PI sampler daemon reads the active SQL Server inventory, samples active requests, writes attempts and session rows to ClickHouse, and exposes AAS/coverage metrics for Prometheus. It is opt-in through the Compose `pi` profile and is not production-validated yet.
Run the unit tests from this directory with `go test ./...`. The portable deployment must provide `CLICKHOUSE_PASSWORD`; the sampler fails closed if it is missing.

## Current slice

- `internal/pi.CalculateAAS` averages per-category active-session counts over successful sample runs. Successful runs with no sessions count as zero; failed and timed-out runs are excluded. Duplicate sample IDs or duplicate sessions in a sample are rejected.
- `sqlserver/active_sessions.sql` is the fast-path query. It returns session and resource fields plus handles for asynchronous enrichment; it does not resolve SQL text or execution plans.
- The collector must persist every sample attempt in `sample_runs`, even when the query succeeds with zero rows. Only successful runs may contribute to the AAS denominator.

Run the unit tests from this directory with `go test ./...` when the Go toolchain is available.

Before enabling collection, validate this query against the supported SQL Server version and monitor principal, measure its execution cost, review wait classification with DBAs, and confirm the sample cadence and coverage under load.