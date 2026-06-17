package category

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/uswuth/vytora-backend/internal/middleware"
	"github.com/uswuth/vytora-backend/internal/services"
)

type NextCodeFunc func(ctx context.Context, entity string) (string, error)

type Handler struct {
	repo       *Repository
	nextCode   NextCodeFunc
	validate   *validator.Validate
}

func NewHandler(repo *Repository, nextCode NextCodeFunc) *Handler {
	return &Handler{
		repo:       repo,
		nextCode:   nextCode,
		validate:   validator.New(),
	}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, `{"error":"validation failed"}`, http.StatusBadRequest)
		return
	}

	existing, _ := h.repo.FindByName(r.Context(), req.Name)
	if existing != nil {
		http.Error(w, `{"error":"category with this name already exists"}`, http.StatusConflict)
		return
	}

	code, err := h.nextCode(r.Context(), "category")
	if err != nil {
		http.Error(w, `{"error":"failed to generate code"}`, http.StatusInternalServerError)
		return
	}

	claims := r.Context().Value(middleware.UserContextKey).(*services.Claims)
	createdBy, _ := uuid.Parse(claims.UserID)

	cat := &Category{
		Code:        code,
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		Status:      req.Status,
		CreatedBy:   &createdBy,
	}

	if err := h.repo.Create(r.Context(), cat); err != nil {
		http.Error(w, `{"error":"failed to create category"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(cat)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	params := ListParams{
		Search: r.URL.Query().Get("search"),
		Status: r.URL.Query().Get("status"),
	}
	if v := r.URL.Query().Get("limit"); v != "" {
		params.Limit, _ = strconv.Atoi(v)
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		params.Offset, _ = strconv.Atoi(v)
	}

	cats, total, err := h.repo.List(r.Context(), params)
	if err != nil {
		http.Error(w, `{"error":"failed to list categories"}`, http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"data":  cats,
		"total": total,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	cat, err := h.repo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"category not found"}`, http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cat)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	cat, err := h.repo.FindByCode(r.Context(), code)
	if err != nil {
		http.Error(w, `{"error":"category not found"}`, http.StatusNotFound)
		return
	}

	var req UpdateCategoryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, `{"error":"validation failed"}`, http.StatusBadRequest)
		return
	}

	claims := r.Context().Value(middleware.UserContextKey).(*services.Claims)
	updatedBy, _ := uuid.Parse(claims.UserID)

	cat.DisplayName = req.DisplayName
	cat.Description = req.Description
	cat.Status = req.Status
	cat.UpdatedBy = &updatedBy

	if err := h.repo.Update(r.Context(), cat); err != nil {
		http.Error(w, `{"error":"failed to update category"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cat)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	if err := h.repo.Delete(r.Context(), code); err != nil {
		http.Error(w, `{"error":"failed to delete category"}`, http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}