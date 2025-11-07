package login

type Request struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type Response struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}
