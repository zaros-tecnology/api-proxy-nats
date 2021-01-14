package rids

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/ahmetb/go-linq"
	"github.com/go-chi/chi"
	"github.com/gofrs/uuid"
)

// BaseRid rids base
type BaseRid interface {
	Routes(service string, key uuid.UUID) *Pattern
	Patterns() []*Pattern
	Name() string
}

// Base rid
type Base struct {
	name  string
	label string
}

// NewRid new rid
func NewRid(name, label string) Base {
	return Base{name, label}
}

// EndpointRest endpoint rest params
type EndpointRest struct {
	method        string
	Endpoint      string
	Params        []string
	Authenticated bool
}

// Pattern estrutura de endpoints
type Pattern struct {
	Label            string
	Service          string
	ServiceLabel     string
	Endpoint         string
	Method           string
	Authenticated    bool
	EndpointNoParams string
	Params           map[string]string
}

type method struct {
	label            string
	service          string
	serviceLabel     string
	endpoint         string
	method           string
	auth             bool
	params           map[string]string
	endpointNoParams string
}

func (p *Pattern) Auth() bool {
	return p.Authenticated
}

func (p *Pattern) EndpointNoMethod() string {
	if p.Endpoint == "" {
		return p.Service
	}
	return fmt.Sprintf("%v.%v", p.Service, p.Endpoint)
}

// EndpointName retorna endpoint do pattern
func (p *Pattern) EndpointName() string {
	if p.Endpoint == "" {
		return fmt.Sprintf("%v.%v", p.Service, p.Method)
	}
	return fmt.Sprintf("%v.%v.%v", p.Service, p.EndpointNoParams, p.Method)
}

func (p *Base) NewMethod(label, endpoint string, params ...string) *method {
	endpointNoParams := endpoint
	sp := strings.Split(endpoint, ".")
	var paramsEndpoint = make(map[string]string)
	for _, param := range params {
		for _, rep := range sp {
			if strings.Contains(rep, "$") {
				paramsEndpoint[rep[1:]] = param
				endpoint = strings.Replace(endpoint, rep, param, 1)
				sp = strings.Split(endpoint, ".")
			}
		}
	}
	return &method{label, p.name, p.label, endpoint, "", true, paramsEndpoint, endpointNoParams}
}

var patterns []*Pattern

// Patterns retorna endpoints registrados
func (p *Base) Patterns() []*Pattern {

	var pt []*Pattern
	linq.From(patterns).WhereT(func(pt *Pattern) bool {
		return pt.Service == p.name
	}).ToSlice(&pt)

	return pt
}

func (p *method) register(method string) *Pattern {
	pa := &Pattern{p.label, p.service, p.serviceLabel, p.endpoint, method, p.auth, p.endpointNoParams, p.params}
	for _, p := range patterns {
		if p.Endpoint == pa.Endpoint && p.Method == pa.Method && p.Service == pa.Service {
			return pa
		}
	}
	patterns = append(patterns, pa)
	return pa
}

func (p *method) NoAuth() *method {
	p.auth = false
	return p
}

func (p *method) Get() *Pattern {
	return p.register("GET")
}

func (p *method) Post() *Pattern {
	return p.register("POST")
}

func (p *method) Put() *Pattern {
	return p.register("PUT")
}

func (p *method) Copy() *Pattern {
	return p.register("COPY")
}

func (p *method) Delete() *Pattern {
	return p.register("DELETE")
}

func (p *method) Internal() *Pattern {
	return &Pattern{p.label, p.service, p.serviceLabel, p.endpoint, "INTERNAL", false, p.endpointNoParams, p.params}
}

func (p *Pattern) register(r *chi.Mux, hc func(endpoint EndpointRest, w http.ResponseWriter, r *http.Request)) {
	if r == nil {
		panic("chi.Mux is null")
	}

	var endpoint string
	if p.Endpoint == "" {
		endpoint = string(p.Service)
	} else {
		endpoint = string(p.Service) + "." + p.Endpoint
	}

	parts := strings.Split(endpoint, ".")

	var params []string
	for i, p := range parts {
		if strings.Contains(p, "$") {
			p = strings.ReplaceAll(p, "$", "")
			parts[i] = fmt.Sprintf("{%v}", p)
			params = append(params, p)
		}
	}

	method := fmt.Sprintf("/api/%v", strings.Join(parts, "/"))
	handler := func(w http.ResponseWriter, r *http.Request) {
		e := EndpointRest{
			Endpoint:      p.EndpointName(),
			Params:        params,
			method:        p.Method,
			Authenticated: p.Auth(),
		}
		hc(e, w, r)
	}

	if p.Method == "GET" {
		fmt.Println("GET ->", method)
		r.Get(method, handler)
	}
	if p.Method == "POST" {
		fmt.Println("POST ->", method)
		r.Post(method, handler)
	}
	if p.Method == "PUT" {
		fmt.Println("PUT ->", method)
		r.Put(method, handler)
	}
	if p.Method == "DELETE" {
		fmt.Println("DELETE ->", method)
		r.Delete(method, handler)
	}
}

// Routes controe todas as rotas
func Routes(patterns []*Pattern, r *chi.Mux, hc func(endpoint EndpointRest, w http.ResponseWriter, r *http.Request)) {
	for _, p := range patterns {
		p.register(r, hc)
	}
}

// Routes rids
func (p *Base) Routes(service string, key uuid.UUID) *Pattern {
	endpoint := fmt.Sprintf("route.%v", key.String())
	m := &method{"", service, "", endpoint, "", false, nil, endpoint}
	return m.Get()
}

// Name retorna nome do serviço
func (p *Base) Name() string {
	return p.name
}
