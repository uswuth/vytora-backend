package risk_assessment

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/entity/vendor"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/services"
)

type NextCodeFunc func(ctx context.Context, entity string) (string, error)

type Handler struct {
	repo         *Repository
	vendorRepo   *vendor.Repository
	nextCode     NextCodeFunc
	validate     *validator.Validate
}

func NewHandler(repo *Repository, vendorRepo *vendor.Repository, nextCode NextCodeFunc) *Handler {
	return &Handler{
		repo:       repo,
		vendorRepo: vendorRepo,
		nextCode:   nextCode,
		validate:   validator.New(),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRiskAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, `{"error":"validation failed"}`, http.StatusBadRequest)
		return
	}

	v, err := h.vendorRepo.FindByCode(r.Context(), req.VendorCode)
	if err != nil {
		http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
		return
	}

	assessmentDate, err := time.Parse("2006-01-02", req.AssessmentDate)
	if err != nil {
		http.Error(w, `{"error":"invalid assessment_date format"}`, http.StatusBadRequest)
		return
	}

	code, err := h.nextCode(r.Context(), "risk_assessment")
	if err != nil {
		http.Error(w, `{"error":"failed to generate code"}`, http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value(middleware.UserContextKey).(*services.Claims)
	reviewedBy, _ := uuid.Parse(claims.UserID)
	now := time.Now()

	ra := &RiskAssessment{
		Code:            code,
		VendorID:        v.ID,
		VendorCode:      v.Code,
		AssessmentDate:  assessmentDate,
		RiskLevel:       req.RiskLevel,
		Findings:        req.Findings,
		Recommendations: req.Recommendations,
		Status:          "Pending",
		ReviewedBy:      &reviewedBy,
		ReviewedAt:      &now,
	}

	if err := h.repo.Create(r.Context(), ra); err != nil {
		http.Error(w, `{"error":"failed to create risk assessment"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ra)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ra, err := h.repo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"risk assessment not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ra)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	vendorCode := r.URL.Query().Get("vendor_code")
	params := ListParams{
		VendorCode: vendorCode,
		RiskLevel:  r.URL.Query().Get("risk_level"),
		Status:     r.URL.Query().Get("status"),
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		params.Limit, _ = strconv.Atoi(v)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		params.Offset, _ = strconv.Atoi(v)
	}

	assessments, total, err := h.repo.ListAll(r.Context(), params)
	if err != nil {
		http.Error(w, `{"error":"failed to list risk assessments"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"data":  assessments,
		"total": total,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Approve(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ra, err := h.repo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"risk assessment not found"}`, http.StatusNotFound)
		return
	}

	ra.Status = "Approved"

	if err := h.repo.Update(r.Context(), ra); err != nil {
		http.Error(w, `{"error":"failed to approve risk assessment"}`, http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}