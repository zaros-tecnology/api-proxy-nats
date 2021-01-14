package socket

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/zaros-tecnology/api-proxy-nats/pkg/models"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/rids"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service/request"

	"github.com/gorilla/websocket"
)

// NewConnectionWS socket
func NewConnectionWS(ws *Conn, srvBase *service.Base) func(w http.ResponseWriter, r *http.Request) {
	srv = srvBase
	ws.msg = make(chan models.MessageDataSocket)
	ws.conn = make(map[string]*websocket.Conn)
	read := func(c *websocket.Conn) {
		for {
			mt, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", message)
			err = c.WriteMessage(mt, message)
			if err != nil {
				log.Println("write:", err)
				break
			}
		}
	}
	go func() {
		var m sync.Mutex
		for msg := range ws.msg {
			for key, c := range ws.conn {
				p := strings.Split(key, ";")[0]
				person, _ := strconv.Atoi(p)
				if person == msg.PersonID {
					data, _ := json.Marshal(msg)
					err := c.WriteMessage(websocket.TextMessage, data)
					if err != nil {
						m.Lock()
						c.Close()
						delete(ws.conn, key)
						m.Unlock()
					}
				}
			}
		}
	}()

	srv.Nats().Subscribe(rids.Route().SocketNewMessage(), func(r *request.CallRequest) {
		var msg models.MessageDataSocket
		r.ParseData(&msg)

		if msg.Data != nil && msg.PersonID > 0 && msg.TypeMessage != "" {
			if connWS != nil {
				connWS.msg <- msg
			}
		}
	})

	return func(w http.ResponseWriter, r *http.Request) {
		upgrader.CheckOrigin = func(r *http.Request) bool {
			return true
		}
		c, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("upgrade:", err)
			return
		}
		go read(c)
		m.Lock()
		defer m.Unlock()
		key := r.Header.Get("Sec-Websocket-key")
		if ws.conn[r.FormValue("person")+";"+key] != nil {
			ws.conn[r.FormValue("person")+";"+key].Close()
		}
		ws.conn[r.FormValue("person")+";"+key] = c
	}
}
