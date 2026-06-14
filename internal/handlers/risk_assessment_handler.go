package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/models"
	"github.com/uswuth/vytora-backend/internal/repository"
	"github.com/uswuth/vytora-backend/internal/services"
)

type RiskAssessmentHandler struct {
	raRepo     *repository.RiskAssessmentRepository
	vendorRepo *repository.VendorRepository
	seqService *services.SequenceService
	validate   *validator.Validate
}

func NewRiskAssessmentHandler(raRepo *repository.RiskAssessmentRepository, vendorRepo *repository.VendorRepository, seqService *services.SequenceService) *RiskAssessmentHandler {
	return &RiskAssessmentHandler{
		raRepo:     raRepo,
		vendorRepo: vendorRepo,
		seqService: seqService,
		validate:   validator.New(),
	}
}

type CreateRiskAssessmentRequest struct {
	VendorCode         string  `json:"vendor_code" validate:"required"`
	SecurityRiskScore  float64 `json:"security_risk_score" validate:"required,min=0,max=100"`
	FinancialRiskScore float64 `json:"financial_risk_score" validate:"required,min=0,max=100"`
	OperationalRiskScore float64 `json:"operational_risk_score" validate:"required,min=0,max=100"`
	LegalRiskScore     float64 `json:"legal_risk_score" validate:"required,min=0,max=100"`
	AssessmentDate     string  `json:"assessment_date" validate:"required"` // YYYY-MM-DD
	Notes              string  `json:"notes"`
}

// Calculate overall score and map to level
func calculateRiskLevel(score float64) string {
	switch {
	case score <= 25:
		return "Low"
	case score <= 50:
		return "Medium"
	case score <= 75:
		return "High"
	default:
		return "Critical"
	}
}
// Create
func (h *RiskAssessmentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRiskAssessmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, `{"error":"validation failed"}`, http.StatusBadRequest)
		return
	}

	// Find vendor by code
	vendor, err := h.vendorRepo.FindByCode(r.Context(), req.VendorCode)
	if err != nil {
		http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
		return
	}

	// Parse date
	assessDate, err := time.Parse("2006-01-02", req.AssessmentDate)
	if err != nil {
		http.Error(w, `{"error":"invalid assessment_date format, use YYYY-MM-DD"}`, http.StatusBadRequest)
		return
	}

	// Calculate
	overall := (req.SecurityRiskScore + req.FinancialRiskScore + req.OperationalRiskScore + req.LegalRiskScore) / 4.0
	level := calculateRiskLevel(overall)

	// Generate code
	code, err := h.seqService.NextCode(r.Context(), "risk_assessment")
	if err != nil {
		http.Error(w, `{"error":"failed to generate code"}`, http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value(middleware.UserContextKey).(*services.Claims)
	assessorID, _ := uuid.Parse(claims.UserID)

	ra := &models.RiskAssessment{
		Code:                code,
		VendorID:            vendor.ID,
		AssessmentDate:      assessDate,
		AssessorID:          assessorID,
		OverallRiskScore:    overall,
		RiskLevel:           level,
		SecurityRiskScore:   req.SecurityRiskScore,
		FinancialRiskScore:  req.FinancialRiskScore,
		OperationalRiskScore: req.OperationalRiskScore,
		LegalRiskScore:      req.LegalRiskScore,
		Status:              "Draft",
		Notes:               req.Notes,
	}

	if err := h.raRepo.Create(r.Context(), ra); err != nil {
		http.Error(w, `{"error":"failed to create risk assessment"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(ra)
}

func (h *RiskAssessmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ra, err := h.raRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"risk assessment not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ra)
}

func (h *RiskAssessmentHandler) List(w http.ResponseWriter, r *http.Request) {
	params := repository.RiskListParams{
		VendorCode: r.URL.Query().Get("vendor_code"),
		RiskLevel:  r.URL.Query().Get("risk_level"),
		Status:     r.URL.Query().Get("status"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		params.Limit, _ = strconv.Atoi(v)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		params.Offset, _ = strconv.Atoi(v)
	}

	assessments, total, err := h.raRepo.List(r.Context(), params)
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

func (h *RiskAssessmentHandler) Approve(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	ra, err := h.raRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"risk assessment not found"}`, http.StatusNotFound)
		return
	}
	if ra.Status != "Draft" && ra.Status != "Reviewed" {
		http.Error(w, `{"error":"only Draft or Reviewed assessments can be approved"}`, http.StatusBadRequest)
		return
	}
	if err := h.raRepo.UpdateStatus(r.Context(), code, "Approved"); err != nil {
		http.Error(w, `{"error":"failed to approve"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}