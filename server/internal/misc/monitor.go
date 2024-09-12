package misc

var Monitor *MonitorStruct

func init() {
	Monitor = &MonitorStruct{}
}

type MonitorStruct struct {
	NumRoomsCreated         int
	NumWebsocketConnections int
	NumUsers                int
	NumCommands             int
}

// var GoRunCounter atomic.Int64
