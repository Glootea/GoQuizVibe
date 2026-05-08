package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/google/uuid"
	ce "github.com/goquizvibe/custom_errors"
	"github.com/goquizvibe/middleware"
	"github.com/goquizvibe/pages"
	"github.com/goquizvibe/services"
)

type DashboardHandler struct {
	dashboardService *services.DashboardService
}

func NewDashboard(ds *services.DashboardService) *DashboardHandler {
	return &DashboardHandler{
		dashboardService: ds,
	}
}

func (h *DashboardHandler) DashboardPage(w http.ResponseWriter, r *http.Request) error {
	ctx := context.Background()
	userID, err := h.dashboardService.GetUserIDFromRequest(r)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrUnauthorized, err), http.StatusUnauthorized)
	}

	data, err := h.dashboardService.GetDashboardData(ctx, userID)
	if err != nil {
		return ce.WithHTTPStatus(errors.Join(ce.ErrInternal, err), http.StatusInternalServerError)
	}

	t := middleware.GetTranslator(r.Context())
	return pages.DashboardPage(*data, t).Render(r.Context(), w)
}

func (h *DashboardHandler) getUserIDFromRequest(r *http.Request) (uuid.UUID, error) {
	return h.dashboardService.GetUserIDFromRequest(r)
}
