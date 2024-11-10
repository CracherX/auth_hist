package app

import (
	"github.com/CracherX/auth_hist/internal/config"
	"github.com/CracherX/auth_hist/internal/logger"
	"github.com/CracherX/auth_hist/internal/storage/db"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type App struct {
	Config *config.Config
	Logger *zap.Logger
	DB     *gorm.DB
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

	return app, nil
}
