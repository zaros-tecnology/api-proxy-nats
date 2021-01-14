package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/zaros-tecnology/api-proxy-nats/pkg/models"

	nats "github.com/nats-io/nats.go"
)

type AccessRequest struct {
	*CallRequest
	err bool
}

// CallRequest handler
type CallRequest struct {
	Params map[string]string
	Data   []byte
	reply  string
	nc     *nats.Conn

	Form     url.Values
	Header   http.Header
	Endpoint string
}

// ErrorRequest error request
type ErrorRequest struct {
	Message string
	Code    int
	Error   error
}

// ParseToken logado
func (c *CallRequest) ParseToken(t interface{}) {
	token := c.Header.Get("Token")
	json.Unmarshal([]byte(token), &t)
}

func (c *CallRequest) RawToken() json.RawMessage {
	return []byte(c.Header.Get("Token"))
}

// User logado
func (c *CallRequest) User() models.User {
	var user models.User
	userStr := c.Header.Get("User")
	json.Unmarshal([]byte(userStr), &user)
	return user
}

// Parse error request
func (e *ErrorRequest) Parse(payload []byte) error {
	return json.Unmarshal(payload, e)
}

// ToJSON error request
func (e ErrorRequest) ToJSON() []byte {
	data, _ := json.Marshal(e)
	return data
}

// EmptyRequest empty
func EmptyRequest() *CallRequest {
	return &CallRequest{}
}

// NewRequest nova instancia da call request
func NewRequest(data interface{}) *CallRequest {
	switch data.(type) {
	case []byte:
		panic("invalid data")
	}
	payload, _ := json.Marshal(data)
	return &CallRequest{
		Params: nil,
		Data:   payload,
	}
}

// CloneRequest clone request
func (c CallRequest) CloneRequest(data interface{}) *CallRequest {
	switch data.(type) {
	case []byte:
		panic("invalid data")
	}
	c.Data, _ = json.Marshal(data)
	return &c
}

// PathParam retorna parametro map string
func (c *CallRequest) PathParam(key string) string {
	return c.Params[key]
}

// ParseData data
func (c *CallRequest) ParseData(v interface{}) error {
	return json.Unmarshal(c.Data, v)
}

// ToJSON CallRequest
func (c *CallRequest) ToJSON() []byte {
	data, _ := json.Marshal(c)
	return data
}

// Unmarshal data
func (c *CallRequest) Unmarshal(data []byte, reply string, nc *nats.Conn) error {
	if len(data) > 0 {
		err := json.Unmarshal(data, c)
		if err != nil {
			return err
		}
	}
	c.nc = nc
	c.reply = reply
	return nil
}

// OK result
func (c *CallRequest) OK(result ...interface{}) {
	if len(result) == 0 {
		c.nc.Publish(c.reply, []byte(""))
		return
	}

	switch result[0].(type) {
	case string:
		str := result[0].(string)
		c.nc.Publish(c.reply, []byte(str))
		return
	case []byte:
		str := result[0].([]byte)
		c.nc.Publish(c.reply, str)
		return
	}

	payload, _ := json.Marshal(result[0])
	c.nc.Publish(c.reply, payload)
}

// Error result
func (c *CallRequest) Error(err error) {
	c.error(ErrorRequest{err.Error(), http.StatusInternalServerError, err})
}

// ErrorRequest result
func (c *CallRequest) ErrorRequest(err *ErrorRequest) {
	if err == nil {
		panic("error request cant be nil")
	}
	c.error(*err)
}

// Error result
func (c *CallRequest) error(err ErrorRequest) {
	c.nc.Publish(c.reply, err.ToJSON())
}

// information already exists
var (
	ErrorInformationAlreadyExists = ErrorRequest{"information already exists", http.StatusUnauthorized, fmt.Errorf("information already exists")}
	ErrorNotFound                 = ErrorRequest{"not found", http.StatusNotFound, fmt.Errorf("not found")}
	ErrorStatusUnauthorized       = ErrorRequest{"not authorized", http.StatusUnauthorized, fmt.Errorf("not authorized")}
	ErrorInvalidParams            = ErrorRequest{"invalid params", http.StatusUnauthorized, fmt.Errorf("invalid params")}
	ErrorInternalServerError      = ErrorRequest{"internal error", http.StatusUnauthorized, fmt.Errorf("internal error")}
	ErrorAccessDenied             = ErrorRequest{"access denied", http.StatusUnauthorized, fmt.Errorf("access denied")}
)

// Error result
func (c *CallRequest) NotFound() {
	c.error(ErrorNotFound)
}

func (c *AccessRequest) AccessDenied() {
	c.err = true
}

func (c *AccessRequest) AccessGranted() {
	c.err = false
}

func InternalError(err error) *ErrorRequest {
	return &ErrorRequest{
		Message: err.Error(),
		Code:    http.StatusInternalServerError,
		Error:   err,
	}
}

// Search query de search
func (c *CallRequest) Search(columns ...string) string {
	search := c.Form.Get("search")
	if search == "" {
		return "deleted_at IS NULL"
	}

	for i := range columns {
		columns[i] = columns[i] + " LIKE ?"
	}

	words := strings.Split(search, " ")

	var andConditions []string
	for i := 0; i < len(words); i++ {
		andConditions = append(andConditions, "("+strings.Join(columns, " OR ")+")")
	}

	var wordsParams []string
	for _, w := range words {
		for range columns {
			wordsParams = append(wordsParams, "'%"+w+"%'")
		}
	}

	where := "deleted_at IS NULL AND (" + strings.Join(andConditions, " AND ") + ")"
	for strings.Contains(where, "?") {
		where = strings.Replace(where, "?", wordsParams[0], 1)
		wordsParams = wordsParams[1:]
	}

	fmt.Println(where)

	return where
}
