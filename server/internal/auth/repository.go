package auth

import (
	"time"

	"gorm.io/gorm"
)

type AuthRepository interface {
}

type HttpSessionDb struct {
	ID         int64     `gorm:"primary_key"`
	Key        string    `gorm:"column:key"`
	Data       string    `gorm:"column:data"`
	CreatedOn  time.Time `gorm:"column:created_on"`
	ExpiresOn  time.Time `gorm:"column:expires_on"`
	ModifiedOn time.Time `gorm:"column:modified_on"`
}

func (h *HttpSessionDb) TableName() string {
	return "http_sessions"
}

func (h *HttpSessionDb) Save(db *gorm.DB) error {
	return db.Save(h).Error
}
