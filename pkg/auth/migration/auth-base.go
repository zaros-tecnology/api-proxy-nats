package migration

import (
	"github.com/zaros-tecnology/api-proxy-nats/pkg/models/migration"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/rids"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service/request"

	"gorm.io/gorm"
)

type mBase struct {
	migration.Base
}

// NewAuth migration
func NewAuth() migration.Migration {
	return &mBase{}
}

func (a *mBase) Versions() []migration.Versions {
	return []migration.Versions{
		&v1{},
	}
}

// Migrate balancer
func (a *mBase) Magration(db *gorm.DB, nc *request.NatsConn) {
	a.Migrate(db, nc, rids.Auth().Name(), a.Versions())
}
