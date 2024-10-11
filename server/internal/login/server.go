package login

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/auth"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/users"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/csrf"
	"golang.org/x/oauth2"
)

type LoginServer struct {
	assetDir    string
	authService *auth.AuthService
	usersRepo   users.UserRepository
}

func NewLoginServer(
	assetDir misc.StaticAssetDirString,
	authService *auth.AuthService,
	usersRepo users.UserRepository,
) (*LoginServer, error) {
	log.Println("NewLoginServer created")
	return &LoginServer{
		assetDir:    string(assetDir),
		authService: authService,
		usersRepo:   usersRepo,
	}, nil
}

func (s *LoginServer) middlewareNoAuthCheck(h misc.HandlerWithErr) http.Handler {
	return misc.MiddlewareWithTimeout(h, 5*time.Second)
}

func (s *LoginServer) RegisterHandlers(rootMux *http.ServeMux) error {
	mux := http.NewServeMux()
	// mux.Handle("GET /auth/login/{$}", s.middlewareNoAuthCheck(s.HandleLoginPage))
	mux.Handle("GET /auth/login/twitch/{$}", s.middlewareNoAuthCheck(s.HandleLoginTwitch))
	mux.Handle("GET /auth/twitch/callback/{$}", s.middlewareNoAuthCheck(s.HandleOAuthCallback))
	mux.Handle("GET /auth/twitch/token/{$}", s.middlewareNoAuthCheck(s.HandleGetToken))
	mux.Handle("GET /auth/twitch/token/validate/{$}", s.middlewareNoAuthCheck(s.HandleValidateToken))
	mux.Handle("POST /auth/logout/{$}", s.middlewareNoAuthCheck(s.HandleLogout))
	mux.Handle("GET  /auth/check/{$}", s.middlewareNoAuthCheck(s.HandleAuthCheck))
	rootMux.Handle("/auth/", mux)
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
		return fmt.Errorf("couldn't generate a session! %s", err)
	}
	var jwtNonceBytes [128]byte
	if _, err := rand.Read(jwtNonceBytes[:]); err != nil {
		return fmt.Errorf("couldn't generate a session! %s", err)
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
	forceVerify := oauth2.SetAuthURLParam("force_verify", "true")
	http.Redirect(w, r, s.authService.Oauth2Config.AuthCodeURL(state, claims, nonce, forceVerify), http.StatusTemporaryRedirect)
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
		session.Flashes(auth.STATE_CALLBACK_KEY)
		session.Flashes(auth.STATE_JWT_NONCE_KEY)
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
		http.Redirect(w, r, "/login/callback?status=failed", http.StatusTemporaryRedirect)
		return fmt.Errorf("couldn't verify your confirmation, please try again. %s", err)
	}

	if r.FormValue("error") != "" {
		http.Redirect(w, r, "/login/callback?status=failed", http.StatusTemporaryRedirect)
		return fmt.Errorf("couldn't verify your confirmation, please try again. %s", r.FormValue("error"))
	}

	// Use the code to exchange for the OAUTH access token
	token, err := s.authService.Oauth2Config.Exchange(context.Background(), r.FormValue("code"))
	if err != nil {
		log.Println("Failed to exchange token")
		http.Redirect(w, r, "/login/callback?status=failed", http.StatusTemporaryRedirect)
		return err
	}
	log.Printf("@@@@ accessToken = %+v", token)

	// Verify the ID token
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		log.Println("can't extract id token from access token")
		http.Redirect(w, r, "/login/callback?status=failed", http.StatusTemporaryRedirect)
		return fmt.Errorf("can't extract id token from access token")
	}
	log.Println("@@@@ rawIDToken: ", rawIDToken)
	idToken, err := s.authService.Verifier.Verify(context.Background(), rawIDToken)
	log.Printf("@@@@ idToken: %+v", idToken)
	if err != nil {
		log.Println("can't verify ID token")
		http.Redirect(w, r, "/login/callback?status=failed", http.StatusTemporaryRedirect)
		return fmt.Errorf("can't verify ID token")
	}
	var claims auth.TwitchClaims
	if err := idToken.Claims(&claims); err != nil {
		log.Println("claims are invalid")
		http.Redirect(w, r, "/login/callback?status=failed", http.StatusTemporaryRedirect)
		return err
	}
	log.Printf("@@@@ claims: %+v", claims)
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
		http.Redirect(w, r, "/login/callback?status=failed", http.StatusTemporaryRedirect)
		return fmt.Errorf("couldn't verify your confirmation, please try again. %s", err)
	}

	// Create the User if they don't exists in our DB
	err = s.createOrInsertUser(r.Context(), claims.Sub)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/login/callback?status=failed", http.StatusTemporaryRedirect)
		return err
	}

	if oldToken, ok := session.Values[auth.OAUTH_TOKEN_KEY].(*oauth2.Token); ok {
		// If there was already an old token in the session just revoke that one.
		s.authService.RevokeToken(oldToken)
	}

	// Set the tokens into the session. These are the credentials needed
	// to know if the User is logged in, and which User they are logged in as
	session.Values[auth.OAUTH_TOKEN_KEY] = token
	// session.Values[auth.OAUTH_JWT_TOKEN_KEY] = rawIDToken
	http.Redirect(w, r, "/login/callback?status=success", http.StatusTemporaryRedirect)
	return nil
}

