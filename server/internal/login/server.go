package login

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"net/http"
	"text/template"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/auth"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
	"golang.org/x/oauth2"
)

type LoginServer struct {
	assetDir    string
	authService *auth.AuthService
	usersRepo   users.UserRepository
}

func NewLoginServer(
	assetDir string,
	authService *auth.AuthService,
	usersRepo users.UserRepository,
) (*LoginServer, error) {

	return &LoginServer{
		assetDir:    assetDir,
		authService: authService,
		usersRepo:   usersRepo,
	}, nil
}

func (s *LoginServer) middlewareWithAuthCheck(h misc.HandlerWithErr) http.Handler {
	return misc.MiddlewareWithTimeout(s.authService.CheckAuth(h), 5*time.Second)
}

func (s *LoginServer) middlewareNoAuthCheck(h misc.HandlerWithErr) http.Handler {
	return misc.MiddlewareWithTimeout(h, 5*time.Second)
}

func (s *LoginServer) RegisterHandlers() error {
	http.Handle("GET /login/{$}", s.middlewareNoAuthCheck(s.HandleLoginPage))
	http.Handle("GET /login/twitch/{$}", s.middlewareNoAuthCheck(s.HandleLoginTwitch))
	http.Handle("GET /login/oauth/cb/{$}", s.middlewareNoAuthCheck(s.HandleOAuthCallback))
	http.Handle("GET /logout/{$}", s.middlewareWithAuthCheck(s.HandleLogout))

	// Test login AuthCheck
	http.Handle("GET /login/in/{$}", s.middlewareWithAuthCheck(s.HandleLoggedIn))
	return nil
}

func (s *LoginServer) HandleLoginPage(w http.ResponseWriter, r *http.Request) error {
	// TODO: User is already logged in. Just redirect to the other page
	// Check if User is already logged in. Redirect to other pages

	t, err := template.ParseFiles(fmt.Sprintf("%s/login/index.html", s.assetDir))
	if err != nil {
		return err
	}
	t.Execute(w, nil)
	return nil
}

func (s *LoginServer) HandleLoginTwitch(w http.ResponseWriter, r *http.Request) error {
	session, err := s.authService.CookieStore.Get(r, auth.OAUTH_SESSION_NAME)
	if err != nil {
		log.Printf("corrupted session %s -- generated new", err)
		err = nil
	}

	// Create the state and nonce values and store them as flashes in the
	// client session cookie
	var tokenBytes [128]byte
	if _, err := rand.Read(tokenBytes[:]); err != nil {
		return fmt.Errorf("Couldn't generate a session! %s", err)
	}
	var jwtNonceBytes [128]byte
	if _, err := rand.Read(jwtNonceBytes[:]); err != nil {
		return fmt.Errorf("Couldn't generate a session! %s", err)
	}
	state := hex.EncodeToString(tokenBytes[:])
	jwtNonce := hex.EncodeToString(jwtNonceBytes[:])
	session.AddFlash(state, auth.STATE_CALLBACK_KEY)
	session.AddFlash(jwtNonce, auth.STATE_JWT_NONCE_KEY)
	if err = session.Save(r, w); err != nil {
		return err
	}

	// Redirect to the oauthService issuer URL
	claims := oauth2.SetAuthURLParam("claims", `{"id_token":{}}`)
	nonce := oauth2.SetAuthURLParam("nonce", jwtNonce)
	http.Redirect(w, r, s.authService.Oauth2Config.AuthCodeURL(state, claims, nonce), http.StatusTemporaryRedirect)
	return nil
}

