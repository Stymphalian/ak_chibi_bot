package spine

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketDebufInfo struct {
	AverageFps *misc.RollingArray[float64]
}
type WebSocketConn struct {
	connectionName   string
	conn             *websocket.Conn
	done             chan struct{}
	remove           bool
	DebugInfo        *WebSocketDebufInfo
	SendChatMsgsFlag bool
}

type SpineBridge struct {
	spineService          *operator.OperatorService
	WebSocketConnections  map[string]*WebSocketConn
	websocketPingerTicker *time.Ticker
	websocketPingerDone   chan bool

	clientResponseCallbackListenersId int
	clientResponseCallbackListeners   map[int]ClientRequestCallback
	// TODO: Might want to add mutex locking for updating websocket connections
}

func NewSpineBridge(spineService *operator.OperatorService) (*SpineBridge, error) {
	s := &SpineBridge{
		spineService:         spineService,
		WebSocketConnections: make(map[string]*WebSocketConn, 0),

		clientResponseCallbackListenersId: 0,
		clientResponseCallbackListeners:   make(map[int]ClientRequestCallback, 0),
	}
	go s.pingWebSockets()
	return s, nil
}

func (s *SpineBridge) pingWebSockets() {
	// misc.GoRunCounter.Add(1)
	// defer misc.GoRunCounter.Add(-1)

	s.websocketPingerTicker = time.NewTicker(time.Duration(30) * time.Second)
	s.websocketPingerDone = make(chan bool)

	for {
		select {
		case <-s.websocketPingerDone:
			log.Println("Closing websocket pinger")
			s.websocketPingerDone = nil
			return
		case <-s.websocketPingerTicker.C:
			for _, websocketConn := range s.WebSocketConnections {
				if websocketConn.conn == nil {
					continue
				}
				websocketConn.conn.WriteControl(
					websocket.PingMessage,
					[]byte{},
					misc.Clock.Now().Add(time.Duration(1)*time.Second),
				)
			}
		}
	}
}

func (s *SpineBridge) Close() error {
	log.Println("SpineBridge::Close() called")
	if (s.websocketPingerDone) != nil {
		close(s.websocketPingerDone)
	}
	// Close

	var wg sync.WaitGroup
	for roomName, websocketConn := range s.WebSocketConnections {
		if websocketConn.conn == nil {
			continue
		}
		err := websocketConn.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		if err != nil {
			log.Printf("write close for websocketConn %s: %v\n", roomName, err)
		}

		wg.Add(1)
		go func() {
			// misc.GoRunCounter.Add(1)
			// defer misc.GoRunCounter.Add(-1)

			select {
			case <-time.After(100 * time.Millisecond):
			case <-websocketConn.done:
			}
			wg.Done()
		}()
	}

	// Wait for the clients to reply
	wg.Wait()

	log.Println("SpineBridge::Close() finished")
	return nil
}

func (s *SpineBridge) clientConnected() bool {
	return len(s.WebSocketConnections) > 0
}

func (s *SpineBridge) handleResponseMessages(connectionId string, message []byte) {
	var data map[string]interface{}
	// log.Println("Received message", string(message))
	err := json.Unmarshal(message, &data)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		return
	}
	if _, ok := data["type_name"]; !ok {
		log.Printf("Invalid JSON: %v", data)
		return
	}
	typeName, ok := data["type_name"].(string)
	if !ok {
		log.Printf("Invalid JSON: type_name is not a string: %v", data)
		return
	}

	switch typeName {
	case RUNTIME_DEBUG_UPDATE:
		var debugUpdateReq RuntimeDebugUpdateRequest
		err := json.Unmarshal(message, &debugUpdateReq)
		if err != nil {
			log.Println(err)
			return
		}
		if _, ok := s.WebSocketConnections[connectionId]; ok {
			s.WebSocketConnections[connectionId].DebugInfo.AverageFps.Add(debugUpdateReq.AverageFps)
		}
	case RUNTIME_ROOM_SETTINGS:
		var req RuntimeRoomSettingsRequest
		err := json.Unmarshal(message, &req)
		if err != nil {
			log.Println(err)
			return
		}
		if _, ok := s.WebSocketConnections[connectionId]; ok {
			s.WebSocketConnections[connectionId].SendChatMsgsFlag = req.ShowChatMessages
		}
	default:
		log.Printf("Unhandled message type: %s", typeName)
	}

	for _, listener := range s.clientResponseCallbackListeners {
		listener(connectionId, typeName, message)
	}
}

