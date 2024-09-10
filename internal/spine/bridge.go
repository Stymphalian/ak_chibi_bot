package spine

import (
	"encoding/json"
	"errors"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/internal/misc"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WebSocketConn struct {
	conn   *websocket.Conn
	done   chan struct{}
	remove bool
}

type ChatUser struct {
	UserName        string
	UserNameDisplay string
	CurrentOperator OperatorInfo
}

type SpineBridge struct {
	Assets                *AssetManager
	ChatUsers             map[string]*ChatUser
	WebSocketConnections  map[string]*WebSocketConn
	websocketPingerTicker *time.Ticker
	websocketPingerDone   chan bool
	// TODO: Might want to add mutex locking for updating websocket connections
}

func NewSpineBridge(assets *AssetManager) (*SpineBridge, error) {
	s := &SpineBridge{
		Assets:               assets,
		ChatUsers:            make(map[string]*ChatUser, 0),
		WebSocketConnections: make(map[string]*WebSocketConn, 0),
	}
	go s.pingWebSockets()
	return s, nil
}

func (s *SpineBridge) pingWebSockets() {
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
					time.Now().Add(time.Duration(1)*time.Second),
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
			select {
			case <-time.After(100 * time.Millisecond):
				wg.Done()
			case <-websocketConn.done:
				wg.Done()
			}
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

func (s *SpineBridge) AddWebsocketConnection(w http.ResponseWriter, r *http.Request) error {
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

	for _, chatUser := range s.ChatUsers {
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

func (s *SpineBridge) HandleAdmin(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *SpineBridge) setInternalSpineOperator(
	UserName string,
	userNameDisplay string,
	info OperatorInfo,
) error {
	assetMap := s.Assets.getAssetMapFromFaction(info.Faction)

	// Validate the setOperator Request
	{
		// log.Println("Request setOperator", info.OperatorId, info.Faction, info.Skin, info.ChibiStance, info.Facing, info.CurrentAnimations)
		log.Println("Request setOperator", info.OperatorId, info.Faction,
			info.Skin, info.ChibiStance, info.Facing, info.CurrentAction)
		err := assetMap.Contains(info.OperatorId, info.Skin, info.ChibiStance,
			info.Facing, info.Action.GetAnimations(info.CurrentAction))
		if err != nil {
			log.Println("Validate setOperator request failed", err)
			return err
		}
	}

	isBase := info.ChibiStance == CHIBI_STANCE_ENUM_BASE
	isFront := info.Facing == CHIBI_FACING_ENUM_FRONT

	atlasFile := ""
	pngFile := ""
	skelFile := ""
	spineData := assetMap.Get(info.OperatorId, info.Skin, isBase, isFront)
	atlasFile = spineData.AtlasFilepath
	pngFile = spineData.PngFilepath
	skelFile = spineData.SkelFilepath
	formatPathFn := func(path string) string {
		return "assets/" + strings.ReplaceAll(path, string(os.PathSeparator), "/")
	}

	// wandering := false
	// if info.TargetPos.IsNone() {
	// 	wandering = true
	// }

	data := map[string]interface{}{
		"type_name":            SET_OPERATOR,
		"user_name":            UserName,
		"user_name_display":    userNameDisplay,
		"operator_id":          info.OperatorId,
		"atlas_file":           formatPathFn(atlasFile),
		"png_file":             formatPathFn(pngFile),
		"skel_file":            formatPathFn(skelFile),
		"start_pos":            info.StartPos,
		"animation_speed":      info.AnimationSpeed,
		"available_animations": info.AvailableAnimations,

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

	chatUser, ok := s.ChatUsers[UserName]
	if !ok {
		s.ChatUsers[UserName] = &ChatUser{
			UserName:        UserName,
			UserNameDisplay: userNameDisplay,
		}
		chatUser = s.ChatUsers[UserName]
	}
	chatUser.UserNameDisplay = userNameDisplay
	chatUser.CurrentOperator = info
	return nil
}

func (s *SpineBridge) getOperatorIds(faction FactionEnum) ([]string, error) {
	assetMap := s.Assets.getAssetMapFromFaction(faction)
	operatorIds := make([]string, 0)
	for operatorId := range assetMap.Data {
		operatorIds = append(operatorIds, operatorId)
	}
	return operatorIds, nil
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

func (s *SpineBridge) GetOperator(req *GetOperatorRequest) (*GetOperatorResponse, error) {
	assetMap := s.Assets.getAssetMapFromFaction(req.Faction)

	operatorData, ok := assetMap.Data[req.OperatorId]
	if !ok {
		return nil, errors.New("No operator with id " + req.OperatorId + " is loaded")
	}

	skinDataMap := make(map[string]SkinData)
	for skinName, skin := range operatorData.Skins {

		baseMapping := make(map[ChibiFacingEnum]AnimationsList, 0)
		battleMapping := make(map[ChibiFacingEnum]AnimationsList, 0)
		for facing, spineData := range skin.Base {
			baseMapping[facing] = spineData.Animations
		}
		for facing, spineData := range skin.Battle {
			battleMapping[facing] = spineData.Animations
		}

		skinDataMap[skinName] = SkinData{
			Stances: map[ChibiStanceEnum]FacingData{
				CHIBI_STANCE_ENUM_BASE: {
					Facings: baseMapping,
				},
				CHIBI_STANCE_ENUM_BATTLE: {
					Facings: battleMapping,
				},
			},
		}
	}

	canonicalName := s.Assets.getCommonNamesFromFaction(req.Faction).GetCanonicalName(req.OperatorId)

	return &GetOperatorResponse{
		SpineResponse: SpineResponse{
			TypeName:   SET_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
		OperatorId:   req.OperatorId,
		OperatorName: canonicalName,
		Skins:        skinDataMap,
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

	// We already don't have an entry for this user, so just return early
	if _, ok := s.ChatUsers[r.UserName]; !ok {
		return successResp, nil
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

	delete(s.ChatUsers, r.UserName)
	return successResp, nil
}

func (s *SpineBridge) GetOperatorIdFromName(name string, faction FactionEnum) (string, []string) {
	commonNames := s.Assets.getCommonNamesFromFaction(faction)

	if operatorId, ok := commonNames.IsMatch(name); ok {
		return operatorId, nil
	}

	matches := commonNames.FindMatchs(name, 5)
	humanMatches := make([]string, 0)
	for _, match := range matches {
		humanMatches = append(humanMatches, commonNames.operatorIdToNames[match][0])
	}
	return "", humanMatches
}

func (s *SpineBridge) CurrentInfo(UserName string) (OperatorInfo, error) {
	chatUser, ok := s.ChatUsers[UserName]
	if !ok {
		return *EmptyOperatorInfo(), NewUserNotFound("User not found: " + UserName)
	}

	return chatUser.CurrentOperator, nil
}

func (s *SpineBridge) SetToDefault(broadcasterName string, opName string, details misc.InitialOperatorDetails) {
	if len(opName) == 0 {
		opName = "Amiya"
	}

	faction := FACTION_ENUM_OPERATOR
	opId, matches := s.GetOperatorIdFromName(opName, FACTION_ENUM_OPERATOR)
	if matches != nil {
		faction = FACTION_ENUM_ENEMY
		opId, matches = s.GetOperatorIdFromName(opName, FACTION_ENUM_ENEMY)
	}
	if matches != nil {
		log.Panic("Failed to get operator id", matches)
	}
	stance, err2 := ChibiStanceEnum_Parse(details.Stance)
	if err2 != nil {
		log.Panic("Failed to parse stance", err2)
	}

	opResp, err := s.GetOperator(&GetOperatorRequest{opId, faction})
	if err != nil {
		log.Panic("Failed to fetch operator info")
	}
	availableAnims := opResp.Skins[details.Skin].Stances[stance].Facings[CHIBI_FACING_ENUM_FRONT]
	availableAnims = FilterAnimations(availableAnims)
	availableSkins := opResp.GetSkinNames()

	opInfo := NewOperatorInfo(
		opResp.OperatorName,
		faction,
		opId,
		details.Skin,
		stance,
		CHIBI_FACING_ENUM_FRONT,
		availableSkins,
		availableAnims,
		1.0,
		misc.NewOption(misc.Vector2{X: details.PositionX, Y: 0.0}),
		ACTION_PLAY_ANIMATION,
		NewActionPlayAnimation(details.Animations),
	)

	s.ChatUsers = map[string]*ChatUser{
		broadcasterName: {
			UserName:        broadcasterName,
			UserNameDisplay: broadcasterName,
			CurrentOperator: opInfo,
		},
	}
}

func (s *SpineBridge) GetRandomOperator() (*OperatorInfo, error) {
	operatorIds, err := s.getOperatorIds(FACTION_ENUM_OPERATOR)
	if err != nil {
		return nil, err
	}

	index := rand.Intn(len(operatorIds))
	operatorId := operatorIds[index]

	faction := FACTION_ENUM_OPERATOR
	operatorData, err := s.GetOperator(&GetOperatorRequest{
		OperatorId: operatorId,
		Faction:    faction,
	})
	if err != nil {
		return nil, err
	}
	chibiStance := CHIBI_STANCE_ENUM_BASE
	skinName := DEFAULT_SKIN_NAME
	stanceMap, ok := operatorData.Skins[skinName].Stances[chibiStance]
	if !ok {
		chibiStance = CHIBI_STANCE_ENUM_BATTLE
	}
	if len(stanceMap.Facings) == 0 {
		chibiStance = CHIBI_STANCE_ENUM_BATTLE
	}
	facing := CHIBI_FACING_ENUM_FRONT
	availableAnimations := operatorData.Skins[skinName].Stances[chibiStance].Facings[facing]
	availableAnimations = FilterAnimations(availableAnimations)
	availableSkins := operatorData.GetSkinNames()

	commonNames := s.Assets.getCommonNamesFromFaction(faction)
	operatorDisplayName := commonNames.GetCanonicalName(operatorId)

	opInfo := NewOperatorInfo(
		operatorDisplayName,
		faction,
		operatorId,
		skinName,
		chibiStance,
		facing,
		availableSkins,
		availableAnimations,
		1.0,
		misc.EmptyOption[misc.Vector2](),
		ACTION_PLAY_ANIMATION,
		NewActionPlayAnimation(
			[]string{GetDefaultAnimForChibiStance(chibiStance)},
		),
	)
	return &opInfo, nil
}

// ----------------------------
// End Spine Client Interface functions
