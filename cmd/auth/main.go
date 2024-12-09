package main

import "github.com/CracherX/auth_hist/pkg/auth/app"

func main() {
	App, err := app.New()
	if err != nil {
		panic("Ошибка запуска приложения")
	}
	App.Run()
}
