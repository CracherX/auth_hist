package models

import "time"

// Users модель пользователей БД.
type Users struct {
	ID       int    `gorm:"type:int;primaryKey"`
	Username string `gorm:"type:varchar(60);uniqueIndex;not null"`
	Password string `gorm:"type:varchar(255);not null"`
	Email    string `gorm:"type:varchar(60);uniqueIndex;not null"`
	About    string `gorm:"type:text"`
}

// RefreshTokens модель Refresh токенов БД.
type RefreshTokens struct {
	ID        int       `gorm:"primaryKey;autoIncrement"`
	Token     string    `gorm:"type:varchar;uniqueIndex;not null"`
	UserId    int       `gorm:"type:int;not null"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null"`
	CreatedAt time.Time `gorm:"type:timestamp;autoCreateTime;not null"`
	UpdatedAt time.Time `gorm:"type:timestamp;autoUpdateTime;not null"`
	IP        string    `gorm:"type:varchar(45);not null"`
	Revoked   bool      `gorm:"type:boolean;not null;default:false"`

	User Users `gorm:"foreignKey:UserId"`
}
