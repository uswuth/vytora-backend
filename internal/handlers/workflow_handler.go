package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/models"
	"github.com/uswuth/vytora-backend/internal/repository"
	"github.com/uswuth/vytora-backend/internal/services"
)

type WorkflowHandler struct {
	vendorRepo *repository.VendorRepository
	auditRepo  *repository.AuditTrailRepository
	seqService *services.SequenceService
}

func NewWorkflowHandler(vendorRepo *repository.VendorRepository, auditRepo *repository.AuditTrailRepository, seqService *services.SequenceService) *WorkflowHandler {
	return &WorkflowHandler{
		vendorRepo: vendorRepo,
		auditRepo:  auditRepo,
		seqService: seqService,
	}
}

// logTransition creates an audit trail entry for a status change.
func (h *WorkflowHandler) logTransition(w http.ResponseWriter, r *http.Request, vendorID uuid.UUID, oldStatus, newStatus string) error {
	claims := r.Context().Value(middleware.UserContextKey).(*services.Claims)
	changedBy, _ := uuid.Parse(claims.UserID)

	code, err := h.seqService.NextCode(r.Context(), "audit_trail")
	if err != nil {
		http.Error(w, `{"error":"failed to generate audit code"}`, http.StatusInternalServerError)
		return err
	}

	audit := &models.AuditTrail{
		Code:      code,
		TableName: "vendors",
		RecordID:  vendorID,
		Action:    "UPDATE",
		FieldName: "status",
		OldValue:  oldStatus,
		NewValue:  newStatus,
		ChangedBy: changedBy,
	}
	return h.auditRepo.Create(r.Context(), audit)
}

// Submit transitions a vendor from Draft -> Submitted (Department Manager)
func (h *WorkflowHandler) Submit(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	vendor, err := h.vendorRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
		return
	}
	if vendor.Status != "Draft" {
		http.Error(w, `{"error":"vendor must be in Draft status"}`, http.StatusBadRequest)
		return
	}
	if err := h.vendorRepo.UpdateStatus(r.Context(), code, "Submitted"); err != nil {
		http.Error(w, `{"error":"failed to update status"}`, http.StatusInternalServerError)
		return
	}
	if err := h.logTransition(w, r, vendor.ID, "Draft", "Submitted"); err != nil {
		return // error already sent
	}
	w.WriteHeader(http.StatusNoContent)
}

// ReviewRisk transitions Submitted -> RiskReview (Risk Manager)
func (h *WorkflowHandler) ReviewRisk(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	vendor, err := h.vendorRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
		return
	}
	if vendor.Status != "Submitted" {
		http.Error(w, `{"error":"vendor must be in Submitted status"}`, http.StatusBadRequest)
		return
	}
	if err := h.vendorRepo.UpdateStatus(r.Context(), code, "RiskReview"); err != nil {
		http.Error(w, `{"error":"failed to update status"}`, http.StatusInternalServerError)
		return
	}
	if err := h.logTransition(w, r, vendor.ID, "Submitted", "RiskReview"); err != nil {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ReviewCompliance transitions RiskReview -> ComplianceReview (Compliance Officer)
func (h *WorkflowHandler) ReviewCompliance(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	vendor, err := h.vendorRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
		return
	}
	if vendor.Status != "RiskReview" {
		http.Error(w, `{"error":"vendor must be in RiskReview status"}`, http.StatusBadRequest)
		return
	}
	if err := h.vendorRepo.UpdateStatus(r.Context(), code, "ComplianceReview"); err != nil {
		http.Error(w, `{"error":"failed to update status"}`, http.StatusInternalServerError)
		return
	}
	if err := h.logTransition(w, r, vendor.ID, "RiskReview", "ComplianceReview"); err != nil {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Approve transitions ComplianceReview -> Approved (System Admin / auto)
func (h *WorkflowHandler) Approve(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	vendor, err := h.vendorRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
		return
	}
	if vendor.Status != "ComplianceReview" {
		http.Error(w, `{"error":"vendor must be in ComplianceReview status"}`, http.StatusBadRequest)
		return
	}
	if err := h.vendorRepo.UpdateStatus(r.Context(), code, "Approved"); err != nil {
		http.Error(w, `{"error":"failed to update status"}`, http.StatusInternalServerError)
		return
	}
	if err := h.logTransition(w, r, vendor.ID, "ComplianceReview", "Approved"); err != nil {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// Reject transitions any state -> Rejected (appropriate role)
func (h *WorkflowHandler) Reject(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	vendor, err := h.vendorRepo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"vendor not found"}`, http.StatusNotFound)
		return
	}
	if vendor.Status == "Rejected" || vendor.Status == "Approved" {
		http.Error(w, `{"error":"vendor cannot be rejected from current status"}`, http.StatusBadRequest)
		return
	}
	oldStatus := vendor.Status
	if err := h.vendorRepo.UpdateStatus(r.Context(), code, "Rejected"); err != nil {
		http.Error(w, `{"error":"failed to update status"}`, http.StatusInternalServerError)
		return
	}
	if err := h.logTransition(w, r, vendor.ID, oldStatus, "Rejected"); err != nil {
		return
	}
	w.WriteHeader(http.StatusNoContent)
}