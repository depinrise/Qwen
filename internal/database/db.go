package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
)

type DB struct {
	conn *sql.DB
}

func NewConnection(dsn string) (*DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Println("✅ Connected to PolarDB MySQL!")

	dbInstance := &DB{conn: db}

	// Create tables if they don't exist
	if err := dbInstance.createTables(); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return dbInstance, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) createTables() error {
	conversationsTable := `
	CREATE TABLE IF NOT EXISTS conversations (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL,
		user_name VARCHAR(255),
		message TEXT NOT NULL,
		response TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_user_id (user_id),
		INDEX idx_created_at (created_at)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	sessionsTable := `
	CREATE TABLE IF NOT EXISTS chat_sessions (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id VARCHAR(255) NOT NULL UNIQUE,
		session_data JSON,
		last_activity TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		INDEX idx_user_id (user_id),
		INDEX idx_last_activity (last_activity)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	memoriesTable := `
	CREATE TABLE IF NOT EXISTS user_memories (
		id BIGINT AUTO_INCREMENT PRIMARY KEY,
		user_id BIGINT NOT NULL,
		memory_key VARCHAR(100) NOT NULL,
		memory_value TEXT NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
		UNIQUE KEY unique_user_memory (user_id, memory_key),
		INDEX idx_user_id (user_id),
		INDEX idx_memory_key (memory_key)
	) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
	`

	if _, err := db.conn.Exec(conversationsTable); err != nil {
		return fmt.Errorf("failed to create conversations table: %w", err)
	}

	if _, err := db.conn.Exec(sessionsTable); err != nil {
		return fmt.Errorf("failed to create chat_sessions table: %w", err)
	}

	if _, err := db.conn.Exec(memoriesTable); err != nil {
		return fmt.Errorf("failed to create user_memories table: %w", err)
	}

	log.Println("✅ Database tables created/verified successfully")
	return nil
}

func (db *DB) GetConnection() *sql.DB {
	return db.conn
}
