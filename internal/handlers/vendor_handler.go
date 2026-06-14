package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/models"
	"github.com/uswuth/vytora-backend/internal/repository"
	"github.com/uswuth/vytora-backend/internal/services"
)

type VendorHandler struct {
	vendorRepo *repository.VendorRepository
	seqService *services.SequenceService
	validate   *validator.Validate
}

func NewVendorHandler(vendorRepo *repository.VendorRepository, seqService *services.SequenceService) *VendorHandler {
	return &VendorHandler{
		vendorRepo: vendorRepo,
		seqService: seqService,
		validate:   validator.New(),
	}
}

// ---------- Request/Response helpers ----------

type CreateVendorRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=255"`
	Category      string `json:"category" validate:"required,max=100"`
	ContactPerson string `json:"contact_person"`
	ContactEmail  string `json:"contact_email" validate:"omitempty,email"`
	Country       string `json:"country"`
	RiskLevel     string `json:"risk_level" validate:"required,oneof=Low Medium High Critical"`
}

type UpdateVendorRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=255"`
	Category      string `json:"category" validate:"required,max=100"`
	ContactPerson string `json:"contact_person"`
	ContactEmail  string `json:"contact_email" validate:"omitempty,email"`
	Country       string `json:"country"`
	RiskLevel     string `json:"risk_level" validate:"required,oneof=Low Medium High Critical"`
	Status        string `json:"status" validate:"required,oneof=Draft Submitted RiskReview ComplianceReview Approved Rejected Active Inactive"`
}

// ---------- Create ----------
func (h *VendorHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateVendorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, `{"error":"validation failed"}`, http.StatusBadRequest)
		return
	}

	code, err := h.seqService.NextCode(r.Context(), "vendor")
	if err != nil {
		http.Error(w, `{"error":"failed to generate vendor code"}`, http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value(middleware.UserContextKey).(*services.Claims)
	createdBy, _ := uuid.Parse(claims.UserID)

	vendor := &models.Vendor{
		Code:          code,
		Name:          req.Name,
		Category:      req.Category,
		ContactPerson: req.ContactPerson,
		ContactEmail:  req.ContactEmail,
		Country:       req.Country,
		RiskLevel:     req.RiskLevel,
		Status:        "Draft",
		CreatedBy:     createdBy,
	}

	if err := h.vendorRepo.Create(r.Context(), vendor); err != nil {
		http.Error(w, `{"error":"failed to create vendor"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(vendor)
}

// ---------- Get by code ----------
func (h *VendorHandler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	vendor, err := h.vendorRepo.FindByCode(r.Context(), code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(vendor)
}

// ---------- List with query params ----------
func (h *VendorHandler) List(w http.ResponseWriter, r *http.Request) {
	params := repository.ListParams{
		Search:    r.URL.Query().Get("search"),
		Category:  r.URL.Query().Get("category"),
		RiskLevel: r.URL.Query().Get("risk_level"),
		Status:    r.URL.Query().Get("status"),
		Country:   r.URL.Query().Get("country"),
		SortBy:    r.URL.Query().Get("sort_by"),
		SortOrder: r.URL.Query().Get("sort_order"),
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		params.Limit, _ = strconv.Atoi(v)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		params.Offset, _ = strconv.Atoi(v)
	}

	vendors, total, err := h.vendorRepo.List(r.Context(), params)
	if err != nil {
		http.Error(w, `{"error":"failed to list vendors"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"data":  vendors,
		"total": total,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// ---------- Update ----------
func (h *VendorHandler) Update(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	var req UpdateVendorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, `{"error":"validation failed"}`, http.StatusBadRequest)
		return
	}

	// Fetch existing vendor (optional, but ensures it exists)
	existing, err := h.vendorRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
		return
	}

	existing.Name = req.Name
	existing.Category = req.Category
	existing.ContactPerson = req.ContactPerson
	existing.ContactEmail = req.ContactEmail
	existing.Country = req.Country
	existing.RiskLevel = req.RiskLevel
	existing.Status = req.Status

	if err := h.vendorRepo.Update(r.Context(), existing); err != nil {
		http.Error(w, `{"error":"failed to update vendor"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(existing)
}

// ---------- Delete ----------
func (h *VendorHandler) Delete(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if err := h.vendorRepo.Delete(r.Context(), code); err != nil {
		http.Error(w, `{"error":"failed to delete vendor"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
