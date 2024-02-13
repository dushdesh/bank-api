package token

import (
	"time"

	"aidanwoods.dev/go-paseto"
	"github.com/google/uuid"
)

type PasetoMaker struct {
	secretKey paseto.V4SymmetricKey
}

func NewPasetoMaker(secret paseto.V4SymmetricKey) Maker {
	secretKey := paseto.NewV4SymmetricKey()
	return &PasetoMaker{secretKey: secretKey}
}

// Create token for a username and duration
func (m *PasetoMaker) CreateToken(username string, duration time.Duration) (string, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", err
	}
	token := paseto.NewToken()
	token.SetIssuedAt(payload.IssuedAt)
	token.SetExpiration(payload.ExpireAt)
	token.SetString("username", payload.Username)
	token.SetString("id", payload.ID.String())

	return token.V4Encrypt(m.secretKey, nil), nil
}
// Verify token if valid or no
func (m *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	parser := paseto.NewParser()
	t, err := parser.ParseV4Local(m.secretKey, token, nil)
	if err != nil {
		return nil, ErrExpiredToken
	}
	username, err := t.GetString("username")
	if err != nil {
		return nil, ErrInvalidToken
	}
	idStr, err := t.GetString("id")
	if err != nil {
		return nil, ErrInvalidToken
	}
	id, err := uuid.Parse(idStr) 
	if err !=nil {
		return nil, ErrInvalidToken
	}
	expireAt, err := t.GetExpiration()
	if err != nil {
		return nil, ErrInvalidToken
	}
	issuedAt, err := t.GetIssuedAt()
	if err != nil {
		return nil, ErrInvalidToken
	}
	payload := &Payload{
		Username: username,
		ExpireAt: expireAt,
		IssuedAt: issuedAt,
		ID: id,
	}
	return payload, nil
}