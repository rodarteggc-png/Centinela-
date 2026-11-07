package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"math"
	"time"

	"centinela/sampler/internal/config"
	"centinela/sampler/internal/model"
	"centinela/sampler/sqlserver"

	_ "github.com/microsoft/go-mssqldb"
)

type Sampler struct {
	database *sql.DB
}

func Open(target config.Target) (*Sampler, error) {
	database, err := sql.Open("sqlserver", target.ConnectionString())
	if err != nil {
		return nil, fmt.Errorf("configure SQL Server connection")
	}
	database.SetMaxOpenConns(1)
	database.SetMaxIdleConns(1)
	database.SetConnMaxLifetime(5 * time.Minute)
	return &Sampler{database: database}, nil
}

func (sampler *Sampler) Close() error {
	return sampler.database.Close()
}

func (sampler *Sampler) DetectCapacity(ctx context.Context) (uint16, uint16, error) {
	var logicalCPUs, visibleSchedulers int64
	err := sampler.database.QueryRowContext(ctx, `
		SELECT
			cpu_count,
			(SELECT COUNT(*) FROM sys.dm_os_schedulers WHERE status = 'VISIBLE ONLINE')
		FROM sys.dm_os_sys_info
	`).Scan(&logicalCPUs, &visibleSchedulers)
	if err != nil {
		return 0, 0, err
	}
	if logicalCPUs < 1 || visibleSchedulers < 1 || logicalCPUs > math.MaxUint16 || visibleSchedulers > math.MaxUint16 {
		return 0, 0, fmt.Errorf("SQL Server returned invalid CPU topology")
	}
	return uint16(logicalCPUs), uint16(visibleSchedulers), nil
}

func (sampler *Sampler) Sample(ctx context.Context, instanceID, sampleID string) ([]model.ActiveSession, time.Time, error) {
	rows, err := sampler.database.QueryContext(ctx, sqlserver.ActiveSessionsQuery)
	if err != nil {
		return nil, time.Time{}, err
	}
	defer rows.Close()

	sessions := make([]model.ActiveSession, 0, 16)
	var sampleTime time.Time
	for rows.Next() {
		var observedAt time.Time
		var sessionID int64
		var databaseName, loginName, hostName, programName sql.NullString
		var command, requestStatus, waitType, waitCategory, waitResource sql.NullString
		var cpuTime, logicalReads, physicalWrites, duration, waitDuration, blockingSession int64
		if err := rows.Scan(
			&observedAt,
			&sessionID,
			&databaseName,
			&loginName,
			&hostName,
			&programName,
			&command,
			&requestStatus,
			&cpuTime,
			&logicalReads,
			&physicalWrites,
			&duration,
			&waitType,
			&waitDuration,
			&waitCategory,
			&blockingSession,
			&waitResource,
		); err != nil {
			return nil, time.Time{}, fmt.Errorf("scan SQL Server active-session row")
		}
		if sessionID < 0 || sessionID > math.MaxUint16 {
			return nil, time.Time{}, fmt.Errorf("SQL Server returned an out-of-range session id")
		}
		if sampleTime.IsZero() {
			sampleTime = observedAt.UTC()
		}
		sessions = append(sessions, model.ActiveSession{
			SampleID:          sampleID,
			InstanceID:        instanceID,
			SampleTime:        model.NewDateTime64(observedAt.UTC()),
			SessionID:         uint16(sessionID),
			DatabaseName:      nullableString(databaseName),
			LoginName:         nullableString(loginName),
			ProgramName:       nullableString(programName),
			HostName:          nullableString(hostName),
			Command:           nullableString(command),
			RequestStatus:     nullableString(requestStatus),
			WaitCategory:      nullableString(waitCategory),
			EffectiveWaitType: nullableString(waitType),
			WaitDurationMS:    nonNegativeUint64(waitDuration),
			CPUTimeMS:         nonNegativeUint32(cpuTime),
			LogicalReads:      nonNegativeUint64(logicalReads),
			PhysicalWrites:    nonNegativeUint32(physicalWrites),
			DurationMS:        nonNegativeUint32(duration),
			BlockingSessionID: int32(blockingSession),
			WaitResource:      nullableString(waitResource),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, time.Time{}, err
	}
	if sampleTime.IsZero() {
		sampleTime = time.Now().UTC()
	}
	return sessions, sampleTime, nil
}

func nullableString(value sql.NullString) string {
	if !value.Valid {
		return ""
	}
	return value.String
}

func nonNegativeUint32(value int64) uint32 {
	if value <= 0 {
		return 0
	}
	if uint64(value) > math.MaxUint32 {
		return math.MaxUint32
	}
	return uint32(value)
}

func nonNegativeUint64(value int64) uint64 {
	if value <= 0 {
		return 0
	}
	return uint64(value)
}