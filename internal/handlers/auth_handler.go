package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/uswuth/vytora-backend/internal/entity/user"
	"github.com/uswuth/vytora-backend/internal/services"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo    *user.Repository
	jwtService  *services.JWTService
	validate    *validator.Validate
}

func NewAuthHandler(userRepo *user.Repository, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		jwtService:  jwtService,
		validate:    validator.New(),
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req user.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusUnauthorized)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, `{"error":"validation failed"}`, http.StatusBadRequest)
		return
	}
	u, err := h.userRepo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, `{"error": "invalid email or password"}`, http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}
	if !u.IsActive {
		http.Error(w, `{"error":"account is deactivated"}`, http.StatusForbidden)
		return
	}
	token, err := h.jwtService.GenerateToken(u)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}
	resp := user.LoginResponse{
		Token: token,
		User:  *u,
	}
	w.Header().Set("Context-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
