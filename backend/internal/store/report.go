package store

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/varadharajaan/tracedeck-agent/backend/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/model"
)

func WeeklyReport(overview model.HostOverview) model.WeeklyReport {
	generatedAt := time.Now().UTC()
	deviceID := strings.TrimSpace(overview.Device.DeviceID)
	summary := overview.Summary
	week := generatedAt.Format("2006-W01")
	highlights := []string{
		fmt.Sprintf("%d study minutes and %d coding minutes recorded.", summary.StudyMinutes, summary.CodingMinutes),
		fmt.Sprintf("Compliance score is %d with %d%% data completeness.", summary.ComplianceScore, summary.DataCompletenessPct),
		fmt.Sprintf("Device health score is %d (%s).", overview.Health.Score, overview.Health.Status),
	}
	risks := []string{
		fmt.Sprintf("%d policy violations, %d anomalies, and %d tamper signals need review.", len(overview.PolicyViolations), len(overview.Anomalies), len(overview.TamperEvents)),
		fmt.Sprintf("Risk score is %d (%s).", overview.RiskScore, overview.RiskLevel),
	}
	if summary.EntertainmentMins > 0 {
		risks = append(risks, fmt.Sprintf("%d entertainment minutes recorded this period.", summary.EntertainmentMins))
	}

	subject := fmt.Sprintf("TraceDeck weekly report for %s", deviceID)
	preview := fmt.Sprintf("Compliance %d, risk %d, health %d.", summary.ComplianceScore, overview.RiskScore, overview.Health.Score)
	emailReady := hasConfirmedEmailDelivery(overview.AlertDeliveries)
	return model.WeeklyReport{
		DeviceID:      deviceID,
		Week:          week,
		Summary:       preview,
		Highlights:    highlights,
		Risks:         risks,
		Generated:     true,
		GeneratedNote: "weekly report generated from current host overview; email proof requires a non-demo delivered email route",
		EmailReady:    emailReady,
		EmailSubject:  subject,
		EmailPreview:  preview,
		PDFReady:      true,
		GeneratedAt:   generatedAt,
	}
}

func hasConfirmedEmailDelivery(deliveries []model.AlertDelivery) bool {
	for _, delivery := range deliveries {
		if delivery.Channel != constants.DeliveryChannelEmail {
			continue
		}
		if delivery.Status != constants.DeliveryStatusDelivered {
			continue
		}
		if delivery.SourceKind == constants.EvidenceSourceDemoSeed {
			continue
		}
		return true
	}
	return false
}

func WeeklyReportPDF(report model.WeeklyReport) []byte {
	lines := []string{
		"TraceDeck Weekly Report",
		"Device: " + report.DeviceID,
		"Week: " + report.Week,
		"Summary: " + report.Summary,
		"Highlights:",
	}
	lines = append(lines, prefixedLines(report.Highlights, "- ")...)
	lines = append(lines, "Risks:")
	lines = append(lines, prefixedLines(report.Risks, "- ")...)
	lines = append(lines, "Email: "+report.EmailSubject)
	return simplePDF(lines)
}

func prefixedLines(values []string, prefix string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, prefix+value)
	}
	return out
}

func simplePDF(lines []string) []byte {
	var content strings.Builder
	content.WriteString("BT\n/F1 12 Tf\n72 760 Td\n14 TL\n")
	for index, line := range lines {
		if index > 0 {
			content.WriteString("T*\n")
		}
		content.WriteString("(")
		content.WriteString(escapePDFText(line))
		content.WriteString(") Tj\n")
	}
	content.WriteString("ET\n")

	objects := []string{
		"<< /Type /Catalog /Pages 2 0 R >>",
		"<< /Type /Pages /Kids [3 0 R] /Count 1 >>",
		"<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] /Resources << /Font << /F1 4 0 R >> >> /Contents 5 0 R >>",
		"<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>",
		fmt.Sprintf("<< /Length %d >>\nstream\n%sendstream", len(content.String()), content.String()),
	}

	var output bytes.Buffer
	output.WriteString("%PDF-1.4\n")
	offsets := make([]int, 0, len(objects)+1)
	offsets = append(offsets, 0)
	for index, object := range objects {
		offsets = append(offsets, output.Len())
		fmt.Fprintf(&output, "%d 0 obj\n%s\nendobj\n", index+1, object)
	}
	xrefOffset := output.Len()
	fmt.Fprintf(&output, "xref\n0 %d\n", len(objects)+1)
	output.WriteString("0000000000 65535 f \n")
	for _, offset := range offsets[1:] {
		fmt.Fprintf(&output, "%010d 00000 n \n", offset)
	}
	fmt.Fprintf(&output, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%%%EOF\n", len(objects)+1, xrefOffset)
	return output.Bytes()
}

func escapePDFText(value string) string {
	value = strings.ReplaceAll(value, `\`, `\\`)
	value = strings.ReplaceAll(value, "(", `\(`)
	value = strings.ReplaceAll(value, ")", `\)`)
	return value
}
