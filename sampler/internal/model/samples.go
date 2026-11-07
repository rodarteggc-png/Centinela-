package model

import (
	"encoding/json"
	"time"
)

type DateTime64 time.Time

func NewDateTime64(value time.Time) DateTime64 {
	return DateTime64(value.UTC())
}

func (value DateTime64) MarshalJSON() ([]byte, error) {
	formatted := time.Time(value).UTC().Format("2006-01-02 15:04:05.000")
	return json.Marshal(formatted)
}

type Instance struct {
	InstanceID   string    `json:"instance_id"`
	InstanceName string    `json:"instance_name"`
	Host         string    `json:"host"`
	Port         uint16    `json:"port"`
	Engine       string    `json:"engine"`
	ClusterName  string    `json:"cluster_name"`
	Environment  string    `json:"environment"`
	LogicalCPUs       *uint16   `json:"logical_cpus"`
	VisibleSchedulers *uint16   `json:"visible_schedulers"`
	UpdatedAt    DateTime64 `json:"updated_at"`
}

type SampleRun struct {
	SampleID              string    `json:"sample_id"`
	InstanceID            string    `json:"instance_id"`
	SampleTime            DateTime64 `json:"sample_time"`
	SampleStatus          string    `json:"sample_status"`
	CollectionDurationMS  uint32    `json:"collection_duration_ms"`
	ObservedSessionCount  uint32    `json:"observed_session_count"`
	ErrorCode             uint16    `json:"error_code"`
}

type ActiveSession struct {
	SampleID          string    `json:"sample_id"`
	InstanceID        string    `json:"instance_id"`
	SampleTime        DateTime64 `json:"sample_time"`
	SessionID         uint16    `json:"session_id"`
	DatabaseName      string    `json:"database_name"`
	LoginName         string    `json:"login_name"`
	ProgramName       string    `json:"program_name"`
	HostName          string    `json:"host_name"`
	Command           string    `json:"command"`
	RequestStatus     string    `json:"request_status"`
	WaitCategory      string    `json:"wait_category"`
	EffectiveWaitType string    `json:"effective_wait_type"`
	WaitDurationMS    uint64    `json:"wait_duration_ms"`
	CPUTimeMS         uint32    `json:"cpu_time_ms"`
	LogicalReads      uint64    `json:"logical_reads"`
	PhysicalWrites    uint32    `json:"physical_writes"`
	DurationMS        uint32    `json:"duration_ms"`
	BlockingSessionID int32     `json:"blocking_session_id"`
	WaitResource      string    `json:"wait_resource"`
	StatementKey      uint64    `json:"statement_key"`
	PlanKey           uint64    `json:"plan_key"`
}