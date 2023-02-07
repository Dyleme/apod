package postgres

import (
	"database/sql"
	"fmt"
	"os"
)

type Config struct {
	UserName string
	Password string
	Host     string
	Port     string
	DBName   string
	SSLMode  string
}

func InitConfig() *Config {
	return &Config{
		UserName: os.Getenv("DB_USERNAME"),
		Password: os.Getenv("DB_PASSWORD"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSL_MODE"),
	}
}

// Constuctor to the postgres database.
func NewDB(conf *Config) (*sql.DB, error) {
	var db *sql.DB

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		conf.Host, conf.Port, conf.UserName, conf.Password, conf.DBName, conf.SSLMode)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("open connection %q: %w", connStr, err)
	}

	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("ping: %w", err)
	}

	return db, nil
}
