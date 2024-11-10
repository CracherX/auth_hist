package config

import (
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"log"
	"net/url"
)

// Config структура конфигурации приложения
type Config struct {
	Server struct {
		AppName string `mapstructure:"app_name"`
		Port    string
		Debug   bool
	}
	Database struct {
		Host     string
		Port     string
		User     string
		Password string
		Name     string
		SslMode  string `mapstructure:"ssl_mode"`
	}
}

func MustLoad() *Config {
	setupFlags()

	pth := pflag.Lookup("configPath").Value.String()
	mustRead(pth)

	updateConfigWithFlags()

	config := parseConfig()

	log.Println("Конфигурация успешно инициализирована")

	return config
}

// MustLoad инициализирует конфигурацию
func mustRead(pth string) {
	viper.SetConfigFile(pth)
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации из файла")
	}
}

// setupFlags назначает флаги командной строки
func setupFlags() {
	pflag.Int("port", 0, "Порт запуска сервера")
	pflag.Bool("debug", false, "Включить режим отладки в терминале")
	pflag.String("db", "", "Строка подключения к базе данных (формат: 'host=localhost port=5432 user=Cracher password=Gleb dbname=test sslmode=disable)")
	pflag.String("configPath", "config/config.yaml", "Путь до файла конфигурации")

	pflag.Parse()
}

// updateConfigWithFlags обновляет конфигурацию на основе флагов командной строки
func updateConfigWithFlags() {
	if pflag.Lookup("port").Changed {
		viper.Set("server.port", pflag.Lookup("port").Value.String())
	}
	if pflag.Lookup("debug").Changed {
		viper.Set("server.debug", pflag.Lookup("debug").Value.String())
	}
	if pflag.Lookup("db").Changed {
		parseDatabaseURL(pflag.Lookup("db").Value.String())
	}
}

// parseDatabaseURL парсит строку подключения к базе данных
func parseDatabaseURL(dsn string) {
	dbURL, err := url.ParseQuery(dsn)
	if err != nil {
		log.Fatalf("Ошибка разбора строки подключения к БД: %v", err)
	}
	viper.Set("database.host", dbURL.Get("host"))
	viper.Set("database.port", dbURL.Get("port"))
	viper.Set("database.user", dbURL.Get("user"))
	viper.Set("database.password", dbURL.Get("password"))
	viper.Set("database.name", dbURL.Get("dbname"))
	viper.Set("database.sslmode", dbURL.Get("sslmode"))
	viper.Set("database.timezone", dbURL.Get("timezone"))
}

// parseConfig парсит конфигурацию в структуру Config
func parseConfig() *Config {
	var config Config
	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("Ошибка записи конфигурации в структуру: %v", err)
	}
	return &config
}
