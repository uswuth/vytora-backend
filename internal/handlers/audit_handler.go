package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/uswuth/vytora-backend/internal/repository"
)

type AuditHandler struct {
	auditRepo *repository.AuditTrailRepository
}

func NewAuditHandler(auditRepo *repository.AuditTrailRepository) *AuditHandler {
	return &AuditHandler{auditRepo: auditRepo}
}

func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	params := repository.AuditListParams{
		TableName:  r.URL.Query().Get("table"),
		RecordCode: r.URL.Query().Get("record_code"),
		Action:     r.URL.Query().Get("action"),
		ChangedBy:  r.URL.Query().Get("changed_by"),
		DateFrom:   r.URL.Query().Get("date_from"),
		DateTo:     r.URL.Query().Get("date_to"),
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		params.Limit, _ = strconv.Atoi(v)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		params.Offset, _ = strconv.Atoi(v)
	}

	audits, total, err := h.auditRepo.List(r.Context(), params)
	if err != nil {
		http.Error(w, `{"error":"failed to list audit trail"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"data":  audits,
		"total": total,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}