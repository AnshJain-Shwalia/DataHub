package models

import "time"

type Chunk struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	FileID    string    `gorm:"column:file_id;type:uuid;not null;index"`
	File      File      `gorm:"foreignKey:FileID;references:ID"`
	Rank      int       `gorm:"column:rank;type:int;not null"`
	Size      int64     `gorm:"column:size;type:bigint;not null"`
	Path      string    `gorm:"column:path;type:text;not null"`
	BranchID  string    `gorm:"column:branch_id;type:uuid;not null;index"`
	Branch    Branch    `gorm:"foreignKey:BranchID;references:ID"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
}
