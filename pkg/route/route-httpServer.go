package route

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/pprof"
	"strings"
	"time"

	"github.com/zaros-tecnology/api-proxy-nats/internal/models"
	"github.com/zaros-tecnology/api-proxy-nats/internal/rids"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/route/handlers"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/route/socket"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/rs/cors"
)

func (s *routeService) serve() chan bool {
	started := make(chan bool)
	go func() {
		r := chi.NewRouter()

		// A good base middleware stack
		r.Use(middleware.RequestID)
		r.Use(middleware.RealIP)
		r.Use(middleware.Logger)
		r.Use(middleware.Recoverer)
		r.Use(handlers.ContentTypeJSONMiddleware)

		cors := cors.New(cors.Options{
			// AllowedOrigins: []string{"https://foo.com"}, // Use this to allow specific origin hosts
			AllowedOrigins: []string{"*"},
			// AllowOriginFunc:  func(r *http.Request, origin string) bool { return true },
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
			ExposedHeaders:   []string{"Link"},
			AllowCredentials: true,
			MaxAge:           300, // Maximum value not ignored by any of major browsers
		})
		r.Use(cors.Handler)

		r.Use(middleware.Timeout(60 * time.Second))
		r.Use(handlers.AuthMiddleware(s.auths...))

		var services []models.Service
		s.DB().Find(&services)

		for _, rid := range services {

			result, err := s.Nats().Conn.Request(rids.Route().Routes(rid.Type, rid.ServiceKey).EndpointName(), nil, time.Millisecond*500)
			if err != nil {
				s.DB().Delete(&rid)
				continue
			}

			var patterns []*rids.Pattern
			err = json.Unmarshal(result.Data, &patterns)
			if err != nil {
				panic("not posible Unmarshal patterns")
			}

			if len(patterns) > 0 {
				rids.Routes(patterns, r, s.handler)

				for _, pattern := range patterns {

					perm := models.Permission{
						Rid:          pattern.EndpointNoMethod(),
						Method:       pattern.Method,
						Nome:         pattern.Label,
						Service:      pattern.Service,
						ServiceLabel: pattern.ServiceLabel,
					}
					if strings.Contains(perm.Rid, "route") {
						continue
					}

					var dbPermissao models.Permission
					err = s.DB().Where(&perm).First(&dbPermissao).Error

					if err != nil {
						s.DB().Save(&perm)

						var admin models.Profile

						s.DB().Where(&models.Profile{Name: "Administrador"}).First(&admin)
						s.DB().Model(&admin).Association("Permissions").Append(&perm)
					}
				}
			}

			for _, auth := range s.auths {
				if auth.Rid.Name() == rid.Type {
					auth.Patterns = patterns
					break
				}
			}
		}

		r.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(http.StatusOK)
		})

		r.HandleFunc("/health", func(rw http.ResponseWriter, r *http.Request) {
			rw.WriteHeader(http.StatusOK)
		})

		r.HandleFunc("/ws", socket.NewConnectionWS(socket.NewInstanceWS(), &s.Base))

		// Register pprof handlers
		r.HandleFunc("/debug/pprof/", pprof.Index)
		r.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		r.HandleFunc("/debug/pprof/profile", pprof.Profile)
		r.HandleFunc("/debug/pprof/symbol", pprof.Symbol)

		r.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		r.Handle("/debug/pprof/heap", pprof.Handler("heap"))
		r.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
		r.Handle("/debug/pprof/block", pprof.Handler("block"))
		r.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
		r.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))

		srv := http.Server{Addr: ":3333", Handler: r}
		ctx, cancel := context.WithCancel(context.Background())
		s.Cancel = cancel

		go func() {
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatal(err)
			}
		}()
		started <- true
		select {
		case <-ctx.Done():
			srv.Shutdown(ctx)
		}
	}()
	return started
}
