package api

import (
	"context"
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/varadharajaan/tracedeck-agent/backend/internal/constants"
	"github.com/varadharajaan/tracedeck-agent/backend/internal/model"
)

type AuthConfig struct {
	APIKey   string
	TenantID string
	ActorID  string
	RoleID   string
}

type authContextKey struct{}

type principal struct {
	TenantID string
	ActorID  string
	RoleID   string
}

func (c AuthConfig) Enabled() bool {
	return strings.TrimSpace(c.APIKey) != ""
}

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	if !s.auth.Enabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if authExempt(r) {
			next.ServeHTTP(w, r)
			return
		}

		supplied := strings.TrimSpace(r.Header.Get(constants.HeaderAPIKey))
		expected := strings.TrimSpace(s.auth.APIKey)
		if supplied == "" || subtle.ConstantTimeCompare([]byte(supplied), []byte(expected)) != 1 {
			writeError(w, http.StatusUnauthorized, "api key is required")
			return
		}

		requestTenantID := strings.TrimSpace(r.Header.Get(constants.HeaderTenantID))
		configTenantID := strings.TrimSpace(s.auth.TenantID)
		if configTenantID != "" && requestTenantID != "" && requestTenantID != configTenantID {
			writeError(w, http.StatusForbidden, "tenant scope is not allowed")
			return
		}
		if requestTenantID == "" {
			requestTenantID = configTenantID
		}

		actorID := strings.TrimSpace(r.Header.Get(constants.HeaderActorID))
		if actorID == "" {
			actorID = strings.TrimSpace(s.auth.ActorID)
		}
		roleID := strings.TrimSpace(s.auth.RoleID)
		if roleID == "" {
			roleID = constants.RoleBusinessManager
		}

		ctx := context.WithValue(r.Context(), authContextKey{}, principal{
			TenantID: requestTenantID,
			ActorID:  actorID,
			RoleID:   roleID,
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func authExempt(r *http.Request) bool {
	if r.URL.Path == constants.RouteHealth {
		return true
	}
	if r.URL.Path == constants.RouteDashboard && r.Method == http.MethodGet {
		return true
	}
	return false
}

func requestPrincipal(ctx context.Context) principal {
	value, _ := ctx.Value(authContextKey{}).(principal)
	return value
}

func tenantAllowed(ctx context.Context, tenantID string) bool {
	tenantID = strings.TrimSpace(tenantID)
	scope := strings.TrimSpace(requestPrincipal(ctx).TenantID)
	return scope == "" || tenantID == "" || tenantID == scope
}

func filterTenantsForPrincipal(ctx context.Context, tenants []model.Tenant) []model.Tenant {
	scope := strings.TrimSpace(requestPrincipal(ctx).TenantID)
	if scope == "" {
		return tenants
	}
	filtered := make([]model.Tenant, 0, len(tenants))
	for _, tenant := range tenants {
		if tenant.TenantID == scope {
			filtered = append(filtered, tenant)
		}
	}
	return filtered
}

func filterDevicesForPrincipal(ctx context.Context, devices []model.Device) []model.Device {
	scope := strings.TrimSpace(requestPrincipal(ctx).TenantID)
	if scope == "" {
		return devices
	}
	filtered := make([]model.Device, 0, len(devices))
	for _, device := range devices {
		if device.TenantID == scope {
			filtered = append(filtered, device)
		}
	}
	return filtered
}

func filterAuditEventsForPrincipal(ctx context.Context, events []model.AuditEvent) []model.AuditEvent {
	scope := strings.TrimSpace(requestPrincipal(ctx).TenantID)
	if scope == "" {
		return events
	}
	filtered := make([]model.AuditEvent, 0, len(events))
	for _, event := range events {
		if event.TenantID == scope {
			filtered = append(filtered, event)
		}
	}
	return filtered
}
