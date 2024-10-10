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

const (
	DATABASE_PASSFILE_FILE_ENV = "DATABASE_PASSWORD_FILE"
	DATABASE_PASSWORD_ENV      = "DATABASE_PASSWORD"
	DATABASE_HOST_ENV          = "DATABASE_HOST"
	DATABASE_PORT_ENV          = "DATABASE_PORT"
	DATABASE_USER_ENV          = "DATABASE_USER"
	DATABASE_DB_ENV            = "DATABASE_DB"
)

type databaseConnInfo struct {
	password string
	hostname string
	port     string
	username string
	dbname   string
}

func getConnectionParamsFromEnv() (*databaseConnInfo, error) {
	dbpassWordFile, ok := os.LookupEnv(DATABASE_PASSFILE_FILE_ENV)
	var password string
	if !ok {
		log.Printf("%s not set", DATABASE_PASSFILE_FILE_ENV)
		log.Printf("Trying to get password from %s", DATABASE_PASSWORD_ENV)
		// Try to get password from env
		password, ok = os.LookupEnv(DATABASE_PASSWORD_ENV)
		if !ok {
			return nil, fmt.Errorf("%s not set", DATABASE_PASSWORD_ENV)
		}
	} else {
		bin, err := os.ReadFile(dbpassWordFile)
		if err != nil {
			return nil, err
		}
		password = string(bin)
	}

	hostname, ok := os.LookupEnv(DATABASE_HOST_ENV)
	if !ok {
		return nil, fmt.Errorf("%s not set", DATABASE_HOST_ENV)
	}
	port, ok := os.LookupEnv(DATABASE_PORT_ENV)
	if !ok {
		return nil, fmt.Errorf("%s not set", DATABASE_PORT_ENV)
	}
	username, ok := os.LookupEnv(DATABASE_USER_ENV)
	if !ok {
		return nil, fmt.Errorf("%s not set", DATABASE_USER_ENV)
	}
	dbname, ok := os.LookupEnv(DATABASE_DB_ENV)
	if !ok {
		return nil, fmt.Errorf("%s not set", DATABASE_DB_ENV)
	}
	return &databaseConnInfo{
		password: password,
		hostname: hostname,
		port:     port,
		username: username,
		dbname:   dbname,
	}, nil
}

func GetConnectionString(connInfo *databaseConnInfo) (string, error) {
	return fmt.Sprint(
		fmt.Sprintf(" host=%s", connInfo.hostname),
		fmt.Sprintf(" port=%s", connInfo.port),
		fmt.Sprintf(" user=%s", connInfo.username),
		fmt.Sprintf(" dbname=%s", connInfo.dbname),
		fmt.Sprintf(" password=%s", connInfo.password),
		" connect_timeout=10",
	), nil
}

func ConnectWithSql(connInfo *databaseConnInfo) (*sql.DB, error) {
	connStr, err := GetConnectionString(connInfo)
	if err != nil {
		return nil, err
	}
	return sql.Open("postgres", connStr)
}

func Connect(connInfo *databaseConnInfo) (*gorm.DB, error) {
	connStr, err := GetConnectionString(connInfo)
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

type DatbaseConn struct {
	DefaultDB *gorm.DB
}

func ProvideDatabaseConn() (*DatbaseConn, error) {
	connInfo, err := getConnectionParamsFromEnv()
	if err != nil {
		return nil, err
	}
	db, err := Connect(connInfo)
	if err != nil {
		log.Fatal(err)
	}
	return &DatbaseConn{
		DefaultDB: db,
	}, nil
}

// TODO: Find a way to setup a test database cleanly that works well with go test
func ProvideTestDatabaseConn() (*DatbaseConn, error) {
	return ProvideDatabaseConn()
	// connInfo := &databaseConnInfo{
	// 	password: "test_user_password",
	// 	hostname: "db",
	// 	port:     "5432",
	// 	username: "test_user",
	// 	dbname:   "test_db",
	// }
	// db, err := Connect(connInfo)
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// return &DatbaseConn{
	// 	DefaultDB: db,
	// }, nil
}
