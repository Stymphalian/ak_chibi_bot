package auth

import (
	"context"
	"encoding/gob"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/Stymphalian/ak_chibi_bot/server/internal/twitch_api"
	"github.com/antonlindstrom/pgstore"
	"github.com/coreos/go-oidc"
	"github.com/gorilla/securecookie"
	"golang.org/x/oauth2"
)

type TwitchClaims struct {
	Iss   string `json:"iss"`
	Sub   string `json:"sub"`
	Aud   string `json:"aud"`
	Exp   int32  `json:"exp"`
	Iat   int32  `json:"iat"`
	Nonce string `json:"nonce"`
	Email string `json:"email"`
}

const (
	STATE_CALLBACK_KEY           = "oauth-state-callback"
	STATE_JWT_NONCE_KEY          = "oauth-state-jwt-nonce"
	OAUTH_SESSION_NAME           = "oauth-oidc-session"
	OAUTH_TOKEN_KEY              = "oauth-token"
	CONTEXT_TWITCH_USER_ID       = "twitch-user-id"
	VALIDATE_OAUTH_TOKENS_PERIOD = 1 * time.Hour
	COOKIE_MAX_AGE               = 6 * time.Hour
)

type AuthService struct {
	// authRepo       AuthRepository
	twitchClientId string
	twitchSecret   string
	cookieSecret   string
	twitchClient   twitch_api.TwitchApiClientInterface
	Oauth2Config   *oauth2.Config
	CookieStore    *pgstore.PGStore
	provider       *oidc.Provider
	Verifier       *oidc.IDTokenVerifier
	shutdownChan   chan struct{}
}

func NewAuthService(
	twitchClientId string,
	twitchSecret string,
	cookieSecret string,
	redirectUrl string,
	twitchClient twitch_api.TwitchApiClientInterface,
) (*AuthService, error) {
	gob.Register(&oauth2.Token{})

	dbPool, err := akdb.DefaultDB.DB()
	if err != nil {
		return nil, err
	}
	cookieStore, err := pgstore.NewPGStoreFromPool(
		dbPool,
		[]byte(cookieSecret),
	)
	cookieStore.Options.MaxAge = int(COOKIE_MAX_AGE.Seconds())
	cookieStore.Options.Secure = true
	if err != nil {
		return nil, err
	}
	// TODO: how to turn this option on?
	// cookieStore.Options.SameSite = http.SameSiteStrictMode

	provider, err := oidc.NewProvider(context.Background(), "https://id.twitch.tv/oauth2")
	if err != nil {
		return nil, err
	}
	verifier := provider.Verifier(&oidc.Config{ClientID: twitchClientId})

	// TODO: Get the authURL/tokenURL from the twitch endpoint
	oauth2Config := &oauth2.Config{
		ClientID:     twitchClientId,
		ClientSecret: twitchSecret,
		Scopes:       []string{oidc.ScopeOpenID},
		Endpoint:     provider.Endpoint(),
		RedirectURL:  redirectUrl,
	}

	return &AuthService{
		twitchClientId: twitchClientId,
		twitchSecret:   twitchSecret,
		cookieSecret:   cookieSecret,
		twitchClient:   twitchClient,
		Oauth2Config:   oauth2Config,
		CookieStore:    cookieStore,
		provider:       provider,
		Verifier:       verifier,
		shutdownChan:   make(chan struct{}),
	}, nil
}

func (s *AuthService) Shutdown() {
	log.Println("AuthService calling Shutdown")
	close(s.shutdownChan)
}

func (r *AuthService) GetShutdownChan() chan struct{} {
	return r.shutdownChan
}

func (s *AuthService) RunLoop() {
	stopTimer := misc.StartTimer(
		"CleanAndValidateSessions",
		VALIDATE_OAUTH_TOKENS_PERIOD,
		func() {
			s.validateAndRefreshOauthTokens()
			s.cleanupExpiredSessions()
		},
	)
	defer stopTimer()
	<-s.shutdownChan
}

