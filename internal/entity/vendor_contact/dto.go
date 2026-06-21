package vendor_contact

type CreateContactRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=255"`
	Email string `json:"email" validate:"omitempty,email"`
	Phone string `json:"phone"`
}

type UpdateContactRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=255"`
	Email string `json:"email" validate:"omitempty,email"`
	Phone string `json:"phone"`
}