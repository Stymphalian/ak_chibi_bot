package auth

import "github.com/golang-jwt/jwt/v5"

type AkChibiBotClaims struct {
	jwt.RegisteredClaims
	UserId         uint   `json:"user_id"`
	TwitchUserId   string `json:"twitch_user_id"`
	TwitchUserName string `json:"twitch_user_name"`
}

const JWT_TOKEN_SECRET = "secret"

type TwitchClaims struct {
	Iss   string `json:"iss"`
	Sub   string `json:"sub"`
	Aud   string `json:"aud"`
	Exp   int32  `json:"exp"`
	Iat   int32  `json:"iat"`
	Nonce string `json:"nonce"`
	Email string `json:"email"`
}

type ContextTwitchUserId string
type ContextTwitchUserName string
type ContextUserId string
