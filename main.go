package apiproxynats

import (
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gofrs/uuid"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/zaros-tecnology/api-proxy-nats/pkg/auth"
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
func NewProxyServer(services []HandlerService, options ProxyOptions) error {

	if err := options.IsValid(); err != nil {
		return err
	}

	servicesType := make(map[string]string)
	for _, srv := range options.Services {
		servicesType[srv] = srv
	}

	if options.Developer {
		s = runDefaultServer()
	}

	sigs := make(chan os.Signal, 1)
	done := make(chan bool, 1)

	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	db, err := connectDatabase(options.ConnectionString)
	if err != nil {
		panic(err)
	}

	key, _ := uuid.NewV4()
	v5 := uuid.NewV5(key, "github.com/zaros-tecnology/api-proxy-nats")

	authService := auth.NewService(db, v5)
	routerService := route.NewService(db, v5,
		&service.AuthRid{Auth: authService, Rid: rids.Auth()},
	)

	if servicesType[routerService.Rid().Name()] == routerService.Rid().Name() || options.Developer {
		routerService.Start()
		authService.Start()
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
		done <- true
	}()

	<-done

	return fmt.Errorf("server closed")
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
