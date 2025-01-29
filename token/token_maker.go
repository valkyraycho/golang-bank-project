package token

import "time"

type TokenMaker interface {
	CreateToken(user_id int64, role string, duration time.Duration) (string, error)
	VerifyToken(token string) (*Payload, error)
}
