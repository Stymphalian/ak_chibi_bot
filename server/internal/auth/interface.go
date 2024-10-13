package auth

import (
	"net/http"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"golang.org/x/oauth2"
)

type AuthServiceInterface interface {
	HasAuthorizedSession(w http.ResponseWriter, r *http.Request) (*AuthorizedInfo, error)
	GetUserFromTwitchId(twitchUserIdStr string) (*misc.UserInfo, error)
	RevokeSessionToken(token *oauth2.Token) error
	ValidateJWTToken(tokenString string) (*AkChibiBotClaims, error)
}
