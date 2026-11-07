package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"centinela/sampler/internal/clickhouse"
	"centinela/sampler/internal/config"
	"centinela/sampler/internal/metrics"
	"centinela/sampler/internal/model"
	"centinela/sampler/internal/pi"
	"centinela/sampler/internal/sqlserver"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	clickHousePassword := os.Getenv("CLICKHOUSE_PASSWORD")
	if clickHousePassword == "" {
		return fmt.Errorf("CLICKHOUSE_PASSWORD must be configured before starting the PI sampler")
	}
	clickHouseUser := envOrDefault("CLICKHOUSE_USER", "centinela_pi")
	clickHouseURL := envOrDefault("CLICKHOUSE_URL", "http://clickhouse:8123")
	inventoryPath := envOrDefault("INVENTORY_FILE", "/etc/centinela/databases_fleet.csv")
	metricsAddress := envOrDefault("METRICS_ADDR", ":9500")
	interval, err := durationOrDefault("SAMPLE_INTERVAL", 3*time.Second)
	if err != nil || interval < time.Second {
		return fmt.Errorf("SAMPLE_INTERVAL must be at least one second")
	}
	queryTimeout, err := durationOrDefault("SQL_QUERY_TIMEOUT", 2*time.Second)
	if err != nil || queryTimeout <= 0 {
		return fmt.Errorf("SQL_QUERY_TIMEOUT must be a positive duration")
	}

	inventoryFile, err := os.Open(inventoryPath)
	if err != nil {
		return fmt.Errorf("open SQL Server inventory")
	}
	targets, loadErr := config.LoadSQLServerTargets(inventoryFile)
	closeErr := inventoryFile.Close()
	if loadErr != nil {
		return loadErr
	}
	if closeErr != nil {
		return fmt.Errorf("close SQL Server inventory")
	}

	writer, err := clickhouse.NewWriter(clickHouseURL, clickHouseUser, clickHousePassword)
	if err != nil {
		return err
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	registry := metrics.NewRegistry(time.Minute)
	server := &http.Server{Addr: metricsAddress, Handler: metricsHandler(registry), ReadHeaderTimeout: 5 * time.Second}
	go func() {
		if serveErr := server.ListenAndServe(); serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
			log.Printf("metrics endpoint stopped")
			stop()
		}
	}()

	instances := make([]model.Instance, 0, len(targets))
	samplers := make([]*sqlserver.Sampler, 0, len(targets))
	activeTargets := make([]config.Target, 0, len(targets))
	for _, target := range targets {
		sampler, openErr := sqlserver.Open(target)
		if openErr != nil {
			// Un target malo no debe tumbar todo el sampler: se registra
			// para que su cobertura se vea en 0 y se sigue con los demás.
			log.Printf("SQL Server target unavailable instance=%s, skipping", target.Name)
			registry.Register(target.Name, target.Cluster, nil)
			continue
		}
		samplers = append(samplers, sampler)
		activeTargets = append(activeTargets, target)

		var logicalCPUs, visibleSchedulers *uint16
		coreCtx, cancel := context.WithTimeout(ctx, queryTimeout)
		logical, visible, coreErr := sampler.DetectCapacity(coreCtx)
		cancel()
		if coreErr == nil {
			logicalCPUs = &logical
			visibleSchedulers = &visible
		} else {
			log.Printf("SQL scheduler capacity unavailable instance=%s", target.Name)
		}
		registry.Register(target.Name, target.Cluster, visibleSchedulers)
		instances = append(instances, model.Instance{
			InstanceID:    target.Name,
			InstanceName:  target.Name,
			Host:          target.Host,
			Port:          target.Port,
			Engine:        "sqlserver",
			ClusterName:   target.Cluster,
			Environment:   target.Environment,
			LogicalCPUs:       logicalCPUs,
			VisibleSchedulers: visibleSchedulers,
			UpdatedAt:     model.NewDateTime64(time.Now().UTC()),
		})
	}
	if err := persist(ctx, 10*time.Second, func(writeCtx context.Context) error {
		return writer.InsertInstances(writeCtx, instances)
	}); err != nil {
		return fmt.Errorf("initialize ClickHouse instance registry: %w", err)
	}

	if len(activeTargets) == 0 {
		return fmt.Errorf("no SQL Server targets could be opened")
	}
	for index, target := range activeTargets {
		go sampleTarget(ctx, target, samplers[index], writer, registry, interval, queryTimeout)
	}
	log.Printf("SQL Server PI sampler started for %d instance(s), %d target(s) skipped", len(activeTargets), len(targets)-len(activeTargets))

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	for _, sampler := range samplers {
		_ = sampler.Close()
	}
	return nil
}

