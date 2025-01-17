//go:build wireinject
// +build wireinject

package server

import (
	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/api"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/auth"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/login"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/room"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/twitch_api"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
	"github.com/google/wire"
)

func InitializeMainServer() (*MainServer, error) {
	wire.Build(NewMainServer,
		// Arguments and Constants
		misc.ProvideCommandLineArgs,
		misc.ProvideStaticAssetDirString,
		misc.ProvideImageAssetDirString,
		// misc.ProvideAddressString,
		// misc.ProvideBotConfigPath,
		misc.ProvideBotConfig,

		// Repositories
		akdb.ProvideDatabaseConn,
		room.NewRoomRepositoryPsql,
		wire.Bind(new(room.RoomRepository), new(*room.RoomRepositoryPsql)),
		users.NewUserRepositoryPsql,
		wire.Bind(new(users.UserRepository), new(*users.UserRepositoryPsql)),
		users.NewChatterRepositoryPsql,
		wire.Bind(new(users.ChatterRepository), new(*users.ChatterRepositoryPsql)),
		users.NewUserPreferencesRepositoryPsql,
		wire.Bind(new(users.UserPreferencesRepository), new(*users.UserPreferencesRepositoryPsql)),
		auth.NewAuthRepositoryPsql,
		wire.Bind(new(auth.AuthRepository), new(*auth.AuthRepositoryPsql)),

		// Services
		operator.NewAssetService,
		operator.NewDefaultOperatorService,
		twitch_api.ProvideTwitchApiClient,
		wire.Bind(new(twitch_api.TwitchApiClientInterface), new(*twitch_api.TwitchApiClient)),
		auth.ProvideAuthService,
		wire.Bind(new(auth.AuthServiceInterface), new(*auth.AuthService)),
		room.NewRoomsManager,

		// API Controllers and Servers
		login.NewLoginServer,
		api.NewApiServer,
	)
	return &MainServer{}, nil
}
