package auth

import (
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
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

func (h *HttpSessionDb) Save() error {
	db := akdb.DefaultDB
	return db.Save(h).Error
}
