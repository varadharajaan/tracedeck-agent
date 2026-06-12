package api

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/varadharajaan/tracedeck-agent/backend/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/model"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/store"
)

//go:embed web/dashboard.html
var dashboardFS embed.FS

type Server struct {
	store     store.Repository
	logger    *slog.Logger
	startedAt time.Time
	auth      AuthConfig
}

func NewServer(repo store.Repository, logger *slog.Logger) *Server {
	if repo == nil {
		repo = store.NewMemory()
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		store:     repo,
		logger:    logger,
		startedAt: time.Now().UTC(),
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
	case len(parts) == 2 && parts[1] == constants.RouteSegmentConsentCenter && r.Method == http.MethodGet:
		s.handleTenantConsentCenter(w, r, tenantID)
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
	writeJSON(w, http.StatusOK, model.ListResponse[model.AlertDelivery]{Items: deliveries, Count: len(deliveries)})
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
	report := store.WeeklyReport(overview)
	data := store.WeeklyReportPDF(report)
	w.Header().Set("Content-Type", constants.ContentTypePDF)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"tracedeck-weekly-%s.pdf\"", report.DeviceID))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
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
	w.Header().Set("Content-Type", constants.ContentTypeJSON)
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
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
