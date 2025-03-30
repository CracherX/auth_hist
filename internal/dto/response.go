package dto

import (
	"encoding/json"
	"net/http"
)

// TokenResponse DTO структура ответа содержащая Access и Refresh токены.
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"`
	Picture  string `json:"picture"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"isAdmin"`
}

type e struct {
	Status  int    `json:"status"`
	Error   string `json:"error"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// Response возвращает сообщение об ошибке клиенту в json формате.
func Response(w http.ResponseWriter, status int, errMsg string, details ...string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	errorResponse := e{
		Status:  status,
		Error:   http.StatusText(status),
		Message: errMsg,
	}
	if len(details) > 0 {
		errorResponse.Details = details[0]
	}
	w.Header().Add("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(errorResponse)
}

type GetUsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"` // Общее количество пользователей
}
