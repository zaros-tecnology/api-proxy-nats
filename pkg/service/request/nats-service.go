package request

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/zaros-tecnology/api-proxy-nats/internal/rids"

	nats "github.com/nats-io/nats.go"
)

var timeout int

func init() {
	t := os.Getenv("TIMEOUT")
	tParse, err := strconv.Atoi(t)
	if err == nil {
		timeout = tParse
	} else {
		timeout = 30
	}
}

type NatsConn struct {
	*nats.Conn
}

func (s *NatsConn) Timeout() time.Duration {
	return time.Duration(timeout) * time.Second
}

// Subscribe endpoint nats
func (s *NatsConn) Subscribe(p *rids.Pattern, hc func(msg *CallRequest)) (*nats.Subscription, error) {
	return s.Conn.Subscribe(p.EndpointName(), func(m *nats.Msg) {
		var msg CallRequest
		msg.Unmarshal(m.Data, m.Reply, s.Conn)
		hc(&msg)
	})
}

// Publish endpoint nats
func (s *NatsConn) Publish(p *rids.Pattern, payload CallRequest) error {
	return s.Conn.Publish(p.EndpointName(), payload.ToJSON())
}

// Request endpoint nats
func (s *NatsConn) Request(p *rids.Pattern, payload CallRequest, rs interface{}) *ErrorRequest {
	if payload.Params == nil {
		payload.Params = make(map[string]string)
	}
	for key := range p.Params {
		payload.Params[key] = p.Params[key]
	}
	result, err := s.Conn.Request(p.EndpointName(), payload.ToJSON(), time.Duration(timeout)*time.Second)
	if err != nil {
		return &ErrorRequest{
			Code:    http.StatusInternalServerError,
			Message: err.Error(),
			Error:   err,
		}
	}

	var eError ErrorRequest
	eError.Parse(result.Data)
	if eError.Code != 200 && eError.Message != "" {
		return &eError
	}

	if rs != nil {
		err = json.Unmarshal(result.Data, rs)
		if err != nil {
			return &ErrorRequest{
				Code:    http.StatusInternalServerError,
				Message: err.Error(),
				Error:   err,
			}
		}
	}
	return nil
}
