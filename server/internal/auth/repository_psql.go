package auth

import "github.com/Stymphalian/ak_chibi_bot/server/internal/akdb"

type AuthRepositoryPsql struct {
	*akdb.DatbaseConn
}

func NewAuthRepositoryPsql(db *akdb.DatbaseConn) *AuthRepositoryPsql {
	return &AuthRepositoryPsql{
		DatbaseConn: db,
	}
}
