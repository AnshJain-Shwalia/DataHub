package models

import "time"

type File struct {
	ID        string    `gorm:"primaryKey;type:uuid"`
	Name      string    `gorm:"column:name;type:text;not null"`
	FolderID  *string   `gorm:"column:folder_id;type:uuid"`
	Folder    *Folder   `gorm:"foreignKey:FolderID;references:ID"`
	Size      int64     `gorm:"column:size;type:bigint;not null"`
	UserID    string    `gorm:"column:user_id;type:uuid;not null;index"`
	User      User      `gorm:"foreignKey:UserID;references:ID"`
	Chunks    []Chunk   `gorm:"-"`
	CreatedAt time.Time `gorm:"column:created_at;type:timestamptz;not null"`
}
