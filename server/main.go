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

	_ "github.com/lib/pq" // add this

	"github.com/Stymphalian/ak_chibi_bot/server/internal/admin"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/room"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/spine"
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

	return &MainStruct{
		*imageAssetDir,
		*staticAssetsDir,
		address,
		botConfigPath,
		botConfig,

		assetService,
		roomManager,
		adminServer,
	}
}

func (s *MainStruct) run() {
	go s.roomManager.RunLoop()
	// s.roomManager.Restore()

	log.Println("Starting server")
	server := &http.Server{Addr: *s.address}
	server.RegisterOnShutdown(s.roomManager.Shutdown)

	log.Printf("Images Assets = %s\n", s.imageAssetDir)
	log.Printf("Static Assets = %s\n", s.staticAssetDir)
	http.Handle("/static/assets/", http.StripPrefix("/static/assets/", http.FileServer(http.Dir(s.imageAssetDir))))
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(s.staticAssetDir))))
	http.Handle("/runtime/{$}", http.StripPrefix("/runtime/", http.FileServer(http.Dir(s.staticAssetDir+"/spine"))))
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
	akdb.Prepare()
	db, err := akdb.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	rows, err := db.Query("SELECT title FROM blog")
	if err != nil {
		log.Fatal(err)
	}
	for rows.Next() {
		var title string
		err = rows.Scan(&title)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("@@@@ found titles:", title)
	}

	// m := NewMainStruct()
	// m.run()
}
