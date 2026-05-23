package internal

import (
	"database/sql"
	"log"

	_ "github.com/tursodatabase/libsql-client-go/libsql"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("libsql", dbPath)
	if err != nil {
		return err
	}
	if err := migrate(DB); err != nil {
		return err
	}
	log.Print("Initialized Turso Database")
	return nil
}

func migrate(db *sql.DB) error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS story_pointing_sessions (
			id TEXT PRIMARY KEY,
			created_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS oknohelp_sessions (
			id TEXT PRIMARY KEY,
			created_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS story_pointing_participants (
			session_id TEXT NOT NULL,
			name TEXT NOT NULL,
			joined_at INTEGER NOT NULL,
			PRIMARY KEY (session_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS oknohelp_participants (
			session_id TEXT NOT NULL,
			name TEXT NOT NULL,
			joined_at INTEGER NOT NULL,
			PRIMARY KEY (session_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS story_pointing_summaries (
			session_id TEXT PRIMARY KEY,
			summary TEXT NOT NULL DEFAULT '',
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS oknohelp_summaries (
			session_id TEXT PRIMARY KEY,
			summary TEXT NOT NULL DEFAULT '',
			updated_at INTEGER NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS story_pointing_votes (
			session_id TEXT NOT NULL,
			name TEXT NOT NULL,
			value TEXT NOT NULL,
			voted_at INTEGER NOT NULL,
			PRIMARY KEY (session_id, name)
		)`,
		`CREATE TABLE IF NOT EXISTS oknohelp_votes (
			session_id TEXT NOT NULL,
			name TEXT NOT NULL,
			value TEXT NOT NULL,
			voted_at INTEGER NOT NULL,
			PRIMARY KEY (session_id, name)
		)`,
	}
	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	addCols := []string{
		`ALTER TABLE story_pointing_summaries ADD COLUMN active INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE oknohelp_summaries ADD COLUMN active INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE story_pointing_summaries ADD COLUMN countdown_until INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE oknohelp_summaries ADD COLUMN countdown_until INTEGER NOT NULL DEFAULT 0`,
		`ALTER TABLE story_pointing_participants ADD COLUMN role TEXT NOT NULL DEFAULT 'play'`,
		`ALTER TABLE oknohelp_participants ADD COLUMN role TEXT NOT NULL DEFAULT 'play'`,
	}
	for _, s := range addCols {
		_, _ = db.Exec(s)
	}
	return nil
}
