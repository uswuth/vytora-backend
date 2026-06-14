package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
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

type CreateVendorRequest struct {
	Name          string `json:"name" validate:"required,min=2,max=255"`
	Category      string `json:"category" validate:"required,max=100"`
	ContactPerson string `json:"contact_person"`
	ContactEmail  string `json:"contact_email" validate:"omitempty,email"`
	Country       string `json:"country"`
	RiskLevel     string `json:"risk_level" validate:"required,oneof=Low Medium High Critical"`
}

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

	claims, ok := r.Context().Value(middleware.UserContextKey).(*services.Claims)
	if !ok {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}
	createdBy, err := uuid.Parse(claims.UserID)
	if err != nil {
		http.Error(w, `{"error":"invalid user id"}`, http.StatusInternalServerError)
		return
	}

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
