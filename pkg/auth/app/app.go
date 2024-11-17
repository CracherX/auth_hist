package app

import (
	"github.com/CracherX/auth_hist/internal/config"
	"github.com/CracherX/auth_hist/internal/logger"
	"github.com/CracherX/auth_hist/internal/middleware"
	"github.com/CracherX/auth_hist/internal/router"
	"github.com/CracherX/auth_hist/internal/storage/db"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"gorm.io/gorm"
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

	app.DB, err = db.Connect(app.Config, 5)
	if err != nil {
		app.Logger.Fatal("Ошибка подключения к БД", zap.Error(err))
	}
	app.Logger.Info("Успешное подключение к БД")

	app.Validator = validator.New()

	app.Router = router.Setup()
	app.Router.Use(middleware.Validate(app.Validator), middleware.Logging(app.Logger))
	router.Auth(app.Router)

	return app, nil
}
