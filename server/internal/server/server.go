package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	_ "github.com/lib/pq" // add this

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

type MainServer struct {
	args      misc.CommandLineArgs
	botConfig *misc.BotConfig

	akDb         *akdb.DatbaseConn
	roomsRepo    room.RoomRepository
	usersRepo    users.UserRepository
	chattersRepo users.ChatterRepository
	authRepo     auth.AuthRepository

	assetService *operator.AssetService
	authService  *auth.AuthService

	roomManager *room.RoomsManager
	apiServer   *api.ApiServer
	loginServer *login.LoginServer
}

func NewMainServer(
	args *misc.CommandLineArgs,
	botConfig *misc.BotConfig,
	assetService *operator.AssetService,
	roomsRepo room.RoomRepository,
	usersRepo users.UserRepository,
	chattersRepo users.ChatterRepository,
	authRepo auth.AuthRepository,
	twitchApiClient twitch_api.TwitchApiClientInterface,
	authService *auth.AuthService,
	loginServer *login.LoginServer,
	roomsManager *room.RoomsManager,
	apiServer *api.ApiServer,
	akDb *akdb.DatbaseConn,
) *MainServer {
	return &MainServer{
		args:      *args,
		botConfig: botConfig,

		akDb:         akDb,
		roomsRepo:    roomsRepo,
		usersRepo:    usersRepo,
		chattersRepo: chattersRepo,
		authRepo:     authRepo,

		assetService: assetService,
		authService:  authService,
		roomManager:  roomsManager,
		apiServer:    apiServer,
		loginServer:  loginServer,
	}
}

func (s *MainServer) Run() {
	go s.roomManager.RunLoop()
	go s.authService.RunLoop()
	s.roomManager.LoadExistingRooms(context.Background())

	log.Println("Starting server")

	mux := http.NewServeMux()
	handler := mux
	server := &http.Server{
		Addr:              s.args.Address,
		Handler:           handler,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	server.RegisterOnShutdown(s.roomManager.Shutdown)
	server.RegisterOnShutdown(s.authService.Shutdown)

	log.Printf("Images Assets = %s\n", s.args.ImageAssetDir)
	log.Printf("Static Assets = %s\n", s.args.StaticAssetDir)

	// Main web app
	webAppFilePath := s.args.StaticAssetDir + "/web_app/build/"
	webAppFileServer := http.FileServer(http.Dir(webAppFilePath))
	mux.Handle("/",
		http.TimeoutHandler(
			http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/" {
					fullPath := webAppFilePath + strings.TrimPrefix(path.Clean(r.URL.Path), "/")
					_, err := os.Stat(fullPath)
					if err != nil {
						if !os.IsNotExist(err) {
							panic(err)
						}
						// Requested file does not exist so we return the default (resolves to index.html)
						r.URL.Path = "/"
					}
				}
				webAppFileServer.ServeHTTP(w, r)
			}),
			DEFAULT_TIMEOUT,
			"",
		),
	)

	// Image File Server
	mux.Handle("/image/assets/",
		http.TimeoutHandler(
			http.StripPrefix("/image/assets/", http.FileServer(http.Dir(s.args.ImageAssetDir))),
			DEFAULT_TIMEOUT,
			"",
		),
	)

	// Public File Server
	mux.Handle("/public/",
		http.TimeoutHandler(
			http.StripPrefix("/public/", http.FileServer(http.Dir(s.args.StaticAssetDir+"/public"))),
			DEFAULT_TIMEOUT,
			"",
		),
	)

	// Spine File Server
	spineMux := http.NewServeMux()
	spineFileServer := http.FileServer(http.Dir(s.args.StaticAssetDir + "/spine/dist"))
	spineMux.Handle("/spine/runtime/{$}",
		http.TimeoutHandler(
			http.StripPrefix("/spine/runtime/", spineFileServer),
			DEFAULT_TIMEOUT,
			"",
		),
	)
	spineMux.Handle("/spine/static/",
		http.TimeoutHandler(
			http.StripPrefix("/spine/static/", spineFileServer),
			DEFAULT_TIMEOUT,
			"",
		),
	)
	mux.Handle("/spine/", spineMux)

	// Bot Server
	mux.Handle("/room/", misc.MiddlewareWithTimeout(s.HandleRoom, DEFAULT_TIMEOUT))
	mux.Handle("/ws/", misc.Middleware(s.HandleSpineWebSocket))

	s.apiServer.RegisterHandlers(mux)
	s.loginServer.RegisterHandlers(mux)

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

func (s *MainServer) WaitForShutdownsWithTimeout(shutdownChans ...chan struct{}) {
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
		sqldb, err := s.akDb.DefaultDB.DB()
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

func (s *MainServer) HandleRoom(w http.ResponseWriter, r *http.Request) error {
	if !r.URL.Query().Has("channelName") {
		return errors.New("invalid connection. Requires channelName query argument")
	}
	channelName := r.URL.Query().Get("channelName")
	channelName = strings.ToLower(channelName)
	extraQueryArgs := ""
	width := r.URL.Query().Get("width")
	if len(width) > 0 {
		widthInt, err := strconv.Atoi(width)
		if err == nil {
			extraQueryArgs += fmt.Sprintf("&width=%d", widthInt)
		}
	}
	height := r.URL.Query().Get("height")
	if len(height) > 0 {
		heightInt, err := strconv.Atoi(height)
		if err == nil {
			extraQueryArgs += fmt.Sprintf("&height=%d", heightInt)
		}
	}
	chibiScale := r.URL.Query().Get("scale")
	if len(chibiScale) > 0 {
		chibiScaleInt, err := strconv.Atoi(chibiScale)
		if err == nil {
			extraQueryArgs += fmt.Sprintf("&scale=%d", chibiScaleInt)
		}
	}

	err := s.roomManager.CreateRoomOrNoOp(r.Context(), channelName)
	if err != nil {
		return err
	}
	http.Redirect(w, r, fmt.Sprintf("/spine/runtime/?channelName=%s%s", channelName, extraQueryArgs), http.StatusSeeOther)
	return nil
}

func (s *MainServer) HandleSpineWebSocket(w http.ResponseWriter, r *http.Request) error {
	if !r.URL.Query().Has("channelName") {
		return errors.New("invalid connection. Requires channelName query argument")
	}
	channelName := r.URL.Query().Get("channelName")
	return s.roomManager.HandleSpineWebSocket(channelName, w, r)
}
