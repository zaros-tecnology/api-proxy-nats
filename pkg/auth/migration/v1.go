package migration

import (
	"github.com/zaros-tecnology/api-proxy-nats/internal/models/migration"

	"gorm.io/gorm"
)

type v1 struct {
	migration.BaseVersion
}

func (b *v1) Migrate(db *gorm.DB) error {
	err := db.Exec(`
	CREATE TABLE oauth2_clients (
		id text NOT NULL,
		secret text NOT NULL,
		"domain" text NOT NULL,
		"data" jsonb NOT NULL,
		CONSTRAINT oauth2_clients_pkey PRIMARY KEY (id)
	);`).Error
	if err != nil {
		panic(err)
	}
	err = db.Exec(`
	CREATE TABLE oauth2_tokens (
		id bigserial NOT NULL,
		created_at timestamptz NOT NULL,
		expires_at timestamptz NOT NULL,
		code text NOT NULL,
		"access" text NOT NULL,
		"refresh" text NOT NULL,
		"data" jsonb NOT NULL,
		CONSTRAINT oauth2_tokens_pkey PRIMARY KEY (id)
	);
	CREATE INDEX idx_oauth2_tokens_access ON public.oauth2_tokens USING btree (access);
	CREATE INDEX idx_oauth2_tokens_code ON public.oauth2_tokens USING btree (code);
	CREATE INDEX idx_oauth2_tokens_expires_at ON public.oauth2_tokens USING btree (expires_at);
	CREATE INDEX idx_oauth2_tokens_refresh ON public.oauth2_tokens USING btree (refresh);`).Error
	if err != nil {
		panic(err)
	}
	return nil
}
