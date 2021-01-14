package models

import (
	"encoding/json"
	"time"

	"github.com/go-playground/validator"
	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// MessageDataSocket mensagem usuario
type MessageDataSocket struct {
	TypeMessage string      `json:"type"`
	PersonID    int         `json:"personId"`
	Data        interface{} `json:"data"`
}

// Base contains common columns for all tables.
type Base struct {
	ID        uuid.UUID `json:"id,omitempty" gorm:"primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `sql:"index" json:"-"`
}

// BeforeCreate will set a UUID rather than numeric ID.
func (base *Base) BeforeCreate(tx *gorm.DB) error {
	nonce, err := uuid.NewV4()
	if err != nil {
		return err
	}
	if base.ID == uuid.Nil {
		base.ID = nonce
	}
	return nil
}

//IsValid varificar se o Atividade é valido
func IsValid(p interface{}) error {
	var validate *validator.Validate
	validate = validator.New()
	return validate.Struct(p)
}

// ToJSON convert json models
func ToJSON(p interface{}) []byte {
	rs, _ := json.Marshal(p)
	return rs
}
