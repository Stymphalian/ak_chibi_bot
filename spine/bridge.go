package spine

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Stymphalian/ak_chibi_bot/misc"
	"github.com/gorilla/websocket"
)

type ChatUser struct {
	userName string

	// TODO: Change to just hold an OperatorInfo object
	currentOperatorName string
	currentOperatorId   string
	currentFaction      FactionEnum
	currentSkin         string
	currentChibiType    ChibiTypeEnum
	currentFacing       ChibiFacingEnum
	currentAnimation    string
	currentPositionX    *float64
}

type SpineBridge struct {
	chatUsers    map[string]*ChatUser
	twitchConfig *misc.TwitchConfig

	conn             *websocket.Conn
	done             chan struct{}
	AssetMap         *SpineAssetMap
	CommonNames      *CommonNames
	EnemyAssetMap    *SpineAssetMap
	EnemyCommonNames *CommonNames
}

func NewSpineBridge(assetDir string, config *misc.TwitchConfig) (*SpineBridge, error) {
	s := &SpineBridge{
		chatUsers:        make(map[string]*ChatUser, 0),
		twitchConfig:     config,
		conn:             nil,
		done:             nil,
		AssetMap:         NewSpineAssetMap(),
		CommonNames:      NewCommonNames(),
		EnemyAssetMap:    NewSpineAssetMap(),
		EnemyCommonNames: NewCommonNames(),
	}

	if err := s.AssetMap.Load(assetDir, "characters"); err != nil {
		return nil, err
	}
	if err := s.CommonNames.Load(filepath.Join(assetDir, "saved_names.json")); err != nil {
		return nil, err
	}
	if err := s.EnemyAssetMap.Load(assetDir, "enemies"); err != nil {
		return nil, err
	}
	if err := s.EnemyCommonNames.Load(filepath.Join(assetDir, "saved_enemy_names.json")); err != nil {
		return nil, err
	}

	// Check for missing assets
	for enemyId, characterIds := range s.EnemyCommonNames.operatorIdToNames {
		if _, ok := s.EnemyAssetMap.Data[enemyId]; !ok {
			if len(characterIds) > 0 {
				log.Println("Missing enemy", enemyId)
			}
		}
	}
	for operatorId, characterIds := range s.CommonNames.operatorIdToNames {
		if _, ok := s.AssetMap.Data[operatorId]; !ok {
			if len(characterIds) > 0 {
				log.Println("Missing operator", operatorId)
			}
		}
	}

	s.resetState()

	return s, nil
}

func (s *SpineBridge) Close() {
	log.Println("SpineBridge::Close() called")
	if s.conn == nil {
		return
	}
	err := s.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
	)
	if err != nil {
		log.Println("write close:", err)
	}
	<-s.done
	log.Println("SpineBridge::Close() finished")
}

func (s *SpineBridge) getOpenWsConnection() *websocket.Conn {
	return s.conn
}

func (s *SpineBridge) clientConnected() bool {
	return s.conn != nil
}

func (s *SpineBridge) HandleForward(w http.ResponseWriter, r *http.Request) error {
	if !s.clientConnected() {
		return errors.New("no client is yet attached to the bridge")
	}
	var upgrader = websocket.Upgrader{} // use default options
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade: ", err)
		return nil
	}
	defer c.Close()

	client := s.getOpenWsConnection()
	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s, type: %d", message, messageType)

		if err := client.WriteMessage(websocket.TextMessage, message); err != nil {
			log.Println("Failed to forward message", err)
		}
	}

	return nil
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
	// switch data["type_name"] {
	// case GET_ANIMATIONS:
	// 	log.Print(GET_ANIMATIONS)
	// 	resp := GetAnimationsResponse{}
	// 	if err := json.NewDecoder(bytes.NewReader(message)).Decode(&resp); err != nil {
	// 		log.Println("Failed to decode message", err)
	// 		break
	// 	}
	// 	log.Print("Processed AnimationsResponse ", resp)
	// 	s.skinAnimations = resp.Animations
	// }
}

