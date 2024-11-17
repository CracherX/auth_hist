package services

import (
	"crypto/rand"
	"encoding/base64"
	"github.com/CracherX/auth_hist/internal/config"
	"github.com/CracherX/auth_hist/internal/storage/models"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"io"
	"net"
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
	err := as.DB.Where("username = ? AND password = ?", usr, pass).First(&user).Error
	if err != nil {
		return 0, gorm.ErrRecordNotFound
	}
	return user.ID, nil
}

func (as *AuthService) CreateRefreshTkn(id int, ip string) (*models.RefreshTokens, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
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
		return nil, err
	}
	return modelToken, nil
}

func (as *AuthService) CreateAccessTkn(id int, rid int, ip string) (string, error) {
	tkn := jwt.NewWithClaims(
		jwt.SigningMethodRS512,
		jwt.MapClaims{
			"iss": net.LookupHost("localhost"),
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
	if err != nil || !token.Valid {
		return "", "", ErrInvalidToken
	}
	claims := token.Claims.(jwt.MapClaims)
	rid := claims["rid"].(string)
	var expectedRefresh models.RefreshTokens
	err = as.DB.Where("id = ?", rid).First(&expectedRefresh).Error
	if expectedRefresh.Revoked || err != nil {
		return "", "", ErrInvalidRefreshToken
	}
	err = bcrypt.CompareHashAndPassword([]byte(expectedRefresh.Token), []byte(refTkn))
	if err != nil {
		return "", "", ErrInvalidRefreshToken
	}
	expectedRefresh.Revoked = true
	as.DB.Save(&expectedRefresh)
	newRefTkn, err := as.CreateRefreshTkn(expectedRefresh.UserId, ip)
	if err != nil {
		return "", "", err
	}
	newAccTkn, err := as.CreateAccessTkn(newRefTkn.UserId, newRefTkn.ID, ip)
	if err != nil {
		return "", "", err
	}
	return newRefTkn.Token, newAccTkn, nil
}

func loadKey(path string, public bool) (key interface{}, err error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

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
