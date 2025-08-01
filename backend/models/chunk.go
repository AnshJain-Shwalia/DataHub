package models

import "time"

type Chunk struct {
	ID        string     `gorm:"primaryKey;type:uuid"`
	FileID    string     `gorm:"column:file_id;type:uuid;not null;index"`
	File      File       `gorm:"foreignKey:FileID;references:ID"`
	Rank      int        `gorm:"column:rank;type:int;not null"`
	Size      int64      `gorm:"column:size;type:bigint;not null"`
	S3Path    *string    `gorm:"column:s3_path;type:text"`
	GitPath   *string    `gorm:"column:git_path;type:text"`
	BranchID  *string    `gorm:"column:branch_id;type:uuid;index"`
	Branch    *Branch    `gorm:"foreignKey:BranchID;references:ID"`
	Status    string     `gorm:"column:status;type:varchar(20);not null;default:'BUFFERED'"`
	CreatedAt time.Time  `gorm:"column:created_at;type:timestamptz;not null"`
	UpdatedAt time.Time  `gorm:"column:updated_at;type:timestamptz;not null"`
}
