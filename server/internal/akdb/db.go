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

func GetRoomFromChannelName(channelName string) (*Room, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var roomId int64
	var isActive bool
	var createdAt time.Time
	err = db.QueryRow(
		"SELECT room_id, is_active, created_at FROM rooms WHERE channel_name = $1",
		channelName,
	).Scan(&roomId, &isActive, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &Room{
		RoomId:      roomId,
		ChannelName: channelName,
		IsActive:    isActive,
		CreatedAt:   createdAt,
		UpdatedAt:   time.Now(),
	}, nil
}

func InsertRoom(channelName string) (*Room, error) {
	db, err := Connect()
	if err != nil {
		return nil, err
	}
	defer db.Close()

	var roomId int64
	var createdAt time.Time
	var updatedAt time.Time
	err = db.QueryRow(
		"INSERT INTO rooms (channel_name, is_active) VALUES ($1, $2) RETURNING room_id, created_at, updated_at;",
		channelName,
		true,
	).Scan(&roomId, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	return &Room{
		RoomId:      roomId,
		ChannelName: channelName,
		IsActive:    true,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}, nil
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
