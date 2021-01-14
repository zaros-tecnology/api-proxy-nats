package service

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/zaros-tecnology/api-proxy-nats/internal/models"
	"github.com/zaros-tecnology/api-proxy-nats/internal/models/migration"
	"github.com/zaros-tecnology/api-proxy-nats/internal/rids"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service/request"

	"github.com/gofrs/uuid"
	nats "github.com/nats-io/nats.go"
	"gorm.io/gorm"
)

// Service base
type Service interface {
	Start()
	Stop()
	Rid() rids.BaseRid
}

// Base service
type Base struct {
	key uuid.UUID
	db  *gorm.DB
	nc  *request.NatsConn
	m   sync.Mutex

	rid rids.BaseRid
}

// NewBaseService nova instancia base
func NewBaseService(db *gorm.DB, key uuid.UUID, rid rids.BaseRid) Base {
	return Base{
		db:  db, //.New(),
		key: key,
		rid: rid,
	}
}

// DeferRecover metodo para defer de verificação de erro
func (s *Base) DeferRecover() {
	if p := recover(); p != nil {
		fmt.Fprintln(os.Stderr, p, string(debug.Stack()))
	}
}

// Lock na instancia geral do serviço
func (s *Base) Lock() {
	s.m.Lock()
}

// Unlock desbloquei uso do serviço
func (s *Base) Unlock() {
	s.m.Unlock()
}

// Nats retorna instancia do nats
func (s *Base) Nats() *request.NatsConn {
	return s.nc
}

// DB retorna instancid do banco
func (s *Base) DB() *gorm.DB {
	return s.db
}

// ServiceKey recupera chave da instancia
func (s *Base) ServiceKey() uuid.UUID {
	return s.key
}

// Init base service
func (s *Base) Init(migration migration.Migration) {
	if s.rid == nil {
		panic("rid not registered")
	}

	nc, _ := nats.Connect(os.Getenv(nats.DefaultURL))
	s.nc = &request.NatsConn{Conn: nc}

	if migration != nil {
		migration.Magration(s.DB(), s.nc)
	}

	if s.Rid().Name() == rids.Route().Name() {
		s.Nats().Subscribe(rids.Route().NewService(), func(msg *request.CallRequest) {
			var srv models.Service
			msg.ParseData(&srv)
			err := s.DB().Create(&srv).Error
			if err != nil {
				msg.Error(err)
				return
			}
			msg.OK(srv)
		})
	}

	_, err := s.Nats().Subscribe(s.rid.Routes(s.rid.Name(), s.key), func(r *request.CallRequest) {
		patterns := s.rid.Patterns()
		r.OK(patterns)
	})

	if err != nil {
		panic(err)
	}

	var srv models.Service
	rError := s.Nats().Request(rids.Route().NewService(), request.NewRequest(models.Service{Type: s.rid.Name(), ServiceKey: s.key}), &srv)
	if rError != nil || srv.ID == uuid.Nil {
		panic(err)
	}

	if s.Rid().Name() != rids.Route().Name() {
		go func() {
			<-time.After(time.Second)
			s.Nats().Publish(rids.Route().Restart(), request.EmptyRequest())
		}()
	}
}

// Stop service
func (s *Base) Stop() {
	s.nc.Drain()
	s.nc.Close()
}

// ParseMessage data
func (s *Base) ParseMessage(data []byte, p interface{}) {
	json.Unmarshal(data, p)
}

// Rid retorna nome do serviço
func (s *Base) Rid() rids.BaseRid {
	return s.rid
}