func (s *SpineBridge) HandleSpine(w http.ResponseWriter, r *http.Request) error {
	if s.clientConnected() {
		return errors.New("client already connected. Only one client allowed at a time")
	}

	var upgrader = websocket.Upgrader{} // use default options
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade: ", err)
		return nil
	}

	// Track that something has connected to the client
	log.Print("Client connected.")
	s.conn = c
	s.done = make(chan struct{})
	defer func() {
		log.Println("Closing connection and done channel.")
		close(s.done)
		s.conn.Close()
		s.conn = nil
	}()

	chatUser := s.chatUsers[s.twitchConfig.Broadcaster]
	s.setInternalSpineOperator(
		chatUser.userName,
		chatUser.userName,
		chatUser.currentOperatorId,
		chatUser.currentFaction,
		chatUser.currentSkin,
		chatUser.currentChibiType,
		chatUser.currentFacing,
		chatUser.currentAnimation,
		chatUser.currentPositionX,
	)
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
		default:
			log.Print("Default")
		}
	}
	s.resetState()
	return nil
}

func (s *SpineBridge) HandleAdmin(w http.ResponseWriter, r *http.Request) error {
	if r.Method != http.MethodPost {
		return nil
	}

	decoder := json.NewDecoder(r.Body)
	var data map[string]interface{}
	err := decoder.Decode(&data)
	if err != nil {
		return err
	}

	switch data["action"].(string) {
	case "remove":
		userName := data["user_name"].(string)
		s.RemoveOperator(&RemoveOperatorRequest{UserName: userName})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
		})
		return nil
	case "add":
		userName := data["user_name"].(string)
		if _, ok := s.chatUsers[userName]; !ok {
			operatorId := "char_002_amiya"
			s.SetOperator(&SetOperatorRequest{
				UserName:        userName,
				UserNameDisplay: userName,
				OperatorId:      operatorId,
				Faction:         FACTION_ENUM_OPERATOR,
				Skin:            "default",
				ChibiType:       CHIBI_TYPE_ENUM_BASE,
				Facing:          CHIBI_FACING_ENUM_FRONT,
				Animation:       "Move",
				PositionX:       nil,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
		})
		return nil
	case "list":
		w.Header().Set("Content-Type", "application/json")
		usernames := make([]string, 0)
		for name := range s.chatUsers {
			usernames = append(usernames, name)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"status": "success",
			"users":  usernames,
		})
		return nil
	default:
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":  "failed",
			"message": "Unknown action",
		})
		return nil
	}
}

func (s *SpineBridge) resetState() {
	opName := "Croissant"
	opId, _ := s.GetOperatorIdFromName(opName, FACTION_ENUM_OPERATOR)

	broadcasterName := s.twitchConfig.Broadcaster
	s.chatUsers = map[string]*ChatUser{
		broadcasterName: {
			userName:            broadcasterName,
			currentOperatorName: opName,
			currentOperatorId:   opId,
			currentFaction:      FACTION_ENUM_OPERATOR,
			currentSkin:         "default",
			currentChibiType:    CHIBI_TYPE_ENUM_BASE,
			currentFacing:       CHIBI_FACING_ENUM_FRONT,
			currentAnimation:    "Move",
			currentPositionX:    nil,
		},
	}
}

