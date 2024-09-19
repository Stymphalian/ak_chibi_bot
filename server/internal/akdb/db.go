package akdb

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	_ "github.com/lib/pq" // add this
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func GetConnectionString() (string, error) {
	dbpassWordFile, ok := os.LookupEnv("DATABASE_PASSWORD_FILE")
	if !ok {
		return "", fmt.Errorf("DATABASE_PASSWORD_FILE not set")
	}
	bin, err := os.ReadFile(dbpassWordFile)
	if err != nil {
		return "", err
	}
	hostname, ok := os.LookupEnv("DATABASE_HOST")
	if !ok {
		return "", fmt.Errorf("DATABASE_HOST not set")
	}
	port, ok := os.LookupEnv("DATABASE_PORT")
	if !ok {
		return "", fmt.Errorf("DATABASE_PORT not set")
	}
	username, ok := os.LookupEnv("DATABASE_USER")
	if !ok {
		return "", fmt.Errorf("DATABASE_USER not set")
	}
	dbname, ok := os.LookupEnv("DATABASE_DB")
	if !ok {
		return "", fmt.Errorf("DATABASE_DB not set")
	}

	return fmt.Sprint(
		fmt.Sprintf(" host=%s", hostname),
		fmt.Sprintf(" port=%s", port),
		fmt.Sprintf(" user=%s", username),
		fmt.Sprintf(" dbname=%s", dbname),
		fmt.Sprintf(" password=%s", string(bin)),
		" connect_timeout=10",
	), nil
}

func ConnectWithSql() (*sql.DB, error) {
	connStr, err := GetConnectionString()
	if err != nil {
		return nil, err
	}
	return sql.Open("postgres", connStr)
}

func Connect() (*gorm.DB, error) {
	connStr, err := GetConnectionString()
	if err != nil {
		return nil, err
	}
	gormDb, err := gorm.Open(
		postgres.Open(connStr),
		&gorm.Config{
			NowFunc: func() time.Time {
				return misc.Clock.Now()
			},
		},
	)
	if err != nil {
		return nil, err
	}
	sqldb, err := gormDb.DB()
	if err != nil {
		return nil, err
	}
	sqldb.SetMaxOpenConns(100)
	sqldb.SetMaxIdleConns(10)
	sqldb.SetConnMaxIdleTime(1 * time.Hour)
	sqldb.SetConnMaxLifetime(1 * time.Hour)
	return gormDb, err
}

var DefaultDB *gorm.DB

func init() {
	db, err := Connect()
	if err != nil {
		log.Fatal(err)
	}
	DefaultDB = db
}
