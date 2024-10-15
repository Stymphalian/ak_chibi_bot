package main

import (
	"bufio"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

// cd text_chat/terminal
// go run main.go
// set_channel stymphalian__
// set_user my_user_name
// !chibi lin
func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:8090", Path: "/ws/text"}
	log.Printf("connecting to %s", u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	go func() {
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			fmt.Println(string(message))
			// log.Printf("recv: %s", message)
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		err = c.WriteMessage(websocket.TextMessage, []byte(text))
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
}
