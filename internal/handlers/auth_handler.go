package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/uswuth/vytora-backend/internal/models"
	"github.com/uswuth/vytora-backend/internal/repository"
	"github.com/uswuth/vytora-backend/internal/services"
	"golang.org/x/crypto/bcrypt"
)

type AuthHandler struct {
	userRepo   *repository.UserRepository
	jwtService *services.JWTService
	validate   *validator.Validate
}

func NewAuthHandler(userRepo *repository.UserRepository, jwtService *services.JWTService) *AuthHandler {
	return &AuthHandler{
		userRepo:   userRepo,
		jwtService: jwtService,
		validate:   validator.New(),
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusUnauthorized)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, `{"error":"validation failed"}`, http.StatusBadRequest)
		return
	}
	user, err := h.userRepo.FindByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, `{"error": "invalid email or password"}`, http.StatusUnauthorized)
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid email or password"}`, http.StatusUnauthorized)
		return
	}
	if !user.IsActive {
		http.Error(w, `{"error":"account is deactivated"}`, http.StatusForbidden)
		return
	}
	token, err := h.jwtService.GenerateToken(user)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}
	resp := models.LoginResponse{
		Token: token,
		User:  *user,
	}
	w.Header().Set("Context-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
