package main

import (
	"github.com/Stymphalian/ak_chibi_bot/server/internal/server"
)

func main() {
	m, err := server.InitializeMainServer()
	if err != nil {
		panic(err)
	}
	m.Run()
}
