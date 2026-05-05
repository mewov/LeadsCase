package storage

import (
	"database/sql"
	"demo-server/internal/libs"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/lib/pq"
)

type (
	Postgres struct {
		db   *sql.DB
		logg *slog.Logger
	}
)

func CreateUrlPostgres(config *libs.Config) string {
	return fmt.Sprintf("postgresql://%s:%s@%s/%s?sslmode=disable",
		config.PostgresUser,
		config.PostgresPassword,
		config.PostgresAddr,
		config.PostgresDb,
	)
}

func CreateAndConnectPostgres(url string, logger *slog.Logger) (*Postgres, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}

	for range 10 {
		err = db.Ping()
		if err == nil {
			break
		}

		fmt.Println("ping postgres...")
		time.Sleep(time.Second)
	}
	if err != nil {
		return nil, err
	}

	logger.Info("connect to postgres", "type_log", "system")
	return &Postgres{
		db:   db,
		logg: logger,
	}, nil
}

func (p *Postgres) Migration() error {
	_, err := p.db.Exec(`
	CREATE EXTENSION IF NOT EXISTS pgcrypto;

	CREATE TABLE IF NOT EXISTS leads (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		client_name TEXT NOT NULL,
		client_title TEXT NOT NULL,
		client_description TEXT NOT NULL,
		client_contact TEXT NOT NULL,
		status VARCHAR(32) NOT NULL DEFAULT 'new',
		created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
		updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	
	CREATE TABLE IF NOT EXISTS admins (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		login VARCHAR(255) NOT NULL UNIQUE,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	
	CREATE TABLE IF NOT EXISTS admin_sessions (
		id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
		admin_id UUID NOT NULL REFERENCES admins(id) ON DELETE CASCADE,
		token TEXT NOT NULL UNIQUE,
		expires_at TIMESTAMPTZ NOT NULL,
		created_at TIMESTAMPTZ NOT NULL DEFAULT now()
	);
	
	CREATE INDEX IF NOT EXISTS idx_leads_status ON leads(status);
	CREATE INDEX IF NOT EXISTS idx_leads_created_at ON leads(created_at);
	CREATE INDEX IF NOT EXISTS idx_admin_sessions_admin_id ON admin_sessions(admin_id);
	CREATE INDEX IF NOT EXISTS idx_admin_sessions_expires_at ON admin_sessions(expires_at);
	`)
	if err != nil {
		p.logg.Error("migration", "type_log", "system", "error", err)
		return err
	}

	p.logg.Info("migration", "type_log", "system")
	return nil
}

func (p *Postgres) Close() error {
	return p.db.Close()
}
