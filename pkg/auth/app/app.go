package app

import (
	"fmt"
	"github.com/CracherX/auth_hist/internal/api/endpoints"
	"github.com/CracherX/auth_hist/internal/config"
	"github.com/CracherX/auth_hist/internal/logger"
	"github.com/CracherX/auth_hist/internal/middleware"
	"github.com/CracherX/auth_hist/internal/router"
	"github.com/CracherX/auth_hist/internal/services"
	"github.com/CracherX/auth_hist/internal/storage/db"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
)

type App struct {
	Config    *config.Config
	Logger    *zap.Logger
	DB        *gorm.DB
	Validator *validator.Validate
	Router    *mux.Router
}

func New() (app *App, err error) {
	app = &App{}

	app.Config = config.MustLoad()
	app.Logger = logger.MustInit(app.Config.Server.Debug)

	app.DB, err = db.Connect(app.Config, app.Config.Database.Retries)
	if err != nil {
		app.Logger.Fatal("Ошибка подключения к БД", zap.Error(err))
	}
	app.Logger.Info("Успешное подключение к БД")

	app.Validator = validator.New()

	app.Router = router.Setup()
	app.Router.Use(middleware.Validate(app.Validator), middleware.Logging(app.Logger))

	as := services.NewAuth(app.DB, app.Config)

	ep := endpoints.New(as, app.Logger, app.Validator)

	router.Auth(app.Router, ep)

	return app, nil
}

// Run запуск приложения.
func (a *App) Run() {
	a.Logger.Info("Запуск приложения", zap.String("Приложение:", a.Config.Server.AppName))
	a.Logger.Debug("Запущен режим отладки для терминала!")
	err := http.ListenAndServe(a.Config.Server.Port, a.Router)
	if err != nil {
		fmt.Println(err)
		a.Logger.Error("Ошибка запуска сервера")
	}
}
