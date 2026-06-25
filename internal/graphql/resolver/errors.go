package resolver

import "fmt"

type InvalidCredentialsError struct{}
func (e *InvalidCredentialsError) Error() string { return "invalid email or password" }
func (e *InvalidCredentialsError) Extensions() map[string]interface{} {
	return map[string]interface{}{"code": "UNAUTHORIZED"}
}

type AccountDeactivatedError struct{}
func (e *AccountDeactivatedError) Error() string { return "account is deactivated" }
func (e *AccountDeactivatedError) Extensions() map[string]interface{} {
	return map[string]interface{}{"code": "FORBIDDEN"}
}

type UnauthorizedError struct{}
func (e *UnauthorizedError) Error() string { return "unauthorized" }
func (e *UnauthorizedError) Extensions() map[string]interface{} {
	return map[string]interface{}{"code": "UNAUTHORIZED"}
}

type ValidationError struct{ Message string }
func (e *ValidationError) Error() string { return e.Message }

type ForbiddenError struct{ Message string }
func (e *ForbiddenError) Error() string { return e.Message }

type NotFoundError struct{ Resource string }
func (e *NotFoundError) Error() string { return fmt.Sprintf("%s not found", e.Resource) }