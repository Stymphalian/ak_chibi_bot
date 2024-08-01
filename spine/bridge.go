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
	"time"

	"github.com/Stymphalian/ak_chibi_bot/misc"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Room struct {
	conn *websocket.Conn
	done chan struct{}
}

type ChatUser struct {
	userName        string
	currentOperator OperatorInfo
}

type SpineBridge struct {
	chatUsers    map[string]*ChatUser
	twitchConfig *misc.TwitchConfig

	Rooms            map[string]*Room
	AssetMap         *SpineAssetMap
	CommonNames      *CommonNames
	EnemyAssetMap    *SpineAssetMap
	EnemyCommonNames *CommonNames
}

func NewSpineBridge(assetDir string, config *misc.TwitchConfig) (*SpineBridge, error) {
	s := &SpineBridge{
		chatUsers:        make(map[string]*ChatUser, 0),
		twitchConfig:     config,
		Rooms:            make(map[string]*Room, 0),
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

	s.resetState(config.InitialOperator, config.OperatorDetails)

	return s, nil
}

func (s *SpineBridge) Close() {
	log.Println("SpineBridge::Close() called")
	for roomName, room := range s.Rooms {
		if room.conn == nil {
			return
		}
		err := room.conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		if err != nil {
			log.Printf("write close for room %s: %v\n", roomName, err)
		}
		<-room.done
	}
	log.Println("SpineBridge::Close() finished")
}

func (s *SpineBridge) clientConnected() bool {
	return len(s.Rooms) > 0
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

func (s *SpineBridge) HandleSpine(w http.ResponseWriter, r *http.Request) error {
	var upgrader = websocket.Upgrader{} // use default options
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade: ", err)
		return nil
	}

	room := &Room{
		conn: c,
		done: make(chan struct{}),
	}
	// Get a uuid string
	roomName := uuid.New().String()
	s.Rooms[roomName] = room

	// Track that something has connected to the client
	log.Print("Client connected.")
	room.conn = c
	room.done = make(chan struct{})
	defer func() {
		log.Println("Closing connection and done channel.")
		close(room.done)
		room.conn.Close()
		room.conn = nil
	}()

	for _, chatUser := range s.chatUsers {
		s.setInternalSpineOperator(
			chatUser.userName,
			chatUser.userName,
			chatUser.currentOperator,
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
		default:
			log.Print("Default")
		}
	}
	s.resetState(s.twitchConfig.InitialOperator, s.twitchConfig.OperatorDetails)
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
	case "debug":
		operatorIdsInterface, ok := data["operator_ids"].([]interface{})
		if !ok {
			return errors.New("operator_ids is not an array")
		}
		operator_ids := make([]string, len(operatorIdsInterface))
		for i, idInterface := range operatorIdsInterface {
			idString, ok := idInterface.(string)
			if !ok {
				return errors.New("operator_ids contains a non-string element")
			}
			operator_ids[i] = idString
		}

		enemyIdsInterface, ok := data["enemy_ids"].([]interface{})
		if !ok {
			return errors.New("operator_ids is not an array")
		}
		enemy_ids := make([]string, len(enemyIdsInterface))
		for i, idInterface := range enemyIdsInterface {
			idString, ok := idInterface.(string)
			if !ok {
				return errors.New("enemy_ids contains a non-string element")
			}
			enemy_ids[i] = idString
		}

		startPosX := 0 + 0.02
		startPosY := 0.04
		for _, operator_id := range operator_ids {
			resp, err := s.GetOperator(&GetOperatorRequest{
				OperatorId: operator_id,
				Faction:    FACTION_ENUM_OPERATOR,
			})
			if err != nil {
				log.Panic("Failed to get operator", err)
			}

			for skin, skinEntry := range resp.Skins {
				if !skinEntry.HasChibiType(CHIBI_TYPE_ENUM_BATTLE) {
					continue
				}
				if skin != DEFAULT_SKIN_NAME {
					continue
				}
				_, err := s.SetOperator(&SetOperatorRequest{
					UserName:        operator_id + "_" + skin,
					UserNameDisplay: operator_id + "_" + skin,
					Operator: OperatorInfo{
						OperatorId:        operator_id,
						Faction:           FACTION_ENUM_OPERATOR,
						Skin:              skin,
						ChibiType:         CHIBI_TYPE_ENUM_BATTLE,
						Facing:            CHIBI_FACING_ENUM_FRONT,
						CurrentAnimations: []string{DEFAULT_ANIM_BATTLE},
						StartPos:          misc.NewOption(misc.Vector2{X: startPosX, Y: startPosY}),
					},
				})
				if err != nil {
					log.Panic("Failed to set operator", err)
				}

				if err == nil {
					startPosX += 0.08
					if startPosX > 1.0 {
						startPosX = 0.02
						startPosY += 0.08
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}

		for _, operator_id := range enemy_ids {
			resp, err := s.GetOperator(&GetOperatorRequest{
				OperatorId: operator_id,
				Faction:    FACTION_ENUM_ENEMY,
			})
			if err != nil {
				log.Panic("Failed to get enemy", err)
			}

			for skin, skinEntry := range resp.Skins {
				if !skinEntry.HasChibiType(CHIBI_TYPE_ENUM_BATTLE) {
					continue
				}

				animation := DEFAULT_ANIM_BATTLE
				for _, anim := range skinEntry.Stances[CHIBI_TYPE_ENUM_BATTLE].Facings[CHIBI_FACING_ENUM_FRONT] {
					if anim == DEFAULT_ANIM_BATTLE {
						animation = DEFAULT_ANIM_BATTLE
						break
					}
					animation = anim
				}

				_, err := s.SetOperator(&SetOperatorRequest{
					UserName:        operator_id + "_" + skin,
					UserNameDisplay: operator_id + "_" + skin,
					Operator: OperatorInfo{
						OperatorId:        operator_id,
						Faction:           FACTION_ENUM_ENEMY,
						Skin:              skin,
						ChibiType:         CHIBI_TYPE_ENUM_BATTLE,
						Facing:            CHIBI_FACING_ENUM_FRONT,
						CurrentAnimations: []string{animation},
						StartPos:          misc.NewOption(misc.Vector2{X: startPosX, Y: startPosY}),
					},
				})
				if err != nil {
					log.Panic("Failed to set enemy", err)
				}

				if err == nil {
					startPosX += 0.08
					if startPosX > 1.0 {
						startPosX = 0.02
						startPosY += 0.08
					}
				}
				time.Sleep(100 * time.Millisecond)
			}
		}

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
				Operator: OperatorInfo{
					OperatorId:        operatorId,
					Faction:           FACTION_ENUM_OPERATOR,
					Skin:              DEFAULT_SKIN_NAME,
					ChibiType:         CHIBI_TYPE_ENUM_BASE,
					Facing:            CHIBI_FACING_ENUM_FRONT,
					CurrentAnimations: []string{DEFAULT_ANIM_BASE},
				},
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

func (s *SpineBridge) resetState(opName string, details misc.InitialOperatorDetails) {
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
	stance, err2 := ChibiTypeEnum_Parse(details.Stance)
	if err2 != nil {
		log.Panic("Failed to parse stance", err2)
	}

	broadcasterName := s.twitchConfig.Broadcaster
	s.chatUsers = map[string]*ChatUser{
		broadcasterName: {
			userName: broadcasterName,
			currentOperator: OperatorInfo{
				DisplayName:       opName,
				OperatorId:        opId,
				Faction:           faction,
				Skin:              details.Skin,
				ChibiType:         stance,
				Facing:            CHIBI_FACING_ENUM_FRONT,
				CurrentAnimations: details.Animations,

				StartPos:   misc.NewOption(misc.Vector2{X: details.PositionX, Y: 0.0}),
				Skins:      nil,
				Animations: nil,
			},
		},
	}
}

func (s *SpineBridge) setInternalSpineOperator(
	userName string,
	userNameDisplay string,
	info OperatorInfo,
) error {
	assetMap := s.getAssetMapFromFaction(info.Faction)
	commonNames := s.getCommonNamesFromFaction(info.Faction)

	// Validate the setOperator Request
	{
		log.Println("Request setOperator", info.OperatorId, info.Faction, info.Skin, info.ChibiType, info.Facing, info.CurrentAnimations)
		err := assetMap.Contains(info.OperatorId, info.Skin, info.ChibiType, info.Facing, info.CurrentAnimations)
		if err != nil {
			log.Println("Validate setOperator request failed", err)
			return err
		}
	}

	isBase := info.ChibiType == CHIBI_TYPE_ENUM_BASE
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
	if info.AnimationSpeed == 0.0 {
		info.AnimationSpeed = 1.0
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
	if info.TargetPos.IsNone() {
		wandering = true
	}

	data := map[string]interface{}{
		"type_name":         SET_OPERATOR,
		"user_name":         userName,
		"user_name_display": userNameDisplay,
		"operator_id":       info.OperatorId,
		"atlas_file":        formatPathFn(atlasFile),
		"png_file":          formatPathFn(pngFile),
		"skel_file":         formatPathFn(skelFile),
		"wandering":         wandering,
		"animations":        info.CurrentAnimations,
		"start_pos":         info.StartPos,
		"target_pos":        info.TargetPos,
		"animation_speed":   info.AnimationSpeed,
		// "atlas_file_base64": atlasFileContentsB64,
		// "skel_file_base64":  skelFileContentsB64,
		// "png_file_base64":   pngFileContentsB64,
	}

	data_json, _ := json.Marshal(data)
	log.Println("setInternalSpineOperator sending: ", string(data_json))
	for _, room := range s.Rooms {
		if room.conn != nil {
			room.conn.WriteJSON(data)
		}
	}

	chatUser, ok := s.chatUsers[userName]
	if !ok {
		s.chatUsers[userName] = &ChatUser{userName: userName}
		chatUser = s.chatUsers[userName]
	}

	chatUser.currentOperator = info
	chatUser.currentOperator.DisplayName = commonNames.GetCanonicalName(info.OperatorId)
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

	canonicalName := s.getCommonNamesFromFaction(req.Faction).GetCanonicalName(req.OperatorId)

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
	for _, room := range s.Rooms {
		if room.conn != nil {
			room.conn.WriteJSON(data)
		}
	}

	delete(s.chatUsers, r.UserName)
	return successResp, nil
}

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

	assetMap := s.getAssetMapFromFaction(chatUser.currentOperator.Faction)

	skins := make([]string, 0)
	for skinName := range assetMap.Data[chatUser.currentOperator.OperatorId].Skins {
		skins = append(skins, skinName)
	}

	animations := make([]string, 0)
	spineData := assetMap.Get(
		chatUser.currentOperator.OperatorId,
		chatUser.currentOperator.Skin,
		chatUser.currentOperator.ChibiType == CHIBI_TYPE_ENUM_BASE,
		chatUser.currentOperator.Facing == CHIBI_FACING_ENUM_FRONT,
	)
	for _, animationName := range spineData.Animations {
		if slices.Contains(excludeAnimations, animationName) {
			continue
		}
		if strings.Contains(animationName, "Default") {
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
		DisplayName:       chatUser.currentOperator.DisplayName,
		OperatorId:        chatUser.currentOperator.OperatorId,
		Faction:           chatUser.currentOperator.Faction,
		Skin:              chatUser.currentOperator.Skin,
		ChibiType:         chatUser.currentOperator.ChibiType,
		Facing:            chatUser.currentOperator.Facing,
		CurrentAnimations: chatUser.currentOperator.CurrentAnimations,

		Skins:      skins,
		Animations: animations,
	}, nil
}

// ----------------------------
// End Spine Client Interfact functions
