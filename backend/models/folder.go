package models

import "time"

type Folder struct {
	ID             string    `gorm:"primaryKey;type:uuid"`
	Name           string    `gorm:"column:name;type:text;not null"`
	ParentFolderID *string   `gorm:"column:parent_folder_id;type:uuid"`
	ParentFolder   *Folder   `gorm:"foreignKey:ParentFolderID;references:ID"`
	UserID         string    `gorm:"column:user_id;type:uuid;not null;index"`
	User           User      `gorm:"foreignKey:UserID;references:ID"`
	Files          []File    `gorm:"-"`
	Subfolders     []Folder  `gorm:"-"`
	CreatedAt      time.Time `gorm:"column:created_at;type:timestamptz;not null"`
}
