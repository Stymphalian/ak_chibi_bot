package auth

import (
	"net/http"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"golang.org/x/oauth2"
)

type AuthServiceInterface interface {
	IsAuthorized(w http.ResponseWriter, r *http.Request) (*AuthorizedInfo, error)
	GetUserFromTwitchId(twitchUserIdStr string) (*misc.UserInfo, error)
	RevokeToken(token *oauth2.Token) error
}