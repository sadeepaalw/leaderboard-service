package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq"
)

func Open() *sql.DB {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	var db *sql.DB
	var err error
	for i := 0; i < 10; i++ {
		db, err = sql.Open("postgres", dsn)
		if err == nil && db.Ping() == nil {
			log.Println("Connected to Postgres database!")
			return db
		}
		log.Printf("Waiting for database to be ready... (%d/10)", i+1)
		time.Sleep(2 * time.Second)
	}
	log.Fatalf("failed to connect to db after retries: %v", err)
	return nil
}

func Close(db *sql.DB) {
	if err := db.Close(); err != nil {
		log.Printf("failed to close db: %v", err)
	}
}
