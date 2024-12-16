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
}

type RefreshRequest struct {
	AccessToken  string `json:"accessToken" validate:"required,jwt"`
	RefreshToken string `json:"refreshToken" validate:"required"`
	IP           string `json:"ip" validate:"required,ip"`
}

type GetUserRequest struct {
	AccessToken string `validate:"required,jwt"`
}

type GetUsersRequest struct {
	Limit  int `json:"limit" validate:"gte=0"`  // Лимит записей
	Offset int `json:"offset" validate:"gte=0"` // Смещение для пагинации
}
