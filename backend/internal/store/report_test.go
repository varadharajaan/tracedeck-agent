package store

import (
	"bytes"
	"testing"
	"time"

	"github.com/varadharajaan/tracedeck-agent/backend/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/model"
)

func TestWeeklyReportGeneratesEmailAndPDFReadyPayload(t *testing.T) {
	t.Parallel()

	report := WeeklyReport(model.HostOverview{
		Device: model.Device{
			DeviceID: "weekly-device",
		},
		Summary: model.DeviceSummary{
			StudyMinutes:        240,
			CodingMinutes:       120,
			EntertainmentMins:   30,
			ComplianceScore:     88,
			DataCompletenessPct: 95,
		},
		RiskScore: 42,
		RiskLevel: constants.RiskLevelLow,
		Health: model.DeviceHealth{
			Score:  91,
			Status: constants.HealthStatusHealthy,
		},
		GeneratedAt: time.Now().UTC(),
	})

	if !report.Generated || !report.EmailReady || !report.PDFReady {
		t.Fatalf("expected generated email/pdf ready report: %+v", report)
	}
	if report.EmailSubject == "" || len(report.Highlights) == 0 || len(report.Risks) == 0 {
		t.Fatalf("expected report content: %+v", report)
	}

	pdf := WeeklyReportPDF(report)
	if !bytes.HasPrefix(pdf, []byte("%PDF-1.4")) {
		t.Fatalf("expected PDF header, got %q", string(pdf[:8]))
	}
	if !bytes.Contains(pdf, []byte("TraceDeck Weekly Report")) {
		t.Fatal("expected report title in PDF")
	}
}
