package api

import (
	"bytes"
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/varadharajaan/tracedeck-agent/backend/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/model"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/store"
)

//go:embed web/*.html
var dashboardFS embed.FS

type Server struct {
	store                    store.Repository
	logger                   *slog.Logger
	startedAt                time.Time
	auth                     AuthConfig
	runtimeSummaryPath       string
	verificationEvidencePath string
	operatorAssurancePath    string
}

func NewServer(repo store.Repository, logger *slog.Logger) *Server {
	if repo == nil {
		repo = store.NewMemory()
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		store:                    repo,
		logger:                   logger,
		startedAt:                time.Now().UTC(),
		runtimeSummaryPath:       constants.DefaultRuntimeSummaryPath,
		verificationEvidencePath: constants.DefaultVerificationEvidencePath,
		operatorAssurancePath:    constants.DefaultOperatorAssurancePath,
	}
}

func NewServerWithAuth(repo store.Repository, logger *slog.Logger, auth AuthConfig) *Server {
	server := NewServer(repo, logger)
	server.auth = auth
	return server
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(constants.RouteDashboard, s.handleDashboard)
	mux.HandleFunc(constants.RouteBrowserActivity, s.handleBrowserActivityPage)
	mux.HandleFunc(constants.RouteHealth, s.handleHealth)
	mux.HandleFunc(constants.RouteVersion, s.handleVersion)
	mux.HandleFunc(constants.RouteDeviceEnroll, s.handleDeviceEnroll)
	mux.HandleFunc(constants.RouteDevices, s.handleDevices)
	mux.HandleFunc(constants.RouteDevices+"/", s.handleDeviceRoutes)
	mux.HandleFunc(constants.RouteTenants, s.handleTenants)
	mux.HandleFunc(constants.RouteTenants+"/", s.handleTenantRoutes)
	mux.HandleFunc(constants.RoutePlans, s.handlePlans)
	mux.HandleFunc(constants.RouteRoles, s.handleRoles)
	mux.HandleFunc(constants.RouteRetentionTiers, s.handleRetentionTiers)
	mux.HandleFunc(constants.RouteAuditEvents, s.handleAuditEvents)
	mux.HandleFunc(constants.RoutePolicyTemplates, s.handlePolicyTemplates)
	mux.HandleFunc(constants.RouteAlertRuleTemplates, s.handleAlertRuleTemplates)
	mux.HandleFunc(constants.RouteAccountPortfolio, s.handleAccountPortfolioIndex)
	mux.HandleFunc(constants.RouteRuntimeStatus, s.handleRuntimeStatusCenter)
	mux.HandleFunc(constants.RouteVerificationCenter, s.handleVerificationEvidenceCenter)
	mux.HandleFunc(constants.RouteOperatorAssurance, s.handleOperatorAssuranceCenter)
	mux.HandleFunc(constants.RouteArchiveStatus, s.handleArchiveStatus)
	return requestLogger(s.logger, s.authMiddleware(mux))
}

func Serve(ctx context.Context, addr string, handler http.Handler, logger *slog.Logger) error {
	if err := validateLocalAddress(addr); err != nil {
		return err
	}

	server := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		logger.Info("backend server started", "addr", addr)
		errCh <- server.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown backend: %w", err)
		}
		return nil
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

func validateLocalAddress(addr string) error {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return fmt.Errorf("invalid backend address %q: %w", addr, err)
	}
	if host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return nil
	}
	return fmt.Errorf("backend address must bind to localhost, got %q", host)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != constants.RouteDashboard {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	data, err := fs.ReadFile(dashboardFS, "web/dashboard.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "dashboard asset unavailable")
		return
	}
	setNoStoreHeaders(w)
	w.Header().Set("Content-Type", constants.ContentTypeHTML)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleBrowserActivityPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != constants.RouteBrowserActivity {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}

	data, err := fs.ReadFile(dashboardFS, "web/browser_activity.html")
	if err != nil {
		writeError(w, http.StatusInternalServerError, "browser activity asset unavailable")
		return
	}
	setNoStoreHeaders(w)
	w.Header().Set("Content-Type", constants.ContentTypeHTML)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, model.Health{
		Status:    constants.StatusOK,
		Service:   constants.BackendName,
		Version:   constants.BackendVersion,
		StartedAt: s.startedAt,
	})
}

func (s *Server) handleVersion(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, model.Version{
		Service: constants.BackendName,
		Version: constants.BackendVersion,
	})
}

