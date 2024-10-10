package room

import (
	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/twitch_api"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
)

func NewFakeRoomsManager() *RoomsManager {
	assetService := operator.NewTestAssetService()
	akDB, _ := akdb.ProvideTestDatabaseConn()
	roomRepo := NewRoomRepositoryPsql(akDB)
	usersRepo := users.NewUserRepositoryPsql(akDB)
	chattersRepo := users.NewChatterRepositoryPsql(akDB)
	botConfig := &misc.BotConfig{
		TwitchClientId:     "test_client_id",
		TwitchAccessToken:  "test_access_token",
		SpineRuntimeConfig: misc.DefaultSpineRuntimeConfig(),
		InitialOperator:    "Amiya",
		OperatorDetails: misc.InitialOperatorDetails{
			Skin:       "default",
			Stance:     "base",
			Animations: []string{"Idle"},
			PositionX:  0.5,
		},
	}
	spineService := operator.NewOperatorService(assetService, botConfig.SpineRuntimeConfig)

	return &RoomsManager{
		Rooms:          make(map[string]*Room),
		assetService:   assetService,
		spineService:   spineService,
		roomRepo:       roomRepo,
		usersRepo:      usersRepo,
		chattersRepo:   chattersRepo,
		botConfig:      botConfig,
		twitchClient:   twitch_api.NewFakeTwitchApiClient(),
		shutdownDoneCh: make(chan struct{}),
		removeRoomCh:   make(chan string),
	}
}
