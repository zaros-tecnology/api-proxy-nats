package models

import "github.com/gofrs/uuid"

// Service serviços que estao sendo executados
type Service struct {
	Base
	ServiceKey uuid.UUID `gorm:"size:45"`
	Type       string    `gorm:"size:45"`
}