func (s *SpineBridge) setInternalSpineOperator(
	userName string,
	userNameDisplay string,
	operatorId string,
	faction FactionEnum,
	skinName string,
	chibiType ChibiTypeEnum,
	facing ChibiFacingEnum,
	animation string,
	positionX *float64,
) error {
	assetMap := s.getAssetMapFromFaction(faction)
	commonNames := s.getCommonNamesFromFaction(faction)

	// Validate the setOperator Request
	{
		log.Println("Request setOperator", operatorId, faction, skinName, chibiType, facing, animation)
		err := assetMap.Contains(operatorId, skinName, chibiType, facing, animation)
		if err != nil {
			log.Println("Validate setOperator request failed", err)
			return err
		}
	}

	isBase := chibiType == CHIBI_TYPE_ENUM_BASE
	isFront := facing == CHIBI_FACING_ENUM_FRONT

	atlasFile := ""
	pngFile := ""
	skelFile := ""
	defaultAnimation := "Idle"
	if isBase {
		defaultAnimation = "Relax"
	}
	spineData := assetMap.Get(operatorId, skinName, isBase, isFront)
	atlasFile = spineData.AtlasFilepath
	pngFile = spineData.PngFilepath
	skelFile = spineData.SkelFilepath
	if len(animation) == 0 {
		animation = defaultAnimation
	}

	formatPathFn := func(path string) string {
		return "assets/" + strings.ReplaceAll(path, string(os.PathSeparator), "/")
	}

	// atlasFileContentsB64 := ""
	// atlasFileBytes, err := os.ReadFile(spineData.AtlasFullFilepath)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// atlasFileContentsB64 = base64.StdEncoding.EncodeToString(atlasFileBytes)

	// skelFileContentsB64 := ""
	// skelFileBytes, err := os.ReadFile(spineData.SkelFullFilepath)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// skelFileContentsB64 = base64.StdEncoding.EncodeToString(skelFileBytes)

	// pngFileContentsB64 := ""
	// pngFileBytes, err := os.ReadFile(spineData.PngFullFilepath)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return err
	// }
	// pngFileContentsB64 = base64.StdEncoding.EncodeToString(pngFileBytes)

	wandering := false
	if positionX == nil {
		wandering = true
	}

	data := map[string]interface{}{
		"type_name":         SET_OPERATOR,
		"user_name":         userName,
		"user_name_display": userNameDisplay,
		"operator_id":       operatorId,
		"atlas_file":        formatPathFn(atlasFile),
		"png_file":          formatPathFn(pngFile),
		"skel_file":         formatPathFn(skelFile),
		"position_x":        positionX,
		"wandering":         wandering,

		// "atlas_file_base64": atlasFileContentsB64,
		// "skel_file_base64":  skelFileContentsB64,
		// "png_file_base64":   pngFileContentsB64,

		"animation": animation,
	}

	data_json, _ := json.Marshal(data)
	log.Println("setInternalSpineOperator sending: ", string(data_json))
	if s.conn != nil {
		s.conn.WriteJSON(data)
	}

	chatUser, ok := s.chatUsers[userName]
	if !ok {
		s.chatUsers[userName] = &ChatUser{userName: userName}
		chatUser = s.chatUsers[userName]
	}

	chatUser.currentOperatorName = commonNames.GetCanonicalName(operatorId)
	chatUser.currentOperatorId = operatorId
	chatUser.currentFaction = faction
	chatUser.currentSkin = skinName
	if isBase {
		chatUser.currentChibiType = CHIBI_TYPE_ENUM_BASE
	} else {
		chatUser.currentChibiType = CHIBI_TYPE_ENUM_BATTLE
	}
	if isFront {
		chatUser.currentFacing = CHIBI_FACING_ENUM_FRONT
	} else {
		chatUser.currentFacing = CHIBI_FACING_ENUM_BACK
	}
	chatUser.currentAnimation = animation
	chatUser.currentPositionX = positionX

	return nil
}

func (s *SpineBridge) getAssetMapFromFaction(faction FactionEnum) *SpineAssetMap {
	switch faction {
	case FACTION_ENUM_OPERATOR:
		return s.AssetMap
	case FACTION_ENUM_ENEMY:
		return s.EnemyAssetMap
	default:
		log.Fatalf("Unknown faction when fetching assetmap: %v", faction)
		return nil
	}
}

func (s *SpineBridge) getCommonNamesFromFaction(faction FactionEnum) *CommonNames {
	switch faction {
	case FACTION_ENUM_OPERATOR:
		return s.CommonNames
	case FACTION_ENUM_ENEMY:
		return s.EnemyCommonNames
	default:
		log.Fatalf("Unknown faction when fetching common names: %v", faction)
		return nil
	}
}

