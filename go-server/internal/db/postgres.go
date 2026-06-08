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

// RecordMatchResult atomically updates both players at the end of a match
func (db *Database) RecordMatchResult(ctx context.Context, winnerID string, loserID string) error {
	// 1. Begin the atomic transaction
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	
	// 2. Defer a rollback. If the function panics or returns an error before tx.Commit(), 
	// this safely cancels all database changes.
	defer tx.Rollback(ctx)

	// 3. Update the Winner
	winnerQuery := `
		UPDATE players 
		SET matches_played = matches_played + 1, 
		    total_wins = total_wins + 1, 
		    total_kills = total_kills + 1
		WHERE id = $1;
	`
	if _, err := tx.Exec(ctx, winnerQuery, winnerID); err != nil {
		return err // Triggers rollback
	}

	// 4. Update the Loser
	loserQuery := `
		UPDATE players 
		SET matches_played = matches_played + 1
		WHERE id = $1;
	`
	if _, err := tx.Exec(ctx, loserQuery, loserID); err != nil {
		return err // Triggers rollback
	}

	// 5. If both succeeded, lock it in!
	return tx.Commit(ctx)
}