func (s *SpineBridge) NumConnections() int {
	return len(s.WebSocketConnections)
}

func (s *SpineBridge) AddConnection(
	w http.ResponseWriter,
	r *http.Request,
	chatters []*ChatterInfo,
) error {
	// misc.GoRunCounter.Add(1)
	// defer misc.GoRunCounter.Add(-1)

	var upgrader = websocket.Upgrader{} // use default options
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade: ", err)
		return nil
	}
	misc.Monitor.NumWebsocketConnections += 1

	// Get a uuid string
	connectionName := uuid.New().String()
	websocketConn := &WebSocketConn{
		connectionName: connectionName,
		conn:           c,
		done:           make(chan struct{}),
		remove:         false,
		DebugInfo: &WebSocketDebufInfo{
			AverageFps: misc.NewRollingArray[float64](10),
		},
		SendChatMsgsFlag: false,
	}
	s.WebSocketConnections[connectionName] = websocketConn

	// Track that something has connected to the client
	log.Print("Client connected.")
	defer func() {
		log.Println("Closing connection and done channel.")
		close(websocketConn.done)
		websocketConn.conn.Close()
		websocketConn.conn = nil
		delete(s.WebSocketConnections, connectionName)
	}()

	for _, chatterInfo := range chatters {
		s.setInternalSpineOperator(
			chatterInfo.Username,
			chatterInfo.UsernameDisplay,
			chatterInfo.OperatorInfo,
			[]string{connectionName},
		)
	}

	for {
		var messageType, message, err = c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		// log.Printf("recv: (%v)%v\n", messageType, string(message))

		switch messageType {
		case websocket.TextMessage:
			// log.Print("TextMessage")
			s.handleResponseMessages(connectionName, message)
		case websocket.PingMessage:
			log.Print("PingMessage")
		case websocket.PongMessage:
			log.Print("PongMessage")
		case websocket.BinaryMessage:
			log.Print("BinaryMessage")
		case websocket.CloseMessage:
			log.Print("CloseMessage")
		default:
			log.Print("Default")
		}
	}
	return nil
}

func (s *SpineBridge) AddListenerToClientRequests(callback ClientRequestCallback) (func(), error) {
	currentId := s.clientResponseCallbackListenersId
	s.clientResponseCallbackListenersId += 1
	s.clientResponseCallbackListeners[currentId] = callback
	return func() {
		delete(s.clientResponseCallbackListeners, currentId)
	}, nil
}

func (s *SpineBridge) setInternalSpineOperator(
	UserName string,
	userNameDisplay string,
	info operator.OperatorInfo,
	connectionIds []string,
) error {
	// Validate the setOperator Request
	if err := s.spineService.ValidateOperatorRequest(&info); err != nil {
		return err
	}

	isBase := info.ChibiStance == operator.CHIBI_STANCE_ENUM_BASE
	isFront := info.Facing == operator.CHIBI_FACING_ENUM_FRONT

	atlasFile := ""
	pngFile := ""
	skelFile := ""
	skelJsonFile := ""
	spritesheetDataFilepath := ""
	spineData := s.spineService.GetSpineData(info.OperatorId, info.Faction, info.Skin, isBase, isFront)
	// TODO: Make the image/assets path a configurable variable
	formatPathFn := func(path string) string {
		// return "/static/assets/" + strings.ReplaceAll(path, string(os.PathSeparator), "/")
		return "/image/assets/" + strings.ReplaceAll(path, string(os.PathSeparator), "/")
	}

	if len(spineData.SpritesheetDataFilepath) > 0 {
		atlasFile = formatPathFn(spineData.AtlasFilepath)
		spritesheetDataFilepath = formatPathFn(spineData.SpritesheetDataFilepath)
	} else {
		atlasFile = formatPathFn(spineData.AtlasFilepath)
		pngFile = formatPathFn(spineData.PngFilepath)
		if len(spineData.SkelFilepath) > 0 {
			skelFile = formatPathFn(spineData.SkelFilepath)
		}
		if len(spineData.SkelJsonFilepath) > 0 {
			skelJsonFile = formatPathFn(spineData.SkelJsonFilepath)
		}
	}

	data := SetOperatorInternalRequest{
		BridgeRequest: BridgeRequest{
			TypeName: SET_OPERATOR,
		},
		UserName:                UserName,
		UserNameDisplay:         userNameDisplay,
		OperatorId:              info.OperatorId,
		AtlasFile:               atlasFile,
		PngFile:                 pngFile,
		SkelFile:                skelFile,
		SkelJsonFile:            skelJsonFile,
		SpritesheetDataFilepath: spritesheetDataFilepath,

		StartPos:            info.StartPos,
		AnimationSpeed:      info.AnimationSpeed,
		AvailableAnimations: info.AvailableAnimations,
		SpriteScale:         info.SpriteScale,
		MaxSpritePixelSize:  s.spineService.GetMaxSpritePixelSize(),
		MovementSpeedPx:     s.spineService.GetReferenceMovementSpeedPx(),
		MovementSpeed:       info.MovementSpeed,
		Action:              info.CurrentAction,
		ActionData:          info.Action,
	}
	data_json, _ := json.Marshal(data)
	log.Println("setInternalSpineOperator sending: ", string(data_json))

	for _, websocketConn := range s.WebSocketConnections {
		if websocketConn.conn == nil {
			continue
		}
		if connectionIds != nil && !slices.Contains(connectionIds, websocketConn.connectionName) {
			continue
		}
		websocketConn.conn.WriteJSON(data)
	}

	return nil
}

