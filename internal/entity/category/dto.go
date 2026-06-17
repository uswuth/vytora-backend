package category

type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=100"`
	DisplayName string `json:"display_name" validate:"required,min=1,max=255"`
	Description string `json:"description"`
	Status      string `json:"status" validate:"required,oneof=Draft Active Inactive"`
}

type UpdateCategoryRequest struct {
	DisplayName string `json:"display_name" validate:"required,min=1,max=255"`
	Description string `json:"description"`
	Status      string `json:"status" validate:"required,oneof=Draft Active Inactive"`
}