func (s *AuthService) validateAndRefreshOauthTokens() error {
	log.Println("Validating oauth tokens")
	db := akdb.DefaultDB
	httpSessions := make([]*HttpSessionDb, 0)
	result := db.Where("expires_on > ?", time.Now()).Find(&httpSessions)
	if result.Error != nil {
		return result.Error
	}

	var err error
	codecs := s.CookieStore.Codecs

	reencodeCookieFn := func(
		name string,
		cookieValues map[interface{}]interface{},
		httpSession *HttpSessionDb,
	) error {
		// Reencode the data and save back to the DB
		newEncodedData, err := securecookie.EncodeMulti(name, cookieValues, codecs...)
		if err != nil {
			return err
		}
		httpSession.Data = newEncodedData
		httpSession.ModifiedOn = time.Now()
		if err = httpSession.Save(); err != nil {
			return err
		}
		return nil
	}

	// Collect some metrics about the cookies
	numCookiesChecked := 0
	numCookiesWithOauthTokens := 0
	numAuthTokensFailedRefresh := 0
	numAuthTokensRefreshed := 0
	numAuthTokensValidated := 0
	numAuthTokensValid := 0
	numAuthTokensInvalid := 0
	for _, httpSession := range httpSessions {
		numCookiesChecked += 1
		data := httpSession.Data
		name := string(OAUTH_SESSION_NAME)

		var cookieValues map[interface{}]interface{}
		err = securecookie.DecodeMulti(name, data, &cookieValues, codecs...)
		if err != nil {
			continue
		}

		token, ok := cookieValues[OAUTH_TOKEN_KEY].(*oauth2.Token)
		if !ok {
			continue
		}

		numCookiesWithOauthTokens += 1
		tokenSource := s.Oauth2Config.TokenSource(context.Background(), token)
		newToken, err := tokenSource.Token()
		if err != nil {
			// Failed to create token source means bad token. Delete it
			delete(cookieValues, OAUTH_TOKEN_KEY)
			reencodeCookieFn(name, cookieValues, httpSession)
			numAuthTokensFailedRefresh += 1
			continue
		}

		if *newToken != *token {
			// The token has been refreshed, reencode into the cookies
			cookieValues[OAUTH_TOKEN_KEY] = newToken
			reencodeCookieFn(name, cookieValues, httpSession)
			numAuthTokensRefreshed += 1
		} else {
			// Validate the token
			_, err = s.twitchClient.ValidateToken(newToken.AccessToken)
			numAuthTokensValidated += 1
			if err != nil {
				// Invalid token. clear from the cookie
				delete(cookieValues, OAUTH_TOKEN_KEY)
				reencodeCookieFn(name, cookieValues, httpSession)
				numAuthTokensInvalid += 1
			} else {
				numAuthTokensValid += 1
			}
		}
	}

	log.Println("Finished validating oauth tokens:")
	log.Println("  NumCookiesChecked:", numCookiesChecked)
	log.Println("  NumCookiesWithOauthTokens:", numCookiesWithOauthTokens)
	log.Println("  NumAuthTokensFailedRefresh:", numAuthTokensFailedRefresh)
	log.Println("  NumAuthTokensRefreshed:", numAuthTokensRefreshed)
	log.Println("  NumAuthTokensValidated:", numAuthTokensValidated)
	log.Println("  NumAuthTokensValid:", numAuthTokensValid)
	log.Println("  NumAuthTokensInvalid:", numAuthTokensInvalid)
	return nil
}