// Start Spine Client Interfact functions
// ----------------------------
func (s *SpineBridge) SetOperator(req *SetOperatorRequest) (*SetOperatorResponse, error) {
	if !s.clientConnected() {
		return nil, errors.New("SpineBridge client is not yet attached")
	}

	err := s.setInternalSpineOperator(
		req.UserName,
		req.UserNameDisplay,
		req.OperatorId,
		req.Faction,
		req.Skin,
		req.ChibiType,
		req.Facing,
		req.Animation,
		req.PositionX,
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
	if !s.clientConnected() {
		return nil, errors.New("SpineBridge client is not yet attached")
	}
	assetMap := s.getAssetMapFromFaction(req.Faction)

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
			Stances: map[ChibiTypeEnum]FacingData{
				CHIBI_TYPE_ENUM_BASE: {
					Facings: baseMapping,
				},
				CHIBI_TYPE_ENUM_BATTLE: {
					Facings: battleMapping,
				},
			},
		}
	}

	return &GetOperatorResponse{
		SpineResponse: SpineResponse{
			TypeName:   SET_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
		OperatorId: req.OperatorId,
		Skins:      skinDataMap,
	}, nil
}

func (s *SpineBridge) RemoveOperator(r *RemoveOperatorRequest) (*RemoveOperatorResponse, error) {
	if !s.clientConnected() {
		return nil, errors.New("SpineBridge client is not yet attached")
	}

	successResp := &RemoveOperatorResponse{
		SpineResponse: SpineResponse{
			TypeName:   REMOVE_OPERATOR,
			ErrorMsg:   "",
			StatusCode: 200,
		},
	}

	// We already don't have an entry for this user, so just return early
	if _, ok := s.chatUsers[r.UserName]; !ok {
		return successResp, nil
	}

	data := map[string]interface{}{
		"type_name": REMOVE_OPERATOR,
		"user_name": r.UserName,
	}

	data_json, _ := json.Marshal(data)
	log.Println("RemoveOperator() sending: ", string(data_json))
	if s.conn != nil {
		s.conn.WriteJSON(data)
	}

	delete(s.chatUsers, r.UserName)
	return successResp, nil
}

// TODO: Add a GetEnemyIds()
func (s *SpineBridge) GetOperatorIds(faction FactionEnum) ([]string, error) {
	assetMap := s.getAssetMapFromFaction(faction)
	operatorIds := make([]string, 0)
	for operatorId := range assetMap.Data {
		operatorIds = append(operatorIds, operatorId)
	}
	return operatorIds, nil
}

func (s *SpineBridge) GetOperatorIdFromName(name string, faction FactionEnum) (string, []string) {
	commonNames := s.getCommonNamesFromFaction(faction)

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

func (s *SpineBridge) CurrentInfo(userName string) (OperatorInfo, error) {
	if !s.clientConnected() {
		return OperatorInfo{}, errors.New("SpineBridge client is not yet attached")
	}
	excludeAnimations := []string{
		"Default",
		"Start",
	}
	chatUser, ok := s.chatUsers[userName]
	if !ok {
		return *EmptyOperatorInfo(), NewUserNotFound("User not found: " + userName)
	}

	assetMap := s.getAssetMapFromFaction(chatUser.currentFaction)

	skins := make([]string, 0)
	for skinName := range assetMap.Data[chatUser.currentOperatorId].Skins {
		skins = append(skins, skinName)
	}

	animations := make([]string, 0)
	spineData := assetMap.Get(
		chatUser.currentOperatorId,
		chatUser.currentSkin,
		chatUser.currentChibiType == CHIBI_TYPE_ENUM_BASE,
		chatUser.currentFacing == CHIBI_FACING_ENUM_FRONT,
	)
	for _, animationName := range spineData.Animations {
		if slices.Contains(excludeAnimations, animationName) {
			continue
		}
		if strings.HasSuffix(animationName, "_Begin") {
			continue
		}
		if strings.HasSuffix(animationName, "_End") {
			continue
		}
		animations = append(animations, animationName)
	}

	// positionX = -1.0
	return OperatorInfo{
		Name:       chatUser.currentOperatorName,
		OperatorId: chatUser.currentOperatorId,
		Faction:    chatUser.currentFaction,
		Skin:       chatUser.currentSkin,
		ChibiType:  chatUser.currentChibiType,
		Facing:     chatUser.currentFacing,
		Animation:  chatUser.currentAnimation,
		PositionX:  nil,

		Skins:      skins,
		Animations: animations,
	}, nil
}

// ----------------------------
// End Spine Client Interfact functions
