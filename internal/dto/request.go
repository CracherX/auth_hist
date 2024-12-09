package dto

type AuthRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
	Password string `json:"password" validate:"required"`
	IP       string `json:"ip" validate:"required,ip"`
}

type RegisterRequest struct {
	Username string `json:"username" validate:"required,min=3,max=20,alphanum"`
	Password string `json:"password" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	About    string `json:"about"`
}

type RefreshRequest struct {
	AccessToken  string `json:"accessToken" validate:"required,jwt"`
	RefreshToken string `json:"refreshToken" validate:"required"`
	IP           string `json:"ip" validate:"required,ip"`
}