func (s *AuthService) cleanupExpiredSessions() error {
	log.Println("Cleaning expired sessions")
	db := akdb.DefaultDB

	httpSessions := make([]*HttpSessionDb, 0)
	tx := db.Where("expires_on < ?", time.Now()).Find(&httpSessions)
	if tx.Error != nil {
		return tx.Error
	}

	name := string(OAUTH_SESSION_NAME)
	codecs := s.CookieStore.Codecs
	for _, httpSession := range httpSessions {
		data := httpSession.Data
		// Decode the data
		// If there is an OAUTH token then revoke it
		var cookieValues map[interface{}]interface{}
		err := securecookie.DecodeMulti(name, data, &cookieValues, codecs...)
		if err != nil {
			continue
		}
		token, ok := cookieValues[OAUTH_TOKEN_KEY].(*oauth2.Token)
		if ok {
			s.RevokeToken(token)
		}
	}

	result, err := db.ConnPool.ExecContext(
		context.Background(),
		"DELETE FROM http_sessions WHERE expires_on < now()",
	)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		log.Printf("Finished cleaning expired sessions %d", rowsAffected)
		return err
	} else {
		log.Println("Finished cleaning expired sessions")
		return nil
	}
}

func (s *AuthService) IsAuthorized(w http.ResponseWriter, r *http.Request) bool {
	session, err := s.CookieStore.Get(r, OAUTH_SESSION_NAME)
	if err != nil {
		return false
	}
	if session.IsNew {
		return false
	}
	token, ok := session.Values[OAUTH_TOKEN_KEY].(*oauth2.Token)
	if !ok {
		return false
	}
	// Verify the tokens
	if token.Expiry.Before(time.Now().UTC()) {
		return false
	}
	_, err = s.twitchClient.ValidateToken(token.AccessToken)

	return err == nil
}

func (s *AuthService) CheckAuth(h misc.HandlerWithErr) misc.HandlerWithErr {
	return func(w http.ResponseWriter, r *http.Request) (err error) {
		session, err := s.CookieStore.Get(r, OAUTH_SESSION_NAME)
		if err != nil {
			http.Redirect(w, r, "/login/", http.StatusTemporaryRedirect)
			return nil
		}
		if session.IsNew {
			http.Redirect(w, r, "/login/", http.StatusTemporaryRedirect)
			return nil
		}

		defer func() {
			if err := session.Save(r, w); err != nil {
				log.Printf("error saving session: %s", err)
			}
		}()

		token, ok := session.Values[OAUTH_TOKEN_KEY].(*oauth2.Token)
		if !ok {
			delete(session.Values, OAUTH_TOKEN_KEY)
			http.Redirect(w, r, "/login/", http.StatusTemporaryRedirect)
			return nil
		}

		// Verify the tokens
		if token.Expiry.Before(time.Now().UTC()) {
			delete(session.Values, OAUTH_TOKEN_KEY)
			http.Redirect(w, r, "/login/", http.StatusTemporaryRedirect)
			return nil
		}

		validateTokenResp, err := s.twitchClient.ValidateToken(token.AccessToken)
		if err != nil {
			delete(session.Values, OAUTH_TOKEN_KEY)
			http.Redirect(w, r, "/login/", http.StatusTemporaryRedirect)
			return nil
		}
		twitchUserIdInt, err := strconv.Atoi(validateTokenResp.UserId)
		if err != nil {
			delete(session.Values, OAUTH_TOKEN_KEY)
			http.Redirect(w, r, "/login/", http.StatusTemporaryRedirect)
			return nil
		}

		*r = *r.WithContext(context.WithValue(r.Context(), CONTEXT_TWITCH_USER_ID, twitchUserIdInt))
		return h(w, r)
	}
}

func (s *AuthService) GetUserFromTwitchId(twitchUserIdStr string) (*misc.UserInfo, error) {
	users, err := s.twitchClient.GetUsersById(twitchUserIdStr)
	if err != nil {
		return nil, err
	}
	if len(users.Data) != 1 {
		return nil, fmt.Errorf("expected 1 user, got %d", len(users.Data))
	}

	return &misc.UserInfo{
		Username:        users.Data[0].Login,
		UsernameDisplay: users.Data[0].DisplayName,
		TwitchUserId:    users.Data[0].Id,
	}, nil
}

func (s *AuthService) RevokeToken(token *oauth2.Token) error {
	return s.twitchClient.RevokeToken(token.AccessToken)
}