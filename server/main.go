package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/Stymphalian/ak_chibi_bot/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/internal/spine"
	"github.com/Stymphalian/ak_chibi_bot/internal/twitchbot"
)

// HumanReadableError represents error information
// that can be fed back to a human user.
//
// This prevents internal state that might be sensitive
// being leaked to the outside world.
type HumanReadableError interface {
	HumanError() string
	HTTPCode() int
}
type HumanReadableWrapper struct {
	error
	ToHuman string
	Code    int
}

func (h HumanReadableWrapper) HumanError() string { return h.ToHuman }
func (h HumanReadableWrapper) HTTPCode() int      { return h.Code }

type HandlerWithErr func(http.ResponseWriter, *http.Request) error

func annotateError(h HandlerWithErr) HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		// parse POST body, limit request size
		if err = r.ParseForm(); err != nil {
			return HumanReadableWrapper{
				ToHuman: "Something went wrong! Please try again.",
				Code:    http.StatusBadRequest,
				error:   err,
			}
		}

		return h(w, r)
	}
}

func errorHandling(handler HandlerWithErr) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := handler(w, r); err != nil {
			var errorString string = "Something went wrong! Please try again."
			var errorCode int = 500

			if v, ok := err.(HumanReadableError); ok {
				errorString, errorCode = v.HumanError(), v.HTTPCode()
			}

			log.Println(err)
			w.Write([]byte(errorString))
			w.WriteHeader(errorCode)
			return
		}
	})
}

type MainStruct struct {
	assetDir         *string
	imageDir         string
	spineAssetDir    string
	address          *string
	twitchConfigPath *string

	spineServer *spine.SpineBridge
	chibiActor  *twitchbot.ChibiActor
	twitchBot   *twitchbot.TwitchBot
}

func (s *MainStruct) run() {
	defer s.spineServer.Close()
	defer s.twitchBot.Close()
	go s.twitchBot.ReadPump()

	log.Println("Starting server")
	server := &http.Server{Addr: *s.address}

	log.Println(s.imageDir)
	log.Println(s.spineAssetDir)
	http.Handle("/runtime/assets/", http.StripPrefix("/runtime/assets/", http.FileServer(http.Dir(s.imageDir))))
	http.Handle("/runtime/", http.StripPrefix("/runtime/", http.FileServer(http.Dir(s.spineAssetDir))))
	http.Handle("/room/", errorHandling(annotateError(s.HandleRoom)))
	http.Handle("/ws/", errorHandling(annotateError(s.spineServer.HandleSpine)))
	// http.Handle("/admin", errorHandling(annotateError(s.spineServer.HandleAdmin)))

	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		log.Println("Signal interrupt received, shutting down")
		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
			os.Exit(1)
		}
	}()

	fmt.Println(server.ListenAndServe())
}

func (s *MainStruct) HandleRoom(w http.ResponseWriter, r *http.Request) error {
	if !r.URL.Query().Has("channelName") {
		log.Println("invalid connection. Requires channelName query argument")
		return nil
	}
	channelName := r.URL.Query().Get("channelName")

	s.twitchBot.JoinChannel(channelName)
	http.Redirect(w, r, fmt.Sprintf("/runtime/?channelName=%s", channelName), http.StatusSeeOther)
	log.Println(r.URL.Path)
	return nil
}

func NewMainStruct() *MainStruct {
	assetDir := flag.String("assetdir", "/ak_chibi_assets", "Asset directory")
	address := flag.String("address", ":7001", "Server address")
	twitchConfigPath := flag.String("twitch_config", "twitch_config.json", "Twitch config filepath containig channel names and tokens")
	flag.Parse()
	log.Println("-assetdir: ", *assetDir)
	log.Println("-address: ", *address)
	log.Println("-twitch_config:", *twitchConfigPath)

	if *twitchConfigPath == "" {
		log.Fatal("Must specify -twitch_config")
	}

	twitchConfig, err := misc.LoadTwitchConfig(*twitchConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	imageDir := fmt.Sprintf("%s/%s", *assetDir, "assets")
	spineAssetDir := fmt.Sprintf("%s/%s", *assetDir, "spine-ts")

	spineServer, err := spine.NewSpineBridge(imageDir, twitchConfig)
	if err != nil {
		log.Fatal(err)
	}
	chibiActor := twitchbot.NewChibiActor(spineServer, twitchConfig)
	twitchBot, err := twitchbot.NewTwitchBot(chibiActor, twitchConfig)
	if err != nil {
		log.Fatal(err)
	}

	return &MainStruct{
		assetDir,
		imageDir,
		spineAssetDir,
		address,
		twitchConfigPath,
		spineServer,
		chibiActor,
		twitchBot,
	}
}

func main() {
	m := NewMainStruct()
	m.run()
}
