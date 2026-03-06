package health

import (
	"testing"

	"github.com/PeterHiroshi/cfmon/internal/api"
)

func TestDetermineStatus(t *testing.T) {
	tests := []struct {
		total  int
		want   string
	}{
		{100, "excellent"},
		{95, "excellent"},
		{90, "excellent"},
		{89, "good"},
		{80, "good"},
		{75, "good"},
		{74, "fair"},
		{60, "fair"},
		{50, "fair"},
		{49, "poor"},
		{30, "poor"},
		{25, "poor"},
		{24, "critical"},
		{10, "critical"},
		{0, "critical"},
	}

	for _, tt := range tests {
		got := determineStatus(tt.total)
		if got != tt.want {
			t.Errorf("determineStatus(%d) = %q, want %q", tt.total, got, tt.want)
		}
	}
}

func TestCheckWorkerHealth_NoWorkers(t *testing.T) {
	// Mock client that returns empty workers
	_ = &api.Client{}
	// Note: This won't actually work without mocking, but tests the logic

	// With no workers, should return full credit
	// (This test is more of a structural test - real testing would need mocks)
	// Skipping actual test execution as it requires API mocking
	t.Skip("Requires API mocking")
}

func TestCheckWorkerHealth_AllHealthy(t *testing.T) {
	// Test logic: if all workers have low error rates, should get full score
	// This would require proper mocking of the API client
	t.Skip("Requires API mocking")
}

func TestCheckContainerHealth_NoContainers(t *testing.T) {
	// Mock client that returns empty containers
	_ = &api.Client{}
	// With no containers, should return full credit
	// (This test is more of a structural test - real testing would need mocks)
	// Skipping actual test execution as it requires API mocking
	t.Skip("Requires API mocking")
}

func TestScore_Structure(t *testing.T) {
	score := &Score{
		Total:              85,
		APIConnectivity:    30,
		WorkerHealth:       30,
		ContainerHealth:    25,
		APIConnectivityMax: APIConnectivityMax,
		WorkerHealthMax:    WorkerHealthMax,
		ContainerHealthMax: ContainerHealthMax,
		Status:             "good",
		Message:            "Test message",
		Timestamp:          "2024-01-01T00:00:00Z",
	}

	if score.Total != 85 {
		t.Errorf("Total = %d, want 85", score.Total)
	}

	if score.Status != "good" {
		t.Errorf("Status = %q, want %q", score.Status, "good")
	}

	if score.APIConnectivityMax != 30 {
		t.Errorf("APIConnectivityMax = %d, want 30", score.APIConnectivityMax)
	}

	if score.WorkerHealthMax != 35 {
		t.Errorf("WorkerHealthMax = %d, want 35", score.WorkerHealthMax)
	}

	if score.ContainerHealthMax != 35 {
		t.Errorf("ContainerHealthMax = %d, want 35", score.ContainerHealthMax)
	}
}

func TestConstants(t *testing.T) {
	if APIConnectivityMax != 30 {
		t.Errorf("APIConnectivityMax = %d, want 30", APIConnectivityMax)
	}

	if WorkerHealthMax != 35 {
		t.Errorf("WorkerHealthMax = %d, want 35", WorkerHealthMax)
	}

	if ContainerHealthMax != 35 {
		t.Errorf("ContainerHealthMax = %d, want 35", ContainerHealthMax)
	}

	if MaxScore != 100 {
		t.Errorf("MaxScore = %d, want 100", MaxScore)
	}

	total := APIConnectivityMax + WorkerHealthMax + ContainerHealthMax
	if total != MaxScore {
		t.Errorf("Sum of max scores = %d, want %d", total, MaxScore)
	}
}
