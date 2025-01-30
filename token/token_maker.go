package token

import "time"

type TokenMaker interface {
	CreateToken(user_id int64, role string, duration time.Duration) (string, *Payload, error)
	VerifyToken(token string) (*Payload, error)
}
