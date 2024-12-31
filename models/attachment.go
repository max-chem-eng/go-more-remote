package models

import "gorm.io/gorm"

type Attachment struct {
	gorm.Model
	FileName    string `json:"file_name" gorm:"not null"`
	ContentType string
	FileSize    int64
	Content     []byte
	ModelName   string
	ModelID     uint
}

func (a *Attachment) Save() error {
	return Db.Save(a).Error
}
