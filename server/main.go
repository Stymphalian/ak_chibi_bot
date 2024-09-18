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
	"syscall"
	"time"

	_ "github.com/lib/pq" // add this

	"github.com/Stymphalian/ak_chibi_bot/server/internal/admin"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/api"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/room"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
)

const (
	DEFAULT_TIMEOUT = 5 * time.Second
)

type MainStruct struct {
	imageAssetDir  string
	staticAssetDir string
	address        *string
	botConfigPath  *string
	botConfig      *misc.BotConfig

	assetService *spine.AssetService
	roomManager  *room.RoomsManager
	adminServer  *admin.AdminServer
	apiServer    *api.ApiServer
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
	assetService, err := spine.NewAssetService(*imageAssetDir)
	if err != nil {
		log.Fatal(err)
	}
	roomManager := room.NewRoomsManager(assetService, botConfig)
	adminServer := admin.NewAdminServer(roomManager, botConfig, *staticAssetsDir)
	apiServer := api.NewApiServer(roomManager)

	return &MainStruct{
		*imageAssetDir,
		*staticAssetsDir,
		address,
		botConfigPath,
		botConfig,

		assetService,
		roomManager,
		adminServer,
		apiServer,
	}
}

func (s *MainStruct) run() {
	go s.roomManager.RunLoop()
	s.roomManager.LoadExistingRooms()
	// s.roomManager.Restore()

	log.Println("Starting server")
	server := &http.Server{
		Addr:              *s.address,
		ReadTimeout:       1 * time.Second,
		ReadHeaderTimeout: 1 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       30 * time.Second,
	}
	server.RegisterOnShutdown(s.roomManager.Shutdown)

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
	http.Handle("/rooms/settings/",
		http.TimeoutHandler(
			http.StripPrefix("/rooms/settings/", http.FileServer(http.Dir(s.staticAssetDir+"/rooms"))),
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
	s.roomManager.WaitForShutdownWithTimeout()
}

func (s *MainStruct) HandleRoom(w http.ResponseWriter, r *http.Request) error {
	if !r.URL.Query().Has("channelName") {
		return errors.New("invalid connection. Requires channelName query argument")
	}
	channelName := r.URL.Query().Get("channelName")
	channelName = strings.ToLower(channelName)

	err := s.roomManager.CreateRoomOrNoOp(channelName, context.Background())
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
