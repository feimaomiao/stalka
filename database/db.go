package database

import (
	"database/sql"
	"fmt"
	"io"
	"os"

	// loads .env file automatically.
	_ "github.com/joho/godotenv/autoload"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func Init(log *zap.SugaredLogger) error {
	connStr := fmt.Sprintf("host=localhost port=5432 user=%s password=%s dbname=esports sslmode=disable",
		os.Getenv("postgres_user"),
		os.Getenv("postgres_password"))
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error(err)
		return err
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		log.Error(err)
		return err
	}
	log.Info("Connected to database")
	log.Info("Running initial setup")
	// open the file static/init.sql
	file, err := os.Open("static/init.sql")
	if err != nil {
		log.Error(err)
		return err
	}
	defer file.Close()
	// read the file
	data, err := io.ReadAll(file)
	if err != nil {
		log.Error(err)
		return err
	}
	sqlCommand := string(data)

	_, err = db.Exec(fmt.Sprintf(sqlCommand, os.Getenv("reader_password"), os.Getenv("writer_password")))
	log.Info("Ran initial setup")
	if err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func Connect(user string, password string) (*sql.DB, error) {
	// Connect to the database
	connStr := fmt.Sprintf(
		"host=localhost port=5432 user=%s password=%s dbname=esports sslmode=disable",
		user,
		password,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}
