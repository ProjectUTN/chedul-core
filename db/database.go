package db

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/uptrace/bun"
)

var Bun *bun.DB

func CreateDatabase() (*sql.DB, error) {
	godotenv.Load()

	var (
		db_name     = os.Getenv("DB_NAME")
		db_user     = os.Getenv("DB_USER")
		db_password = os.Getenv("DB_PASSWORD")
		db_host     = os.Getenv("DB_HOST")
		uri         = fmt.Sprintf("user=%s dbname=%s password=%s host=%s port=5432", db_user, db_name, db_password, db_host)
	)

	conn, err := sql.Open("postgres", uri)
	if err != nil {
		return nil, err
	}

	return conn, nil
}
