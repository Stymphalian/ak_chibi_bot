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
	"syscall"

	"github.com/Stymphalian/ak_chibi_bot/internal/admin"
	"github.com/Stymphalian/ak_chibi_bot/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/internal/room"
	"github.com/Stymphalian/ak_chibi_bot/internal/spine"
)

type MainStruct struct {
	imageAssetDir    string
	spineAssetDir    string
	adminAssetDir    string
	address          *string
	twitchConfigPath *string

	twitchConfig *misc.TwitchConfig
	assetManager *spine.AssetManager
	roomManager  *room.RoomsManager
	adminServer  *admin.AdminServer
}

func NewMainStruct() *MainStruct {
	imageAssetDir := flag.String("image_assetdir", "/ak_chibi_assets/assets", "Image Asset Directory")
	spineAssetDir := flag.String("spine_assetdir", "/ak_chibi_assets/spine-ts", "Spine Asset Directory")
	adminAssetDir := flag.String("admin_assetdir", "/ak_chibi_assets/admin", "Admin Asset Directory")
	address := flag.String("address", ":8080", "Server address")
	twitchConfigPath := flag.String("twitch_config", "twitch_config.json", "Twitch config filepath containing channel names and tokens")
	flag.Parse()
	log.Println("-image_assetdir: ", *imageAssetDir)
	log.Println("-spine_assetdir: ", *spineAssetDir)
	log.Println("-admin_assetdir: ", *adminAssetDir)
	log.Println("-address: ", *address)
	log.Println("-twitch_config:", *twitchConfigPath)

	if *twitchConfigPath == "" {
		log.Fatal("Must specify -twitch_config")
	}

	twitchConfig, err := misc.LoadTwitchConfig(*twitchConfigPath)
	if err != nil {
		log.Fatal(err)
	}
	assetManager, err := spine.NewAssetManager(*imageAssetDir)
	if err != nil {
		log.Fatal(err)
	}
	roomManager := room.NewRoomsManager(assetManager, twitchConfig)
	adminServer := admin.NewAdminServer(roomManager, twitchConfig, *adminAssetDir)

	return &MainStruct{
		*imageAssetDir,
		*spineAssetDir,
		*adminAssetDir,
		address,
		twitchConfigPath,

		twitchConfig,
		assetManager,
		roomManager,
		adminServer,
	}
}

func (s *MainStruct) run() {
	go s.roomManager.RunLoop()

	log.Println("Starting server")
	server := &http.Server{Addr: *s.address}
	server.RegisterOnShutdown(s.roomManager.Shutdown)

	log.Println(s.imageAssetDir)
	log.Println(s.spineAssetDir)
	http.Handle("/runtime/assets/", http.StripPrefix("/runtime/assets/", http.FileServer(http.Dir(s.imageAssetDir))))
	http.Handle("/runtime/", http.StripPrefix("/runtime/", http.FileServer(http.Dir(s.spineAssetDir))))
	http.Handle("/room/", misc.Middleware(s.HandleRoom))
	http.Handle("/ws/", misc.Middleware(s.HandleSpineWebSocket))
	s.adminServer.RegisterAdmin()

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
