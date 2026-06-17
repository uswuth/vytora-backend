package user

type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	FullName string `json:"full_name" validate:"required,min=2,max=255"`
	Role     string `json:"role" validate:"required,oneof=system_admin risk_manager compliance_officer department_manager auditor"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=system_admin risk_manager compliance_officer department_manager auditor"`
}