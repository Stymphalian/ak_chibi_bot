package chat

type ChatMessage struct {
	Username        string
	UserDisplayName string
	Message         string
}

type ChatMessageHandler interface {
	HandleMessage(msg ChatMessage) (string, error)
}
