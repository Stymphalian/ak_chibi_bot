package chatbot

type ChatBotter interface {
	Close() error
	ReadLoop() error
}