func (s *LoginServer) HandleOAuthCallback(w http.ResponseWriter, r *http.Request) error {
	session, err := s.authService.CookieStore.Get(r, auth.OAUTH_SESSION_NAME)
	if err != nil {
		log.Printf("corrupted session %s -- generated new", err)
		err = nil
	}

	// ensure we flush the csrf challenge even if the request is ultimately unsuccessful
	defer func() {
		if err := session.Save(r, w); err != nil {
			log.Printf("error saving session: %s", err)
		}
	}()

	stateFlash := session.Flashes(auth.STATE_CALLBACK_KEY)
	stateFromForm := r.FormValue("state")
	switch stateChallenge, state := stateFlash, stateFromForm; {
	case state == "", len(stateChallenge) < 1:
		err = errors.New("missing state challenge")
	case state != stateChallenge[0]:
		err = fmt.Errorf("invalid oauth state, expected '%s', got '%s'", state, stateChallenge[0])
	}
	if err != nil {
		return fmt.Errorf("Couldn't verify your confirmation, please try again. %s", err)
	}

	if r.FormValue("error") != "" {
		return fmt.Errorf("Couldn't verify your confirmation, please try again. %s", r.FormValue("error"))
	}

	// Use the code to exchange for the OAUTH access token
	token, err := s.authService.Oauth2Config.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		log.Println("Failed to exchange token")
		return err
	}

	// Verify the ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Println("can't extract id token from access token")
		return fmt.Errorf("can't extract id token from access token")
	}
	idToken, err := s.authService.Verifier.Verify(context.Background(), rawIDToken)
	if err != nil {
		log.Println("can't verify ID token")
		return fmt.Errorf("can't verify ID token")
	}
	var claims auth.TwitchClaims
	if err := idToken.Claims(&claims); err != nil {
		log.Println("claims are invalid")
		return err
	}
	// Validate the ID token's nonce
	JwtNonceFlash := session.Flashes(auth.STATE_JWT_NONCE_KEY)
	switch nonceChallenge, nonce := JwtNonceFlash, claims.Nonce; {
	case nonce == "", len(nonceChallenge) < 1:
		err = errors.New("missing nonce challenge")
	case nonce != nonceChallenge[0]:
		err = fmt.Errorf("invalid nonce , expected '%s', got '%s'", nonce, nonceChallenge[0])
	}
	if err != nil {
		log.Println(err)
		return fmt.Errorf("Couldn't verify your confirmation, please try again. %s", err)
	}

	// Create the User if they don't exists in our DB
	err = s.createOrInsertUser(r.Context(), claims.Sub)
	if err != nil {
		log.Println(err)
		return err
	}

	if oldToken, ok := session.Values[auth.OAUTH_TOKEN_KEY].(*oauth2.Token); ok {
		// If there was already an old token in the session just revoke that one.
		s.authService.RevokeToken(oldToken)
	}

	// Set the tokens into the session. These are the credentials needed
	// to know if the User is logged in, and which User they are logged in as
	session.Values[auth.OAUTH_TOKEN_KEY] = token
	http.Redirect(w, r, "/login/in/", http.StatusTemporaryRedirect)

	return nil
}

func (s *LoginServer) createOrInsertUser(ctx context.Context, twitchUserIdStr string) error {
	// Insert into Users table
	userInfo, err := s.authService.GetUserFromTwitchId(twitchUserIdStr)
	if err != nil {
		return err
	}
	_, err = s.usersRepo.GetOrInsertUser(ctx, *userInfo)
	if err != nil {
		return err
	}
	return nil
}

func (s *LoginServer) HandleLoggedIn(w http.ResponseWriter, r *http.Request) error {
	w.Write([]byte("Logged IN"))
	return nil
}

func (s *LoginServer) HandleLogout(w http.ResponseWriter, r *http.Request) error {
	session, err := s.authService.CookieStore.Get(r, auth.OAUTH_SESSION_NAME)
	if err != nil {
		log.Printf("corrupted session %s -- generated new", err)
		return nil
	}
	defer func() {
		if err := session.Save(r, w); err != nil {
			log.Printf("error saving session: %s", err)
		}
	}()

	delete(session.Values, auth.OAUTH_TOKEN_KEY)
	w.Write([]byte("Logged OUT"))
	return nil
}
