package spine

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketConn struct {
	conn   *websocket.Conn
	done   chan struct{}
	remove bool
}

type SpineBridge struct {
	spineService          *SpineService
	WebSocketConnections  map[string]*WebSocketConn
	websocketPingerTicker *time.Ticker
	websocketPingerDone   chan bool
	// TODO: Might want to add mutex locking for updating websocket connections
}

func NewSpineBridge(spineService *SpineService) (*SpineBridge, error) {
	s := &SpineBridge{
		spineService:         spineService,
		WebSocketConnections: make(map[string]*WebSocketConn, 0),
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
	close(s.websocketPingerDone)

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

func (s *SpineBridge) handleResponseMessages(message []byte) {
	var data map[string]interface{}
	log.Println("Received message", string(message))
	err := json.Unmarshal(message, &data)
	if err != nil {
		log.Printf("Error decoding JSON: %v", err)
		return
	}
	log.Println("data", data)
}

func (s *SpineBridge) NumConnections() int {
	return len(s.WebSocketConnections)
}

func (s *SpineBridge) AddConnection(
	w http.ResponseWriter,
	r *http.Request,
	chatters []*ChatUser,
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

	websocketConn := &WebSocketConn{
		conn:   c,
		done:   make(chan struct{}),
		remove: false,
	}
	// Get a uuid string
	connectionName := uuid.New().String()
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

	for _, chatUser := range chatters {
		s.setInternalSpineOperator(
			chatUser.UserName,
			chatUser.UserNameDisplay,
			chatUser.CurrentOperator,
		)
	}

	for {
		var messageType, message, err = c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: (%v)%v\n", messageType, string(message))

		switch messageType {
		case websocket.TextMessage:
			log.Print("TextMessage")
			s.handleResponseMessages(message)
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

func (s *SpineBridge) setInternalSpineOperator(
	UserName string,
	userNameDisplay string,
	info OperatorInfo,
) error {
	// Validate the setOperator Request
	if err := s.spineService.ValidateOperatorRequest(&info); err != nil {
		return err
	}

	isBase := info.ChibiStance == CHIBI_STANCE_ENUM_BASE
	isFront := info.Facing == CHIBI_FACING_ENUM_FRONT

	atlasFile := ""
	pngFile := ""
	skelFile := ""
	spineData := s.spineService.GetSpineData(info.OperatorId, info.Faction, info.Skin, isBase, isFront)
	atlasFile = spineData.AtlasFilepath
	pngFile = spineData.PngFilepath
	skelFile = spineData.SkelFilepath
	formatPathFn := func(path string) string {
		return "/static/assets/" + strings.ReplaceAll(path, string(os.PathSeparator), "/")
	}

	data := map[string]interface{}{
		"type_name":             SET_OPERATOR,
		"user_name":             UserName,
		"user_name_display":     userNameDisplay,
		"operator_id":           info.OperatorId,
		"atlas_file":            formatPathFn(atlasFile),
		"png_file":              formatPathFn(pngFile),
		"skel_file":             formatPathFn(skelFile),
		"start_pos":             info.StartPos,
		"animation_speed":       info.AnimationSpeed,
		"available_animations":  info.AvailableAnimations,
		"sprite_scale":          info.SpriteScale,
		"max_sprite_pixel_size": s.spineService.GetMaxSpritePixelSize(),
		"movement_speed_px":     s.spineService.GetReferenceMovementSpeedPx(),
		"movement_speed":        info.MovementSpeed,

		"action":      info.CurrentAction,
		"action_data": info.Action,
	}

	data_json, _ := json.Marshal(data)
	log.Println("setInternalSpineOperator sending: ", string(data_json))

	for _, websocketConn := range s.WebSocketConnections {
		if websocketConn.conn != nil {
			websocketConn.conn.WriteJSON(data)
		}
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
	)
	if err != nil {
		return nil, err
	}

	return &SetOperatorResponse{
		SpineResponse: SpineResponse{
			TypeName:   SET_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}, nil
}

func (s *SpineBridge) RemoveOperator(r *RemoveOperatorRequest) (*RemoveOperatorResponse, error) {
	successResp := &RemoveOperatorResponse{
		SpineResponse: SpineResponse{
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

	// delete(s.ChatUsers, r.UserName)
	return successResp, nil
}

// ----------------------------
// End Spine Client Interface functions
