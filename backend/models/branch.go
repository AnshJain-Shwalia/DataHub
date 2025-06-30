package models

import "time"

type Branch struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	Name      string    `gorm:"column:name;type:text;not null"`
	RepoID    string    `gorm:"column:repo_id;type:uuid;not null;index"`
	Repo      Repo      `gorm:"foreignKey:RepoID;references:ID"`
	Chunks    []Chunk   `gorm:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
}