// Start Spine Client Interface functions
// ----------------------------
func (s *SpineBridge) SetOperator(req *SetOperatorRequest) (*SetOperatorResponse, error) {
	err := s.setInternalSpineOperator(
		req.UserName,
		req.UserNameDisplay,
		req.Operator,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &SetOperatorResponse{
		BridgeResponse: BridgeResponse{
			TypeName:   SET_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}, nil
}

func (s *SpineBridge) RemoveOperator(r *RemoveOperatorRequest) (*RemoveOperatorResponse, error) {
	successResp := &RemoveOperatorResponse{
		BridgeResponse: BridgeResponse{
			TypeName:   REMOVE_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}

	data := map[string]interface{}{
		"type_name": REMOVE_OPERATOR,
		"user_name": r.UserName,
	}

	if s.clientConnected() {
		data_json, _ := json.Marshal(data)
		log.Println("RemoveOperator() sending: ", string(data_json))
		for _, websocketConn := range s.WebSocketConnections {
			if websocketConn.conn != nil {
				websocketConn.conn.WriteJSON(data)
			}
		}
	}

	return successResp, nil
}

func (s *SpineBridge) ShowChatMessage(r *ShowChatMessageRequest) (*ShowChatMessageResponse, error) {
	if s.clientConnected() {
		data := ShowChatMessageInternalRequest{
			BridgeRequest: BridgeRequest{
				TypeName: SHOW_CHAT_MESSAGE,
			},
			UserName: r.UserName,
			// WARNING: This is unsanitized input we are sending to the FE
			// Should be fine because we never render this message in the DOM
			// and only directly as text in the TextCanvas
			Message: r.Message,
		}
		// data_json, _ := json.Marshal(data)
		// log.Println("ShowChatMessage() sending: ", string(data_json))
		for _, websocketConn := range s.WebSocketConnections {
			if websocketConn.conn != nil && websocketConn.SendChatMsgsFlag {
				websocketConn.conn.WriteJSON(data)
			}
		}
	}

	successResp := &ShowChatMessageResponse{
		BridgeResponse: BridgeResponse{
			TypeName:   SHOW_CHAT_MESSAGE,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}
	return successResp, nil
}

func (s *SpineBridge) FindOperator(r *FindOperatorRequest) (*FindOperatorResponse, error) {
	if s.clientConnected() {
		data := FindOperatorInternalRequest{
			BridgeRequest: BridgeRequest{
				TypeName: FIND_OPERATOR,
			},
			UserName: r.UserName,
		}
		// data_json, _ := json.Marshal(data)
		// log.Println("FindOperator() sending: ", string(data_json))
		for _, websocketConn := range s.WebSocketConnections {
			if websocketConn.conn != nil {
				err := websocketConn.conn.WriteJSON(data)
				if err != nil {
					log.Println("Error sending FindOperatorInternalRequest: ", err)
				}
			}
		}
	}

	successResp := &FindOperatorResponse{
		BridgeResponse: BridgeResponse{
			TypeName:   FIND_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}
	return successResp, nil
}

// ----------------------------
// End Spine Client Interface functions
