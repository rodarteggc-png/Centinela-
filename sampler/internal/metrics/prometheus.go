package metrics

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"centinela/sampler/internal/pi"
)

type timedRun struct {
	at  time.Time
	run pi.SampleRun
}

type instanceState struct {
	runs              []timedRun
	lastLatencyMS     float64
	errorsTotal       uint64
	cluster           string
	visibleSchedulers *uint16
}

type Registry struct {
	mu        sync.RWMutex
	instances map[string]*instanceState
	window    time.Duration
}

func NewRegistry(window time.Duration) *Registry {
	return &Registry{instances: make(map[string]*instanceState), window: window}
}

func (registry *Registry) Register(instance, cluster string, visibleSchedulers *uint16) {
	registry.mu.Lock()
	defer registry.mu.Unlock()
	state := registry.instances[instance]
	if state == nil {
		state = &instanceState{}
		registry.instances[instance] = state
	}
	state.cluster = cluster
	state.visibleSchedulers = visibleSchedulers
}

func (registry *Registry) Record(instance string, run pi.SampleRun, latency time.Duration) {
	now := time.Now().UTC()
	registry.mu.Lock()
	defer registry.mu.Unlock()
	state := registry.instances[instance]
	if state == nil {
		state = &instanceState{}
		registry.instances[instance] = state
	}
	state.runs = append(state.runs, timedRun{at: now, run: run})
	state.lastLatencyMS = float64(latency.Microseconds()) / 1000
	if run.Status != pi.SampleSuccess {
		state.errorsTotal++
	}
	cutoff := now.Add(-registry.window)
	firstRecent := 0
	for firstRecent < len(state.runs) && state.runs[firstRecent].at.Before(cutoff) {
		firstRecent++
	}
	state.runs = append([]timedRun(nil), state.runs[firstRecent:]...)
}

func (registry *Registry) ServeHTTP(writer http.ResponseWriter, _ *http.Request) {
	registry.mu.RLock()
	defer registry.mu.RUnlock()

	writer.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	fmt.Fprintln(writer, "# HELP centinela_aas_total Average active sessions from successful samples in the rolling window.")
	fmt.Fprintln(writer, "# TYPE centinela_aas_total gauge")
	fmt.Fprintln(writer, "# HELP centinela_sampler_success_ratio Successful sample attempts in the rolling window.")
	fmt.Fprintln(writer, "# TYPE centinela_sampler_success_ratio gauge")
	fmt.Fprintln(writer, "# HELP centinela_sampler_latency_ms Last SQL Server collection duration in milliseconds.")
	fmt.Fprintln(writer, "# TYPE centinela_sampler_latency_ms gauge")
	fmt.Fprintln(writer, "# HELP centinela_sampler_errors_total Failed SQL Server sample attempts.")
	fmt.Fprintln(writer, "# TYPE centinela_sampler_errors_total counter")
	fmt.Fprintln(writer, "# HELP centinela_sql_visible_schedulers SQL Server schedulers available for user workload.")
	fmt.Fprintln(writer, "# TYPE centinela_sql_visible_schedulers gauge")

	instances := make([]string, 0, len(registry.instances))
	for instance := range registry.instances {
		instances = append(instances, instance)
	}
	sort.Strings(instances)
	for _, instance := range instances {
		state := registry.instances[instance]
		runs := make([]pi.SampleRun, 0, len(state.runs))
		for _, timed := range state.runs {
			runs = append(runs, timed.run)
		}
		summary, err := pi.CalculateAAS(runs)
		if err != nil {
			continue
		}
		label := escapeLabel(instance)
		cluster := escapeLabel(state.cluster)
		labels := fmt.Sprintf("target=\"%s\",cluster=\"%s\"", label, cluster)
		fmt.Fprintf(writer, "centinela_sampler_success_ratio{%s} %s\n", labels, formatFloat(summary.CoverageRatio))
		fmt.Fprintf(writer, "centinela_sampler_latency_ms{%s} %s\n", labels, formatFloat(state.lastLatencyMS))
		fmt.Fprintf(writer, "centinela_sampler_errors_total{%s} %d\n", labels, state.errorsTotal)
		if state.visibleSchedulers != nil {
			fmt.Fprintf(writer, "centinela_sql_visible_schedulers{%s} %d\n", labels, *state.visibleSchedulers)
		}
		if summary.SuccessfulRuns == 0 {
			continue
		}
		for category, value := range summary.ByCategory {
			fmt.Fprintf(writer, "centinela_aas_total{%s,category=\"%s\"} %s\n", labels, escapeLabel(string(category)), formatFloat(value))
		}
	}
}

func escapeLabel(value string) string {
	return strings.NewReplacer(`\\`, `\\\\`, `"`, `\\"`, "\n", `\\n`).Replace(value)
}

func formatFloat(value float64) string {
	return fmt.Sprintf("%g", value)
}
