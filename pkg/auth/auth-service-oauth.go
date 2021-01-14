package auth

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx"
	pg "github.com/vgarvardt/go-oauth2-pg"
	"github.com/vgarvardt/go-pg-adapter/pgxadapter"
	"gopkg.in/oauth2.v3/errors"
	"gopkg.in/oauth2.v3/manage"
	"gopkg.in/oauth2.v3/models"
	"gopkg.in/oauth2.v3/server"
)

// Create Client store
func (a *srv) Create(info *models.Client) (err error) {
	err = a.clientStore.Create(info)
	return
}

func connect() *srv {
	connection := os.Getenv("CONNECTION_STRING")
	if os.Getenv("UNIT_TEST") == "true" {
		connection = os.Getenv("TEST_DB_CONN")
	}
	fmt.Println("CONNECTION_STRING > ", connection)
	pgxConnConfig, err := pgx.ParseConnectionString(connection)
	if err != nil {
		fmt.Println(err)
	}

	fmt.Println("PGX Config ->", pgxConnConfig)

	channel := make(chan func() (
		*server.Server,
		*manage.Manager,
		*pg.ClientStore,
		*pgx.Conn))

	go func() {
		fmt.Println(">> Conectando PGX Oauth2")
		_, err := pgx.NewConnPool(pgx.ConnPoolConfig{
			ConnConfig:     pgxConnConfig,
			AcquireTimeout: time.Minute * 5,
			AfterConnect: func(conn *pgx.Conn) error {

				manager := manage.NewDefaultManager()
				manager.SetAuthorizeCodeTokenCfg(manage.DefaultAuthorizeCodeTokenCfg)

				adapter := pgxadapter.NewConn(conn)
				tokenStore, _ := pg.NewTokenStore(adapter, pg.WithTokenStoreGCInterval(time.Minute))
				defer tokenStore.Close()

				clientStore, _ := pg.NewClientStore(adapter)

				manager.MapTokenStorage(tokenStore)
				manager.MapClientStorage(clientStore)

				srv := server.NewDefaultServer(manager)
				srv.SetAllowGetAccessRequest(true)
				srv.SetClientInfoHandler(server.ClientFormHandler)
				manager.SetRefreshTokenCfg(manage.DefaultRefreshTokenCfg)

				srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
					log.Println("Internal Error:", err.Error())
					return
				})

				srv.SetResponseErrorHandler(func(re *errors.Response) {
					log.Println("Response Error:", re.Error.Error())
				})

				channel <- func() (*server.Server, *manage.Manager, *pg.ClientStore, *pgx.Conn) {
					return srv, manager, clientStore, conn
				}

				return nil
			},
		})
		if err != nil {
			panic(err)
		}
	}()

	server, manager, clientStore, pgxConn := (<-channel)()
	fmt.Println("<< Conectado PGX Oauth2")

	return &srv{
		clientStore: clientStore,
		manager:     manager,
		server:      server,
		conn:        pgxConn,
	}
}
