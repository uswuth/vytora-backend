package report

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/services"
)

type Handler struct {
	reportRepo *Repository
}

func NewHandler(reportRepo *Repository) *Handler {
	return &Handler{
		reportRepo: reportRepo,
	}
}

// getUserDeptID returns the assigned department manager UUID if the user is a department manager,
// otherwise nil (meaning see all vendors).
func getUserDeptID(r *http.Request) *uuid.UUID {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*services.Claims)
	if !ok {
		return nil
	}
	if claims.Role == "department_manager" {
		uid, err := uuid.Parse(claims.UserID)
		if err == nil {
			return &uid
		}
	}
	return nil
}

func (h *Handler) Summary(w http.ResponseWriter, r *http.Request) {
	deptID := getUserDeptID(r)
	summary, err := h.reportRepo.GetSummary(r.Context(), deptID)
	if err != nil {
		http.Error(w, `{"error":"failed to generate summary"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(summary)
}

func (h *Handler) MonthlyOnboarding(w http.ResponseWriter, r *http.Request) {
	deptID := getUserDeptID(r)
	data, err := h.reportRepo.GetMonthlyOnboarding(r.Context(), deptID)
	if err != nil {
		http.Error(w, `{"error":"failed to get monthly onboarding"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}