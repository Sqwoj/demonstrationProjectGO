package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

func initDB() error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_PORT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to open database connection: %v", err)
	}
	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to connect to the database: %v", err)
	}
	//если нужно снести распечатай..
	// _, err = db.Exec(`DROP TABLE IF EXISTS events;`)
	// if err != nil {
	// 	return fmt.Errorf("failed to drop events table: %v", err)
	// }
	// _, err = db.Exec(`DROP TABLE IF EXISTS devices;`)
	// if err != nil {
	// 	return fmt.Errorf("failed to drop devices table: %v", err)
	// }
	// _, err = db.Exec(`DROP TABLE IF EXISTS users;`)
	// if err != nil {
	// 	return fmt.Errorf("failed to drop users table: %v", err)
	// }

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS devices (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			ip VARCHAR(255) NOT NULL,
			status VARCHAR(50) DEFAULT 'inactive',
			model VARCHAR(255),
			location VARCHAR(255),
			serial_number VARCHAR(255),
			last_seen TIMESTAMPTZ,
			uptime VARCHAR(100),
			description TEXT
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create devices table: %v", err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS events (
			id SERIAL PRIMARY KEY,
			device_id INTEGER REFERENCES devices(id),
			type VARCHAR(100) DEFAULT 'info',
			severity VARCHAR(50) DEFAULT 'info',
			description TEXT,
			created_at TIMESTAMPTZ DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create events table: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password VARCHAR(255) NOT NULL
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create users table: %v", err)
	}

	return nil
}
