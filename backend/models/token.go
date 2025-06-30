package models

import "time"

type Token struct {
	ID                   string     `gorm:"primaryKey;type:uuid"`
	UserID               string     `gorm:"column:user_id;type:uuid;not null;index"`
	User                 User       `gorm:"foreignKey:UserID;references:ID"`
	Platform             string     `gorm:"column:platform;type:varchar(50);not null;index"` // as of now can be "GOOGLE" or "GITHUB"
	AccessToken          string     `gorm:"column:access_token;type:text;not null"`
	AccessTokenExpiry    *time.Time `gorm:"column:access_token_expiry;type:timestamptz"`
	RefreshToken         *string    `gorm:"column:refresh_token;type:text"`
	RefreshTokenExpiry   *time.Time `gorm:"column:refresh_token_expiry;type:timestamptz"`
	AccessTokenIssuedAt  time.Time  `gorm:"column:access_token_issued_at;type:timestamptz;not null"`
	RefreshTokenIssuedAt *time.Time `gorm:"column:refresh_token_issued_at;type:timestamptz"`
	CreatedAt            time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
}
