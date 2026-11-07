package pi

import "testing"

func TestCalculateAASIncludesSuccessfulEmptySamplesAndExcludesFailures(t *testing.T) {
	runs := []SampleRun{
		{ID: "sample-1", Status: SampleSuccess},
		{
			ID:     "sample-2",
			Status: SampleSuccess,
			Sessions: []ActiveSession{
				{SessionID: 51, Category: CategoryCPU},
			},
		},
		{
			ID:     "sample-3",
			Status: SampleTimeout,
			Sessions: []ActiveSession{
				{SessionID: 52, Category: CategoryLock},
			},
		},
	}

	summary, err := CalculateAAS(runs)
	if err != nil {
		t.Fatalf("CalculateAAS returned an error: %v", err)
	}
	if summary.Attempts != 3 || summary.SuccessfulRuns != 2 {
		t.Fatalf("unexpected sample counts: %+v", summary)
	}
	if summary.ByCategory[CategoryCPU] != 0.5 {
		t.Fatalf("CPU AAS = %v, want 0.5", summary.ByCategory[CategoryCPU])
	}
	if summary.ByCategory[CategoryLock] != 0 {
		t.Fatalf("Lock AAS = %v, want 0", summary.ByCategory[CategoryLock])
	}
	if summary.CoverageRatio != 2.0/3.0 {
		t.Fatalf("coverage = %v, want %v", summary.CoverageRatio, 2.0/3.0)
	}
}

func TestCalculateAASCountsEachSessionOncePerSample(t *testing.T) {
	runs := []SampleRun{
		{
			ID:     "sample-1",
			Status: SampleSuccess,
			Sessions: []ActiveSession{
				{SessionID: 51, Category: CategoryCPU},
				{SessionID: 52, Category: CategoryCPU},
				{SessionID: 53, Category: CategoryLock},
			},
		},
		{
			ID:     "sample-2",
			Status: SampleSuccess,
			Sessions: []ActiveSession{
				{SessionID: 51, Category: CategoryCPU},
			},
		},
	}

	summary, err := CalculateAAS(runs)
	if err != nil {
		t.Fatalf("CalculateAAS returned an error: %v", err)
	}
	if summary.ByCategory[CategoryCPU] != 1.5 || summary.ByCategory[CategoryLock] != 0.5 {
		t.Fatalf("unexpected AAS by category: %v", summary.ByCategory)
	}
}

func TestCalculateAASRejectsDuplicateSessionAndSampleIDs(t *testing.T) {
	duplicateSession := []SampleRun{{
		ID:     "sample-1",
		Status: SampleSuccess,
		Sessions: []ActiveSession{
			{SessionID: 51, Category: CategoryCPU},
			{SessionID: 51, Category: CategoryCPU},
		},
	}}
	if _, err := CalculateAAS(duplicateSession); err == nil {
		t.Fatal("expected duplicate session error")
	}

	duplicateSample := []SampleRun{
		{ID: "sample-1", Status: SampleSuccess},
		{ID: "sample-1", Status: SampleSuccess},
	}
	if _, err := CalculateAAS(duplicateSample); err == nil {
		t.Fatal("expected duplicate sample ID error")
	}
}

func TestCalculateAASReturnsZeroesWhenNoAttemptSucceeded(t *testing.T) {
	summary, err := CalculateAAS([]SampleRun{
		{ID: "sample-1", Status: SampleError},
	})
	if err != nil {
		t.Fatalf("CalculateAAS returned an error: %v", err)
	}
	if summary.SuccessfulRuns != 0 || summary.CoverageRatio != 0 {
		t.Fatalf("unexpected summary for failed sample: %+v", summary)
	}
	for _, value := range summary.ByCategory {
		if value != 0 {
			t.Fatalf("AAS = %v with no successful samples, want 0", value)
		}
	}
}