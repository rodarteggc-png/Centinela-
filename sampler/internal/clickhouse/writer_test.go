package clickhouse

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"centinela/sampler/internal/model"
)

func TestInsertRunsSendsAuthenticatedJSONEachRow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		username, password, ok := request.BasicAuth()
		if !ok || username != "pi_user" || password != "test-secret" {
			t.Error("ClickHouse authentication was not sent correctly")
		}
		if request.URL.Query().Get("query") != "INSERT INTO centinela_pi.sample_runs FORMAT JSONEachRow" {
			t.Errorf("unexpected insert query: %s", request.URL.Query().Get("query"))
		}
		var run struct {
			SampleID     string `json:"sample_id"`
			InstanceID   string `json:"instance_id"`
			SampleTime   string `json:"sample_time"`
			SampleStatus string `json:"sample_status"`
		}
		if err := json.NewDecoder(request.Body).Decode(&run); err != nil {
			t.Errorf("decode JSONEachRow body: %v", err)
		}
		if run.SampleID != "sample-id" || run.InstanceID != "node-1" || run.SampleStatus != "success" {
			t.Errorf("unexpected sample run payload: %+v", run)
		}
		if run.SampleTime != "2026-11-04 12:00:00.000" {
			t.Errorf("ClickHouse timestamp = %q, want DateTime64 format", run.SampleTime)
		}
		writer.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	writer, err := NewWriter(server.URL, "pi_user", "test-secret")
	if err != nil {
		t.Fatalf("NewWriter returned an error: %v", err)
	}
	err = writer.InsertRuns(context.Background(), []model.SampleRun{{
		SampleID:             "sample-id",
		InstanceID:           "node-1",
		SampleTime:           model.NewDateTime64(time.Date(2026, 11, 4, 12, 0, 0, 0, time.UTC)),
		SampleStatus:         "success",
		CollectionDurationMS: 12,
		ObservedSessionCount: 0,
	}})
	if err != nil {
		t.Fatalf("InsertRuns returned an error: %v", err)
	}
}

func TestInsertRunsRejectsClickHouseErrorsWithoutReturningBody(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusBadRequest)
		_, _ = io.WriteString(writer, "Code: 62. DB::Exception: sample schema mismatch (query: INSERT secret-payload)")
	}))
	defer server.Close()

	client, err := NewWriter(server.URL, "pi_user", "test-secret")
	if err != nil {
		t.Fatalf("NewWriter returned an error: %v", err)
	}
	err = client.InsertRuns(context.Background(), []model.SampleRun{{SampleID: "sample-id"}})
	if err == nil {
		t.Fatal("expected ClickHouse HTTP error")
	}
	if !strings.Contains(err.Error(), "sample schema mismatch") || strings.Contains(err.Error(), "secret-payload") {
		t.Fatal("error should preserve the safe diagnostic and omit the query payload")
	}
}