func sampleTarget(ctx context.Context, target config.Target, sampler *sqlserver.Sampler, writer *clickhouse.Writer, registry *metrics.Registry, interval, queryTimeout time.Duration) {
	defer sampler.Close()
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			sampleOnce(ctx, target, sampler, writer, registry, queryTimeout)
		}
	}
}

func sampleOnce(parent context.Context, target config.Target, sampler *sqlserver.Sampler, writer *clickhouse.Writer, registry *metrics.Registry, queryTimeout time.Duration) {
	sampleID, err := newUUID()
	if err != nil {
		log.Printf("cannot create sample id for %s", target.Name)
		return
	}
	started := time.Now().UTC()
	queryCtx, cancel := context.WithTimeout(parent, queryTimeout)
	sessions, sampleTime, sampleErr := sampler.Sample(queryCtx, target.Name, sampleID)
	queryContextErr := queryCtx.Err()
	cancel()
	if sampleTime.IsZero() {
		sampleTime = started
	}
	duration := time.Since(started)

	status := pi.SampleSuccess
	errorCode := uint16(0)
	if sampleErr != nil {
		status = pi.SampleError
		errorCode = 2
		if errors.Is(sampleErr, context.DeadlineExceeded) || errors.Is(queryContextErr, context.DeadlineExceeded) {
			status = pi.SampleTimeout
			errorCode = 1
		}
	} else if len(sessions) > 0 {
		if err := persist(context.Background(), 5*time.Second, func(writeCtx context.Context) error {
			return writer.InsertSessions(writeCtx, sessions)
		}); err != nil {
			status = pi.SampleError
			errorCode = 3
		}
	}

	run := model.SampleRun{
		SampleID:             sampleID,
		InstanceID:           target.Name,
		SampleTime:            model.NewDateTime64(sampleTime),
		SampleStatus:          string(status),
		CollectionDurationMS:  durationMS(duration),
		ObservedSessionCount:  uint32(len(sessions)),
		ErrorCode:             errorCode,
	}
	runWriteErr := persist(context.Background(), 5*time.Second, func(writeCtx context.Context) error {
		return writer.InsertRuns(writeCtx, []model.SampleRun{run})
	})
	if runWriteErr != nil {
		status = pi.SampleError
		errorCode = 4
	}

	piSessions := make([]pi.ActiveSession, 0, len(sessions))
	for _, session := range sessions {
		piSessions = append(piSessions, pi.ActiveSession{SessionID: session.SessionID, Category: pi.WaitCategory(session.WaitCategory)})
	}
	registry.Record(target.Name, pi.SampleRun{ID: sampleID, Status: status, Sessions: piSessions}, duration)
	if status != pi.SampleSuccess {
		log.Printf("sample failed instance=%s status=%s error_code=%d", target.Name, status, errorCode)
	}
}

func metricsHandler(registry *metrics.Registry) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/metrics", registry)
	mux.HandleFunc("/-/healthy", func(writer http.ResponseWriter, _ *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, _ = writer.Write([]byte("healthy\n"))
	})
	return mux
}

func persist(parent context.Context, timeout time.Duration, write func(context.Context) error) error {
	ctx, cancel := context.WithTimeout(parent, timeout)
	defer cancel()
	return write(ctx)
}

func durationMS(value time.Duration) uint32 {
	milliseconds := value.Milliseconds()
	if milliseconds < 0 {
		return 0
	}
	if milliseconds > int64(^uint32(0)) {
		return ^uint32(0)
	}
	return uint32(milliseconds)
}

func durationOrDefault(name string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(name)
	if value == "" {
		return fallback, nil
	}
	return time.ParseDuration(value)
}

func envOrDefault(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

func newUUID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	encoded := hex.EncodeToString(value[:])
	return encoded[:8] + "-" + encoded[8:12] + "-" + encoded[12:16] + "-" + encoded[16:20] + "-" + encoded[20:], nil
}