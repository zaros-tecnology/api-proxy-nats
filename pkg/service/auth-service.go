package service

import (
	"net/http"

	"github.com/zaros-tecnology/api-proxy-nats/pkg/rids"
)

// Auth interface
type Auth interface {
	HandleTokenRequest(w http.ResponseWriter, r *http.Request) error
	RemoveAccessToken(token string) error
	ValidationBearerToken(r *http.Request, w http.ResponseWriter) bool
}

// AuthRid auth service
type AuthRid struct {
	Rid      rids.BaseRid
	Auth     Auth
	Patterns []*rids.Pattern
}