func (s *Server) handleRuntimeStatusCenter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != constants.RouteRuntimeStatus {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	center, err := s.runtimeStatusCenter()
	if err != nil {
		s.logger.Error("runtime status center lookup failed", "error", err)
		writeError(w, http.StatusInternalServerError, "runtime status center lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, center)
}

func (s *Server) handleVerificationEvidenceCenter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != constants.RouteVerificationCenter {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	center, err := s.verificationEvidenceCenter()
	if err != nil {
		s.logger.Error("verification evidence center lookup failed", "error", err)
		writeError(w, http.StatusInternalServerError, "verification evidence center lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, center)
}

func (s *Server) handleOperatorAssuranceCenter(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != constants.RouteOperatorAssurance {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	center, err := s.operatorAssuranceCenter()
	if err != nil {
		s.logger.Error("operator assurance center lookup failed", "error", err)
		writeError(w, http.StatusInternalServerError, "operator assurance center lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, center)
}

func (s *Server) runtimeStatusCenter() (model.RuntimeStatusCenter, error) {
	summaryPath := strings.TrimSpace(s.runtimeSummaryPath)
	if summaryPath == "" {
		summaryPath = constants.DefaultRuntimeSummaryPath
	}

	data, err := os.ReadFile(summaryPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return missingRuntimeStatusCenter(summaryPath), nil
		}
		return model.RuntimeStatusCenter{}, fmt.Errorf("read runtime summary: %w", err)
	}

	var artifact model.RuntimeSummaryArtifact
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if err := json.Unmarshal(data, &artifact); err != nil {
		return model.RuntimeStatusCenter{}, fmt.Errorf("decode runtime summary: %w", err)
	}
	return runtimeStatusCenterFromArtifact(summaryPath, artifact), nil
}

func missingRuntimeStatusCenter(summaryPath string) model.RuntimeStatusCenter {
	now := time.Now().UTC()
	return model.RuntimeStatusCenter{
		SummaryAvailable: false,
		SummaryPath:      summaryPath,
		Source:           constants.RuntimeStatusSourcePhase97Summary,
		Summary: model.RuntimeStatusSummary{
			Status:             constants.StatusAttention,
			Headline:           "Runtime summary has not been generated.",
			Detail:             "Run the scripted summary command before using this page for operator proof.",
			CanContinue:        false,
			DoctorOverall:      constants.StatusPending,
			DoctorLocal:        constants.StatusPending,
			SchedulerReadback:  constants.StatusPending,
			SummaryGeneratedAt: "",
		},
		Proof: []model.RuntimeStatusProof{
			{
				ID:            constants.RuntimeStatusProofBackendID,
				Label:         "Backend Runtime",
				Value:         "summary missing",
				Detail:        "No runtime summary artifact was found at the configured local path.",
				Status:        constants.StatusAttention,
				EvidenceScope: constants.EvidenceScopeMetadataOnly,
			},
		},
		Actions: []model.RuntimeStatusAction{
			{
				ID:       constants.RuntimeStatusActionSummaryID,
				Title:    "Generate runtime summary",
				Detail:   "Create the local JSON and text operator proof files under data/local/output.",
				Command:  constants.RuntimeSummaryCommand,
				Severity: constants.SeverityHigh,
				Status:   constants.StatusPending,
			},
		},
		PrivacyBoundary: constants.RuntimeStatusPrivacyNote,
		GeneratedAt:     now,
	}
}

func runtimeStatusCenterFromArtifact(summaryPath string, artifact model.RuntimeSummaryArtifact) model.RuntimeStatusCenter {
	status := runtimeStatusValue(artifact)
	headline := strings.TrimSpace(artifact.Verdict.Headline)
	if headline == "" {
		headline = "Runtime proof is available."
	}
	detail := runtimeStatusDetail(artifact)
	actions := runtimeStatusActions(artifact)

	return model.RuntimeStatusCenter{
		SummaryAvailable: true,
		SummaryPath:      summaryPath,
		Source:           constants.RuntimeStatusSourcePhase97Summary,
		BaseURL:          artifact.BaseURL,
		Summary: model.RuntimeStatusSummary{
			Status:             status,
			Headline:           headline,
			Detail:             detail,
			CanContinue:        artifact.Verdict.CanContinue,
			RuntimeOK:          artifact.Backend.RuntimeOK,
			HealthOK:           artifact.Backend.HealthOK,
			SchedulerReadback:  artifact.Backend.SchedulerReadback,
			LaunchTaskVerified: artifact.Backend.LaunchTaskVerified,
			DoctorOverall:      artifact.Doctor.Overall,
			DoctorLocal:        artifact.Doctor.Local,
			TrackedContentDiff: artifact.Git.TrackedContentDiff,
			FrontendURLPresent: artifact.Frontend.URLPresent,
			SummaryGeneratedAt: artifact.GeneratedAt,
		},
		Proof:           runtimeStatusProof(artifact),
		Actions:         actions,
		Artifact:        artifact,
		PrivacyBoundary: constants.RuntimeStatusPrivacyNote,
		GeneratedAt:     time.Now().UTC(),
	}
}

func runtimeStatusValue(artifact model.RuntimeSummaryArtifact) string {
	if !artifact.Verdict.CanContinue || !artifact.Backend.RuntimeOK || !artifact.Backend.HealthOK {
		return constants.StatusAttention
	}
	if artifact.Doctor.Overall != "" && artifact.Doctor.Overall != constants.StatusOK && artifact.Doctor.Overall != "skipped" {
		return constants.StatusAttention
	}
	if artifact.Git.TrackedContentDiff || artifact.Backend.Advisory.Severity == constants.StatusWatch || artifact.Backend.SchedulerReadback == "denied" {
		return constants.StatusWatch
	}
	return constants.StatusOK
}

func runtimeStatusDetail(artifact model.RuntimeSummaryArtifact) string {
	advisory := strings.TrimSpace(artifact.Backend.Advisory.Headline)
	if advisory == "" {
		advisory = "No task advisory was recorded."
	}
	return fmt.Sprintf("Backend runtime=%t health=%t, Scheduler=%s, doctor=%s, git tracked diff=%t. %s",
		artifact.Backend.RuntimeOK,
		artifact.Backend.HealthOK,
		emptyFallback(artifact.Backend.SchedulerReadback, "unknown"),
		emptyFallback(artifact.Doctor.Overall, "unknown"),
		artifact.Git.TrackedContentDiff,
		advisory,
	)
}

func runtimeStatusProof(artifact model.RuntimeSummaryArtifact) []model.RuntimeStatusProof {
	backendStatus := constants.StatusAttention
	if artifact.Backend.RuntimeOK && artifact.Backend.HealthOK {
		backendStatus = constants.StatusOK
	}
	schedulerStatus := constants.StatusWatch
	if artifact.Backend.LaunchTaskVerified {
		schedulerStatus = constants.StatusOK
	} else if !artifact.Backend.TaskPresent {
		schedulerStatus = constants.StatusAttention
	}
	doctorStatus := constants.StatusAttention
	if artifact.Doctor.Overall == constants.StatusOK || artifact.Doctor.Overall == "skipped" {
		doctorStatus = constants.StatusOK
	}
	frontendStatus := constants.StatusWatch
	if artifact.Frontend.URLPresent {
		frontendStatus = constants.StatusOK
	}
	gitStatus := constants.StatusOK
	if artifact.Git.TrackedContentDiff {
		gitStatus = constants.StatusAttention
	}
	privacyStatus := constants.StatusAttention
	if artifact.Privacy.MetadataOnly && artifact.Privacy.SensitiveCollection == "denied" {
		privacyStatus = constants.StatusOK
	}

	return []model.RuntimeStatusProof{
		{
			ID:            constants.RuntimeStatusProofBackendID,
			Label:         "Backend Runtime",
			Value:         boolRuntimeLabel(artifact.Backend.RuntimeOK && artifact.Backend.HealthOK),
			Detail:        fmt.Sprintf("pid=%d running=%t evidence=%s ready_file=%t", artifact.Backend.PID, artifact.Backend.PIDRunning, emptyFallback(artifact.Backend.RuntimeEvidence, "unknown"), artifact.Backend.ReadyFilePresent),
			Status:        backendStatus,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.RuntimeStatusProofSchedulerID,
			Label:         "Scheduler Readback",
			Value:         emptyFallback(artifact.Backend.SchedulerReadback, "unknown"),
			Detail:        fmt.Sprintf("task=%s state=%s verified=%t", emptyFallback(artifact.Backend.TaskName, "unknown"), emptyFallback(artifact.Backend.TaskState, "unknown"), artifact.Backend.LaunchTaskVerified),
			Status:        schedulerStatus,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.RuntimeStatusProofDoctorID,
			Label:         "Runtime Doctor",
			Value:         emptyFallback(artifact.Doctor.Overall, "unknown"),
			Detail:        fmt.Sprintf("local=%s skipped=%t report=%s", emptyFallback(artifact.Doctor.Local, "unknown"), artifact.Doctor.Skipped, emptyFallback(artifact.Doctor.ReportJSON, "not recorded")),
			Status:        doctorStatus,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.RuntimeStatusProofFrontendID,
			Label:         "Frontend URL",
			Value:         boolRuntimeLabel(artifact.Frontend.URLPresent),
			Detail:        emptyFallback(artifact.Frontend.URL, "No Function URL recorded."),
			Status:        frontendStatus,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.RuntimeStatusProofGitID,
			Label:         "Git Hygiene",
			Value:         boolCleanLabel(!artifact.Git.TrackedContentDiff),
			Detail:        fmt.Sprintf("branch=%s head=%s tracked_diff_count=%d", emptyFallback(artifact.Git.Branch, "unknown"), emptyFallback(artifact.Git.Head, "unknown"), artifact.Git.TrackedContentDiffCount),
			Status:        gitStatus,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.RuntimeStatusProofPrivacyID,
			Label:         "Privacy Boundary",
			Value:         boolRuntimeLabel(artifact.Privacy.MetadataOnly && artifact.Privacy.SensitiveCollection == "denied"),
			Detail:        fmt.Sprintf("metadata_only=%t sensitive_collection=%s", artifact.Privacy.MetadataOnly, emptyFallback(artifact.Privacy.SensitiveCollection, "unknown")),
			Status:        privacyStatus,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
	}
}

func runtimeStatusActions(artifact model.RuntimeSummaryArtifact) []model.RuntimeStatusAction {
	seen := map[string]bool{}
	actions := make([]model.RuntimeStatusAction, 0, len(artifact.Verdict.NextActions)+2)
	addAction := func(id string, title string, detail string, command string, severity string, status string) {
		key := strings.TrimSpace(title + "|" + detail + "|" + command)
		if key == "" || seen[key] {
			return
		}
		seen[key] = true
		actions = append(actions, model.RuntimeStatusAction{
			ID:       id,
			Title:    title,
			Detail:   detail,
			Command:  command,
			Severity: severity,
			Status:   status,
		})
	}

	if !artifact.Backend.RuntimeOK || !artifact.Backend.HealthOK {
		addAction(constants.RuntimeStatusActionRestartID, "Restart backend task", "Restart the hidden local backend task and rerun the summary.", constants.RuntimeTaskRestartCommand, constants.SeverityHigh, constants.StatusPending)
	}
	for index, action := range artifact.Verdict.NextActions {
		trimmed := strings.TrimSpace(action)
		if trimmed == "" {
			continue
		}
		addAction(fmt.Sprintf("%s-%d", constants.RuntimeStatusActionReviewID, index+1), "Runtime summary action", trimmed, constants.RuntimeSummaryCommand, constants.SeverityInfo, constants.StatusWatch)
	}
	if strings.TrimSpace(artifact.Backend.Advisory.OperatorAction) != "" {
		addAction(constants.RuntimeStatusActionReviewID, "Task advisory action", artifact.Backend.Advisory.OperatorAction, constants.RuntimeSummaryCommand, artifact.Backend.Advisory.Severity, constants.StatusWatch)
	}
	if len(actions) == 0 {
		addAction(constants.RuntimeStatusActionSummaryID, "Keep runtime proof fresh", "Refresh the local runtime summary before demos, deploys, or post-merge checks.", constants.RuntimeSummaryCommand, constants.SeverityInfo, constants.StatusOK)
	}
	return actions
}

func emptyFallback(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func boolRuntimeLabel(value bool) string {
	if value {
		return "ready"
	}
	return "attention"
}

func boolCleanLabel(value bool) string {
	if value {
		return "clean"
	}
	return "dirty"
}

func (s *Server) verificationEvidenceCenter() (model.VerificationEvidenceCenter, error) {
	evidencePath := strings.TrimSpace(s.verificationEvidencePath)
	if evidencePath == "" {
		evidencePath = constants.DefaultVerificationEvidencePath
	}

	data, err := os.ReadFile(evidencePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return missingVerificationEvidenceCenter(evidencePath), nil
		}
		return model.VerificationEvidenceCenter{}, fmt.Errorf("read verification evidence: %w", err)
	}

	var artifact model.VerificationEvidenceArtifact
	data = bytes.TrimPrefix(data, []byte{0xEF, 0xBB, 0xBF})
	if err := json.Unmarshal(data, &artifact); err != nil {
		return model.VerificationEvidenceCenter{}, fmt.Errorf("decode verification evidence: %w", err)
	}
	return verificationEvidenceCenterFromArtifact(evidencePath, artifact), nil
}

func missingVerificationEvidenceCenter(evidencePath string) model.VerificationEvidenceCenter {
	now := time.Now().UTC()
	return model.VerificationEvidenceCenter{
		EvidenceAvailable: false,
		EvidencePath:      evidencePath,
		Source:            constants.VerificationEvidenceSourcePhase99,
		Summary: model.VerificationEvidenceSummary{
			Status:            constants.StatusAttention,
			Headline:          "Verification evidence has not been generated.",
			Detail:            "Run the scripted evidence command so the dashboard can show local gate proof.",
			CanPromote:        false,
			EvidenceGenerated: "",
		},
		Proof: []model.VerificationEvidenceProof{
			{
				ID:            constants.VerificationEvidenceProofScriptedID,
				Label:         "Scripted Evidence",
				Value:         "missing",
				Detail:        "No verification evidence artifact was found at the configured local path.",
				Status:        constants.StatusAttention,
				EvidenceScope: constants.EvidenceScopeMetadataOnly,
			},
		},
		Actions: []model.VerificationEvidenceAction{
			{
				ID:       constants.VerificationEvidenceActionGenerateID,
				Title:    "Generate verification evidence",
				Detail:   "Create the local JSON evidence file under data/local/output.",
				Command:  constants.VerificationEvidenceCommand,
				Severity: constants.SeverityHigh,
				Status:   constants.StatusPending,
			},
		},
		PrivacyBoundary: constants.VerificationEvidencePrivacyNote,
		GeneratedAt:     now,
	}
}

func verificationEvidenceCenterFromArtifact(evidencePath string, artifact model.VerificationEvidenceArtifact) model.VerificationEvidenceCenter {
	status, okCount, watchCount, attentionCount := verificationEvidenceStatus(artifact)
	phase := emptyFallback(artifact.Phase, "current")
	headline := fmt.Sprintf("%s verification evidence is %s.", phase, status)
	if artifact.CanPromote && status == constants.StatusOK {
		headline = fmt.Sprintf("%s scripted gates are ready for publish or post-merge review.", phase)
	}
	detail := fmt.Sprintf("%d gates ok, %d watch, %d attention. Branch=%s head=%s.",
		okCount,
		watchCount,
		attentionCount,
		emptyFallback(artifact.Branch, "unknown"),
		emptyFallback(artifact.Head, "unknown"),
	)

	return model.VerificationEvidenceCenter{
		EvidenceAvailable: true,
		EvidencePath:      evidencePath,
		Source:            constants.VerificationEvidenceSourcePhase99,
		BaseURL:           artifact.BaseURL,
		Summary: model.VerificationEvidenceSummary{
			Status:            status,
			Headline:          headline,
			Detail:            detail,
			Phase:             phase,
			CanPromote:        artifact.CanPromote,
			GatesTotal:        len(artifact.Gates),
			GatesOK:           okCount,
			GatesWatch:        watchCount,
			GatesAttention:    attentionCount,
			ArtifactsTotal:    len(artifact.Artifacts),
			EvidenceGenerated: artifact.GeneratedAt,
		},
		Gates:           artifact.Gates,
		Proof:           verificationEvidenceProof(artifact, status, okCount, watchCount, attentionCount),
		Artifacts:       artifact.Artifacts,
		Actions:         verificationEvidenceActions(artifact, status),
		Artifact:        artifact,
		PrivacyBoundary: constants.VerificationEvidencePrivacyNote,
		GeneratedAt:     time.Now().UTC(),
	}
}

func verificationEvidenceStatus(artifact model.VerificationEvidenceArtifact) (string, int, int, int) {
	okCount := 0
	watchCount := 0
	attentionCount := 0
	for _, gate := range artifact.Gates {
		switch strings.ToLower(strings.TrimSpace(gate.Status)) {
		case constants.StatusOK, "passed", "pass", "success":
			okCount++
		case constants.StatusWatch, constants.StatusPending, "":
			watchCount++
		default:
			attentionCount++
		}
	}
	if attentionCount > 0 || !artifact.CanPromote || len(artifact.Gates) == 0 {
		return constants.StatusAttention, okCount, watchCount, attentionCount
	}
	if watchCount > 0 || artifact.OverallStatus == constants.StatusWatch {
		return constants.StatusWatch, okCount, watchCount, attentionCount
	}
	return constants.StatusOK, okCount, watchCount, attentionCount
}

func verificationEvidenceProof(artifact model.VerificationEvidenceArtifact, status string, okCount int, watchCount int, attentionCount int) []model.VerificationEvidenceProof {
	artifactStatus := constants.StatusOK
	if len(artifact.Artifacts) == 0 {
		artifactStatus = constants.StatusWatch
	}
	privacyStatus := constants.StatusAttention
	if artifact.Privacy.MetadataOnly && artifact.Privacy.SensitiveCollection == "denied" {
		privacyStatus = constants.StatusOK
	}
	gitStatus := constants.StatusOK
	if strings.TrimSpace(artifact.Branch) == "" || strings.TrimSpace(artifact.Head) == "" {
		gitStatus = constants.StatusWatch
	}

	return []model.VerificationEvidenceProof{
		{
			ID:            constants.VerificationEvidenceProofScriptedID,
			Label:         "Scripted Gates",
			Value:         fmt.Sprintf("%d/%d ok", okCount, len(artifact.Gates)),
			Detail:        fmt.Sprintf("watch=%d attention=%d can_promote=%t", watchCount, attentionCount, artifact.CanPromote),
			Status:        status,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.VerificationEvidenceProofReportsID,
			Label:         "Reports And Logs",
			Value:         fmt.Sprintf("%d artifacts", len(artifact.Artifacts)),
			Detail:        "Only local log/report paths are exposed; file contents are not embedded.",
			Status:        artifactStatus,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.VerificationEvidenceProofGitID,
			Label:         "Git Head",
			Value:         emptyFallback(artifact.Head, "unknown"),
			Detail:        fmt.Sprintf("branch=%s", emptyFallback(artifact.Branch, "unknown")),
			Status:        gitStatus,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.VerificationEvidenceProofPrivacyID,
			Label:         "Privacy Boundary",
			Value:         boolRuntimeLabel(privacyStatus == constants.StatusOK),
			Detail:        fmt.Sprintf("metadata_only=%t sensitive_collection=%s", artifact.Privacy.MetadataOnly, emptyFallback(artifact.Privacy.SensitiveCollection, "unknown")),
			Status:        privacyStatus,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
	}
}

func verificationEvidenceActions(artifact model.VerificationEvidenceArtifact, status string) []model.VerificationEvidenceAction {
	actions := make([]model.VerificationEvidenceAction, 0, len(artifact.Actions)+2)
	for _, action := range artifact.Actions {
		actions = append(actions, model.VerificationEvidenceAction{
			ID:       action.ID,
			Title:    action.Title,
			Detail:   action.Detail,
			Command:  action.Command,
			Severity: action.Severity,
			Status:   action.Status,
		})
	}
	if status != constants.StatusOK {
		actions = append(actions, model.VerificationEvidenceAction{
			ID:       constants.VerificationEvidenceActionRunGatesID,
			Title:    "Run phase verifier",
			Detail:   "Rerun the active phase verifier, then regenerate verification evidence.",
			Command:  constants.VerificationEvidenceVerifyCommand,
			Severity: constants.SeverityHigh,
			Status:   constants.StatusPending,
		})
	}
	if len(actions) == 0 {
		actions = append(actions, model.VerificationEvidenceAction{
			ID:       constants.VerificationEvidenceActionGenerateID,
			Title:    "Refresh verification evidence",
			Detail:   "Refresh the evidence artifact before publishing or handing off.",
			Command:  constants.VerificationEvidenceCommand,
			Severity: constants.SeverityInfo,
			Status:   constants.StatusOK,
		})
	}
	return actions
}

func (s *Server) operatorAssuranceCenter() (model.OperatorAssuranceCenter, error) {
	runtimeCenter, err := s.runtimeStatusCenter()
	if err != nil {
		return model.OperatorAssuranceCenter{}, fmt.Errorf("runtime status for operator assurance: %w", err)
	}
	evidenceCenter, err := s.verificationEvidenceCenter()
	if err != nil {
		return model.OperatorAssuranceCenter{}, fmt.Errorf("verification evidence for operator assurance: %w", err)
	}
	exportPath := strings.TrimSpace(s.operatorAssurancePath)
	if exportPath == "" {
		exportPath = constants.DefaultOperatorAssurancePath
	}
	return operatorAssuranceCenterFromSources(exportPath, runtimeCenter, evidenceCenter), nil
}

func operatorAssuranceCenterFromSources(exportPath string, runtimeCenter model.RuntimeStatusCenter, evidenceCenter model.VerificationEvidenceCenter) model.OperatorAssuranceCenter {
	runtimeReady := runtimeCenter.SummaryAvailable && runtimeCenter.Summary.CanContinue && runtimeCenter.Summary.RuntimeOK && runtimeCenter.Summary.HealthOK
	evidenceReady := evidenceCenter.EvidenceAvailable && evidenceCenter.Summary.CanPromote && evidenceCenter.Summary.Status == constants.StatusOK
	gitClean := runtimeCenter.SummaryAvailable && !runtimeCenter.Summary.TrackedContentDiff
	frontendCacheStatus := "artifact_cache_miss"
	frontendCacheHitPct := 0
	if runtimeCenter.Summary.FrontendURLPresent {
		frontendCacheStatus = "artifact_cache_hit"
		frontendCacheHitPct = 100
	}
	status := operatorAssuranceStatus(runtimeCenter, evidenceCenter, runtimeReady, evidenceReady, gitClean)
	canContinue := runtimeReady && evidenceReady && gitClean
	schedulerExplanation := operatorAssuranceSchedulerExplanation(runtimeCenter.Summary.SchedulerReadback, runtimeReady, runtimeCenter.Summary.LaunchTaskVerified)
	actions := operatorAssuranceActions(runtimeCenter, evidenceCenter, runtimeReady, evidenceReady, canContinue)
	nextStep := "Keep runtime summary and verification evidence fresh before handoff."
	if len(actions) > 0 {
		nextStep = actions[0].Detail
	}
	headline := "Operator assurance is ready for handoff."
	if status == constants.StatusWatch {
		headline = "Operator assurance can continue with watch items."
	} else if status == constants.StatusAttention {
		headline = "Operator assurance needs attention before handoff."
	}
	detail := fmt.Sprintf("Runtime=%s, Scheduler=%s, verification=%s, frontend cache=%s, git clean=%t.",
		boolRuntimeLabel(runtimeReady),
		emptyFallback(runtimeCenter.Summary.SchedulerReadback, "unknown"),
		emptyFallback(evidenceCenter.Summary.Status, "unknown"),
		frontendCacheStatus,
		gitClean,
	)
	updatedAt := runtimeCenter.Summary.SummaryGeneratedAt
	if strings.TrimSpace(evidenceCenter.Summary.EvidenceGenerated) != "" {
		updatedAt = evidenceCenter.Summary.EvidenceGenerated
	}

	return model.OperatorAssuranceCenter{
		Source:             constants.OperatorAssuranceSourcePhase100,
		BaseURL:            emptyFallback(runtimeCenter.BaseURL, evidenceCenter.BaseURL),
		RuntimeAvailable:   runtimeCenter.SummaryAvailable,
		EvidenceAvailable:  evidenceCenter.EvidenceAvailable,
		RuntimeSummaryPath: runtimeCenter.SummaryPath,
		EvidencePath:       evidenceCenter.EvidencePath,
		ExportPath:         exportPath,
		Summary: model.OperatorAssuranceSummary{
			Status:               status,
			Headline:             headline,
			Detail:               detail,
			CanContinue:          canContinue,
			CanPromote:           evidenceReady,
			RuntimeReady:         runtimeReady,
			SchedulerReadback:    emptyFallback(runtimeCenter.Summary.SchedulerReadback, "unknown"),
			SchedulerExplanation: schedulerExplanation,
			VerificationStatus:   emptyFallback(evidenceCenter.Summary.Status, "unknown"),
			FrontendCacheStatus:  frontendCacheStatus,
			FrontendCacheHitPct:  frontendCacheHitPct,
			GitClean:             gitClean,
			ExportPath:           exportPath,
			OperatorNextStep:     nextStep,
			UpdatedAt:            updatedAt,
		},
		Cards:           operatorAssuranceCards(runtimeCenter, evidenceCenter, runtimeReady, evidenceReady, gitClean, frontendCacheStatus),
		Actions:         actions,
		PrivacyBoundary: constants.OperatorAssurancePrivacyNote,
		GeneratedAt:     time.Now().UTC(),
	}
}

func operatorAssuranceStatus(runtimeCenter model.RuntimeStatusCenter, evidenceCenter model.VerificationEvidenceCenter, runtimeReady bool, evidenceReady bool, gitClean bool) string {
	if !runtimeCenter.SummaryAvailable || !evidenceCenter.EvidenceAvailable || !runtimeReady || !evidenceReady || !gitClean {
		return constants.StatusAttention
	}
	if runtimeCenter.Summary.Status == constants.StatusAttention || evidenceCenter.Summary.Status == constants.StatusAttention {
		return constants.StatusAttention
	}
	if runtimeCenter.Summary.Status == constants.StatusWatch || evidenceCenter.Summary.Status == constants.StatusWatch || !runtimeCenter.Summary.FrontendURLPresent {
		return constants.StatusWatch
	}
	return constants.StatusOK
}

func operatorAssuranceSchedulerExplanation(readback string, runtimeReady bool, launchTaskVerified bool) string {
	readback = strings.ToLower(strings.TrimSpace(readback))
	if launchTaskVerified || readback == "verified" {
		return "Scheduler readback is verified and the backend runtime proof is healthy."
	}
	if readback == "denied" && runtimeReady {
		return "Backend runtime is healthy; this non-elevated shell cannot read Scheduler metadata. Use elevated PowerShell only when service readback proof is required."
	}
	if runtimeReady {
		return "Backend runtime is healthy, but service-manager readback still needs a fresh local proof."
	}
	return "Backend runtime proof is not ready, so service-manager readback cannot be trusted yet."
}

func operatorAssuranceCards(runtimeCenter model.RuntimeStatusCenter, evidenceCenter model.VerificationEvidenceCenter, runtimeReady bool, evidenceReady bool, gitClean bool, frontendCacheStatus string) []model.OperatorAssuranceCard {
	schedulerStatus := constants.StatusAttention
	if runtimeCenter.Summary.LaunchTaskVerified || runtimeCenter.Summary.SchedulerReadback == "verified" {
		schedulerStatus = constants.StatusOK
	} else if runtimeReady && runtimeCenter.Summary.SchedulerReadback == "denied" {
		schedulerStatus = constants.StatusWatch
	}
	frontendStatus := constants.StatusWatch
	if runtimeCenter.Summary.FrontendURLPresent {
		frontendStatus = constants.StatusOK
	}
	privacyStatus := constants.StatusAttention
	if strings.Contains(strings.ToLower(runtimeCenter.PrivacyBoundary), "metadata-only") && strings.Contains(strings.ToLower(evidenceCenter.PrivacyBoundary), "metadata-only") {
		privacyStatus = constants.StatusOK
	}
	gitStatus := constants.StatusOK
	if !gitClean {
		gitStatus = constants.StatusAttention
	}
	verificationStatus := constants.StatusAttention
	if evidenceReady {
		verificationStatus = constants.StatusOK
	} else if evidenceCenter.EvidenceAvailable {
		verificationStatus = constants.StatusWatch
	}
	runtimeStatus := constants.StatusAttention
	if runtimeReady {
		runtimeStatus = constants.StatusOK
	}

	return []model.OperatorAssuranceCard{
		{
			ID:            constants.OperatorAssuranceCardRuntimeID,
			Label:         "Runtime Health",
			Value:         boolRuntimeLabel(runtimeReady),
			Detail:        fmt.Sprintf("backend=%t health=%t can_continue=%t", runtimeCenter.Summary.RuntimeOK, runtimeCenter.Summary.HealthOK, runtimeCenter.Summary.CanContinue),
			Status:        runtimeStatus,
			Source:        constants.RuntimeStatusSourcePhase97Summary,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.OperatorAssuranceCardSchedulerID,
			Label:         "Scheduler Readback",
			Value:         emptyFallback(runtimeCenter.Summary.SchedulerReadback, "unknown"),
			Detail:        operatorAssuranceSchedulerExplanation(runtimeCenter.Summary.SchedulerReadback, runtimeReady, runtimeCenter.Summary.LaunchTaskVerified),
			Status:        schedulerStatus,
			Source:        constants.RuntimeStatusSourcePhase97Summary,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.OperatorAssuranceCardVerificationID,
			Label:         "Verification Gates",
			Value:         fmt.Sprintf("%d/%d ok", evidenceCenter.Summary.GatesOK, evidenceCenter.Summary.GatesTotal),
			Detail:        fmt.Sprintf("watch=%d attention=%d can_promote=%t", evidenceCenter.Summary.GatesWatch, evidenceCenter.Summary.GatesAttention, evidenceCenter.Summary.CanPromote),
			Status:        verificationStatus,
			Source:        constants.VerificationEvidenceSourcePhase99,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.OperatorAssuranceCardFrontendCacheID,
			Label:         "Frontend Cache",
			Value:         frontendCacheStatus,
			Detail:        "Frontend status is read from the local runtime summary artifact; this endpoint does not perform a live cloud request.",
			Status:        frontendStatus,
			Source:        constants.RuntimeStatusSourcePhase97Summary,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.OperatorAssuranceCardGitID,
			Label:         "Git Hygiene",
			Value:         boolCleanLabel(gitClean),
			Detail:        fmt.Sprintf("tracked_content_diff=%t", runtimeCenter.Summary.TrackedContentDiff),
			Status:        gitStatus,
			Source:        constants.RuntimeStatusSourcePhase97Summary,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
		{
			ID:            constants.OperatorAssuranceCardPrivacyID,
			Label:         "Privacy Boundary",
			Value:         boolRuntimeLabel(privacyStatus == constants.StatusOK),
			Detail:        "Only local metadata proof is exposed; sensitive collection categories remain denied.",
			Status:        privacyStatus,
			Source:        constants.OperatorAssuranceSourcePhase100,
			EvidenceScope: constants.EvidenceScopeMetadataOnly,
		},
	}
}

func operatorAssuranceActions(runtimeCenter model.RuntimeStatusCenter, evidenceCenter model.VerificationEvidenceCenter, runtimeReady bool, evidenceReady bool, canContinue bool) []model.OperatorAssuranceAction {
	actions := []model.OperatorAssuranceAction{
		{
			ID:       constants.OperatorAssuranceActionRefreshID,
			Title:    "Refresh assurance pack",
			Detail:   "Regenerate the local operator assurance JSON and text handoff files.",
			Command:  constants.OperatorAssuranceCommand,
			Severity: constants.SeverityInfo,
			Status:   constants.StatusOK,
		},
	}
	if !runtimeReady {
		actions = append(actions, model.OperatorAssuranceAction{
			ID:       constants.OperatorAssuranceActionRestartID,
			Title:    "Restart backend runtime",
			Detail:   "Restart the local backend task, then refresh runtime summary and assurance.",
			Command:  constants.RuntimeTaskRestartCommand,
			Severity: constants.SeverityHigh,
			Status:   constants.StatusPending,
		})
	}
	if runtimeCenter.Summary.SchedulerReadback == "denied" && runtimeReady {
		actions = append(actions, model.OperatorAssuranceAction{
			ID:       constants.OperatorAssuranceActionAdminReadID,
			Title:    "Run elevated Scheduler readback",
			Detail:   "Use an elevated PowerShell session only if you need service-manager readback proof; runtime proof is already healthy.",
			Command:  constants.RuntimeTaskStatusCommand,
			Severity: constants.SeverityInfo,
			Status:   constants.StatusWatch,
		})
	}
	if !evidenceReady {
		actions = append(actions, model.OperatorAssuranceAction{
			ID:       constants.OperatorAssuranceActionRunVerifierID,
			Title:    "Run Phase 100 verifier",
			Detail:   "Run the full Phase 100 verifier and regenerate verification evidence.",
			Command:  constants.OperatorAssuranceVerifyCommand,
			Severity: constants.SeverityHigh,
			Status:   constants.StatusPending,
		})
	}
	if !runtimeCenter.SummaryAvailable || runtimeCenter.Summary.Status != constants.StatusOK {
		actions = append(actions, model.OperatorAssuranceAction{
			ID:       constants.OperatorAssuranceActionRuntimeID,
			Title:    "Refresh runtime summary",
			Detail:   "Refresh local runtime proof before sharing the operator assurance pack.",
			Command:  constants.RuntimeSummaryCommand,
			Severity: constants.SeverityInfo,
			Status:   constants.StatusWatch,
		})
	}
	if !evidenceCenter.EvidenceAvailable || evidenceCenter.Summary.Status != constants.StatusOK {
		actions = append(actions, model.OperatorAssuranceAction{
			ID:       constants.OperatorAssuranceActionEvidenceID,
			Title:    "Refresh verification evidence",
			Detail:   "Refresh local gate evidence after any verification run.",
			Command:  constants.VerificationEvidenceCommand,
			Severity: constants.SeverityInfo,
			Status:   constants.StatusWatch,
		})
	}
	if canContinue && len(actions) == 1 {
		actions[0].Status = constants.StatusOK
	}
	return actions
}

func (s *Server) handleDeviceEnroll(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeMethodNotAllowed(w)
		return
	}

	var req model.EnrollDeviceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid device enrollment json")
		return
	}
	if err := validateEnrollDeviceRequest(req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !tenantAllowed(r.Context(), req.TenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}

	device, err := s.store.EnrollDevice(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "device enrollment failed")
		return
	}
	writeJSON(w, http.StatusCreated, device)
}

func (s *Server) handleDevices(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != constants.RouteDevices {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	devices := s.store.ListDevices(r.Context())
	devices = filterDevicesForPrincipal(r.Context(), devices)
	writeJSON(w, http.StatusOK, model.ListResponse[model.Device]{
		Items: devices,
		Count: len(devices),
	})
}

func (s *Server) handleDeviceRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, constants.RouteDevices+"/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		http.NotFound(w, r)
		return
	}

	deviceID := parts[0]
	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		s.handleDevice(w, r, deviceID)
	case len(parts) == 3 && parts[1] == constants.RouteSegmentSummary && parts[2] == constants.RouteSegmentDaily && r.Method == http.MethodGet:
		s.handleDailySummary(w, r, deviceID)
	case len(parts) == 4 && parts[1] == constants.RouteSegmentReports && parts[2] == constants.RouteSegmentWeekly && parts[3] == constants.RouteSegmentPDF && r.Method == http.MethodGet:
		if !s.deviceAllowed(w, r, deviceID) {
			return
		}
		s.handleWeeklyReportPDF(w, r, deviceID)
	case len(parts) == 3 && parts[1] == constants.RouteSegmentReports && parts[2] == constants.RouteSegmentWeekly && r.Method == http.MethodGet:
		if !s.deviceAllowed(w, r, deviceID) {
			return
		}
		s.handleWeeklyReport(w, r, deviceID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentOverview && r.Method == http.MethodGet:
		s.handleHostOverview(w, r, deviceID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentHealth && r.Method == http.MethodGet:
		s.handleDeviceHealth(w, r, deviceID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentPolicyEvents && r.Method == http.MethodGet:
		s.handlePolicyViolations(w, r, deviceID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentAnomalies && r.Method == http.MethodGet:
		s.handleAnomalies(w, r, deviceID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentTamperEvents && r.Method == http.MethodGet:
		s.handleTamperEvents(w, r, deviceID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentAlertDelivery && r.Method == http.MethodGet:
		s.handleAlertDeliveries(w, r, deviceID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentTelemetry && r.Method == http.MethodPost:
		s.handleTelemetryIngest(w, r, deviceID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentTelemetryStatus && r.Method == http.MethodGet:
		s.handleTelemetryStatus(w, r, deviceID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleTenants(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != constants.RouteTenants {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodGet:
		tenants := s.store.ListTenants(r.Context())
		tenants = filterTenantsForPrincipal(r.Context(), tenants)
		writeJSON(w, http.StatusOK, model.ListResponse[model.Tenant]{
			Items: tenants,
			Count: len(tenants),
		})
	case http.MethodPost:
		var req model.CreateTenantRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid tenant json")
			return
		}
		if err := validateCreateTenantRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		if !tenantAllowed(r.Context(), req.TenantID) {
			writeError(w, http.StatusForbidden, "tenant scope is not allowed")
			return
		}
		tenant, err := s.store.CreateTenant(r.Context(), req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "tenant creation failed")
			return
		}
		writeJSON(w, http.StatusCreated, tenant)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleAccountPortfolioIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != constants.RouteAccountPortfolio {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	tenants := filterTenantsForPrincipal(r.Context(), s.store.ListTenants(r.Context()))
	tenantIDs := make([]string, 0, len(tenants))
	for _, tenant := range tenants {
		tenantIDs = append(tenantIDs, tenant.TenantID)
	}
	if strings.TrimSpace(requestPrincipal(r.Context()).TenantID) != "" && len(tenantIDs) == 0 {
		tenantIDs = append(tenantIDs, requestPrincipal(r.Context()).TenantID)
	}
	index, err := s.store.AccountPortfolioIndex(r.Context(), tenantIDs)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "account portfolio index lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, index)
}

func (s *Server) handleTenantRoutes(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, constants.RouteTenants+"/")
	parts := strings.Split(strings.Trim(rest, "/"), "/")
	if len(parts) == 0 || strings.TrimSpace(parts[0]) == "" {
		http.NotFound(w, r)
		return
	}

	tenantID := parts[0]
	switch {
	case len(parts) == 1 && r.Method == http.MethodGet:
		s.handleTenant(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentAuditEvents && r.Method == http.MethodGet:
		s.handleTenantAuditEvents(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentAlertRules:
		s.handleTenantAlertRules(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentNotifications:
		s.handleTenantNotificationRoutes(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentNotificationPref:
		s.handleTenantNotificationPreferences(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentConsentCenter && r.Method == http.MethodGet:
		s.handleTenantConsentCenter(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentAlertInbox && r.Method == http.MethodGet:
		s.handleTenantAlertInbox(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentOperations && r.Method == http.MethodGet:
		s.handleTenantOperationsSummary(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentMonetization && r.Method == http.MethodGet:
		s.handleTenantMonetizationSummary(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentBusinessDash && r.Method == http.MethodGet:
		s.handleTenantBusinessDashboard(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentOnboardingCenter && r.Method == http.MethodGet:
		s.handleTenantOnboardingCenter(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentCustomerSettings && r.Method == http.MethodGet:
		s.handleTenantCustomerSettingsCenter(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentRevenueOps && r.Method == http.MethodGet:
		s.handleTenantRevenueOperationsCenter(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentDeploymentReady && r.Method == http.MethodGet:
		s.handleTenantDeploymentReadinessCenter(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentPremiumOps && r.Method == http.MethodGet:
		s.handleTenantPremiumOperationsHub(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentRoleExperience && r.Method == http.MethodGet:
		s.handleTenantRoleExperiences(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentCustomerControl && r.Method == http.MethodGet:
		s.handleTenantCustomerControlRoom(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentSuccessPacket && r.Method == http.MethodGet:
		s.handleTenantCustomerSuccessPacket(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentPushActivation && r.Method == http.MethodGet:
		s.handleTenantPushActivationCenter(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentPortfolioCenter && r.Method == http.MethodGet:
		s.handleTenantPortfolioCenter(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentExecutiveConsole && r.Method == http.MethodGet:
		s.handleTenantExecutiveConsole(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentNotificationRev && r.Method == http.MethodGet:
		s.handleTenantNotificationRevenueCockpit(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentProviderSim:
		s.handleTenantProviderSimulationLab(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentProviderSetup && r.Method == http.MethodGet:
		s.handleTenantNotificationProviderSetup(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentPackageBilling && r.Method == http.MethodGet:
		s.handleTenantPackageBillingReadiness(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentNotificationCmd && r.Method == http.MethodGet:
		s.handleTenantNotificationCommandCenter(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentDeliveryTimeline && r.Method == http.MethodGet:
		s.handleTenantDeliveryTimeline(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentDeliveryAssure && r.Method == http.MethodGet:
		s.handleTenantDeliveryAssurance(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentDeliveryDrill:
		s.handleTenantDeliveryDrilldown(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentDeliveryRemedy:
		s.handleTenantDeliveryRemediation(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentSyncHealth && r.Method == http.MethodGet:
		s.handleTenantSyncHealth(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentActivityFeed && r.Method == http.MethodGet:
		s.handleTenantActivityFeed(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentBrowserActivity && r.Method == http.MethodGet:
		s.handleTenantBrowserActivity(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentActivityViews:
		s.handleTenantActivityViews(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentDataExports:
		s.handleTenantDataExports(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentDeleteRequests:
		s.handleTenantDeleteRequests(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentDeviceGroups:
		s.handleTenantDeviceGroups(w, r, tenantID)
	case len(parts) == 2 && parts[1] == constants.RouteSegmentPolicyAssign:
		s.handleTenantPolicyAssignments(w, r, tenantID)
	default:
		http.NotFound(w, r)
	}
}

func (s *Server) handleTenant(w http.ResponseWriter, r *http.Request, tenantID string) {
	tenant, err := s.store.GetTenant(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant lookup failed")
		return
	}
	if !tenantAllowed(r.Context(), tenant.TenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	writeJSON(w, http.StatusOK, tenant)
}

func (s *Server) handleTenantAuditEvents(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	if _, err := s.store.GetTenant(r.Context(), tenantID); err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant lookup failed")
		return
	}
	events := s.store.ListAuditEvents(r.Context(), tenantID)
	events = filterAuditEventsForPrincipal(r.Context(), events)
	writeJSON(w, http.StatusOK, model.ListResponse[model.AuditEvent]{
		Items: events,
		Count: len(events),
	})
}

func (s *Server) handleTenantAlertRules(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	if _, err := s.store.GetTenant(r.Context(), tenantID); err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant lookup failed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		rules := s.store.ListAlertRules(r.Context(), tenantID)
		writeJSON(w, http.StatusOK, model.ListResponse[model.AlertRule]{
			Items: rules,
			Count: len(rules),
		})
	case http.MethodPost:
		var req model.CreateAlertRuleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid alert rule json")
			return
		}
		if err := validateCreateAlertRuleRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		rule, err := s.store.CreateAlertRule(r.Context(), tenantID, req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "alert rule creation failed")
			return
		}
		writeJSON(w, http.StatusCreated, rule)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleTenantNotificationRoutes(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	if _, err := s.store.GetTenant(r.Context(), tenantID); err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant lookup failed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		routes := s.store.ListNotificationRoutes(r.Context(), tenantID)
		writeJSON(w, http.StatusOK, model.ListResponse[model.NotificationRoute]{
			Items: routes,
			Count: len(routes),
		})
	case http.MethodPost:
		var req model.CreateNotificationRouteRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid notification route json")
			return
		}
		if err := validateCreateNotificationRouteRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		route, err := s.store.CreateNotificationRoute(r.Context(), tenantID, req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "notification route creation failed")
			return
		}
		writeJSON(w, http.StatusCreated, route)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleTenantNotificationPreferences(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	if _, err := s.store.GetTenant(r.Context(), tenantID); err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant lookup failed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		preferences, err := s.store.TenantNotificationPreferences(r.Context(), tenantID)
		if err != nil {
			if errors.Is(err, store.ErrTenantNotFound) {
				writeError(w, http.StatusNotFound, "tenant not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "tenant notification preference lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, preferences)
	case http.MethodPost:
		var req model.UpdateNotificationPreferencesRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid notification preference json")
			return
		}
		if err := validateUpdateNotificationPreferencesRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		preferences, err := s.store.UpdateTenantNotificationPreferences(r.Context(), tenantID, req)
		if err != nil {
			if errors.Is(err, store.ErrTenantNotFound) {
				writeError(w, http.StatusNotFound, "tenant not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "tenant notification preference update failed")
			return
		}
		writeJSON(w, http.StatusOK, preferences)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleTenantConsentCenter(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	tenant, err := s.store.GetTenant(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant lookup failed")
		return
	}
	auditEvents := s.store.ListAuditEvents(r.Context(), tenantID)
	auditEvents = filterAuditEventsForPrincipal(r.Context(), auditEvents)
	rules := s.store.ListAlertRules(r.Context(), tenantID)
	writeJSON(w, http.StatusOK, buildConsentCenter(tenant, auditEvents, rules))
}

func (s *Server) handleTenantOperationsSummary(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	summary, err := s.store.TenantOperationsSummary(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant operations lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleTenantAlertInbox(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	inbox, err := s.store.TenantAlertInbox(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant alert inbox lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, inbox)
}

func (s *Server) handleTenantMonetizationSummary(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	summary, err := s.store.TenantMonetizationSummary(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant monetization lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleTenantBusinessDashboard(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	dashboard, err := s.store.TenantBusinessDashboard(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant business dashboard lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, dashboard)
}

func (s *Server) handleTenantRoleExperiences(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	experience, err := s.store.TenantRoleExperiences(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant role experiences lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, experience)
}

func (s *Server) handleTenantOnboardingCenter(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	center, err := s.store.TenantOnboardingCenter(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant onboarding center lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, center)
}

func (s *Server) handleTenantCustomerSettingsCenter(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	center, err := s.store.TenantCustomerSettingsCenter(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant customer settings center lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, center)
}

func (s *Server) handleTenantRevenueOperationsCenter(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	center, err := s.store.TenantRevenueOperationsCenter(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant revenue operations center lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, center)
}

func (s *Server) handleTenantDeploymentReadinessCenter(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	center, err := s.store.TenantDeploymentReadinessCenter(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		s.logger.Error("tenant deployment readiness center lookup failed", "tenant_id", tenantID, "error", err)
		writeError(w, http.StatusInternalServerError, "tenant deployment readiness center lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, center)
}

func (s *Server) handleTenantPremiumOperationsHub(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	hub, err := s.store.TenantPremiumOperationsHub(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant premium operations hub lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, hub)
}

func (s *Server) handleTenantExecutiveConsole(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	console, err := s.store.TenantExecutiveConsole(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant executive console lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, console)
}

func (s *Server) handleTenantCustomerControlRoom(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	room, err := s.store.TenantCustomerControlRoom(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant customer control room lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, room)
}

func (s *Server) handleTenantCustomerSuccessPacket(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	packet, err := s.store.TenantCustomerSuccessPacket(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant customer success packet lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, packet)
}

func (s *Server) handleTenantPushActivationCenter(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	center, err := s.store.TenantPushActivationCenter(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant push activation center lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, center)
}

func (s *Server) handleTenantPortfolioCenter(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	center, err := s.store.TenantPortfolioCenter(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant portfolio center lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, center)
}

func (s *Server) handleTenantNotificationRevenueCockpit(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	cockpit, err := s.store.TenantNotificationRevenueCockpit(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant notification revenue cockpit lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, cockpit)
}

func (s *Server) handleTenantNotificationCommandCenter(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	commandCenter, err := s.store.TenantNotificationCommandCenter(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant notification command center lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, commandCenter)
}

func (s *Server) handleTenantDeliveryTimeline(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	filter := deliveryTimelineFilterFromQuery(r)
	if filter.Channel != "" && !knownChannels([]string{filter.Channel}) {
		writeError(w, http.StatusBadRequest, "unknown delivery timeline channel")
		return
	}
	if filter.Status != "" && !knownDeliveryTimelineStatus(filter.Status) {
		writeError(w, http.StatusBadRequest, "unknown delivery timeline status")
		return
	}
	if filter.Provider != "" && !knownDeliveryProvider(filter.Provider) {
		writeError(w, http.StatusBadRequest, "unknown delivery timeline provider")
		return
	}
	timeline, err := s.store.TenantDeliveryTimeline(r.Context(), tenantID, filter)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant delivery timeline lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, timeline)
}

func (s *Server) handleTenantDeliveryAssurance(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	filter := deliveryAssuranceFilterFromQuery(r)
	if filter.Channel != "" && !knownChannels([]string{filter.Channel}) {
		writeError(w, http.StatusBadRequest, "unknown delivery assurance channel")
		return
	}
	if filter.AssuranceState != "" && !knownDeliveryAssuranceState(filter.AssuranceState) {
		writeError(w, http.StatusBadRequest, "unknown delivery assurance state")
		return
	}
	assurance, err := s.store.TenantDeliveryAssurance(r.Context(), tenantID, filter)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant delivery assurance lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, assurance)
}

func (s *Server) handleTenantDeliveryDrilldown(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		drilldown, err := s.store.TenantDeliveryDrilldown(r.Context(), tenantID)
		if err != nil {
			if errors.Is(err, store.ErrTenantNotFound) {
				writeError(w, http.StatusNotFound, "tenant not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "tenant delivery drilldown lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, drilldown)
	case http.MethodPost:
		var req model.RunDeliveryDrilldownRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid delivery drilldown json")
			return
		}
		if err := validateRunDeliveryDrilldownRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		drilldown, err := s.store.RunTenantDeliveryDrilldown(r.Context(), tenantID, req)
		if err != nil {
			if errors.Is(err, store.ErrTenantNotFound) {
				writeError(w, http.StatusNotFound, "tenant not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "tenant delivery drilldown failed")
			return
		}
		writeJSON(w, http.StatusAccepted, drilldown)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleTenantDeliveryRemediation(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		remediation, err := s.store.TenantDeliveryRemediation(r.Context(), tenantID)
		if err != nil {
			if errors.Is(err, store.ErrTenantNotFound) {
				writeError(w, http.StatusNotFound, "tenant not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "tenant delivery remediation lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, remediation)
	case http.MethodPost:
		var req model.RunDeliveryRemediationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid delivery remediation json")
			return
		}
		if err := validateRunDeliveryRemediationRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		remediation, err := s.store.RunTenantDeliveryRemediation(r.Context(), tenantID, req)
		if err != nil {
			if errors.Is(err, store.ErrTenantNotFound) {
				writeError(w, http.StatusNotFound, "tenant not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "tenant delivery remediation failed")
			return
		}
		writeJSON(w, http.StatusAccepted, remediation)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleTenantSyncHealth(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	summary, err := s.store.TenantSyncHealth(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant sync health lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleTenantActivityFeed(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	feed, err := s.store.TenantActivityFeed(r.Context(), tenantID, activityFeedFilterFromQuery(r))
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant activity feed lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, feed)
}

func (s *Server) handleTenantBrowserActivity(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	filter, err := browserActivityFilterFromQuery(r)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if filter.Browser != "" && !knownBrowser(filter.Browser) {
		writeError(w, http.StatusBadRequest, "unknown browser activity browser")
		return
	}
	if filter.Category != "" && !knownBrowserCategory(filter.Category) {
		writeError(w, http.StatusBadRequest, "unknown browser activity category")
		return
	}
	viewer, err := s.store.TenantBrowserActivity(r.Context(), tenantID, filter)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant browser activity lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, viewer)
}

func (s *Server) handleTenantActivityViews(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	if _, err := s.store.GetTenant(r.Context(), tenantID); err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant lookup failed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		views := s.store.ListTenantActivityViews(r.Context(), tenantID)
		writeJSON(w, http.StatusOK, model.ListResponse[model.TenantActivityView]{
			Items: views,
			Count: len(views),
		})
	case http.MethodPost:
		var req model.CreateTenantActivityViewRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid activity view json")
			return
		}
		if err := validateCreateTenantActivityViewRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		view, err := s.store.CreateTenantActivityView(r.Context(), tenantID, req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "activity view creation failed")
			return
		}
		writeJSON(w, http.StatusCreated, view)
	default:
		writeMethodNotAllowed(w)
	}
}

func activityFeedFilterFromQuery(r *http.Request) model.TenantActivityFeedFilter {
	query := r.URL.Query()
	limit := 0
	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil {
			limit = parsed
		}
	}
	return model.TenantActivityFeedFilter{
		DeviceID:    strings.TrimSpace(query.Get("device_id")),
		Kind:        strings.TrimSpace(query.Get("kind")),
		Severity:    strings.TrimSpace(query.Get("severity")),
		Channel:     strings.TrimSpace(query.Get("channel")),
		Status:      strings.TrimSpace(query.Get("status")),
		Query:       strings.TrimSpace(query.Get("q")),
		IncludeDemo: requestIncludesDemoEvidence(r),
		Limit:       limit,
	}
}

func browserActivityFilterFromQuery(r *http.Request) (model.TenantBrowserActivityFilter, error) {
	query := r.URL.Query()
	limit := 0
	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil {
			limit = parsed
		}
	}
	var studySafe *bool
	if rawStudySafe := strings.TrimSpace(query.Get("study_safe")); rawStudySafe != "" {
		parsed, err := strconv.ParseBool(rawStudySafe)
		if err != nil {
			return model.TenantBrowserActivityFilter{}, fmt.Errorf("invalid study_safe value")
		}
		studySafe = &parsed
	}
	return model.TenantBrowserActivityFilter{
		DeviceID:  strings.TrimSpace(query.Get("device_id")),
		Browser:   strings.TrimSpace(query.Get("browser")),
		Category:  strings.TrimSpace(query.Get("category")),
		Domain:    strings.TrimSpace(query.Get("domain")),
		StudySafe: studySafe,
		Query:     strings.TrimSpace(query.Get("q")),
		Limit:     limit,
	}, nil
}

func (s *Server) handleTenantProviderSimulationLab(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		lab, err := s.store.TenantProviderSimulationLab(r.Context(), tenantID)
		if err != nil {
			if errors.Is(err, store.ErrTenantNotFound) {
				writeError(w, http.StatusNotFound, "tenant not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "tenant provider simulation lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, lab)
	case http.MethodPost:
		var req model.RunProviderSimulationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid provider simulation json")
			return
		}
		if err := validateRunProviderSimulationRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		lab, err := s.store.RunTenantProviderSimulation(r.Context(), tenantID, req)
		if err != nil {
			if errors.Is(err, store.ErrTenantNotFound) {
				writeError(w, http.StatusNotFound, "tenant not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "tenant provider simulation failed")
			return
		}
		writeJSON(w, http.StatusAccepted, lab)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleTenantNotificationProviderSetup(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	setup, err := s.store.TenantNotificationProviderSetup(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant notification provider setup lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, setup)
}

func (s *Server) handleTenantPackageBillingReadiness(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	readiness, err := s.store.TenantPackageBillingReadiness(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant package billing readiness lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, readiness)
}

func deliveryTimelineFilterFromQuery(r *http.Request) model.TenantDeliveryTimelineFilter {
	query := r.URL.Query()
	limit := 0
	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil {
			limit = parsed
		}
	}
	return model.TenantDeliveryTimelineFilter{
		DeviceID: strings.TrimSpace(query.Get("device_id")),
		Channel:  strings.ToLower(strings.TrimSpace(query.Get("channel"))),
		Status:   strings.ToLower(strings.TrimSpace(query.Get("status"))),
		Provider: strings.ToLower(strings.TrimSpace(query.Get("provider"))),
		Query:    strings.TrimSpace(query.Get("q")),
		Limit:    limit,
	}
}

func deliveryAssuranceFilterFromQuery(r *http.Request) model.TenantDeliveryAssuranceFilter {
	query := r.URL.Query()
	limit := 0
	if rawLimit := strings.TrimSpace(query.Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err == nil {
			limit = parsed
		}
	}
	return model.TenantDeliveryAssuranceFilter{
		DeviceID:       strings.TrimSpace(query.Get("device_id")),
		Channel:        strings.ToLower(strings.TrimSpace(query.Get("channel"))),
		AssuranceState: strings.ToLower(strings.TrimSpace(query.Get("assurance_state"))),
		Limit:          limit,
	}
}

func (s *Server) handleTenantDataExports(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	if _, err := s.store.GetTenant(r.Context(), tenantID); err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant lookup failed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		exports := s.store.ListTenantDataExports(r.Context(), tenantID)
		writeJSON(w, http.StatusOK, model.ListResponse[model.TenantDataExport]{
			Items: exports,
			Count: len(exports),
		})
	case http.MethodPost:
		var req model.CreateTenantDataExportRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid data export json")
			return
		}
		if err := validateCreateTenantDataExportRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		export, err := s.store.CreateTenantDataExport(r.Context(), tenantID, req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "data export creation failed")
			return
		}
		writeJSON(w, http.StatusCreated, export)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleTenantDeleteRequests(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	if _, err := s.store.GetTenant(r.Context(), tenantID); err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant lookup failed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		requests := s.store.ListDeleteRequests(r.Context(), tenantID)
		writeJSON(w, http.StatusOK, model.ListResponse[model.DeleteRequest]{
			Items: requests,
			Count: len(requests),
		})
	case http.MethodPost:
		var req model.CreateDeleteRequestRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid delete request json")
			return
		}
		if err := validateCreateDeleteRequestRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		deleteRequest, err := s.store.CreateDeleteRequest(r.Context(), tenantID, req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "delete request creation failed")
			return
		}
		writeJSON(w, http.StatusCreated, deleteRequest)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleTenantDeviceGroups(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	if _, err := s.store.GetTenant(r.Context(), tenantID); err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant lookup failed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		groups := s.store.ListDeviceGroups(r.Context(), tenantID)
		writeJSON(w, http.StatusOK, model.ListResponse[model.DeviceGroup]{
			Items: groups,
			Count: len(groups),
		})
	case http.MethodPost:
		var req model.CreateDeviceGroupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid device group json")
			return
		}
		if err := validateCreateDeviceGroupRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		group, err := s.store.CreateDeviceGroup(r.Context(), tenantID, req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "device group creation failed")
			return
		}
		writeJSON(w, http.StatusCreated, group)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleTenantPolicyAssignments(w http.ResponseWriter, r *http.Request, tenantID string) {
	if !tenantAllowed(r.Context(), tenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	if _, err := s.store.GetTenant(r.Context(), tenantID); err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tenant lookup failed")
		return
	}

	switch r.Method {
	case http.MethodGet:
		assignments := s.store.ListPolicyAssignments(r.Context(), tenantID)
		writeJSON(w, http.StatusOK, model.ListResponse[model.PolicyAssignment]{
			Items: assignments,
			Count: len(assignments),
		})
	case http.MethodPost:
		var req model.CreatePolicyAssignmentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid policy assignment json")
			return
		}
		if err := validateCreatePolicyAssignmentRequest(req); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		assignment, err := s.store.CreatePolicyAssignment(r.Context(), tenantID, req)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "policy assignment creation failed")
			return
		}
		writeJSON(w, http.StatusCreated, assignment)
	default:
		writeMethodNotAllowed(w)
	}
}

func (s *Server) handleDevice(w http.ResponseWriter, r *http.Request, deviceID string) {
	device, err := s.store.GetDevice(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "device lookup failed")
		return
	}
	if !tenantAllowed(r.Context(), device.TenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	writeJSON(w, http.StatusOK, device)
}

func (s *Server) handleDailySummary(w http.ResponseWriter, r *http.Request, deviceID string) {
	if !s.deviceAllowed(w, r, deviceID) {
		return
	}
	summary, err := s.store.DailySummary(r.Context(), deviceID, r.URL.Query().Get("date"))
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "daily summary lookup failed")
		return
	}
	if !requestIncludesDemoEvidence(r) {
		summary = withoutDemoDeviceSummary(summary)
	}
	writeJSON(w, http.StatusOK, summary)
}

func (s *Server) handleHostOverview(w http.ResponseWriter, r *http.Request, deviceID string) {
	if !s.deviceAllowed(w, r, deviceID) {
		return
	}
	overview, err := s.store.HostOverview(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "host overview lookup failed")
		return
	}
	if !requestIncludesDemoEvidence(r) {
		overview = withoutDemoHostOverviewEvidence(overview)
	}
	writeJSON(w, http.StatusOK, overview)
}

func (s *Server) handlePolicyViolations(w http.ResponseWriter, r *http.Request, deviceID string) {
	if !s.deviceAllowed(w, r, deviceID) {
		return
	}
	events, err := s.store.ListPolicyViolations(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "policy violation lookup failed")
		return
	}
	if !requestIncludesDemoEvidence(r) {
		events = withoutDemoRiskEvents(events)
	}
	writeJSON(w, http.StatusOK, model.ListResponse[model.RiskEvent]{Items: events, Count: len(events)})
}

func (s *Server) handleDeviceHealth(w http.ResponseWriter, r *http.Request, deviceID string) {
	if !s.deviceAllowed(w, r, deviceID) {
		return
	}
	health, err := s.store.DeviceHealth(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "device health lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, health)
}

func (s *Server) handleAnomalies(w http.ResponseWriter, r *http.Request, deviceID string) {
	if !s.deviceAllowed(w, r, deviceID) {
		return
	}
	events, err := s.store.ListAnomalies(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "anomaly lookup failed")
		return
	}
	if !requestIncludesDemoEvidence(r) {
		events = withoutDemoRiskEvents(events)
	}
	writeJSON(w, http.StatusOK, model.ListResponse[model.RiskEvent]{Items: events, Count: len(events)})
}

func (s *Server) handleTamperEvents(w http.ResponseWriter, r *http.Request, deviceID string) {
	if !s.deviceAllowed(w, r, deviceID) {
		return
	}
	events, err := s.store.ListTamperEvents(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "tamper event lookup failed")
		return
	}
	if !requestIncludesDemoEvidence(r) {
		events = withoutDemoRiskEvents(events)
	}
	writeJSON(w, http.StatusOK, model.ListResponse[model.RiskEvent]{Items: events, Count: len(events)})
}

func (s *Server) handleAlertDeliveries(w http.ResponseWriter, r *http.Request, deviceID string) {
	if !s.deviceAllowed(w, r, deviceID) {
		return
	}
	deliveries, err := s.store.ListAlertDeliveries(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "alert delivery lookup failed")
		return
	}
	if !requestIncludesDemoEvidence(r) {
		deliveries = withoutDemoAlertDeliveries(deliveries)
	}
	writeJSON(w, http.StatusOK, model.ListResponse[model.AlertDelivery]{Items: deliveries, Count: len(deliveries)})
}

func (s *Server) handleTelemetryIngest(w http.ResponseWriter, r *http.Request, deviceID string) {
	var req model.IngestTelemetryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid telemetry ingest json")
		return
	}
	if err := validateIngestTelemetryRequest(deviceID, req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if !tenantAllowed(r.Context(), req.TenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return
	}
	response, err := s.store.IngestTelemetryEvents(r.Context(), deviceID, req)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrTenantNotFound):
			writeError(w, http.StatusNotFound, "tenant not found")
		case errors.Is(err, store.ErrDeviceNotFound):
			writeError(w, http.StatusNotFound, "device not found")
		default:
			writeError(w, http.StatusInternalServerError, "telemetry ingest failed")
		}
		return
	}
	writeJSON(w, http.StatusAccepted, response)
}

func (s *Server) handleTelemetryStatus(w http.ResponseWriter, r *http.Request, deviceID string) {
	if !s.deviceAllowed(w, r, deviceID) {
		return
	}
	status, err := s.store.TelemetryIngestStatus(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "telemetry status lookup failed")
		return
	}
	writeJSON(w, http.StatusOK, status)
}

func (s *Server) handleWeeklyReport(w http.ResponseWriter, r *http.Request, deviceID string) {
	overview, err := s.store.HostOverview(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "weekly report lookup failed")
		return
	}
	if !requestIncludesDemoEvidence(r) {
		overview = withoutDemoHostOverviewEvidence(overview)
	}
	writeJSON(w, http.StatusOK, store.WeeklyReport(overview))
}

func (s *Server) handleWeeklyReportPDF(w http.ResponseWriter, r *http.Request, deviceID string) {
	overview, err := s.store.HostOverview(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "weekly report pdf lookup failed")
		return
	}
	if !requestIncludesDemoEvidence(r) {
		overview = withoutDemoHostOverviewEvidence(overview)
	}
	report := store.WeeklyReport(overview)
	data := store.WeeklyReportPDF(report)
	w.Header().Set("Content-Type", constants.ContentTypePDF)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"tracedeck-weekly-%s.pdf\"", safeAttachmentFilenamePart(report.DeviceID)))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data) // #nosec G705 -- PDF bytes are generated server-side and text is escaped by store.WeeklyReportPDF.
}

func safeAttachmentFilenamePart(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "device"
	}
	var builder strings.Builder
	for _, r := range value {
		switch {
		case r >= 'a' && r <= 'z':
			builder.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			builder.WriteRune(r)
		case r >= '0' && r <= '9':
			builder.WriteRune(r)
		case r == '-', r == '_', r == '.':
			builder.WriteRune(r)
		default:
			builder.WriteByte('-')
		}
	}
	clean := strings.Trim(builder.String(), "-_.")
	if clean == "" {
		return "device"
	}
	return clean
}

func requestIncludesDemoEvidence(r *http.Request) bool {
	value := strings.ToLower(strings.TrimSpace(r.URL.Query().Get(constants.QueryIncludeDemo)))
	switch value {
	case constants.QueryValueTrue, constants.QueryValueOne, constants.QueryValueYes:
		return true
	default:
		return false
	}
}

func withoutDemoHostOverviewEvidence(overview model.HostOverview) model.HostOverview {
	originalTamperCount := len(overview.TamperEvents)
	overview.PolicyViolations = withoutDemoRiskEvents(overview.PolicyViolations)
	overview.Anomalies = withoutDemoRiskEvents(overview.Anomalies)
	overview.TamperEvents = withoutDemoRiskEvents(overview.TamperEvents)
	overview.AlertDeliveries = withoutDemoAlertDeliveries(overview.AlertDeliveries)
	overview.Summary = withoutDemoDeviceSummary(overview.Summary)
	overview.Summary.PolicyViolations = len(overview.PolicyViolations)
	overview.Summary.AlertsRaised = len(overview.AlertDeliveries)
	overview.RiskScore = riskScoreForVisibleEvents(overview.PolicyViolations, overview.Anomalies, overview.TamperEvents)
	overview.RiskLevel = riskLevelForVisibleScore(overview.RiskScore)
	overview.Summary.ComplianceScore = complianceScoreForVisibleRisk(overview.RiskScore)
	if originalTamperCount > 0 && len(overview.TamperEvents) == 0 {
		overview.Summary.ArchiveBacklog = constants.RiskScoreNone
		overview.Archive.PendingBatches = constants.RiskScoreNone
		overview.Archive.Status = constants.StatusHealthy
	}
	return overview
}

func withoutDemoDeviceSummary(summary model.DeviceSummary) model.DeviceSummary {
	summary.StudyMinutes = constants.SummaryMetricNone
	summary.CodingMinutes = constants.SummaryMetricNone
	summary.EntertainmentMins = constants.SummaryMetricNone
	summary.PolicyViolations = constants.SummaryMetricNone
	summary.ArchiveBacklog = constants.SummaryMetricNone
	summary.AlertsRaised = constants.SummaryMetricNone
	summary.ComplianceScore = constants.ComplianceScoreClean
	summary.DataCompletenessPct = constants.DataCompletenessUnknownPct
	return summary
}

func withoutDemoRiskEvents(events []model.RiskEvent) []model.RiskEvent {
	filtered := make([]model.RiskEvent, 0, len(events))
	for _, event := range events {
		if isDemoEvidence(event.SourceKind) {
			continue
		}
		filtered = append(filtered, event)
	}
	return filtered
}

func withoutDemoAlertDeliveries(deliveries []model.AlertDelivery) []model.AlertDelivery {
	filtered := make([]model.AlertDelivery, 0, len(deliveries))
	for _, delivery := range deliveries {
		if isDemoEvidence(delivery.SourceKind) {
			continue
		}
		filtered = append(filtered, delivery)
	}
	return filtered
}

func isDemoEvidence(sourceKind string) bool {
	return strings.EqualFold(strings.TrimSpace(sourceKind), constants.EvidenceSourceDemoSeed)
}

func riskScoreForVisibleEvents(groups ...[]model.RiskEvent) int {
	score := constants.RiskScoreNone
	count := constants.RiskScoreNone
	for _, events := range groups {
		for _, event := range events {
			if event.Status == constants.RiskStatusResolved {
				continue
			}
			count++
			eventScore := riskScoreForSeverity(event.Severity)
			if eventScore > score {
				score = eventScore
			}
		}
	}
	if count > 1 {
		score += (count - 1) * constants.RiskScoreCountPenalty
	}
	if score > constants.RiskScoreMaximum {
		return constants.RiskScoreMaximum
	}
	return score
}

func riskScoreForSeverity(severity string) int {
	switch severity {
	case constants.SeverityCritical:
		return constants.RiskScoreCritical
	case constants.SeverityHigh:
		return constants.RiskScoreHigh
	case constants.SeverityMedium:
		return constants.RiskScoreMedium
	case constants.SeverityLow:
		return constants.RiskScoreLow
	case constants.SeverityInfo:
		return constants.RiskScoreInfo
	default:
		return constants.RiskScoreNone
	}
}

func riskLevelForVisibleScore(score int) string {
	switch {
	case score >= constants.RiskScoreHighThreshold:
		return constants.RiskLevelHigh
	case score >= constants.RiskScoreMediumThreshold:
		return constants.RiskLevelMedium
	default:
		return constants.RiskLevelLow
	}
}

func complianceScoreForVisibleRisk(score int) int {
	compliance := constants.ComplianceScoreClean - score
	if compliance < constants.RiskScoreNone {
		return constants.RiskScoreNone
	}
	return compliance
}

func (s *Server) handlePolicyTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	templates := store.PolicyTemplates()
	writeJSON(w, http.StatusOK, model.ListResponse[model.PolicyTemplate]{
		Items: templates,
		Count: len(templates),
	})
}

func (s *Server) handleAlertRuleTemplates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	templates := store.AlertRuleTemplates()
	writeJSON(w, http.StatusOK, model.ListResponse[model.AlertRuleTemplate]{
		Items: templates,
		Count: len(templates),
	})
}

func (s *Server) handlePlans(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	plans := store.Plans()
	writeJSON(w, http.StatusOK, model.ListResponse[model.Plan]{
		Items: plans,
		Count: len(plans),
	})
}

func (s *Server) handleRoles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	roles := store.Roles()
	writeJSON(w, http.StatusOK, model.ListResponse[model.Role]{
		Items: roles,
		Count: len(roles),
	})
}

func (s *Server) handleRetentionTiers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	tiers := store.RetentionTiers()
	writeJSON(w, http.StatusOK, model.ListResponse[model.RetentionTier]{
		Items: tiers,
		Count: len(tiers),
	})
}

func (s *Server) handleAuditEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	events := s.store.ListAuditEvents(r.Context(), "")
	events = filterAuditEventsForPrincipal(r.Context(), events)
	writeJSON(w, http.StatusOK, model.ListResponse[model.AuditEvent]{
		Items: events,
		Count: len(events),
	})
}

func (s *Server) handleArchiveStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, store.ArchiveStatus())
}

func (s *Server) deviceAllowed(w http.ResponseWriter, r *http.Request, deviceID string) bool {
	device, err := s.store.GetDevice(r.Context(), deviceID)
	if err != nil {
		if errors.Is(err, store.ErrDeviceNotFound) {
			writeError(w, http.StatusNotFound, "device not found")
			return false
		}
		writeError(w, http.StatusInternalServerError, "device lookup failed")
		return false
	}
	if !tenantAllowed(r.Context(), device.TenantID) {
		writeError(w, http.StatusForbidden, "tenant scope is not allowed")
		return false
	}
	return true
}

func validateCreateTenantRequest(req model.CreateTenantRequest) error {
	switch {
	case strings.TrimSpace(req.TenantID) == "":
		return errors.New("tenant_id is required")
	case strings.TrimSpace(req.Name) == "":
		return errors.New("name is required")
	case strings.TrimSpace(req.PlanID) == "":
		return errors.New("plan_id is required")
	case !store.KnownPlanID(req.PlanID):
		return errors.New("plan_id is unknown")
	case strings.TrimSpace(req.RetentionTierID) == "":
		return errors.New("retention_tier_id is required")
	case !store.KnownRetentionTierID(req.RetentionTierID):
		return errors.New("retention_tier_id is unknown")
	case strings.TrimSpace(req.PrimaryProfile) == "":
		return errors.New("primary_profile is required")
	default:
		return nil
	}
}

func validateEnrollDeviceRequest(req model.EnrollDeviceRequest) error {
	switch {
	case strings.TrimSpace(req.TenantID) == "":
		return errors.New("tenant_id is required")
	case strings.TrimSpace(req.DeviceID) == "":
		return errors.New("device_id is required")
	case strings.TrimSpace(req.HostName) == "":
		return errors.New("host_name is required")
	case strings.TrimSpace(req.Profile) == "":
		return errors.New("profile is required")
	default:
		return nil
	}
}

func validateCreateAlertRuleRequest(req model.CreateAlertRuleRequest) error {
	switch {
	case strings.TrimSpace(req.TemplateID) == "":
		return errors.New("template_id is required")
	case !store.KnownAlertRuleTemplateID(req.TemplateID):
		return errors.New("template_id is unknown")
	case strings.TrimSpace(req.Name) == "":
		return errors.New("name is required")
	case strings.TrimSpace(req.Trigger) == "":
		return errors.New("trigger is required")
	case strings.TrimSpace(req.Severity) == "":
		return errors.New("severity is required")
	case !knownSeverity(req.Severity):
		return errors.New("severity is unknown")
	case len(req.Channels) == 0:
		return errors.New("at least one channel is required")
	case !knownChannels(req.Channels):
		return errors.New("channel is unknown")
	case strings.TrimSpace(req.Condition.Subject) == "":
		return errors.New("condition.subject is required")
	case strings.TrimSpace(req.Condition.Operator) == "":
		return errors.New("condition.operator is required")
	default:
		return nil
	}
}

func validateCreateNotificationRouteRequest(req model.CreateNotificationRouteRequest) error {
	switch {
	case strings.TrimSpace(req.Channel) == "":
		return errors.New("channel is required")
	case !knownChannels([]string{req.Channel}):
		return errors.New("channel is unknown")
	case strings.TrimSpace(req.Provider) == "":
		return errors.New("provider is required")
	case !knownDeliveryProvider(req.Provider):
		return errors.New("provider is unknown")
	case !providerAllowedForChannel(req.Provider, req.Channel):
		return errors.New("provider is not valid for channel")
	case strings.TrimSpace(req.RecipientLabel) == "":
		return errors.New("recipient_label is required")
	case strings.TrimSpace(req.Status) != "" && !knownRouteStatus(req.Status):
		return errors.New("status is unknown")
	default:
		return nil
	}
}

func validateUpdateNotificationPreferencesRequest(req model.UpdateNotificationPreferencesRequest) error {
	if strings.TrimSpace(req.DigestCadence) != "" && !knownNotificationDigestCadence(req.DigestCadence) {
		return errors.New("digest_cadence is unknown")
	}
	if err := validateNotificationQuietHours(req.QuietHours); err != nil {
		return err
	}
	if err := validateNotificationEscalation(req.Escalation); err != nil {
		return err
	}
	for index, rule := range req.Rules {
		if err := validateNotificationPreferenceRule(index, rule); err != nil {
			return err
		}
	}
	return nil
}

func validateNotificationQuietHours(quiet model.NotificationQuietHours) error {
	if strings.TrimSpace(quiet.StartLocal) == "" && strings.TrimSpace(quiet.EndLocal) == "" && strings.TrimSpace(quiet.Timezone) == "" {
		return nil
	}
	switch {
	case strings.TrimSpace(quiet.StartLocal) == "":
		return errors.New("quiet_hours.start_local is required")
	case strings.TrimSpace(quiet.EndLocal) == "":
		return errors.New("quiet_hours.end_local is required")
	case strings.TrimSpace(quiet.Timezone) == "":
		return errors.New("quiet_hours.timezone is required")
	default:
		return nil
	}
}

func validateNotificationEscalation(escalation model.NotificationEscalationPolicy) error {
	if escalation.AfterMinutes == 0 && escalation.RepeatEveryMins == 0 && escalation.MaxRepeats == 0 && len(escalation.Channels) == 0 && strings.TrimSpace(escalation.Owner) == "" {
		return nil
	}
	switch {
	case escalation.AfterMinutes < 0:
		return errors.New("escalation.after_minutes cannot be negative")
	case escalation.RepeatEveryMins < 0:
		return errors.New("escalation.repeat_every_minutes cannot be negative")
	case escalation.MaxRepeats < 0:
		return errors.New("escalation.max_repeats cannot be negative")
	case len(escalation.Channels) > 0 && !knownChannels(escalation.Channels):
		return errors.New("escalation channel is unknown")
	default:
		return nil
	}
}

func validateNotificationPreferenceRule(index int, rule model.NotificationPreferenceRule) error {
	switch {
	case strings.TrimSpace(rule.Name) == "":
		return fmt.Errorf("rules[%d].name is required", index)
	case strings.TrimSpace(rule.EventType) == "":
		return fmt.Errorf("rules[%d].event_type is required", index)
	case strings.TrimSpace(rule.Severity) == "":
		return fmt.Errorf("rules[%d].severity is required", index)
	case !knownSeverity(rule.Severity):
		return fmt.Errorf("rules[%d].severity is unknown", index)
	case len(rule.Channels) == 0:
		return fmt.Errorf("rules[%d].channels is required", index)
	case !knownChannels(rule.Channels):
		return fmt.Errorf("rules[%d].channel is unknown", index)
	case strings.TrimSpace(rule.Mode) == "":
		return fmt.Errorf("rules[%d].mode is required", index)
	case !knownNotificationPreferenceMode(rule.Mode):
		return fmt.Errorf("rules[%d].mode is unknown", index)
	case strings.TrimSpace(rule.RecipientGroup) == "":
		return fmt.Errorf("rules[%d].recipient_group is required", index)
	default:
		return nil
	}
}

func validateRunDeliveryDrilldownRequest(req model.RunDeliveryDrilldownRequest) error {
	mode := strings.TrimSpace(req.Mode)
	channel := strings.TrimSpace(req.Channel)
	switch {
	case mode == "":
		return errors.New("mode is required")
	case !knownDeliveryDrillMode(mode):
		return errors.New("mode is unknown")
	case channel != "" && !knownChannels([]string{channel}):
		return errors.New("channel is unknown")
	default:
		return nil
	}
}

func validateRunDeliveryRemediationRequest(req model.RunDeliveryRemediationRequest) error {
	mode := strings.TrimSpace(req.Mode)
	channel := strings.TrimSpace(req.Channel)
	action := strings.TrimSpace(req.Action)
	switch {
	case mode == "":
		return errors.New("mode is required")
	case !knownDeliveryRemediationMode(mode):
		return errors.New("mode is unknown")
	case channel != "" && !knownChannels([]string{channel}):
		return errors.New("channel is unknown")
	case action != "" && !knownDeliveryRemediationAction(action):
		return errors.New("action is unknown")
	default:
		return nil
	}
}

func validateRunProviderSimulationRequest(req model.RunProviderSimulationRequest) error {
	mode := strings.TrimSpace(req.Mode)
	channel := strings.TrimSpace(req.Channel)
	switch {
	case mode == "":
		return errors.New("mode is required")
	case !knownProviderSimulationMode(mode):
		return errors.New("mode is unknown")
	case channel != "" && !knownChannels([]string{channel}):
		return errors.New("channel is unknown")
	default:
		return nil
	}
}

func validateCreateTenantActivityViewRequest(req model.CreateTenantActivityViewRequest) error {
	filter := req.Filter
	switch {
	case strings.TrimSpace(req.Name) == "":
		return errors.New("name is required")
	case strings.TrimSpace(req.PaidTier) != "" && !store.KnownPlanID(req.PaidTier):
		return errors.New("paid_tier is unknown")
	case strings.TrimSpace(filter.Kind) != "" && !knownActivityFeedKind(filter.Kind):
		return errors.New("filter.kind is unknown")
	case strings.TrimSpace(filter.Severity) != "" && !knownSeverity(filter.Severity):
		return errors.New("filter.severity is unknown")
	case strings.TrimSpace(filter.Channel) != "" && !knownChannels([]string{filter.Channel}):
		return errors.New("filter.channel is unknown")
	case strings.TrimSpace(filter.Status) != "" && !knownActivityFeedStatus(filter.Status):
		return errors.New("filter.status is unknown")
	case filter.Limit > constants.ActivityFeedMaxLimit:
		return fmt.Errorf("filter.limit exceeds maximum %d", constants.ActivityFeedMaxLimit)
	default:
		return nil
	}
}

func validateIngestTelemetryRequest(deviceID string, req model.IngestTelemetryRequest) error {
	switch {
	case strings.TrimSpace(req.TenantID) == "":
		return errors.New("tenant_id is required")
	case strings.TrimSpace(req.DeviceID) == "":
		return errors.New("device_id is required")
	case strings.TrimSpace(deviceID) != strings.TrimSpace(req.DeviceID):
		return errors.New("device_id must match route device")
	case strings.TrimSpace(req.HostName) == "":
		return errors.New("host_name is required")
	case strings.TrimSpace(req.Profile) == "":
		return errors.New("profile is required")
	case len(req.Events) == 0:
		return errors.New("at least one telemetry event is required")
	case len(req.Events) > constants.TelemetryIngestMaxEvents:
		return fmt.Errorf("events exceeds maximum batch size %d", constants.TelemetryIngestMaxEvents)
	default:
		for index, evt := range req.Events {
			if strings.TrimSpace(evt.Type) == "" {
				return fmt.Errorf("events[%d].type is required", index)
			}
			if strings.TrimSpace(evt.Source) == "" {
				return fmt.Errorf("events[%d].source is required", index)
			}
			if strings.TrimSpace(evt.TenantID) != "" && strings.TrimSpace(evt.TenantID) != strings.TrimSpace(req.TenantID) {
				return fmt.Errorf("events[%d].tenant_id must match request tenant_id", index)
			}
			if strings.TrimSpace(evt.DeviceID) != "" && strings.TrimSpace(evt.DeviceID) != strings.TrimSpace(req.DeviceID) {
				return fmt.Errorf("events[%d].device_id must match request device_id", index)
			}
		}
		return nil
	}
}

func validateCreateDeviceGroupRequest(req model.CreateDeviceGroupRequest) error {
	switch {
	case strings.TrimSpace(req.Name) == "":
		return errors.New("name is required")
	case strings.TrimSpace(req.Profile) == "":
		return errors.New("profile is required")
	case strings.TrimSpace(req.PolicyTemplateID) == "":
		return errors.New("policy_template_id is required")
	case !store.KnownPolicyTemplateID(req.PolicyTemplateID):
		return errors.New("policy_template_id is unknown")
	default:
		return nil
	}
}

func validateCreatePolicyAssignmentRequest(req model.CreatePolicyAssignmentRequest) error {
	switch {
	case strings.TrimSpace(req.Name) == "":
		return errors.New("name is required")
	case strings.TrimSpace(req.TargetType) == "":
		return errors.New("target_type is required")
	case !knownPolicyAssignmentTarget(req.TargetType):
		return errors.New("target_type is unknown")
	case strings.TrimSpace(req.TargetID) == "":
		return errors.New("target_id is required")
	case strings.TrimSpace(req.PolicyTemplateID) == "":
		return errors.New("policy_template_id is required")
	case !store.KnownPolicyTemplateID(req.PolicyTemplateID):
		return errors.New("policy_template_id is unknown")
	case strings.TrimSpace(req.Mode) == "":
		return errors.New("mode is required")
	case !knownPolicyAssignmentMode(req.Mode):
		return errors.New("mode is unknown")
	default:
		return nil
	}
}

func validateCreateTenantDataExportRequest(req model.CreateTenantDataExportRequest) error {
	switch {
	case strings.TrimSpace(req.Format) == "":
		return errors.New("format is required")
	case !knownDataExportFormat(req.Format):
		return errors.New("format is unknown")
	case strings.TrimSpace(req.Scope) == "":
		return errors.New("scope is required")
	case !knownDataExportScope(req.Scope):
		return errors.New("scope is unknown")
	default:
		return nil
	}
}

func validateCreateDeleteRequestRequest(req model.CreateDeleteRequestRequest) error {
	switch {
	case strings.TrimSpace(req.Scope) == "":
		return errors.New("scope is required")
	case !knownDeleteRequestScope(req.Scope):
		return errors.New("scope is unknown")
	case strings.TrimSpace(req.Reason) == "":
		return errors.New("reason is required")
	default:
		return nil
	}
}

func knownSeverity(severity string) bool {
	switch strings.TrimSpace(severity) {
	case constants.SeverityInfo, constants.SeverityLow, constants.SeverityMedium, constants.SeverityHigh, constants.SeverityCritical:
		return true
	default:
		return false
	}
}

func knownDataExportFormat(format string) bool {
	switch strings.TrimSpace(format) {
	case constants.DataExportFormatJSON, constants.DataExportFormatPDF:
		return true
	default:
		return false
	}
}

func knownDataExportScope(scope string) bool {
	switch strings.TrimSpace(scope) {
	case constants.DataExportScopeTenant, constants.DataExportScopeDevice:
		return true
	default:
		return false
	}
}

func knownDeleteRequestScope(scope string) bool {
	switch strings.TrimSpace(scope) {
	case constants.DeleteRequestScopeTenant, constants.DeleteRequestScopeDevice:
		return true
	default:
		return false
	}
}

func knownPolicyAssignmentTarget(target string) bool {
	switch strings.TrimSpace(target) {
	case constants.PolicyAssignmentTargetTenant, constants.PolicyAssignmentTargetDeviceGroup, constants.PolicyAssignmentTargetDevice:
		return true
	default:
		return false
	}
}

func knownPolicyAssignmentMode(mode string) bool {
	switch strings.TrimSpace(mode) {
	case constants.PolicyAssignmentModeAudit, constants.PolicyAssignmentModeActive:
		return true
	default:
		return false
	}
}

func knownChannels(channels []string) bool {
	for _, channel := range channels {
		switch strings.TrimSpace(channel) {
		case constants.DeliveryChannelEmail, constants.DeliveryChannelPush, constants.DeliveryChannelDashboard:
		default:
			return false
		}
	}
	return true
}

func knownDeliveryProvider(provider string) bool {
	switch strings.TrimSpace(provider) {
	case constants.DeliveryProviderSMTP, constants.DeliveryProviderWebPush, constants.DeliveryProviderLocalFeed:
		return true
	default:
		return false
	}
}

func knownBrowser(browser string) bool {
	switch strings.ToLower(strings.TrimSpace(browser)) {
	case constants.BrowserNameChrome, constants.BrowserNameEdge, constants.BrowserNameBrave:
		return true
	default:
		return false
	}
}

func knownBrowserCategory(category string) bool {
	switch strings.ToLower(strings.TrimSpace(category)) {
	case constants.BrowserCategoryStudy,
		constants.BrowserCategoryVideoStreaming,
		constants.BrowserCategorySocialMedia,
		constants.BrowserCategoryGaming,
		constants.BrowserCategoryShopping,
		constants.BrowserCategoryBlocked,
		constants.BrowserCategoryUnknown:
		return true
	default:
		return false
	}
}

func knownNotificationPreferenceMode(mode string) bool {
	switch strings.TrimSpace(mode) {
	case constants.NotificationPreferenceModeImmediate, constants.NotificationPreferenceModeDigest, constants.NotificationPreferenceModeSilent:
		return true
	default:
		return false
	}
}

func knownNotificationDigestCadence(cadence string) bool {
	switch strings.TrimSpace(cadence) {
	case constants.NotificationDigestCadenceDaily, constants.NotificationDigestCadenceWeekly:
		return true
	default:
		return false
	}
}

func providerAllowedForChannel(provider string, channel string) bool {
	switch strings.TrimSpace(channel) {
	case constants.DeliveryChannelEmail:
		return strings.TrimSpace(provider) == constants.DeliveryProviderSMTP
	case constants.DeliveryChannelPush:
		return strings.TrimSpace(provider) == constants.DeliveryProviderWebPush
	case constants.DeliveryChannelDashboard:
		return strings.TrimSpace(provider) == constants.DeliveryProviderLocalFeed
	default:
		return false
	}
}

func knownDeliveryDrillMode(mode string) bool {
	return strings.TrimSpace(mode) == constants.DeliveryDrillModeDryRun
}

func knownProviderSimulationMode(mode string) bool {
	return strings.TrimSpace(mode) == constants.ProviderSimulationModeDryRun
}

func knownDeliveryRemediationMode(mode string) bool {
	return strings.TrimSpace(mode) == constants.DeliveryRemediationModeDryRun
}

func knownDeliveryRemediationAction(action string) bool {
	switch strings.TrimSpace(action) {
	case constants.DeliveryRemediationActionRetryPlan,
		constants.DeliveryRemediationActionOwnerAck,
		constants.DeliveryRemediationActionSLAWatch,
		constants.DeliveryRemediationActionEnable,
		constants.DeliveryRemediationActionFix,
		constants.DeliveryRemediationActionRehearsal,
		constants.DeliveryRemediationActionMaintain:
		return true
	default:
		return false
	}
}

func knownRouteStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case constants.StatusHealthy, constants.StatusWatch, constants.StatusAttention, constants.StatusPending:
		return true
	default:
		return false
	}
}

func knownActivityFeedKind(kind string) bool {
	switch strings.TrimSpace(kind) {
	case constants.ActivityFeedKindRisk, constants.ActivityFeedKindDelivery, constants.ActivityFeedKindTelemetry:
		return true
	default:
		return false
	}
}

func knownActivityFeedStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case constants.RiskStatusOpen,
		constants.RiskStatusAcknowledged,
		constants.RiskStatusResolved,
		constants.DeliveryStatusDelivered,
		constants.DeliveryStatusPending,
		constants.DeliveryStatusRetrying,
		constants.DeliveryStatusFailed,
		constants.DeliveryStatusSuppressed,
		constants.StatusOK,
		constants.StatusHealthy,
		constants.StatusWatch,
		constants.StatusAttention:
		return true
	default:
		return false
	}
}

func knownDeliveryTimelineStatus(status string) bool {
	switch strings.TrimSpace(status) {
	case constants.DeliveryStatusDelivered,
		constants.DeliveryStatusPending,
		constants.DeliveryStatusRetrying,
		constants.DeliveryStatusFailed,
		constants.DeliveryStatusSuppressed:
		return true
	default:
		return false
	}
}

func knownDeliveryAssuranceState(state string) bool {
	switch strings.TrimSpace(state) {
	case constants.DeliveryAssuranceProviderConfirmed,
		constants.DeliveryAssuranceDryRunRehearsed,
		constants.DeliveryAssuranceDashboardVisible,
		constants.DeliveryAssuranceDemoOnly,
		constants.DeliveryAssuranceRetrying,
		constants.DeliveryAssuranceFailed,
		constants.DeliveryAssuranceRouteDisabled,
		constants.DeliveryAssurancePendingProvider:
		return true
	default:
		return false
	}
}

func buildConsentCenter(tenant model.Tenant, auditEvents []model.AuditEvent, rules []model.AlertRule) model.ConsentCenter {
	alertRecipients := []string{
		"configured email recipients from endpoint policy",
		"parent mobile push subscriptions",
		"local dashboard feed",
	}
	if len(rules) == 0 {
		alertRecipients = []string{"local dashboard feed"}
	}
	return model.ConsentCenter{
		TenantID:           tenant.TenantID,
		MonitoringVisible:  true,
		PauseControls:      "service manager can stop/start local collection where policy and law allow",
		DataExportReady:    true,
		DeleteRequestReady: true,
		AlertRecipients:    alertRecipients,
		Collection: []model.ConsentCollectionItem{
			{
				Name:        constants.ConsentCollectionAppUsage,
				Status:      constants.ConsentStatusCollected,
				Description: "process name, category, severity, and policy metadata for productivity and risk reporting",
				Retention:   "bounded by selected retention tier",
			},
			{
				Name:        constants.ConsentCollectionBrowserDomains,
				Status:      constants.ConsentStatusCollected,
				Description: "domain/category activity without raw URLs or page titles",
				Retention:   "bounded by selected retention tier",
			},
			{
				Name:        constants.ConsentCollectionDeviceHealth,
				Status:      constants.ConsentStatusDerived,
				Description: "CPU, memory, disk, battery, startup, crash, and agent heartbeat health score",
				Retention:   "bounded by selected retention tier",
			},
			{
				Name:        constants.ConsentCollectionArchiveHealth,
				Status:      constants.ConsentStatusDerived,
				Description: "S3 archive provider, backlog, upload key, and retry health metadata",
				Retention:   "bounded by selected retention tier",
			},
			{
				Name:        constants.ConsentCollectionPasswords,
				Status:      constants.ConsentStatusDenied,
				Description: "passwords, credentials, auth tokens, cookies, and keystrokes are not collected",
				Retention:   "not collected",
			},
			{
				Name:        constants.ConsentCollectionScreenshots,
				Status:      constants.ConsentStatusDenied,
				Description: "screenshots and hidden screen content are not collected",
				Retention:   "not collected",
			},
			{
				Name:        constants.ConsentCollectionPrivateContent,
				Status:      constants.ConsentStatusDenied,
				Description: "private messages, camera, microphone, raw URLs, and page titles are not collected",
				Retention:   "not collected",
			},
		},
		AuditEvents: auditEvents,
		UpdatedAt:   time.Now().UTC(),
	}
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	setNoStoreHeaders(w)
	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func setNoStoreHeaders(w http.ResponseWriter) {
	w.Header().Set(constants.HeaderCache, constants.CacheNoStore)
	w.Header().Set(constants.HeaderPragma, constants.PragmaNoCache)
	w.Header().Set(constants.HeaderExpires, constants.ExpiresNow)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, model.ErrorResponse{Error: message})
}

func writeMethodNotAllowed(w http.ResponseWriter) {
	writeError(w, http.StatusMethodNotAllowed, "method not allowed")
}

func requestLogger(logger *slog.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		logger.Info("backend request",
			"method", r.Method,
			"path", r.URL.Path,
			"duration_ms", time.Since(startedAt).Milliseconds(),
		)
	})
}
