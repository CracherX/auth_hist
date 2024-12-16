package router

import (
	"github.com/gorilla/mux"
	"net/http"
)

type AuthEndpoint interface {
	Auth(w http.ResponseWriter, r *http.Request)
	Refresh(w http.ResponseWriter, r *http.Request)
	RegisterUser(w http.ResponseWriter, r *http.Request)
	GetUser(w http.ResponseWriter, r *http.Request)
	GetUsers(w http.ResponseWriter, r *http.Request)
}

// Setup устанавливает главный роутер
func Setup() *mux.Router {
	r := mux.NewRouter()
	return r
}

// Auth - устанавливает auth/ саброутер для главного mr роутера
func Auth(mr *mux.Router, ep AuthEndpoint) *mux.Router {
	authRoute := mr.PathPrefix("/auth").Subrouter()
	authRoute.HandleFunc("/access", ep.Auth).Methods(http.MethodPost)
	authRoute.HandleFunc("/refresh", ep.Refresh).Methods(http.MethodPost)
	authRoute.HandleFunc("/register", ep.RegisterUser).Methods(http.MethodPost)
	authRoute.HandleFunc("/profile", ep.GetUser).Methods(http.MethodGet)
	authRoute.HandleFunc("/all", ep.GetUsers).Methods(http.MethodGet)
	return authRoute
}
