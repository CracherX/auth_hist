package endpoints

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/CracherX/auth_hist/internal/dto"
	"github.com/CracherX/auth_hist/internal/services"
	"github.com/CracherX/auth_hist/internal/storage/models"
	"github.com/golang-jwt/jwt/v4"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"net/http"
	"strconv"
)

type AuthService interface {
	Login(usr string, pass string) (int, error)
	CreateRefreshTkn(id int, ip string) (*models.RefreshTokens, string, error)
	CreateAccessTkn(id int, rid int, ip string) (string, error)
	RefreshTkns(accTkn string, refTkn string, ip string) (string, string, error)
	Register(request *dto.RegisterRequest) error
	GetUser(dto *dto.GetUserRequest) (*models.Users, error)
	GetUsers(req *dto.GetUsersRequest) (*dto.GetUsersResponse, error)
}

type Logger interface {
	Error(msg string, args ...zap.Field)
	Info(msg string, args ...zap.Field)
}

type Validator interface {
	Struct(s interface{}) error
}

type Endpoint struct {
	Service   AuthService
	Logger    Logger
	Validator Validator
}

func New(ser AuthService, log Logger, valid Validator) *Endpoint {
	ep := &Endpoint{
		Service:   ser,
		Logger:    log,
		Validator: valid,
	}
	return ep
}

func (ep *Endpoint) Auth(w http.ResponseWriter, r *http.Request) {
	var reqDat dto.AuthRequest

	err := json.NewDecoder(r.Body).Decode(&reqDat)
	err = ep.Validator.Struct(&reqDat)
	if err != nil {
		ep.Logger.Info("Bad Request на запрос авторизации", zap.String("IP", reqDat.IP))
		dto.Response(w, http.StatusBadRequest, "Bad Request", "Обратитесь к документации и заполните тело запроса правильно")
		return
	}

	id, err := ep.Service.Login(reqDat.Username, reqDat.Password)
	if err != nil {
		ep.Logger.Info("Unauthorized на запрос авторизации", zap.String("IP", reqDat.IP))
		dto.Response(w, http.StatusUnauthorized, "Unauthorized", "Введенный логин и пароль недействительны")
		return
	}

	refMod, refTkn, err := ep.Service.CreateRefreshTkn(id, reqDat.IP)
	if err != nil {
		if errors.Is(err, driver.ErrBadConn) {
			ep.Logger.Error("Ошибка подключения к БД при выполнении запроса авторизации", zap.String("IP", reqDat.IP))
			dto.Response(w, http.StatusBadGateway, "Bad Gateway", "Ошибка в работе внешних сервисов")
			return
		} else {
			ep.Logger.Error("Ошибка создания Refresh-токена", zap.String("Ошибка", err.Error()), zap.String("IP", reqDat.IP))
			dto.Response(w, http.StatusInternalServerError, "Internal Server Response", "Внутренняя ошибка сервера, обратитесь к техническому специалисту")
		}
		return
	}

	accTkn, err := ep.Service.CreateAccessTkn(id, refMod.ID, reqDat.IP)
	if err != nil {
		ep.Logger.Error("Ошибка создания Access-токена", zap.String("Ошибка", err.Error()), zap.String("IP", reqDat.IP))
		dto.Response(w, http.StatusInternalServerError, "Internal Server Response", "Внутренняя ошибка сервера, обратитесь к техническому специалисту")
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	err = json.NewEncoder(w).Encode(&dto.TokenResponse{
		AccessToken:  accTkn,
		RefreshToken: refTkn,
	})
	if err != nil {
		ep.Logger.Error("Ошибка записи в Response Writer во время выполнения запроса авторизации", zap.String("IP", reqDat.IP))
		dto.Response(w, http.StatusInternalServerError, "Internal Server Response", "Внутренняя ошибка сервера, обратитесь к техническому специалисту")
		return
	}
}

func (ep *Endpoint) Refresh(w http.ResponseWriter, r *http.Request) {
	var reqDat dto.RefreshRequest

	err := json.NewDecoder(r.Body).Decode(&reqDat)
	err = ep.Validator.Struct(&reqDat)
	if err != nil {
		fmt.Println(err.Error())
		ep.Logger.Info("Bad Request на запрос рефреш операции", zap.String("IP", reqDat.IP))
		dto.Response(w, http.StatusBadRequest, "Bad Request", "Обратитесь к документации и заполните тело запроса правильно")
		return
	}
	accTkn, refTkn, err := ep.Service.RefreshTkns(reqDat.AccessToken, reqDat.RefreshToken, reqDat.IP)
	if err != nil {
		switch {
		case errors.Is(err, driver.ErrBadConn):
			ep.Logger.Error("Ошибка подключения к БД при выполнении запроса рефреш операции", zap.String("IP", reqDat.IP))
			dto.Response(w, http.StatusBadGateway, "Bad Gateway", "Ошибка в работе внешних сервисов")
		case errors.Is(err, services.ErrInvalidRefreshToken) || errors.Is(err, services.ErrInvalidToken):
			ep.Logger.Info(err.Error(), zap.String("IP", reqDat.IP))
			dto.Response(w, http.StatusUnauthorized, "Unauthorized", "Пара Refresh/Access токен не валидна")
		case errors.Is(err, jwt.ErrTokenExpired):
			ep.Logger.Info(err.Error(), zap.String("IP", reqDat.IP))
			dto.Response(w, http.StatusUnauthorized, "Unauthorized", "Срок действия токена истек")
		default:
			ep.Logger.Error("Ошибка при выполнении рефреш операции", zap.String("IP", reqDat.IP), zap.String("Ошибка", err.Error()))
			dto.Response(w, http.StatusInternalServerError, "Internal Server Response", "Внутренняя ошибка сервера, обратитесь к техническому специалисту")
		}
		return
	}

	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	err = json.NewEncoder(w).Encode(&dto.TokenResponse{
		AccessToken:  accTkn,
		RefreshToken: refTkn,
	})
	if err != nil {
		ep.Logger.Error("Ошибка записи в Response Writer во время выполнения запроса рефреш операции", zap.String("IP", reqDat.IP))
		dto.Response(w, http.StatusInternalServerError, "Internal Server Response", "Внутренняя ошибка сервера, обратитесь к техническому специалисту")
		return
	}
}

func (ep *Endpoint) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var reqDat dto.RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&reqDat)
	err = ep.Validator.Struct(&reqDat)
	if err != nil {
		ep.Logger.Info("Bad Request на запрос регистрации")
		dto.Response(w, http.StatusBadRequest, "Bad Request", "Обратитесь к документации и заполните тело запроса правильно")
		return
	}
	err = ep.Service.Register(&reqDat)
	if err != nil {
		switch {
		case errors.Is(err, driver.ErrBadConn):
			ep.Logger.Error("Ошибка подключения к БД при выполнении запроса рефреш операции")
			dto.Response(w, http.StatusBadGateway, "Bad Gateway", "Ошибка в работе внешних сервисов")
		case errors.Is(err, gorm.ErrRegistered):
			ep.Logger.Info("Попытка регистрации уже существующего пользователя")
			dto.Response(w, http.StatusConflict, "Conflict", "Пользователь с таким Email или именем уже существует")
		default:
			ep.Logger.Error("Ошибка при выполнении регистрации нового пользователя", zap.String("Ошибка", err.Error()))
			dto.Response(w, http.StatusInternalServerError, "Internal Server Response", "Внутренняя ошибка сервера, обратитесь к техническому специалисту")
		}
		return
	}
	dto.Response(w, http.StatusCreated, "Created", "Новый пользователь успешно зарегистрирован!")
}

