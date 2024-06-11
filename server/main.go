package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"

	"github.com/Stymphalian/ak_chibi_bot/misc"
	"github.com/Stymphalian/ak_chibi_bot/spine"
	"github.com/Stymphalian/ak_chibi_bot/twitchbot"
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

func main() {
	assetDir := flag.String("assetdir", "/ak_chibi_assets/assets", "Asset directory")
	address := flag.String("address", ":7001", "Server address")
	twitchConfigPath := flag.String("twitch_config", "twitch_config.json", "Twitch config filepath containig channel names and tokens")
	flag.Parse()
	log.Println("-assetdir: ", *assetDir)
	log.Println("-address: ", *address)
	log.Println("-twitch_config:", *twitchConfigPath)

	if *twitchConfigPath == "" {
		log.Fatal("Must specify -twitch_config")
		return
	}

	twitchConfig, err := misc.LoadTwitchConfig(*twitchConfigPath)
	if err != nil {
		log.Fatal(err)
	}

	spineServer, err := spine.NewSpineBridge(*assetDir, twitchConfig)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer spineServer.Close()

	chibiActor := twitchbot.NewChibiActor(spineServer, twitchConfig)
	twitchBot, err := twitchbot.NewTwitchBot(chibiActor, twitchConfig)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer twitchBot.Close()
	go twitchBot.ReadPump()

	log.Println("Starting server")
	server := &http.Server{Addr: *address}
	http.Handle("/", http.FileServer(http.Dir("./spine-ts")))
	http.Handle(
		"/player/example/assets/",
		http.StripPrefix("/player/example/assets/", http.FileServer(http.Dir(*assetDir))),
	)
	http.Handle("/spine", errorHandling(annotateError(spineServer.HandleSpine)))
	http.Handle("/forward", errorHandling(annotateError(spineServer.HandleForward)))
	http.Handle("/admin", errorHandling(annotateError(spineServer.HandleAdmin)))

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
