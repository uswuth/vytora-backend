package contract

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/uswuth/vytora-backend/internal/entity/vendor"
)

type Handler struct {
	contractRepo *Repository
	vendorRepo   *vendor.Repository
	nextCode     NextCodeFunc
	validate     *validator.Validate
}

type NextCodeFunc func(ctx context.Context, entity string) (string, error)

func NewHandler(contractRepo *Repository, vendorRepo *vendor.Repository, nextCode NextCodeFunc) *Handler {
	return &Handler{
		contractRepo: contractRepo,
		vendorRepo:   vendorRepo,
		nextCode:     nextCode,
		validate:     validator.New(),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
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

	code, err := h.nextCode(r.Context(), "contract")
	if err != nil {
		http.Error(w, `{"error":"failed to generate code"}`, http.StatusInternalServerError)
		return
	}

	contract := &Contract{
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

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	contract, err := h.contractRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"contract not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(contract)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) Expiring(w http.ResponseWriter, r *http.Request) {
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