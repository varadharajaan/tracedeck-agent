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
	store     *store.Memory
	logger    *slog.Logger
	startedAt time.Time
}

func NewServer(repo *store.Memory, logger *slog.Logger) *Server {
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

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(constants.RouteDashboard, s.handleDashboard)
	mux.HandleFunc(constants.RouteHealth, s.handleHealth)
	mux.HandleFunc(constants.RouteVersion, s.handleVersion)
	mux.HandleFunc(constants.RouteDeviceEnroll, s.handleDeviceEnroll)
	mux.HandleFunc(constants.RouteDevices, s.handleDevices)
	mux.HandleFunc(constants.RouteDevices+"/", s.handleDeviceRoutes)
	mux.HandleFunc(constants.RoutePolicyTemplates, s.handlePolicyTemplates)
	mux.HandleFunc(constants.RouteArchiveStatus, s.handleArchiveStatus)
	return requestLogger(s.logger, mux)
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
	case len(parts) == 3 && parts[1] == "summary" && parts[2] == "daily" && r.Method == http.MethodGet:
		s.handleDailySummary(w, r, deviceID)
	case len(parts) == 3 && parts[1] == "reports" && parts[2] == "weekly" && r.Method == http.MethodGet:
		s.handleWeeklyReport(w, r, deviceID)
	case len(parts) == 2 && parts[1] == "policy-violations" && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, model.ListResponse[string]{Items: []string{}, Count: 0})
	case len(parts) == 2 && parts[1] == "anomalies" && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, model.ListResponse[string]{Items: []string{}, Count: 0})
	case len(parts) == 2 && parts[1] == "tamper-events" && r.Method == http.MethodGet:
		writeJSON(w, http.StatusOK, model.ListResponse[string]{Items: []string{}, Count: 0})
	default:
		http.NotFound(w, r)
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
	writeJSON(w, http.StatusOK, device)
}

func (s *Server) handleDailySummary(w http.ResponseWriter, r *http.Request, deviceID string) {
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

func (s *Server) handleWeeklyReport(w http.ResponseWriter, _ *http.Request, deviceID string) {
	writeJSON(w, http.StatusOK, store.WeeklyReport(deviceID))
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

func (s *Server) handleArchiveStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeMethodNotAllowed(w)
		return
	}
	writeJSON(w, http.StatusOK, store.ArchiveStatus())
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
