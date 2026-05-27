package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/goquizvibe/backend/feature/dashboard/services"
	pages "github.com/goquizvibe/backend/feature/dashboard/ui"
	ce "github.com/goquizvibe/backend/shared/custom_errors"
	"github.com/goquizvibe/backend/shared/locales"
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

	t := locales.GetTranslator(r.Context())
	return pages.DashboardPage(*data, t).Render(r.Context(), w)
}
