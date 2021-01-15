package apiproxynats

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofrs/uuid"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/rids"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/route"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/service"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var s *server.Server

// ProxyOptions options
type ProxyOptions struct {
	Developer        bool
	Services         []string
	ConnectionString string
	LocalNats        bool
	NatsURL          string
}

// HandlerService handler
type HandlerService func(db *gorm.DB, key uuid.UUID) service.Service

// HandlerAuthService required instance auth service
type HandlerAuthService func(db *gorm.DB, key uuid.UUID) service.Auth

// IsValid validate options
func (p *ProxyOptions) IsValid() error {
	if len(p.ConnectionString) == 0 {
		return fmt.Errorf("ConnectionString is required")
	}
	if !p.LocalNats && len(p.NatsURL) == 0 {
		return fmt.Errorf("NatsURL is required")
	}
	if len(p.Services) == 0 {
		return fmt.Errorf("Services is required")
	}
	return nil
}

// NewProxyServer new service proxy
func NewProxyServer(services []HandlerService, handerAuth HandlerAuthService, options ProxyOptions) (ctx context.Context, connected chan bool, err error) {

	err = options.IsValid()
	if err != nil {
		return
	}

	servicesType := make(map[string]string)
	for _, srv := range options.Services {
		servicesType[srv] = srv
	}

	if options.Developer {
		s = runDefaultServer()
	}

	db, err := connectDatabase(options.ConnectionString)
	if err != nil {
		return
	}

	ctx, cancel := context.WithCancel(context.Background())

	connected = make(chan bool)

	go func() {
		sigs := make(chan os.Signal, 1)

		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		key, _ := uuid.NewV4()
		v5 := uuid.NewV5(key, "github.com/zaros-tecnology/api-proxy-nats")

		authService := handerAuth(db, v5)
		routerService := route.NewService(db, v5,
			&service.AuthRid{Auth: authService, Rid: rids.Auth()},
		)

		if servicesType[routerService.Rid().Name()] == routerService.Rid().Name() || options.Developer {
			routerService.Start()
		}

		var wg sync.WaitGroup
		wg.Add(len(services))

		for _, srv := range services {
			go func(srv HandlerService) {

				defer wg.Done()
				s := srv(db, v5)

				if servicesType[s.Rid().Name()] == s.Rid().Name() || options.Developer {

					routerService.AddAuthRid(&service.AuthRid{Auth: authService, Rid: s.Rid()})
					s.Start()

					fmt.Println(">> Starting ", s.Rid().Name())
				}

			}(srv)
		}

		wg.Wait()

		go func() {
			sig := <-sigs
			fmt.Println(sig)
			cancel()
		}()

		connected <- true

		<-ctx.Done()
	}()

	return
}

func connectDatabase(connectionString string) (*gorm.DB, error) {

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  os.Getenv(connectionString),
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// defaultNatsOptions are default options for the unit tests.
var defaultNatsOptions = server.Options{
	Host: "127.0.0.1",
	Port: 4222,
}

// RunDefaultServer starts a new Go routine based server using the default options
func runDefaultServer() *server.Server {
	return runServer(&defaultNatsOptions)
}

// RunServer starts a new Go routine based server
func runServer(opts *server.Options) *server.Server {
	if opts == nil {
		opts = &defaultNatsOptions
	}
	s, err := server.NewServer(opts)
	if err != nil {
		panic(err)
	}

	s.ConfigureLogger()

	// Run server in Go routine.
	go s.Start()

	// Wait for accept loop(s) to be started
	if !s.ReadyForConnections(10 * time.Second) {
		panic("Unable to start NATS Server in Go Routine")
	}
	return s
}
