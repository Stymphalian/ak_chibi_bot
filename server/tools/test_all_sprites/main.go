package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/api"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
)

type Context struct {
	nextX            float32
	nextY            float32
	positionX        float32
	positionY        float32
	count            int
	faction          operator.FactionEnum
	authToken        string
	weirdOperatorIds map[string]bool
	weirdEnemyIds    map[string]bool
}

func NewContext(authToken string) *Context {
	weirdOperatorIds := map[string]bool{
		"char_261_sddrag":  true,
		"char_260_durnar":  true,
		"char_2015_dusk":   true,
		"char_4048_doroth": true,
		"char_017_huang":   true,
		"char_1020_reed2":  true,
		"char_469_indigo":  true,
		"char_475_akafyu":  true,
		"char_242_otter":   true,
		"char_4054_malist": true,
		"char_4071_peper":  true,
		"char_388_mint":    true,
		"char_455_nothin":  true,
		"char_431_ashlok":  true,
		"char_4047_pianst": true,
		"char_492_quercu":  true,
		"char_103_angel":   true,
		"char_340_shwaz":   true,
		"char_2023_ling":   true,
		"char_003_kalts":   true,
		"char_4009_irene":  true,

		// Operators who float in the air in battle stance
		// This is because part of their model goes below their feet
		"char_164_nightm": true,
		// "char_2023_ling":  true,
		"char_188_helage": true,
	}

	weirdEnemeyIds := map[string]bool{
		// Flying enemies with weird y positions
		"enemy_1005_yokai":    true,
		"enemy_1005_yokai_2":  true,
		"enemy_1005_yokai_3":  true,
		"enemy_1040_bombd":    true,
		"enemy_1040_bombd_2":  true,
		"enemy_1070_iced":     true,
		"enemy_1105_tyokai":   true,
		"enemy_1105_tyokai_2": true,
		"enemy_1106_byokai":   true,
		"enemy_1188_krgdrn":   true,
		"enemy_1188_krgdrn_2": true,
		"enemy_1269_nhfly":    true,
		"enemy_1269_nhfly_2":  true,
		"enemy_1321_wdarft":   true,
		"enemy_1321_wdarft_2": true,
		"enemy_4002_syokai":   true,
		"enemy_4002_syokai_2": true,
		"enemy_6011_planty":   true,

		// enemies with y position incorrect after fix
		"enemy_1101_plkght":   true,
		"enemy_1010_demon":    true,
		"enemy_1010_demon_2":  true,
		"enemy_1100_scorpn":   true,
		"enemy_1100_scorpn_2": true,
		"enemy_1215_ptrarc":   true,
		"enemy_1215_ptrarc_2": true,
		"enemy_1202_msfzhi":   true,
		"enemy_1202_msfzhi_2": true,
		"enemy_1293_duswrd_3": true,
		"enemy_1301_ymcnon":   true,
		"enemy_1301_ymcnon_2": true,
		"enemy_1308_mheagl":   true,
		"enemy_1308_mheagl_2": true,
		"enemy_1337_bhrknf":   true,
		"enemy_1337_bhrknf_2": true,
		"enemy_1338_bhrjst":   true,
		"enemy_1338_bhrjst_2": true,
		"enemy_1340_bthtbw":   true,
		"enemy_1340_bthtbw_2": true,
		"enemy_1341_bthtms":   true,
		"enemy_1341_bthtms_2": true,
		"enemy_7010_bldrgn":   true,

		// Enemies with weird y positions
		"enemy_1550_dhnzzh": true,

		// Large enemies
		"enemy_1526_sfsui6": true,
		"enemy_2054_smdeer": true,
		"enemy_2018_csdoll": true,
		"enemy_1533_stmkgt": true,
	}

	return &Context{
		nextX:            0.025,
		nextY:            0.05,
		positionX:        0.025,
		positionY:        0.0,
		count:            0,
		authToken:        authToken,
		faction:          operator.FACTION_ENUM_OPERATOR,
		weirdOperatorIds: weirdOperatorIds,
		weirdEnemyIds:    weirdEnemeyIds,
	}
}

// ?channelName=stymphalian__&debug=true&scale=0.5&width=4000&height=2000
func (c *Context) Run(opId string, skin string, stance operator.ChibiStanceEnum, facing operator.ChibiFacingEnum, spineData *operator.SpineData) {

	if c.faction == operator.FACTION_ENUM_OPERATOR {
		if !c.weirdOperatorIds[opId] {
			return
		}
	} else if c.faction == operator.FACTION_ENUM_ENEMY {
		if !c.weirdEnemyIds[opId] {
			return
		}
	}

	log.Println(opId, skin, stance, facing, spineData.SkelFilepath)

	jsonObj := api.RoomSetOperatorRequest{
		ChannelName:     "stymphalian__",
		Username:        fmt.Sprintf("%s-%s-%s-%s", opId, skin, stance, facing),
		UserDisplayName: fmt.Sprintf("%s-%s-%s-%s", opId, skin, stance, facing),
		Faction:         c.faction,
		OperatorId:      opId,
		Skin:            skin,
		Stance:          stance,
		PositionX:       float64(c.positionX),
		PositionY:       float64(c.positionY),
	}
	jsonBytes, err := json.Marshal(jsonObj)
	if err != nil {
		log.Fatal(err)
	}

	req, err := http.NewRequest(
		"POST",
		"http://localhost:8080/api/rooms/users/set/",
		bytes.NewReader(jsonBytes),
	)
	req.Header.Set("Authorization", "Bearer "+c.authToken)
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		log.Fatal(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Fatal(resp.StatusCode)
	}

	c.count += 1
	c.positionX += c.nextX
	if c.count%(int(1/c.nextX)) == 0 {
		c.positionX = c.nextX
		c.positionY += c.nextY
	}
}

func main() {
	var err error
	assetDir := flag.String("assetDir", "", "path to the assets")
	authToken := flag.String("authToken", "", "auth token for the bot")
	flag.Parse()
	log.Println("-assetDir: ", *assetDir)

	assetMap := operator.NewSpineAssetMap()
	enemyAssetMap := operator.NewSpineAssetMap()
	err = assetMap.LoadFromIndex(filepath.Join(*assetDir, "characters_index.json"))
	if err != nil {
		log.Fatal(err)
	}
	err = enemyAssetMap.LoadFromIndex(filepath.Join(*assetDir, "enemy_index.json"))
	if err != nil {
		log.Fatal(err)
	}

	context := NewContext(*authToken)
	context.faction = operator.FACTION_ENUM_OPERATOR
	assetMap.Iterate(context.Run)
	context.faction = operator.FACTION_ENUM_ENEMY
	enemyAssetMap.Iterate(context.Run)
}
