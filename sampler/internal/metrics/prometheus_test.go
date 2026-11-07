package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"centinela/sampler/internal/pi"
)

func TestRegistryOmitsAASWhenThereIsNoSuccessfulSample(t *testing.T) {
	registry := NewRegistry(time.Minute)
	registry.Register("node-1", "cluster-a", nil)
	registry.Record("node-1", pi.SampleRun{ID: "failed", Status: pi.SampleError}, time.Millisecond)
	response := httptest.NewRecorder()
	registry.ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
	body := response.Body.String()
	if strings.Contains(body, "centinela_aas_total{") {
		t.Fatal("AAS must be absent when sample coverage is zero")
	}
	if !strings.Contains(body, `centinela_sampler_success_ratio{target="node-1",cluster="cluster-a"} 0`) {
		t.Fatal("zero coverage metric was not exported")
	}
}

func TestRegistryPublishesZeroAASForSuccessfulEmptySample(t *testing.T) {
	registry := NewRegistry(time.Minute)
	registry.Register("node-1", "cluster-a", nil)
	registry.Record("node-1", pi.SampleRun{ID: "empty", Status: pi.SampleSuccess}, time.Millisecond)
	response := httptest.NewRecorder()
	registry.ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
	body := response.Body.String()
	if !strings.Contains(body, `centinela_aas_total{target="node-1",cluster="cluster-a",category="CPU"} 0`) {
		t.Fatal("successful empty sample must export zero AAS")
	}
	if !strings.Contains(body, `centinela_sampler_success_ratio{target="node-1",cluster="cluster-a"} 1`) {
		t.Fatal("successful sample coverage was not exported")
	}
}

func TestRegistryExportsVisibleSchedulerCapacityPerTarget(t *testing.T) {
	registry := NewRegistry(time.Minute)
	visibleSchedulers := uint16(8)
	registry.Register("node-1", "cluster-a", &visibleSchedulers)
	registry.Record("node-1", pi.SampleRun{ID: "empty", Status: pi.SampleSuccess}, time.Millisecond)
	response := httptest.NewRecorder()
	registry.ServeHTTP(response, httptest.NewRequest("GET", "/metrics", nil))
	if !strings.Contains(response.Body.String(), `centinela_sql_visible_schedulers{target="node-1",cluster="cluster-a"} 8`) {
		t.Fatal("visible SQL scheduler capacity was not exported")
	}
}
