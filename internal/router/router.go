package router

import "github.com/gorilla/mux"

// Setup устанавливает главный роутер
func Setup() *mux.Router {
	r := mux.NewRouter()
	return r
}

// Auth - устанавливает auth/ саброутер для главного mr роутера
func Auth(mr *mux.Router, ep *api.TokenEndpoint) *mux.Router {
	authRoute := mr.PathPrefix("/auth").Subrouter()
	authRoute.HandleFunc("/token/{GUID}", ep.Access).Methods("GET")
	authRoute.HandleFunc("/refresh", ep.Refresh).Methods("POST")
	return authRoute
}
