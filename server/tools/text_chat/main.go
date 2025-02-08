package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chatbot"
	"github.com/gorilla/websocket"
)

// This must be kept in sync with chatbot.chatbot.CliMessage struct
// type chatbot.CliMessage struct {
// 	Username string `json:"username"`
// 	Message  string `json:"message"`
// }

type Chat struct {
	channel string
	conn    *websocket.Conn
	read    chan string
	write   chan chatbot.CliMessage
	done    chan bool
}

func (m *Chat) Run() error {
	conn := m.conn

	go func() {
		defer log.Println("Ping ticker done for", m.channel)
		ticker := time.NewTicker(10 * time.Second)
		for {
			select {
			case <-ticker.C:
				conn.WriteMessage(websocket.PingMessage, []byte("PING"))
			case <-m.done:
				ticker.Stop()
				return
			}
		}
	}()

	go func() {
		defer log.Println("Read pump done for", m.channel)
		for {
			// Read
			_, message, err := conn.ReadMessage()
			if err != nil {
				fmt.Println("read:", err)
				m.done <- true
				return
			}
			fmt.Println(string(message))
			m.read <- string(message)
		}

	}()

	go func() {
		defer log.Println("Write pump done for", m.channel)
		for {
			select {
			case message := <-m.write:
				jsonMsg, err := json.Marshal(message)
				if err != nil {
					fmt.Println("write:", err)
					return
				}
				// fmt.Println(string(jsonMsg))
				if err := conn.WriteMessage(websocket.TextMessage, jsonMsg); err != nil {
					fmt.Println("write:", err)
					return
				}
			case <-m.done:
				return
			}
		}
	}()

	<-m.done
	return nil
}

func (c *Chat) Close() error {
	log.Println("Closed Chat for", c.channel)
	c.conn.Close()
	close(c.done)
	return nil
}

type Main struct {
	currentChannel         string
	currentUser            string
	currentUserDisplayName string
	chats                  map[string]*Chat
}

func (m *Main) HandleConnection(w http.ResponseWriter, r *http.Request) {
	channelQuery := r.URL.Query().Get("channel")
	if channelQuery == "" {
		fmt.Println("Query must be provided in ws/ connection")
		return
	}

	var upgrader = websocket.Upgrader{} // use default options
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()

	chat := &Chat{
		channel: channelQuery,
		conn:    conn,
		read:    make(chan string, 10),
		write:   make(chan chatbot.CliMessage, 10),
		done:    make(chan bool),
	}
	m.chats[channelQuery] = chat
	defer func() {
		chat.Close()
		delete(m.chats, channelQuery)
	}()

	if err := chat.Run(); err != nil {
		fmt.Println(err)
	}
}

func (m *Main) HandleTextConnection(w http.ResponseWriter, r *http.Request) {
	var upgrader = websocket.Upgrader{} // use default options
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer conn.Close()
	done := make(chan bool)

	change := make(chan bool)
	go func() {
		for {
			select {
			case text := <-m.chats[m.currentChannel].read:
				conn.WriteMessage(websocket.TextMessage, []byte(text))
			case <-change:
				continue
			case <-done:
				return
			}
		}
	}()
	conn.WriteMessage(websocket.TextMessage, []byte(
		fmt.Sprintf("%s/%s", m.currentChannel, m.currentUser),
	))

	for {
		// Read
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			break
		}
		text := strings.TrimSpace(string(message))

		if strings.HasPrefix(text, "set_user_display") {
			m.currentUserDisplayName = strings.TrimPrefix(text, "set_user_display ")
			conn.WriteMessage(websocket.TextMessage, []byte(
				fmt.Sprintf("%s/%s", m.currentChannel, m.currentUser),
			))
			continue
		} else if strings.HasPrefix(text, "set_user") {
			m.currentUser = strings.TrimPrefix(text, "set_user ")
			m.currentUserDisplayName = ""
			conn.WriteMessage(websocket.TextMessage, []byte(
				fmt.Sprintf("%s/%s", m.currentChannel, m.currentUser),
			))
			continue
		} else if strings.HasPrefix(text, "set_channel") {
			m.currentChannel = strings.TrimPrefix(text, "set_channel ")
			conn.WriteMessage(websocket.TextMessage, []byte(
				fmt.Sprintf("%s/%s", m.currentChannel, m.currentUser),
			))
			change <- true
			continue
		}

		chat, ok := m.chats[m.currentChannel]
		if !ok {
			fmt.Println("Channel not connected.")
			continue
		}
		chat.write <- chatbot.CliMessage{
			Username:        m.currentUser,
			UserDisplayName: m.currentUserDisplayName,
			Message:         strings.TrimSpace(text),
		}
	}

	close(done)
}

func (m *Main) Run(address string) error {
	// Connect to the bot-server
	// Start an http server which will accept an websocket connection
	http.Handle("/ws", http.HandlerFunc(m.HandleConnection))
	http.Handle("/ws/text", http.HandlerFunc(m.HandleTextConnection))
	if err := http.ListenAndServe(address, nil); err != nil {
		log.Fatal(err)
	}
	return nil
}

func main() {
	channelFlag := flag.String("channel", "stymphalian__", "The spine runtime channel to connect to")
	userFlag := flag.String("user", "stymphalian__", "The User in which to chat as.")
	addressFlag := flag.String("address", ":8090", "Port to connect to")
	flag.Parse()
	fmt.Println("-channel: ", *channelFlag)
	fmt.Println("-user: ", *userFlag)
	fmt.Println("-address: ", *addressFlag)

	mainCli := Main{
		currentChannel: *channelFlag,
		currentUser:    *userFlag,
		chats:          make(map[string]*Chat),
	}
	mainCli.Run(*addressFlag)
}
