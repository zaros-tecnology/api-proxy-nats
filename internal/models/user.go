package models

import (
	"encoding/json"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
)

// User struct
type User struct {
	Base
	Email         string    `json:"email" validate:"required"`
	Password      string    `json:"password" validate:"required"`
	CheckPassword string    `json:"checkPassword" gorm:"-"`
	Nome          string    `json:"nome" validate:"required"`
	Token         uuid.UUID `json:"-"`
	ExpiresIN     int64     `json:"-"`

	Profiles []*Profile `gorm:"many2many:user_profile;"`
}

// Token login
type Token struct {
	AccessToken string `json:"access_token"`
	ExpiresIN   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
	Name        string `json:"name"`
}

// Oauth2Token struct
type Oauth2Tokens struct {
	Base
	Access  string
	Data    string
	Expired bool

	DataToken Oauth2TokenData `gorm:"-"`
}

// QueryByToken query
func (Oauth2Tokens) QueryByToken() string {
	return "select access, data, expires_at < now() expired from oauth2_tokens where access = ? limit 1"
}

// AfterFind depois de buscar
func (u *Oauth2Tokens) AfterFind(tx *gorm.DB) (err error) {
	return json.Unmarshal([]byte(u.Data), &u.DataToken)
}

// Oauth2TokenData struct
type Oauth2TokenData struct {
	Code             string    `json:"Code"`
	Scope            string    `json:"Scope"`
	Access           string    `json:"Access"`
	UserID           string    `json:"UserID"`
	Refresh          string    `json:"Refresh"`
	ClientID         string    `json:"ClientID"`
	RedirectURI      string    `json:"RedirectURI"`
	CodeCreateAt     time.Time `json:"CodeCreateAt"`
	CodeExpiresIn    int       `json:"CodeExpiresIn"`
	AccessCreateAt   time.Time `json:"AccessCreateAt"`
	AccessExpiresIn  int64     `json:"AccessExpiresIn"`
	RefreshCreateAt  time.Time `json:"RefreshCreateAt"`
	RefreshExpiresIn int       `json:"RefreshExpiresIn"`
}

// Oauth2Clients struct
type Oauth2Clients struct {
	ID     string
	Secret string
	Data   string

	Oauth2ClientsData Oauth2ClientsData `gorm:"-"`
}

// QueryByID query
func (Oauth2Clients) QueryByID() string {
	return "select id, secret, data from oauth2_clients where id = ? limit 1"
}

// AfterFind depois de buscar
func (u *Oauth2Clients) AfterFind(tx *gorm.DB) (err error) {
	return json.Unmarshal([]byte(u.Data), &u.Oauth2ClientsData)
}

// Oauth2ClientsData struct
type Oauth2ClientsData struct {
	ID     string `json:"ID"`
	Domain string `json:"Domain"`
	Secret string `json:"Secret"`
	UserID string `json:"UserID"`
}

// HavePermissionRequest request
type HavePermissionRequest struct {
	Service  string
	Endpoint string
	Method   string
}

// QueryPermission get permissions
func (HavePermissionRequest) QueryPermission() string {
	return `select p.* from permissions p 
	join profile_permission pp on p.id = pp.permissao_id 
	join user_profile up on pp.perfil_id = up.perfil_id
	join usuario_admins ua on up.usuario_admin_id = ua.id 
	where ua.id = ? and p.service = ? and p.rid = ? and p."method" = ?`
}
