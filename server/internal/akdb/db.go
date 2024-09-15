package akdb

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/lib/pq" // add this
)

func Connect() (*sql.DB, error) {

	bin, err := os.ReadFile("/run/secrets/db-password.txt")
	if err != nil {
		return nil, err
	}

	log.Println("@@@@ connecting")
	return sql.Open(
		"postgres",
		fmt.Sprint(
			//" host=psql.jordanyu.com",
			" host=db",
			" port=5432",
			" user=postgres",
			" dbname=akdb",
			fmt.Sprintf(" password=%s", string(bin)),
			" connect_timeout=10",
			//" sslmode=verify-full",
			//" sslrootcert=/work/secrets/CertAuth2.crt",
			//" sslkey=/work/secrets/postgres.key",
			//" sslcert=/work/secrets/postgres.crt",
		),
	)

	// username := "postgres"
	// dbHost := "db"
	// dbPort := "5432"
	// dbName := "akdb"
	// return sql.Open("postgres",
	// 	fmt.Sprintf(
	// 		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
	// 		// "postgres://%s:%s@%s:%s/%s",
	// 		username,
	// 		string(bin),
	// 		dbHost,
	// 		dbPort,
	// 		dbName,
	// 	),
	// )
}

func Prepare() error {
	db, err := Connect()
	if err != nil {
		return err
	}
	defer db.Close()

	for i := 0; i < 60; i++ {
		if err := db.Ping(); err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if _, err := db.Exec("DROP TABLE IF EXISTS blog"); err != nil {
		return err
	}

	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS blog (id SERIAL, title VARCHAR)"); err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		if _, err := db.Exec("INSERT INTO blog (title) VALUES ($1);", fmt.Sprintf("Blog post #%d", i)); err != nil {
			return err
		}
	}
	return nil
}
