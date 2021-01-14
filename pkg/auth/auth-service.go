package auth

import (
	"sync"
	"time"

	"github.com/zaros-tecnology/api-proxy-nats/pkg/auth/migration"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/models"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/rids"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service"

	"github.com/gofrs/uuid"
	"github.com/jackc/pgx"
	pg "github.com/vgarvardt/go-oauth2-pg"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/server"
	"gorm.io/gorm"
)

type srv struct {
	*service.Base
	userToken    map[string]*models.User
	tokenExpires map[string]*time.Time

	server      *server.Server
	manager     *manage.Manager
	clientStore *pg.ClientStore
	conn        *pgx.Conn
	m           sync.Mutex
}

// NewService instance route service
func NewService(db *gorm.DB, key uuid.UUID) *srv {
	base := service.NewBaseService(db, key, rids.Auth())
	auth := &srv{}
	auth.Base = &base

	auth.userToken = make(map[string]*models.User)
	auth.tokenExpires = make(map[string]*time.Time)

	return auth
}

// Start service route
func (s *srv) Start() {
	s.Init(migration.NewAuth())

	s.Nats().Subscribe(rids.Auth().Login(), s.login)
	s.Nats().Subscribe(rids.Auth().HavePermission(), s.userHavePermission)

	auth := connect()
	s.clientStore = auth.clientStore
	s.manager = auth.manager
	s.server = auth.server
	s.conn = auth.conn
}
