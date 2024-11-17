package middleware

import (
	"context"
	"github.com/go-playground/validator/v10"
	"net/http"
)

const validContext = "validator"

// Validate добавляет middleware передающий validator приложения в контекст запроса.
func Validate(valid *validator.Validate) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), validContext, valid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetValidator получает validator переданный в контекст запроса.
func GetValidator(ctx context.Context) *validator.Validate {
	valid, ok := ctx.Value(validContext).(*validator.Validate)
	if !ok {
		panic("Не могу получить валидатор из контекста")
		return nil
	}
	return valid
}
