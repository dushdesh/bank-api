package token

import "time"

type Maker interface {
	// Create token for a username and duration
	CreateToken(username string, duration time.Duration) (string, error)
	// Verify token if valid or no
	VerifyToken(token string) (*Payload, error)
}