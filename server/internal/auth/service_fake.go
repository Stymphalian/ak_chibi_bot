package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

type FakeAuthService struct {
	IsAuthenticated bool
	IsJwtTokenValid bool
	Username        string
	TwitchUserId    string
	UserId          uint
}

func NewFakeAuthService() *FakeAuthService {
	return &FakeAuthService{}
}

func (s *FakeAuthService) HasAuthorizedSession(w http.ResponseWriter, r *http.Request) (*AuthorizedInfo, error) {
	if !s.IsAuthenticated {
		return nil, fmt.Errorf("not authorized")
	} else {
		return &AuthorizedInfo{
			Authenticated: s.IsAuthenticated,
			User: AuthUserInfo{
				Username:     s.Username,
				TwitchUserId: s.TwitchUserId,
				IsAdmin:      false,
			},
		}, nil
	}
}
func (s *FakeAuthService) GetUserFromTwitchId(twitchUserIdStr string) (*misc.UserInfo, error) {
	return nil, nil
}

func (s *FakeAuthService) RevokeSessionToken(token *oauth2.Token) error {
	return nil
}

func (s *FakeAuthService) ValidateJWTToken(tokenString string) (*AkChibiBotClaims, error) {
	if !s.IsJwtTokenValid {
		return nil, fmt.Errorf("invalid token")
	} else {
		return &AkChibiBotClaims{
			UserId:         s.UserId,
			TwitchUserId:   s.TwitchUserId,
			TwitchUserName: s.Username,
			RegisteredClaims: jwt.RegisteredClaims{
				Issuer:    JWT_ISSUER,
				Subject:   fmt.Sprintf("%d", s.UserId),
				Audience:  jwt.ClaimStrings{JWT_AUDIENCE},
				IssuedAt:  jwt.NewNumericDate(time.Now()),
				NotBefore: jwt.NewNumericDate(time.Now()),
				ExpiresAt: jwt.NewNumericDate(time.Now().Add(JWT_TOKEN_VALID_DURATION)),
			},
		}, nil
	}
}
