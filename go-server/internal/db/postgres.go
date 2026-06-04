package db

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Database wraps our connection pool
type Database struct {
	Pool *pgxpool.Pool
}

// InitDB connects to Postgres and returns our Database struct
func InitDB(dsn string) *Database {
	// Context with a 10-second timeout for the initial connection
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}

	// Verify the connection is actually alive
	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("Database ping failed: %v\n", err)
	}

	log.Println("✦ Connected to PostgreSQL database")
	return &Database{Pool: pool}
}

// CreatePlayer atomically inserts a new guest into the database
func (db *Database) CreatePlayer(ctx context.Context, id string, displayName string) error {
	query := `
		INSERT INTO players (id, display_name) 
		VALUES ($1, $2)
		ON CONFLICT (id) DO NOTHING;
	`
	_, err := db.Pool.Exec(ctx, query, id, displayName)
	return err
}

// UpdateLastLogin touches the timestamp when they connect to the WebSocket
func (db *Database) UpdateLastLogin(ctx context.Context, id string) {
	query := `UPDATE players SET last_login = CURRENT_TIMESTAMP WHERE id = $1;`
	db.Pool.Exec(ctx, query, id)
}