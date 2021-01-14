package migration

import (
	"github.com/zaros-tecnology/api-proxy-nats/internal/models/migration"
	"github.com/zaros-tecnology/api-proxy-nats/internal/rids"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service/request"

	"gorm.io/gorm"
)

type srv1Migration struct {
	migration.Base
}

// NewRoute migration
func NewRoute() migration.Migration {
	return &srv1Migration{}
}

func (a *srv1Migration) Versions() []migration.Versions {
	return []migration.Versions{}
}

// Migrate balancer
func (a *srv1Migration) Magration(db *gorm.DB, nc *request.NatsConn) {
	a.Migrate(db, nc, rids.Route().Name(), a.Versions())
}
