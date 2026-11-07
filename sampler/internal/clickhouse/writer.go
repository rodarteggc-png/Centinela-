package clickhouse

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"centinela/sampler/internal/model"
)

type Writer struct {
	endpoint *url.URL
	username string
	password string
	client   *http.Client
}

func NewWriter(endpoint, username, password string) (*Writer, error) {
	if username == "" || password == "" {
		return nil, fmt.Errorf("ClickHouse username and password are required")
	}
	parsed, err := url.Parse(endpoint)
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return nil, fmt.Errorf("invalid ClickHouse endpoint")
	}
	return &Writer{
		endpoint: parsed,
		username: username,
		password: password,
		client:   &http.Client{Timeout: 5 * time.Second},
	}, nil
}

func (writer *Writer) InsertInstances(ctx context.Context, rows []model.Instance) error {
	return insertRows(writer, ctx, "instances", rows)
}

func (writer *Writer) InsertRuns(ctx context.Context, rows []model.SampleRun) error {
	return insertRows(writer, ctx, "sample_runs", rows)
}

func (writer *Writer) InsertSessions(ctx context.Context, rows []model.ActiveSession) error {
	return insertRows(writer, ctx, "active_session_samples", rows)
}

func insertRows[T any](writer *Writer, ctx context.Context, table string, rows []T) error {
	var body bytes.Buffer
	encoder := json.NewEncoder(&body)
	encoder.SetEscapeHTML(true)
	for _, row := range rows {
		if err := encoder.Encode(row); err != nil {
			return fmt.Errorf("encode ClickHouse row")
		}
	}
	if body.Len() == 0 {
		return nil
	}

	requestURL := *writer.endpoint
	query := requestURL.Query()
	query.Set("query", "INSERT INTO centinela_pi."+table+" FORMAT JSONEachRow")
	requestURL.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL.String(), &body)
	if err != nil {
		return fmt.Errorf("build ClickHouse request: %w", err)
	}
	request.SetBasicAuth(writer.username, writer.password)
	request.Header.Set("Content-Type", "application/json")

	response, err := writer.client.Do(request)
	if err != nil {
		return fmt.Errorf("ClickHouse request failed")
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		message, readErr := io.ReadAll(io.LimitReader(response.Body, 4096))
		if readErr != nil {
			return fmt.Errorf("ClickHouse returned HTTP %d", response.StatusCode)
		}
		diagnostic := strings.SplitN(string(message), "(query:", 2)[0]
		diagnostic = strings.Join(strings.Fields(diagnostic), " ")
		if len(diagnostic) > 240 {
			diagnostic = diagnostic[:240]
		}
		return fmt.Errorf("ClickHouse returned HTTP %d: %s", response.StatusCode, diagnostic)
	}
	return nil
}