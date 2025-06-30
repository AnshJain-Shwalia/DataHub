package models

import "time"

type Repo struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	GithubID  *string   `gorm:"column:github_id;type:text"`
	TokenID   string    `gorm:"column:token_id;type:uuid;not null;index"`
	Token     Token     `gorm:"foreignKey:TokenID;references:ID"`
	Name      string    `gorm:"column:name;type:text;not null"`
	Branches  []Branch  `gorm:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
}
