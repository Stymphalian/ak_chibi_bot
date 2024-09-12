package chatbot

import "io"

type ChatBotter interface {
	io.Closer
	ReadLoop() error
}
