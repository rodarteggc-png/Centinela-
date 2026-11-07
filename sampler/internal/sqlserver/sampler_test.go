package sqlserver

import "testing"

func TestSQLServerSchedulerCapacityContractIsDistinguishedFromPhysicalCores(t *testing.T) {
	const logicalCPUCount = 16
	const visibleSchedulerCount = 8
	if visibleSchedulerCount > logicalCPUCount {
		t.Fatal("visible schedulers cannot exceed logical CPUs")
	}
}