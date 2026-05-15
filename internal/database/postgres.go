package database

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	Pool *pgxpool.Pool
}

func New(dsn string) (*DB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("connect to database: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("ping database: %w", err)
	}

	return &DB{Pool: pool}, nil
}

func (db *DB) Close() {
	db.Pool.Close()
}

func (db *DB) RunMigrations(ctx context.Context) error {
	migrations := []string{
		migrationCreateUsers,
		migrationCreateServers,
		migrationCreateUserServers,
		migrationCreateSubscriptions,
		migrationCreateTrafficLogs,
		migrationCreateDomains,
	}

	for _, m := range migrations {
		if _, err := db.Pool.Exec(ctx, m); err != nil {
			return fmt.Errorf("run migration: %w", err)
		}
	}
	return nil
}

const migrationCreateUsers = `
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    uuid VARCHAR(36) UNIQUE NOT NULL,
    username VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    email VARCHAR(255) DEFAULT '',
    status VARCHAR(20) DEFAULT 'active',
    traffic_used BIGINT DEFAULT 0,
    traffic_limit BIGINT DEFAULT 107374182400,
    subscription_url TEXT DEFAULT '',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const migrationCreateServers = `
CREATE TABLE IF NOT EXISTS servers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    port INTEGER NOT NULL DEFAULT 443,
    api_port INTEGER NOT NULL DEFAULT 9443,
    api_key VARCHAR(255) DEFAULT '',
    location VARCHAR(100) DEFAULT '',
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const migrationCreateUserServers = `
CREATE TABLE IF NOT EXISTS user_servers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    server_id BIGINT REFERENCES servers(id) ON DELETE CASCADE,
    port INTEGER NOT NULL DEFAULT 0,
    key VARCHAR(255) NOT NULL DEFAULT '',
    traffic_used BIGINT DEFAULT 0,
    traffic_limit BIGINT DEFAULT 107374182400,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, server_id)
);`

const migrationCreateSubscriptions = `
CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    user_server_id BIGINT REFERENCES user_servers(id) ON DELETE CASCADE,
    link TEXT NOT NULL DEFAULT '',
    uuid VARCHAR(36) NOT NULL DEFAULT '',
    traffic_used BIGINT DEFAULT 0,
    traffic_limit BIGINT DEFAULT 107374182400,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const migrationCreateTrafficLogs = `
CREATE TABLE IF NOT EXISTS traffic_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    server_id BIGINT REFERENCES servers(id) ON DELETE CASCADE,
    upload BIGINT DEFAULT 0,
    download BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`

const migrationCreateDomains = `
CREATE TABLE IF NOT EXISTS domains (
    id BIGSERIAL PRIMARY KEY,
    domain VARCHAR(255) UNIQUE NOT NULL,
    server_id BIGINT REFERENCES servers(id) ON DELETE SET NULL,
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);`