type TokenResponse struct {
	Token string `json:"token"`
}

func (s *LoginServer) HandleValidateToken(w http.ResponseWriter, r *http.Request) error {
	// Get token from query string
	tokenString := r.URL.Query().Get("token")
	if tokenString == "" {
		return fmt.Errorf("no token provided")
	}
	claims, err := s.authService.ValidateJWTToken(tokenString)
	log.Println("@@@@ validate tokens claims", claims)
	return err
}

func (s *LoginServer) HandleGetToken(w http.ResponseWriter, r *http.Request) error {
	authInfo, err := s.authService.HasAuthorizedSession(w, r)
	if err != nil {
		return err
	}
	if !authInfo.Authenticated {
		return fmt.Errorf("not authenticated")
	}

	typeStr := r.URL.Query().Get("type")
	var myClaims auth.AkChibiBotClaims
	if typeStr == "bad" {
		myClaims = auth.AkChibiBotClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "ak-chibi-bot",
				Subject:   fmt.Sprintf("%d", authInfo.User.UserId),
				Audience:  jwt.ClaimStrings{"ak-chibi-bot"},
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
			},
			UserId:         authInfo.User.UserId,
			TwitchUserId:   authInfo.User.TwitchUserId,
			TwitchUserName: authInfo.User.Username,
		}
	} else {
		myClaims = auth.AkChibiBotClaims{
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    "ak-chibi-bot",
				Subject:   fmt.Sprintf("%d", authInfo.User.UserId),
				Audience:  jwt.ClaimStrings{"ak-chibi-bot"},
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
			},
			UserId:         authInfo.User.UserId,
			TwitchUserId:   authInfo.User.TwitchUserId,
			TwitchUserName: authInfo.User.Username,
		}
	}

	tokenString, err := s.authService.CreateJWTToken(&myClaims)
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(&TokenResponse{
		Token: tokenString,
	})
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

type LoggedInResponse struct {
	Authenticated bool   `json:"authenticated"`
	Username      string `json:"username"`
	TwitchUserId  string `json:"userId"`
	IsAdmin       bool   `json:"is_admin"`
	CsrfToken     string `json:"csrf_token"`
}

func (s *LoginServer) HandleAuthCheck(w http.ResponseWriter, r *http.Request) error {
	info, err := s.authService.HasAuthorizedSession(w, r)
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-CSRF-Token", csrf.Token(r))
	if err != nil {
		return json.NewEncoder(w).Encode(&LoggedInResponse{
			Authenticated: false,
			IsAdmin:       false,
			Username:      "",
			TwitchUserId:  "",
		})
	} else {
		return json.NewEncoder(w).Encode(&LoggedInResponse{
			Authenticated: info.Authenticated,
			Username:      info.User.Username,
			IsAdmin:       info.User.IsAdmin,
			TwitchUserId:  info.User.TwitchUserId,
			CsrfToken:     csrf.Token(r),
		})
	}
}

func (s *LoginServer) HandleLogout(w http.ResponseWriter, r *http.Request) error {
	log.Println("logout called")
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

	if oldToken, ok := session.Values[auth.OAUTH_TOKEN_KEY].(*oauth2.Token); ok {
		// If there was already an old token in the session just revoke that one.
		s.authService.RevokeToken(oldToken)
	}
	delete(session.Values, auth.OAUTH_TOKEN_KEY)
	delete(session.Values, auth.OAUTH_JWT_TOKEN_KEY)
	return nil
}
