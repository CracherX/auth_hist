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
	json.NewEncoder(w).Encode(errorResponse)
}
