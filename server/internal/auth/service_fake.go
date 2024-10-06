package auth

import (
	"fmt"
	"net/http"

	"github.com/Stymphalian/ak_chibi_bot/server/internal/misc"
	"golang.org/x/oauth2"
)

type FakeAuthService struct {
	IsAuthenticated bool
	Username        string
	TwitchUserId    string
}

func NewFakeAuthService() *FakeAuthService {
	return &FakeAuthService{}
}

func (s *FakeAuthService) IsAuthorized(w http.ResponseWriter, r *http.Request) (*AuthorizedInfo, error) {
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

func (s *FakeAuthService) RevokeToken(token *oauth2.Token) error {
	return nil
}
