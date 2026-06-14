package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/uswuth/vytora-backend/internal/models"
	"github.com/uswuth/vytora-backend/internal/repository"
	"github.com/uswuth/vytora-backend/internal/services"
)

type ContractHandler struct {
	contractRepo *repository.ContractRepository
	vendorRepo   *repository.VendorRepository
	seqService   *services.SequenceService
	validate     *validator.Validate
}

func NewContractHandler(contractRepo *repository.ContractRepository, vendorRepo *repository.VendorRepository, seqService *services.SequenceService) *ContractHandler {
	return &ContractHandler{
		contractRepo: contractRepo,
		vendorRepo:   vendorRepo,
		seqService:   seqService,
		validate:     validator.New(),
	}
}

type CreateContractRequest struct {
	VendorCode     string  `json:"vendor_code" validate:"required"`
	ContractNumber string  `json:"contract_number" validate:"required,max=100"`
	StartDate      string  `json:"start_date" validate:"required"` // YYYY-MM-DD
	EndDate        string  `json:"end_date" validate:"required"`
	ContractValue  *float64 `json:"contract_value"`
	RenewalStatus  string  `json:"renewal_status" validate:"required,oneof=Auto-Renew Manual Expiring"`
}

func (h *ContractHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateContractRequest
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

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		http.Error(w, `{"error":"invalid start_date format"}`, http.StatusBadRequest)
		return
	}
	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		http.Error(w, `{"error":"invalid end_date format"}`, http.StatusBadRequest)
		return
	}

	code, err := h.seqService.NextCode(r.Context(), "contract")
	if err != nil {
		http.Error(w, `{"error":"failed to generate code"}`, http.StatusInternalServerError)
		return
	}

	contract := &models.Contract{
		Code:           code,
		VendorID:       vendor.ID,
		VendorCode:     vendor.Code,
		ContractNumber: req.ContractNumber,
		StartDate:      start,
		EndDate:        end,
		ContractValue:  req.ContractValue,
		RenewalStatus:  req.RenewalStatus,
	}

	if err := h.contractRepo.Create(r.Context(), contract); err != nil {
		http.Error(w, `{"error":"failed to create contract"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(contract)
}

func (h *ContractHandler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	contract, err := h.contractRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"contract not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contract)
}

func (h *ContractHandler) List(w http.ResponseWriter, r *http.Request) {
	vendorCode := r.URL.Query().Get("vendor_code")
	vendor, err := h.vendorRepo.FindByCode(r.Context(), vendorCode)
	if err != nil {
		http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
		return
	}
	contracts, err := h.contractRepo.ListByVendor(r.Context(), vendor.ID.String())
	if err != nil {
		http.Error(w, `{"error":"failed to list contracts"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contracts)
}

func (h *ContractHandler) Expiring(w http.ResponseWriter, r *http.Request) {
	daysStr := r.URL.Query().Get("days")
	if daysStr == "" {
		daysStr = "30"
	}
	days, _ := strconv.Atoi(daysStr)

	contracts, err := h.contractRepo.Expiring(r.Context(), days)
	if err != nil {
		http.Error(w, `{"error":"failed to get expiring contracts"}`, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contracts)
}