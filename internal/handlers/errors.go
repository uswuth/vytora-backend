package handlers

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type AppError struct {
	HTTPCode int
	Message  string
}

func MapError(err error) *AppError {
	if err == nil {
		return nil
	}

	if errors.Is(err, pgx.ErrNoRows) {
		return &AppError{HTTPCode: fiber.StatusNotFound, Message: "resource not found"}
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23503": // foreign key violation
			return &AppError{HTTPCode: fiber.StatusConflict, Message: "referenced resource not found"}
		case "23505": // unique violation
			return &AppError{HTTPCode: fiber.StatusConflict, Message: "duplicate entry already exists"}
		case "23514": // check violation
			return &AppError{HTTPCode: fiber.StatusBadRequest, Message: "invalid value for field"}
		}
	}

	msg := err.Error()
	if strings.Contains(msg, "invalid input syntax") {
		return &AppError{HTTPCode: fiber.StatusBadRequest, Message: "invalid input value"}
	}

	return &AppError{HTTPCode: fiber.StatusInternalServerError, Message: "internal server error"}
}

func HandleError(c *fiber.Ctx, err error) error {
	appErr := MapError(err)
	return c.Status(appErr.HTTPCode).JSON(fiber.Map{"error": appErr.Message})
}