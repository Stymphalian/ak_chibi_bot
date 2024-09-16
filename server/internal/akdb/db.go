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

	// if _, err := db.Exec("DROP TABLE IF EXISTS blog"); err != nil {
	// 	return err
	// }

	// if _, err := db.Exec("CREATE TABLE IF NOT EXISTS blog (id SERIAL, title VARCHAR)"); err != nil {
	// 	return err
	// }

	for i := 0; i < 10; i++ {
		if _, err := db.Exec("INSERT INTO blog (title) VALUES ($1);", fmt.Sprintf("Blog post #%d", i)); err != nil {
			log.Println(err)
			return err
		}
	}
	return nil
}
