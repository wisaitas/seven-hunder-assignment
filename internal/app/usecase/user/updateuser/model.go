package updateuser

type Request struct {
	Name  *string `json:"name" validate:"omitempty"`
	Email *string `json:"email" validate:"omitempty,email"`
}
