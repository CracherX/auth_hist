package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/CracherX/auth_hist/internal/config"
	"github.com/CracherX/auth_hist/internal/dto"
	"github.com/CracherX/auth_hist/internal/storage/models"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"io"
	"os"
	"time"
)

type AuthService struct {
	DB     *gorm.DB
	Config *config.Config
}

func NewAuth(db *gorm.DB, cfg *config.Config) *AuthService {
	ser := &AuthService{
		DB:     db,
		Config: cfg,
	}
	return ser
}

func (as *AuthService) Login(usr string, pass string) (int, error) {
	var user models.Users
	err := as.DB.Where("username = ?", usr).First(&user).Error
	if err != nil {
		return 0, gorm.ErrRecordNotFound
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(pass))
	if err != nil {
		return 0, gorm.ErrRecordNotFound
	}
	return user.ID, nil
}

func (as *AuthService) CreateRefreshTkn(id int, ip string) (*models.RefreshTokens, string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, "", err
	}
	tkn := base64.StdEncoding.EncodeToString(bytes)

	htkn, err := bcrypt.GenerateFromPassword([]byte(tkn), bcrypt.DefaultCost)

	modelToken := &models.RefreshTokens{
		Token:     string(htkn),
		UserId:    id,
		ExpiresAt: time.Now().Add(24 * time.Hour * 7),
		IP:        ip,
	}
	err = as.DB.Create(modelToken).Error
	if err != nil {
		return nil, "", err
	}
	return modelToken, tkn, nil
}

func (as *AuthService) CreateAccessTkn(id int, rid int, ip string) (string, error) {
	tkn := jwt.NewWithClaims(
		jwt.SigningMethodRS512,
		jwt.MapClaims{
			"sub": id,
			"iat": time.Now().Unix(),
			"exp": time.Now().Add(time.Hour).Unix(),
			"ip":  ip,
			"rid": rid,
		},
	)
	key, err := loadKey(as.Config.Server.JwtSecretPath, false)
	if err != nil {
		return "", err
	}
	sigTkn, err := tkn.SignedString(key)
	if err != nil {
		return "", err
	}
	return sigTkn, nil
}

func (as *AuthService) RefreshTkns(accTkn string, refTkn string, ip string) (string, string, error) {
	token, err := jwt.Parse(
		accTkn,
		func(token *jwt.Token) (interface{}, error) {
			return loadKey(as.Config.Server.JwtPublicPath, true)
		},
	)
	if (err != nil || !token.Valid) && !errors.Is(err, jwt.ErrTokenExpired) {
		return "", "", ErrInvalidToken
	}
	claims := token.Claims.(jwt.MapClaims)
	rid := claims["rid"]
	var expectedRefresh models.RefreshTokens
	err = as.DB.Where("id = ?", rid).First(&expectedRefresh).Error
	if expectedRefresh.Revoked || err != nil {
		return "", "", ErrInvalidRefreshToken
	}
	if time.Now().After(expectedRefresh.ExpiresAt) {
		expectedRefresh.Revoked = true
		as.DB.Save(&expectedRefresh)
		return "", "", jwt.ErrTokenExpired
	}
	err = bcrypt.CompareHashAndPassword([]byte(expectedRefresh.Token), []byte(refTkn))
	if err != nil {
		return "", "", ErrInvalidRefreshToken
	}
	expectedRefresh.Revoked = true
	as.DB.Save(&expectedRefresh)
	newRefTkn, _, err := as.CreateRefreshTkn(expectedRefresh.UserId, ip)
	if err != nil {
		return "", "", err
	}
	newAccTkn, err := as.CreateAccessTkn(newRefTkn.UserId, newRefTkn.ID, ip)
	if err != nil {
		return "", "", err
	}
	return newAccTkn, newRefTkn.Token, nil
}

func (as *AuthService) Register(dto *dto.RegisterRequest) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(dto.Password), bcrypt.DefaultCost)
	newUser := models.Users{
		Username: dto.Username,
		Password: string(hash),
		Email:    dto.Email,
	}
	err = as.DB.Create(&newUser).Error
	if err != nil {
		return err
	}
	return nil
}

func (as *AuthService) GetUser(dto *dto.GetUserRequest) (*models.Users, error) {
	var user models.Users

	tkn, err := jwt.Parse(dto.AccessToken,
		func(token *jwt.Token) (interface{}, error) {
			return loadKey(as.Config.Server.JwtPublicPath, true)
		})
	if err != nil {
		return nil, err
	}
	if !tkn.Valid {
		return nil, ErrInvalidToken
	}
	claims := tkn.Claims.(jwt.MapClaims)

	id := claims["sub"]

	err = as.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func loadKey(path string, public bool) (key interface{}, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func(file *os.File) {
		_ = file.Close()
	}(file)

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	if public {
		key, err = jwt.ParseRSAPublicKeyFromPEM(data)
	} else {
		key, err = jwt.ParseRSAPrivateKeyFromPEM(data)
	}
	if err != nil {
		return nil, err
	}
	return key, nil
}
