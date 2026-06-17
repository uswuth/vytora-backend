package compliance_record

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

type Handler struct {
	compRepo   *Repository
	vendorRepo *vendor.Repository
	nextCode   NextCodeFunc
	validate   *validator.Validate
}

type NextCodeFunc func(ctx context.Context, entity string) (string, error)

func NewHandler(compRepo *Repository, vendorRepo *vendor.Repository, nextCode NextCodeFunc) *Handler {
	return &Handler{
		compRepo:   compRepo,
		vendorRepo: vendorRepo,
		nextCode:   nextCode,
		validate:   validator.New(),
	}
}

type CreateComplianceRequest struct {
	VendorCode        string `json:"vendor_code" validate:"required"`
	CertificationType string `json:"certification_type" validate:"required,oneof=ISO27001 SOC2 GDPR PCI_DSS"`
	ValidFrom         string `json:"valid_from" validate:"required"`      // YYYY-MM-DD
	ValidUntil        string `json:"valid_until" validate:"required"`
	IssuedBy          string `json:"issued_by"`
	EvidenceURL       string `json:"evidence_url"`
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateComplianceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, `{"error":"validation failed"}`, http.StatusBadRequest)
		return
	}

	vendor, err := h.vendorRepo.FindByCode(r.Context(), req.VendorCode)
	if err != nil {
		http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
		return
	}

	from, err := time.Parse("2006-01-02", req.ValidFrom)
	if err != nil {
		http.Error(w, `{"error":"invalid valid_from format"}`, http.StatusBadRequest)
		return
	}
	until, err := time.Parse("2006-01-02", req.ValidUntil)
	if err != nil {
		http.Error(w, `{"error":"invalid valid_until format"}`, http.StatusBadRequest)
		return
	}

	// Auto-compute status based on dates
	status := "Pending"
	if time.Now().After(until) {
		status = "Expired"
	} else if time.Now().After(from) || time.Now().Equal(from) {
		status = "Approved"
	}

	code, err := h.nextCode(r.Context(), "compliance_record")
	if err != nil {
		http.Error(w, `{"error":"failed to generate code"}`, http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value(middleware.UserContextKey).(*services.Claims)
	reviewedBy, _ := uuid.Parse(claims.UserID)

	cr := &ComplianceRecord{
		Code:              code,
		VendorID:          vendor.ID,
		CertificationType: req.CertificationType,
		Status:            status,
		ValidFrom:         &from,
		ValidUntil:        &until,
		IssuedBy:          req.IssuedBy,
		EvidenceURL:       req.EvidenceURL,
		ReviewedBy:        &reviewedBy,
	}

	if err := h.compRepo.Create(r.Context(), cr); err != nil {
		http.Error(w, `{"error":"failed to create compliance record"}`, http.StatusInternalServerError)
		return
	}

	cr.VendorCode = vendor.Code

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cr)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	cr, err := h.compRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"compliance record not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cr)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	existing, err := h.compRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"compliance record not found"}`, http.StatusNotFound)
		return
	}

	var req struct {
		CertificationType string `json:"certification_type" validate:"required,oneof=ISO27001 SOC2 GDPR PCI_DSS"`
		Status            string `json:"status" validate:"required,oneof=Pending Approved Expired"`
		ValidFrom         string `json:"valid_from"`
		ValidUntil        string `json:"valid_until"`
		IssuedBy          string `json:"issued_by"`
		EvidenceURL       string `json:"evidence_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, `{"error":"validation failed"}`, http.StatusBadRequest)
		return
	}

	if req.ValidFrom != "" {
		t, _ := time.Parse("2006-01-02", req.ValidFrom)
		existing.ValidFrom = &t
	}
	if req.ValidUntil != "" {
		t, _ := time.Parse("2006-01-02", req.ValidUntil)
		existing.ValidUntil = &t
	}
	existing.CertificationType = req.CertificationType
	existing.Status = req.Status
	existing.IssuedBy = req.IssuedBy
	existing.EvidenceURL = req.EvidenceURL

	if err := h.compRepo.Update(r.Context(), existing); err != nil {
		http.Error(w, `{"error":"failed to update compliance record"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	vendorCode := r.URL.Query().Get("vendor_code")
	vendor, err := h.vendorRepo.FindByCode(r.Context(), vendorCode)
	if err != nil {
		http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
		return
	}
	records, err := h.compRepo.ListByVendor(r.Context(), vendor.ID.String())
	if err != nil {
		http.Error(w, `{"error":"failed to list compliance records"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}

func (h *Handler) Expiring(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	if daysStr == "" {
		daysStr = "30"
	}
	days, _ := strconv.Atoi(daysStr)

	records, err := h.compRepo.Expiring(r.Context(), days)
	if err != nil {
		http.Error(w, `{"error":"failed to get expiring certifications"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(records)
}