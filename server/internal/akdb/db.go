package akdb

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/lib/pq" // add this
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func ConnectWithSql() (*sql.DB, error) {
	dbpassWordFile, ok := os.LookupEnv("DATABASE_PASSWORD_FILE")
	if !ok {
		return nil, fmt.Errorf("DATABASE_PASSWORD_FILE not set")
	}
	bin, err := os.ReadFile(dbpassWordFile)
	if err != nil {
		return nil, err
	}
	hostname, ok := os.LookupEnv("DATABASE_HOST")
	if !ok {
		return nil, fmt.Errorf("DATABASE_HOST not set")
	}
	port, ok := os.LookupEnv("DATABASE_PORT")
	if !ok {
		return nil, fmt.Errorf("DATABASE_PORT not set")
	}
	username, ok := os.LookupEnv("DATABASE_USER")
	if !ok {
		return nil, fmt.Errorf("DATABASE_USER not set")
	}
	dbname, ok := os.LookupEnv("DATABASE_DB")
	if !ok {
		return nil, fmt.Errorf("DATABASE_DB not set")
	}

	return sql.Open(
		"postgres",
		fmt.Sprint(
			fmt.Sprintf(" host=%s", hostname),
			fmt.Sprintf(" port=%s", port),
			fmt.Sprintf(" user=%s", username),
			fmt.Sprintf(" dbname=%s", dbname),
			fmt.Sprintf(" password=%s", string(bin)),
			" connect_timeout=10",
		),
	)
}

func Connect() (*gorm.DB, error) {
	dbpassWordFile, ok := os.LookupEnv("DATABASE_PASSWORD_FILE")
	if !ok {
		return nil, fmt.Errorf("DATABASE_PASSWORD_FILE not set")
	}
	bin, err := os.ReadFile(dbpassWordFile)
	if err != nil {
		return nil, err
	}
	hostname, ok := os.LookupEnv("DATABASE_HOST")
	if !ok {
		return nil, fmt.Errorf("DATABASE_HOST not set")
	}
	port, ok := os.LookupEnv("DATABASE_PORT")
	if !ok {
		return nil, fmt.Errorf("DATABASE_PORT not set")
	}
	username, ok := os.LookupEnv("DATABASE_USER")
	if !ok {
		return nil, fmt.Errorf("DATABASE_USER not set")
	}
	dbname, ok := os.LookupEnv("DATABASE_DB")
	if !ok {
		return nil, fmt.Errorf("DATABASE_DB not set")
	}

	connString := fmt.Sprint(
		fmt.Sprintf(" host=%s", hostname),
		fmt.Sprintf(" port=%s", port),
		fmt.Sprintf(" user=%s", username),
		fmt.Sprintf(" dbname=%s", dbname),
		fmt.Sprintf(" password=%s", string(bin)),
		" connect_timeout=10",
	)
	db, err := gorm.Open(postgres.Open(connString), &gorm.Config{})
	return db, err
}
