package db

import (
	"database/sql"
	"fmt"
	"go-tutuplapak-file/config"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB(cfg *config.Config) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.DbHost, cfg.DbPort, cfg.DbUser, cfg.DbPass, cfg.DbName, cfg.DbSSL)

	var err error
	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return
	}

	if err := DB.Ping(); err != nil {
		log.Fatalf("Database is unreachable: %v", err)
		return
	}

	log.Println("Database connected successfully!")
}
