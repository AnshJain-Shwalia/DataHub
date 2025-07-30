package models

import "time"

type Token struct {
	ID                   string     `gorm:"primaryKey;type:uuid"`
	UserID               string     `gorm:"column:user_id;type:uuid;not null;index;uniqueIndex:idx_user_platform_account,priority:1"`
	User                 User       `gorm:"foreignKey:UserID;references:ID"`
	Platform             string     `gorm:"column:platform;type:varchar(50);not null;index;uniqueIndex:idx_user_platform_account,priority:2"` // as of now can be "GOOGLE" or "GITHUB"
	AccountIdentifier    *string    `gorm:"column:account_identifier;type:varchar(255);index;uniqueIndex:idx_user_platform_account,priority:3"` // GitHub username or Google email - used to prevent duplicate tokens per account
	AccessToken          string     `gorm:"column:access_token;type:text;not null"`
	AccessTokenExpiry    *time.Time `gorm:"column:access_token_expiry;type:timestamptz"`
	RefreshToken         *string    `gorm:"column:refresh_token;type:text"`
	RefreshTokenExpiry   *time.Time `gorm:"column:refresh_token_expiry;type:timestamptz"`
	AccessTokenIssuedAt  time.Time  `gorm:"column:access_token_issued_at;type:timestamptz;not null"`
	RefreshTokenIssuedAt *time.Time `gorm:"column:refresh_token_issued_at;type:timestamptz"`
	CreatedAt            time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt            time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
}
