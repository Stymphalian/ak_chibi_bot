package chatbot

import (
	"encoding/json"
	"log"
	"net/url"
	"strings"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/chat"
	"github.com/gorilla/websocket"
)

// Needs more testing
type CliChatBot struct {
	*websocketClient
	chatMessageHandler chat.ChatMessageHandler
	channelName        string
}

type websocketClient struct {
	conn      *websocket.Conn
	closeDone chan struct{}
	write     chan string
	read      chan string
}

func NewCliChatBot(
	chatMessageHandler chat.ChatMessageHandler,
	twitchChannelName string) (*CliChatBot, error) {

	rawQuery := url.Values{}
	rawQuery.Set("channel", twitchChannelName)
	u := url.URL{
		Scheme:   "ws",
		Host:     "text_chat:8090",
		Path:     "/ws",
		RawQuery: rawQuery.Encode(),
	}
	log.Printf("connecting to %s", u.String())
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return nil, err
	}
	log.Println("connected to " + u.String())

	self := &CliChatBot{
		chatMessageHandler: chatMessageHandler,
		channelName:        twitchChannelName,
		websocketClient: &websocketClient{
			closeDone: make(chan struct{}),
			write:     make(chan string, 10),
			read:      make(chan string, 10),
			conn:      conn,
		},
	}
	return self, nil
}

func (t *CliChatBot) Close() error {
	log.Println("CliChatBot::Close() called")

	err := t.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		log.Println("Failed to send close message", err)
	} else {
		select {
		case <-t.closeDone:
			log.Println("CliChatBot: Received close message")
		case <-time.After(100 * time.Millisecond):
			log.Println("CliChatBot: Timed out waiting for close message")
		}
	}
	t.conn.Close()

	log.Println("CliChatBot::Close() finished")
	return nil
}

func (t *CliChatBot) HandlePrivateMessage(m chat.ChatMessage) {
	var trimmed = strings.TrimSpace(m.Message)
	if len(trimmed) == 0 {
		return
	}

	outputMsg, err := t.chatMessageHandler.HandleMessage(m)
	if err == nil && len(outputMsg) > 0 {
		t.conn.WriteMessage(websocket.TextMessage, []byte(outputMsg))
	}
}

type CliMessage struct {
	Username string `json:"username"`
	Message  string `json:"message"`
}

func (t *CliChatBot) ReadLoop() error {
	defer log.Println("CLIChatBot Read pump done for ", t.channelName)

	t.conn.SetPingHandler(func(a string) error {
		t.conn.WriteMessage(websocket.PongMessage, []byte(a))
		return nil
	})

	for {
		msgType, message, err := t.conn.ReadMessage()
		if err != nil {
			if err == websocket.ErrCloseSent {
				log.Println("CLIChatBot websocket closed")
				close(t.closeDone)
				return nil
			}
			log.Println("CLIChatBot read:", err)
			return err
		}
		log.Println("CLIChatBot read:", string(message))
		switch msgType {
		case websocket.PingMessage:
			fallthrough
		case websocket.PongMessage:
			fallthrough
		case websocket.BinaryMessage:
			continue
		case websocket.CloseMessage:
			log.Println("CLIChatBot websocket closed")
			close(t.closeDone)
			return nil
		case websocket.TextMessage:
			var msg CliMessage
			err = json.Unmarshal(message, &msg)
			if err != nil {
				log.Println("read:", err)
				return err
			}
			chatMessage := chat.ChatMessage{
				Username:        msg.Username,
				UserDisplayName: msg.Username,
				TwitchUserId:    "twitch-" + msg.Username,
				Message:         msg.Message,
			}
			t.HandlePrivateMessage(chatMessage)
		}
	}
}
