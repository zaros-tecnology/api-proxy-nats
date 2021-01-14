package route

import (
	"context"
	"fmt"
	"time"

	"github.com/zaros-tecnology/api-proxy-nats/internal/rids"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/route/migration"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service/request"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

type routeService struct {
	service.Base

	auths  []*service.AuthRid
	Cancel context.CancelFunc
}

// NewService instance route service
func NewService(db *gorm.DB, key uuid.UUID, auths ...*service.AuthRid) *routeService {
	return &routeService{
		service.NewBaseService(db, key, rids.Route()), auths, nil}
}

func (s *routeService) AddAuthRid(auth *service.AuthRid) {
	s.auths = append(s.auths, auth)
}

// Start service route
func (s *routeService) Start() {
	s.Init(migration.NewRoute())

	s.Nats().Subscribe(rids.Route().Restart(), func(msg *request.CallRequest) {
		s.Lock()
		defer s.Unlock()

		s.Cancel()

		fmt.Println("Restart Http server in 1 secound")
		<-time.After(time.Second)

		started := s.serve()
		<-started
		close(started)
	})

	started := s.serve()
	<-started
	close(started)
}
