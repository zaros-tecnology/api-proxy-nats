package request

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/zaros-tecnology/api-proxy-nats/pkg/rids"

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
func (s *NatsConn) Subscribe(p *rids.Pattern,
	hc func(msg *CallRequest),
	access ...func(msg *AccessRequest)) (*nats.Subscription, error) {
	return s.Conn.Subscribe(p.EndpointName(), func(m *nats.Msg) {
		var msg CallRequest
		msg.Unmarshal(m.Data, m.Reply, s.Conn)
		if len(access) > 0 {
			acr := AccessRequest{
				CallRequest: &msg,
				err:         true,
			}

			access[0](&acr)
			if acr.err {
				msg.error(ErrorStatusUnauthorized)
				return
			}

			if acr.get != nil && *acr.get && p.Method != "GET" {
				msg.error(ErrorStatusUnauthorized)
				return
			}
			if acr.get != nil && !*acr.get {
				match := false
				for _, m := range acr.methods {
					match = match || strings.Contains(p.EndpointName(), m)
				}
				if !match {
					msg.error(ErrorStatusUnauthorized)
					return
				}
			}
		}
		hc(&msg)
	})
}

// Publish endpoint nats
func (s *NatsConn) Publish(p *rids.Pattern, payload CallRequest, token ...interface{}) error {
	return s.Conn.Publish(p.EndpointName(), payload.ToJSON())
}

// Get endpoint nats
func (s *NatsConn) Get(p *rids.Pattern, rs interface{}, token ...json.RawMessage) *ErrorRequest {
	return s.Request(p, nil, rs, token...)
}

// Request endpoint nats
func (s *NatsConn) Request(p *rids.Pattern, payload *CallRequest, rs interface{}, token ...json.RawMessage) *ErrorRequest {
	if payload == nil {
		payload = &CallRequest{}
	}
	if payload.Params == nil {
		payload.Params = make(map[string]string)
	}
	for key := range p.Params {
		payload.Params[key] = p.Params[key]
	}
	if len(token) > 0 {
		payload.Header.Set("Token", string(token[0]))
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
