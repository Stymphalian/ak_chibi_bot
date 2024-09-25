package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq" // add this

	"github.com/Stymphalian/ak_chibi_bot/server/internal/admin"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/api"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/auth"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/login"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/operator"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/room"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/twitch_api"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
)

const (
	DEFAULT_TIMEOUT         = 5 * time.Second
	FORCEFUL_SHUTDOWN_DELAY = 10 * time.Second
)

type MainStruct struct {
	imageAssetDir  string
	staticAssetDir string
	address        *string
	botConfigPath  *string
	botConfig      *misc.BotConfig

	roomsRepo    room.RoomRepository
	usersRepo    users.UserRepository
	chattersRepo users.ChatterRepository
	authRepo     auth.AuthRepository

	assetService *operator.AssetService
	authService  *auth.AuthService

	roomManager *room.RoomsManager
	adminServer *admin.AdminServer
	apiServer   *api.ApiServer
	loginServer *login.LoginServer
}

func NewMainStruct() *MainStruct {
	imageAssetDir := flag.String("image_assetdir", "../static/assets", "Image Asset Directory")
	staticAssetsDir := flag.String("static_dir", "../static", "Static assets folder")
	address := flag.String("address", ":8080", "Server address")
	botConfigPath := flag.String("bot_config", "bot_config.json", "Config filepath containing channel names and tokens")
	flag.Parse()
	log.Println("-image_assetdir: ", *imageAssetDir)
	log.Println("-static_dir: ", *staticAssetsDir)
	log.Println("-address: ", *address)
	log.Println("-bot_config:", *botConfigPath)

	if *botConfigPath == "" {
		log.Fatal("Must specify -bot_config")
	}

	botConfig, err := misc.LoadBotConfig(*botConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	assetService, err := operator.NewAssetService(*imageAssetDir)
	if err != nil {
		log.Fatal(err)
	}
	roomsRepo := room.NewPostgresRoomRepository()
	usersRepo := users.NewUserRepositoryPsql()
	chattersRepo := users.NewChatterRepositoryPsql()
	authRepo := auth.NewAuthRepositoryPsql()

	authService, err := auth.NewAuthService(
		botConfig.TwitchClientId,
		botConfig.TwitchClientSecret,
		botConfig.CookieSecret,
		botConfig.TwitchOauthRedirectUrl,
		twitch_api.NewTwitchApiClient(
			botConfig.TwitchClientId,
			botConfig.TwitchAccessToken,
		),
		authRepo,
	)
	if err != nil {
		log.Fatal(err)
	}
	loginServer, err := login.NewLoginServer(*staticAssetsDir, authService, usersRepo)
	if err != nil {
		log.Fatal(err)
	}
	roomManager := room.NewRoomsManager(assetService, roomsRepo, usersRepo, chattersRepo, botConfig)
	adminServer := admin.NewAdminServer(roomManager, botConfig, *staticAssetsDir)
	apiServer := api.NewApiServer(roomManager)

	return &MainStruct{
		imageAssetDir:  *imageAssetDir,
		staticAssetDir: *staticAssetsDir,
		address:        address,
		botConfigPath:  botConfigPath,
		botConfig:      botConfig,

		roomsRepo:    roomsRepo,
		usersRepo:    usersRepo,
		chattersRepo: chattersRepo,
		authRepo:     authRepo,

		assetService: assetService,
		authService:  authService,
		roomManager:  roomManager,
		adminServer:  adminServer,
		apiServer:    apiServer,
		loginServer:  loginServer,
	}
}

func (s *MainStruct) run() {
	go s.roomManager.RunLoop()
	go s.authService.RunLoop()
	s.roomManager.LoadExistingRooms(context.Background())

	log.Println("Starting server")
	server := &http.Server{
		Addr:              *s.address,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	server.RegisterOnShutdown(s.roomManager.Shutdown)
	server.RegisterOnShutdown(s.authService.Shutdown)

	log.Printf("Images Assets = %s\n", s.imageAssetDir)
	log.Printf("Static Assets = %s\n", s.staticAssetDir)
	http.Handle("/static/assets/",
		http.TimeoutHandler(
			http.StripPrefix("/static/assets/", http.FileServer(http.Dir(s.imageAssetDir))),
			DEFAULT_TIMEOUT,
			"",
		),
	)
	http.Handle("/static/",
		http.TimeoutHandler(
			http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticAssetDir))),
			DEFAULT_TIMEOUT,
			"",
		),
	)
	http.Handle("/room/settings/",
		http.TimeoutHandler(
			http.StripPrefix("/room/settings/", http.FileServer(http.Dir(s.staticAssetDir+"/rooms"))),
			DEFAULT_TIMEOUT,
			"",
		),
	)
	http.Handle("/runtime/{$}",
		http.TimeoutHandler(
			http.StripPrefix("/runtime/", http.FileServer(http.Dir(s.staticAssetDir+"/spine"))),
			DEFAULT_TIMEOUT,
			"",
		),
	)
	http.Handle("/room/", misc.MiddlewareWithTimeout(s.HandleRoom, DEFAULT_TIMEOUT))
	http.Handle("/ws/", misc.Middleware(s.HandleSpineWebSocket))
	s.adminServer.RegisterHandlers()
	s.apiServer.RegisterHandlers()
	s.loginServer.RegisterHandlers()

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt, syscall.SIGTERM)
		<-sigint
		signal.Stop(sigint)

		log.Println("Signal interrupt received, shutting down")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
			os.Exit(1)
		}
	}()

	fmt.Println(server.ListenAndServe())
	s.WaitForShutdownsWithTimeout(
		s.roomManager.GetShutdownChan(),
		s.authService.GetShutdownChan(),
	)
}

func (s *MainStruct) WaitForShutdownsWithTimeout(shutdownChans ...chan struct{}) {
	log.Println("Waiting for shutdown")
	allChanDones := make(chan struct{})

	go func() {
		var wg sync.WaitGroup
		wg.Add(len(shutdownChans))
		for _, ch := range shutdownChans {
			go func(ch chan struct{}) {
				defer wg.Done()
				<-ch
			}(ch)
		}
		wg.Wait()

		// close the default DB, after everythign else is closed.
		sqldb, err := akdb.DefaultDB.DB()
		if err == nil {
			sqldb.Close()
		}
		close(allChanDones)
	}()

	select {
	case <-time.After(FORCEFUL_SHUTDOWN_DELAY):
		log.Printf("Shutting down forcefully after %v", FORCEFUL_SHUTDOWN_DELAY)
	case <-allChanDones:
		log.Printf("Shutting down gracefully")
	}
}

func (s *MainStruct) HandleRoom(w http.ResponseWriter, r *http.Request) error {
	if !r.URL.Query().Has("channelName") {
		return errors.New("invalid connection. Requires channelName query argument")
	}
	channelName := r.URL.Query().Get("channelName")
	channelName = strings.ToLower(channelName)

	err := s.roomManager.CreateRoomOrNoOp(r.Context(), channelName)
	if err != nil {
		return err
	}
	http.Redirect(w, r, fmt.Sprintf("/runtime/?channelName=%s", channelName), http.StatusSeeOther)
	return nil
}

func (s *MainStruct) HandleSpineWebSocket(w http.ResponseWriter, r *http.Request) error {
	if !r.URL.Query().Has("channelName") {
		return errors.New("invalid connection. Requires channelName query argument")
	}
	channelName := r.URL.Query().Get("channelName")
	return s.roomManager.HandleSpineWebSocket(channelName, w, r)
}

func main() {
	m := NewMainStruct()
	m.run()
}
