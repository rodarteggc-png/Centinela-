package pi

import "fmt"

type WaitCategory string

const (
	CategoryCPU     WaitCategory = "CPU"
	CategoryIO      WaitCategory = "IO"
	CategoryLock    WaitCategory = "Lock"
	CategoryMemory  WaitCategory = "Memory"
	CategoryNetwork WaitCategory = "Network"
	CategoryOther   WaitCategory = "Other"
)

var waitCategories = [...]WaitCategory{
	CategoryCPU,
	CategoryIO,
	CategoryLock,
	CategoryMemory,
	CategoryNetwork,
	CategoryOther,
}

type SampleStatus string

const (
	SampleSuccess SampleStatus = "success"
	SampleTimeout SampleStatus = "timeout"
	SampleError   SampleStatus = "error"
)

type ActiveSession struct {
	SessionID uint16
	Category  WaitCategory
}

type SampleRun struct {
	ID       string
	Status   SampleStatus
	Sessions []ActiveSession
}

type AASSummary struct {
	Attempts       int
	SuccessfulRuns int
	CoverageRatio  float64
	ByCategory     map[WaitCategory]float64
}

func CalculateAAS(runs []SampleRun) (AASSummary, error) {
	result := AASSummary{
		Attempts:   len(runs),
		ByCategory: make(map[WaitCategory]float64, len(waitCategories)),
	}
	for _, category := range waitCategories {
		result.ByCategory[category] = 0
	}

	seenRunIDs := make(map[string]struct{}, len(runs))
	categoryTotals := make(map[WaitCategory]uint64, len(waitCategories))
	for _, category := range waitCategories {
		categoryTotals[category] = 0
	}

	for runIndex, run := range runs {
		if run.ID == "" {
			return AASSummary{}, fmt.Errorf("sample run %d has an empty ID", runIndex)
		}
		if _, exists := seenRunIDs[run.ID]; exists {
			return AASSummary{}, fmt.Errorf("duplicate sample run ID %q", run.ID)
		}
		seenRunIDs[run.ID] = struct{}{}

		switch run.Status {
		case SampleSuccess:
			result.SuccessfulRuns++
		case SampleTimeout, SampleError:
			continue
		default:
			return AASSummary{}, fmt.Errorf("sample run %q has unsupported status %q", run.ID, run.Status)
		}

		seenSessions := make(map[uint16]WaitCategory, len(run.Sessions))
		for _, session := range run.Sessions {
			if !isKnownCategory(session.Category) {
				return AASSummary{}, fmt.Errorf("sample run %q has unsupported category %q", run.ID, session.Category)
			}
			if previousCategory, exists := seenSessions[session.SessionID]; exists {
				return AASSummary{}, fmt.Errorf(
					"sample run %q contains session %d more than once (%s, %s)",
					run.ID,
					session.SessionID,
					previousCategory,
					session.Category,
				)
			}
			seenSessions[session.SessionID] = session.Category
			categoryTotals[session.Category]++
		}
	}

	if result.SuccessfulRuns == 0 {
		return result, nil
	}
	result.CoverageRatio = float64(result.SuccessfulRuns) / float64(result.Attempts)
	for category, total := range categoryTotals {
		result.ByCategory[category] = float64(total) / float64(result.SuccessfulRuns)
	}
	return result, nil
}

func isKnownCategory(category WaitCategory) bool {
	for _, knownCategory := range waitCategories {
		if category == knownCategory {
			return true
		}
	}
	return false
}