func (ep *Endpoint) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := r.URL.Query()
	data := &dto.GetUserRequest{AccessToken: vars.Get("jwt")}
	err := ep.Validator.Struct(data)
	if err != nil {
		ep.Logger.Info("Bad Request на запрос получения данных о пользователях", zap.String("Ошибка", err.Error()))
		dto.Response(w, http.StatusBadRequest, "Bad Request", "Обратитесь к документации и заполните тело запроса правильно")
		return
	}
	user, err := ep.Service.GetUser(data)
	if err != nil {
		switch {
		case errors.Is(err, driver.ErrBadConn):
			ep.Logger.Error("Ошибка подключения к БД при выполнении запроса рефреш операции")
			dto.Response(w, http.StatusBadGateway, "Bad Gateway", "Ошибка в работе внешних сервисов")
		case errors.Is(err, services.ErrInvalidToken):
			ep.Logger.Info(err.Error())
			dto.Response(w, http.StatusUnauthorized, "Unauthorized", "Access токен не валиден")
		case errors.Is(err, jwt.ErrTokenExpired):
			ep.Logger.Info(err.Error())
			dto.Response(w, http.StatusUnauthorized, "Unauthorized", "Срок действия токена истек")
		default:
			ep.Logger.Error("Ошибка при выполнении операции получения данных пользователя", zap.String("Ошибка", err.Error()))
			dto.Response(w, http.StatusInternalServerError, "Internal Server Response", "Внутренняя ошибка сервера, обратитесь к техническому специалисту")
		}
		return
	}
	res := dto.UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Picture:  user.Picture,
		Password: user.Password,
		Email:    user.Email,
		IsAdmin:  user.IsAdmin,
	}
	w.Header().Add("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	err = json.NewEncoder(w).Encode(&res)
	if err != nil {
		ep.Logger.Error("Ошибка записи в Response Writer во время выполнения запроса получения данных пользователя")
		dto.Response(w, http.StatusInternalServerError, "Internal Server Response", "Внутренняя ошибка сервера, обратитесь к техническому специалисту")
		return
	}
}

func (ep *Endpoint) GetUsers(w http.ResponseWriter, r *http.Request) {
	// Получение параметров лимита и смещения из запроса
	vars := r.URL.Query()
	limitStr := vars.Get("limit")
	offsetStr := vars.Get("offset")

	// Преобразование параметров в числа
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 10 // Значение по умолчанию
	}
	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0 // Значение по умолчанию
	}

	// Формирование DTO запроса
	req := &dto.GetUsersRequest{
		Limit:  limit,
		Offset: offset,
	}

	// Валидация DTO
	if err := ep.Validator.Struct(req); err != nil {
		ep.Logger.Info("Bad Request для запроса всех пользователей", zap.String("Ошибка", err.Error()))
		dto.Response(w, http.StatusBadRequest, "Bad Request", "Обратитесь к документации и корректно укажите параметры запроса")
		return
	}

	// Выполнение метода UseCase для получения пользователей
	usersResponse, err := ep.Service.GetUsers(req)
	if err != nil {
		ep.Logger.Error("Ошибка получения данных о пользователях", zap.String("Ошибка", err.Error()))
		dto.Response(w, http.StatusInternalServerError, "Internal Server Error", "Не удалось получить данные пользователей")
		return
	}

	// Успешный ответ
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if err := json.NewEncoder(w).Encode(usersResponse); err != nil {
		ep.Logger.Error("Ошибка записи в Response Writer при возврате списка пользователей", zap.String("Ошибка", err.Error()))
		dto.Response(w, http.StatusInternalServerError, "Internal Server Error", "Ошибка на стороне сервера")
	}
}
