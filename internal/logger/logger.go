package logger

import (
	"go.uber.org/zap"
	"log"
)

func MustInit(deb bool) (logger *zap.Logger) {
	var err error
	if deb {
		logger, err = zap.NewDevelopment()
		if err != nil {
			log.Fatalf("Ошибка инициализации логгера разработки")
		}
	} else {
		logger, err = zap.NewProduction()
		if err != nil {
			log.Fatalf("Ошибка инициализации логгера")
		}
	}
	logger.Info("Логгер успешно инициализирован")
	return logger
}
