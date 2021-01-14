package socket

import (
	"sync"

	"github.com/zaros-tecnology/api-proxy-nats/pkg/models"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service"

	"github.com/gorilla/websocket"
)

var srv *service.Base
var upgrader = websocket.Upgrader{}
var m sync.Mutex
var connWS *Conn
var tokenNumbers map[string]map[string]string

func init() {
	tokenNumbers = make(map[string]map[string]string)
}

// Conn struct
type Conn struct {
	conn map[string]*websocket.Conn
	msg  chan models.MessageDataSocket
}

// Connections retorna quantidade reconexao
func (ws *Conn) Connections() int {
	return len(ws.conn)
}

// NewInstanceWS ws intance
func NewInstanceWS() *Conn {
	connWS = new(Conn)
	return connWS
}
