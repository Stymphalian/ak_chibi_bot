package room

import (
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/twitch_api"
)

func NewFakeRoomsManager() *RoomsManager {
	assetService := spine.NewTestAssetService()
	botConfig := &misc.BotConfig{
		TwitchClientId:     "test_client_id",
		TwitchAccessToken:  "test_access_token",
		SpineRuntimeConfig: misc.DefaultSpineRuntimeConfig(),
		OperatorDetails: misc.InitialOperatorDetails{
			Skin:       "default",
			Stance:     "base",
			Animations: []string{"Idle"},
			PositionX:  0.5,
		},
	}
	spineService := spine.NewSpineService(assetService, botConfig.SpineRuntimeConfig)

	return &RoomsManager{
		Rooms:          make(map[string]*Room),
		assetService:   assetService,
		spineService:   spineService,
		botConfig:      botConfig,
		twitchClient:   twitch_api.NewFakeTwitchApiClient(),
		shutdownDoneCh: make(chan struct{}),
	}
}
