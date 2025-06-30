package models

import "time"

type User struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	Name      string    `gorm:"column:name;type:varchar(100);not null"`
	Email     string    `gorm:"column:email;type:varchar(255);not null;unique;index"`
	Tokens    []Token   `gorm:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;type:timestamptz;not null"`
}
